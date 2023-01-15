// Package bitdotio provides a Go SDK for bit.io database connections and developer APIs.
package bitdotio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

//
// Public constants
//

const (
	// APIVersion is the currently supported API version
	APIVersion string = "v2beta"

	// APIURL is the URL of the API service backend.
	APIURL string = "https://api.bit.io"

	// AppName identifies the type of client to bit.io
	AppName string = "go-bitdotio-sdk"

	// ClientVersion is the version of the bitdotio-python library being used.
	ClientVersion string = "0.0.0b"

	// DBHost is the host for database connections.
	DBHost string = "db.bit.io"

	// DBPort is the port for database connections.
	DBPort string = "5432"

	// MaxConnIdleTime is the maximum duration (as a string) that an idle connection will be kept alive
	// Currently one second less than the bit.io server-side timeout for idle connections
	MaxConnIdleTime string = "299s"

	// PoolMinConns is the minimum number of connections to keep alive per pool.
	PoolMinConns int32 = 0

	SSLMode string = "require"

	UserAgent string = "go-bitdotio-sdk/" + ClientVersion
)

//
// BitDotIO
//

type BitDotIO struct {
	accessToken string
	APIClient   APIClient
	lock        sync.RWMutex
	pools       map[string]*pgxpool.Pool
}

// NewBitDotIO constructs a new BitDotIO client.
func NewBitDotIO(accessToken string) *BitDotIO {
	return &BitDotIO{
		accessToken: accessToken,
		APIClient:   NewDefaultAPIClient(accessToken),
		pools:       make(map[string]*pgxpool.Pool),
	}
}

//
// Connection Pool Methods
//

// getConnString generates a pgxpool connection string for a particular bit.io database
func (b *BitDotIO) getConnString(dbName string, maxConns int32) string {

	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_min_conns=%d pool_max_conn_idle_time=%s",
		UserAgent,
		b.accessToken,
		DBHost,
		DBPort,
		dbName,
		SSLMode,
		PoolMinConns,
		MaxConnIdleTime,
	)
	if maxConns != 0 {
		connString += fmt.Sprintf(" pool_max_conns=%d", maxConns)
	}
	return connString
}

// CreatePool establishes a new connection pool for a bit.io database
//
// Note for reviewers: For now, it seems more user-friendly to have a wrapper here,
// but CreatePoolWithMaxConns could be refactored to take a config struct if we
// want to expose multiple configuration options later.
func (b *BitDotIO) CreatePool(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	// 0 maxConnections is a sentinal for "use pgxpool default"
	// Ref: https://pkg.go.dev/github.com/jackc/pgx/v4/pgxpool#ParseConfig
	return b.CreatePoolWithMaxConns(ctx, dbName, 0)
}

// CreatePoolWithMaxConns establishes a new connection pool for a bit.io database with a specified max number of connections
//
// Note for reviewers: I thought about simply having a GetPool that functions as
// a GetOrCreate, as in python-bitdotio. That is an attractive option both as
// a user convenience and because it might enable more performant concurrency
// safe pool creation (instead of the RW locks currently implemented). However,
// it's important to have explicit control over the context of a pool being
// created, which tipped me towards a separate explicit method instead of a
// dual-purpose getter.
func (b *BitDotIO) CreatePoolWithMaxConns(ctx context.Context, dbName string, maxConns int32) (*pgxpool.Pool, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if pool, ok := b.pools[dbName]; ok {
		// Check if pool is still open, only create a new one if not
		// https://github.com/jackc/pgx/issues/891#issuecomment-743775246
		conn, err := pool.Acquire(context.Background())
		if err == nil {
			conn.Release()
			return nil, fmt.Errorf("pool already exists for db '%s'", dbName)
		} else if err.Error() != "closed pool" {
			return nil, fmt.Errorf("found an existing pool for db %s and unable to verify closed state", dbName)
		}
	}

	pool, err := pgxpool.New(ctx, b.getConnString(dbName, maxConns))
	if err != nil {
		return nil, fmt.Errorf("unable to create pool for db %s: %w", dbName, err)
	}

	b.pools[dbName] = pool
	return pool, nil
}

// GetPool retrieves an existing connection pool for a bit.io database
func (b *BitDotIO) GetPool(dbName string) (*pgxpool.Pool, error) {
	b.lock.RLock()
	defer b.lock.RLock()
	if pool, ok := b.pools[dbName]; ok {
		return pool, nil
	}
	return nil, fmt.Errorf("pool does not exist for db %s", dbName)
}

// Connect acquires a connection for a connection pool for a bit.io database
func (b *BitDotIO) Connect(ctx context.Context, dbName string) (*pgxpool.Conn, error) {
	pool, err := b.GetPool(dbName)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire a connection for db %s: %w", dbName, err)
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire a connection for db %s: %w", dbName, err)
	}
	return conn, nil
}

// ClosePool closes a connection pool for a bit.io database
func (b *BitDotIO) ClosePool(dbName string) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if pool, ok := b.pools[dbName]; ok {
		pool.Close()
		delete(b.pools, dbName)
		return nil
	}
	return fmt.Errorf("no open pool found for db %s", dbName)
}

//
// API Methods
//

// ListDatabases lists metadata for all databases that you own or are a collaborator on.
func (b *BitDotIO) ListDatabases() ([]*Database, error) {
	data, err := b.APIClient.Call("GET", "db/", nil)
	if err != nil {
		err = fmt.Errorf("failed to get list of databases: %v", err)
		return nil, err
	}
	var databaseList DatabaseList
	if err = json.Unmarshal(data, &databaseList); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return databaseList.Databases, err
}

// CreateDatabase creates a new database.
func (b *BitDotIO) CreateDatabase(name string, databaseConfig *DatabaseConfig) (*Database, error) {
	body, err := json.Marshal(databaseConfig)
	if err != nil {
		err = fmt.Errorf("failed to serialize new database params: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("POST", "db/", body)
	if err != nil {
		err = fmt.Errorf("failed to create database: %v", err)
		return nil, err
	}
	var database Database
	if err = json.Unmarshal(data, &database); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &database, err
}

// GetDatabase gets metadata about a single database.
func (b *BitDotIO) GetDatabase(username, dbName string) (*Database, error) {
	path, err := url.JoinPath("db/", username, dbName)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("GET", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to get database: %v", err)
		return nil, err
	}
	var database Database
	if err = json.Unmarshal(data, &database); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &database, err
}

// DeleteDatabase deletes a single database.
func (b *BitDotIO) DeleteDatabase(username, dbName string) error {
	path, err := url.JoinPath("db/", username, dbName)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return err
	}

	_, err = b.APIClient.Call("DELETE", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to delete database: %v", err)
		return err
	}
	return err
}

// UpdateDatabase updates the configuration of a database.
func (b *BitDotIO) UpdateDatabase(username, dbName string, databaseConfig *DatabaseConfig) (*Database, error) {
	path, err := url.JoinPath("db/", username, dbName)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	body, err := json.Marshal(databaseConfig)
	if err != nil {
		err = fmt.Errorf("failed to serialize new database params: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("PATCH", path, body)
	if err != nil {
		err = fmt.Errorf("failed to update database: %v", err)
		return nil, err
	}
	var database Database
	if err = json.Unmarshal(data, &database); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &database, err
}

// CreateKey creates a new API key/database password with the same permissions as the requester
func (b *BitDotIO) CreateKey() (*Credentials, error) {
	path := "api-key/"

	data, err := b.APIClient.Call("POST", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to create a new key: %v", err)
		return nil, err
	}
	var credentials Credentials
	if err = json.Unmarshal(data, &credentials); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &credentials, err
}

// ListServiceAccounts lists metadata pertaining to service accounts the requester has created
func (b *BitDotIO) ListServiceAccounts() ([]*ServiceAccount, error) {
	data, err := b.APIClient.Call("GET", "service-account/", nil)
	if err != nil {
		err = fmt.Errorf("failed to get a list of service accounts: %v", err)
		return nil, err
	}
	var serviceAccountList ServiceAccountList
	if err = json.Unmarshal(data, &serviceAccountList); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return serviceAccountList.ServiceAccounts, err
}

// GetServiceAccount gets metadata about a single service account.
func (b *BitDotIO) GetServiceAccount(serviceAccountID string) (*ServiceAccount, error) {
	path, err := url.JoinPath("service-account", serviceAccountID)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("GET", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to get service account: %v", err)
		return nil, err
	}
	var serviceAccount ServiceAccount
	if err = json.Unmarshal(data, &serviceAccount); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &serviceAccount, err
}

// CreateServiceAccountKey creates a new key for a service account.
func (b *BitDotIO) CreateServiceAccountKey(serviceAccountID string) (*Credentials, error) {
	path, err := url.JoinPath("service-account", serviceAccountID, "api-key/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("POST", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to create new service account key: %v", err)
		return nil, err
	}
	var credentials Credentials
	if err = json.Unmarshal(data, &credentials); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &credentials, err
}

// RevokeServiceAccountKeys revokes all keys for a service account.
func (b *BitDotIO) RevokeServiceAccountKeys(serviceAccountID string) error {
	path, err := url.JoinPath("service-account", serviceAccountID, "api-key/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return err
	}

	_, err = b.APIClient.Call("DELETE", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to revoke service account keys: %v", err)
		return err
	}
	return err
}

// CreateImportJob creates a new import job.
// NB: Client is responsible for closing any closable readers passed in for files.
func (b *BitDotIO) CreateImportJob(dbName string, tableName string, config *ImportJobConfig) (*ImportJob, error) {
	// TODO: validate dbName
	if (config.FileURL == "") == (config.File == nil) {
		return nil, fmt.Errorf("Must provide File XOR FileURL")
	}

	path, err := url.JoinPath("db", dbName, "import/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	// Note for reviewers: Could use reflection but it seems using dynamic language
	// features are less idiomatic in a case like this?
	// https://stackoverflow.com/a/42849112
	fields := fieldParts{
		"table_name": strings.NewReader(tableName),
	}
	if v := config.SchemaName; v != "" {
		fields["schema_name"] = strings.NewReader(v)
	}
	if v := config.InferHeader; v != "" {
		// Note for reviewers: Possibly refactor this using a map[string]struct{} set?
		if v != "auto" && v != "first_row" && v != "header" {
			return nil, fmt.Errorf("InferHeader options are 'auto', 'first_row', or 'header', got %s", v)
		}
		fields["infer_header"] = strings.NewReader(v)
	}
	if v := config.FileURL; v != "" {
		fields["schema_name"] = strings.NewReader(v)
	}

	var files fileParts
	if f := config.File; f != nil {
		files = fileParts{"file": &formFile{tableName, f}}
	}

	data, err := b.APIClient.CallMultipart("POST", path, fields, files)
	if err != nil {
		err = fmt.Errorf("failed to create import job: %v", err)
		return nil, err
	}

	var importJob ImportJob
	if err = json.Unmarshal(data, &importJob); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &importJob, err
}

// GetImportJob gets the status for an import job.
func (b *BitDotIO) GetImportJob(importID string) (*ImportJob, error) {
	path, err := url.JoinPath("import", importID)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("GET", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to get import job status: %v", err)
		return nil, err
	}

	var importJob ImportJob
	if err = json.Unmarshal(data, &importJob); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &importJob, err
}

// CreateExportJob creates a new export job.
func (b *BitDotIO) CreateExportJob(dbName string, config *ExportJobConfig) (*ExportJob, error) {
	// TODO: validate dbName
	if (config.QueryString == "") == (config.TableName == "") {
		return nil, fmt.Errorf("Must provide QueryString XOR TableName")
	}

	// Explicit schema name is required by the API, but we can default to "public"
	// here if table_name is given.
	if config.TableName != "" && config.SchemaName == "" {
		config.SchemaName = "public"
	}

	path, err := url.JoinPath("db", dbName, "export/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	body, err := json.Marshal(config)
	if err != nil {
		err = fmt.Errorf("failed to marshal export job config: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("POST", path, body)
	if err != nil {
		err = fmt.Errorf("failed to create export job: %v", err)
		return nil, err
	}

	var exportJob ExportJob
	if err = json.Unmarshal(data, &exportJob); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &exportJob, err
}

// GetExportJob gets the status for an export job.
func (b *BitDotIO) GetExportJob(exportID string) (*ExportJob, error) {
	path, err := url.JoinPath("export", exportID)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	data, err := b.APIClient.Call("GET", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to get export job status: %v", err)
		return nil, err
	}

	var exportJob ExportJob
	if err = json.Unmarshal(data, &exportJob); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &exportJob, err
}
