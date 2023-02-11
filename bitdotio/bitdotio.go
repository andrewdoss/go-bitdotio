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
	// apiVersion is the currently supported API version.
	apiVersion string = "v2beta"

	// apiURL is the URL of the bit.io developer API service.
	apiURL string = "https://api.bit.io"

	// appName identifies the client to bit.io during direct Postgres connections.
	appName string = "go-bitdotio-sdk"

	// clientVersion is the version of the bitdotio-python library being used.
	clientVersion string = "0.0.0b"

	// dbHost is the host for database connections.
	dbHost string = "db.bit.io"

	// dbPort is the port for database connections.
	dbPort string = "5432"

	// maxConnIdleTime is the maximum idle time for a connection in a pool.
	maxConnIdleTime string = "290s"

	// poolMinConns is the minimum number of connections per pool.
	poolMinConns int32 = 0

	// pgSSLMode is the Postgres sslmode for connections to bit.io.
	pgSSLMode string = "require"

	// userAgent identifies the client to bit.io during HTTP requests.
	userAgent string = appName + clientVersion
)

// BitDotIO implements utility methods for usage of the bit.io developer API and
// manages per-database connection pools.
//
// BitDotIO's methods are safe for use across multiple goroutines. In general, a
// program should only create one BitDotIO instance per unique API key required
// for access (often only one).
//
// Some user-only API methods may receive 403 Forbidden responses if called using
// a service account token. See docs.bit.io for the latest API reference and
// further information about service accounts.
type BitDotIO struct {
	accessToken string
	apiClient   APIClient
	// Note for reviewers: debatable whether RW lock is a net benefit over simple mutex given extra overhead
	lock  sync.RWMutex
	pools map[string]*pgxpool.Pool
}

// Note for reviewers: I briefly looked into making an interface to decouple
// this package from pgxpool. I'm not sure that's important for a beta version, and further,
// any interface will have the downsides of:
// 1. Potentially getting out of sync w/ pgxpool
// 2. Limiting to a subset of features OR burdening the client with type assertions to use
//    pgx features that are outside of the interface.

// NewBitDotIO constructs a new BitDotIO client for a provided API key.
func NewBitDotIO(accessToken string) *BitDotIO {
	return &BitDotIO{
		accessToken: accessToken,
		apiClient:   NewDefaultAPIClient(accessToken),
		// Note for reviewers: I briefly looked into making an interface to decouple
		// this package from pgxpool. I'm not sure that's important for a beta version, and further,
		// any interface will have the downsides of:
		// 1. Potentially getting out of sync w/ pgxpool
		// 2. Limiting to a subset of features OR burdening the client with type assertions to use
		//    pgx features that are outside of the interface.
		pools: make(map[string]*pgxpool.Pool),
	}
}

//
// Connection Pool Methods
//

// getConnString generates a pgxpool connection string for a bit.io database.
func (b *BitDotIO) getConnString(dbName string, maxConns int32) string {

	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_min_conns=%d pool_max_conn_idle_time=%s",
		userAgent,
		b.accessToken,
		dbHost,
		dbPort,
		dbName,
		pgSSLMode,
		poolMinConns,
		maxConnIdleTime,
	)
	if maxConns != 0 {
		connString += fmt.Sprintf(" pool_max_conns=%d", maxConns)
	}
	return connString
}

// CreatePool establishes a new connection pool for a bit.io database. dbName
// must be a full, user-qualified database name (e.g. `username/dbname`).
// CreatePool can also be called for a database that previously had a pool that
// has been closed and will handle replacing the closed pool with a new open pool.
func (b *BitDotIO) CreatePool(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	// 0 maxConnections is a sentinal for "use pgxpool default". See ref for
	// default: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ParseConfig
	return b.CreatePoolWithMaxConns(ctx, dbName, 0)
}

// Note for reviewers: CreatePoolWithMaxConns could be refactored to take a
// config struct if we want to expose multiple configuration options later.

// CreatePoolWithMaxConns establishes a new connection pool for a bit.io database
// with a specified max number of connections, maxConns. See CreatePool for other
// documentation.
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
	// Note for reviewers: we could technically make pool creation non-locking by
	// bundling the pools w/ ready channels in the map, but pool creation takes
	// about 1 ms on my 5-year old mid-level mac mini, and I also think our pool
	// management methods are less performance-critical than the pgxpool itself.
	pool, err := pgxpool.New(ctx, b.getConnString(dbName, maxConns))
	if err != nil {
		return nil, fmt.Errorf("unable to create pool for db %s: %w", dbName, err)
	}
	b.pools[dbName] = pool
	return pool, nil
}

// Note for reviewers: I thought about simply having a GetPool that functions as
// a GetOrCreate, as in python-bitdotio. That is an attractive option both as
// a user convenience and because it might enable more performant concurrency-
// safe pool creation (instead of the RW locks currently implemented). However,
// it's important to have explicit control over the context of a pool being
// created, which tipped me towards a separate explicit method instead of a
// dual-purpose getter.

// GetPool retrieves an existing connection pool for a bit.io database.
func (b *BitDotIO) GetPool(dbName string) (*pgxpool.Pool, error) {
	b.lock.RLock()
	defer b.lock.RLock()
	if pool, ok := b.pools[dbName]; ok {
		return pool, nil
	}
	return nil, fmt.Errorf("pool does not exist for db %s", dbName)
}

// Connect acquires a connection from an existing pool for a bit.io database.
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

// ClosePool closes a connection pool for a bit.io database. Pools can be safely
// closed using this BitDotIO method or directly from the pool API.
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
	data, err := b.apiClient.Call("GET", "db/", nil)
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
func (b *BitDotIO) CreateDatabase(databaseConfig *DatabaseConfig) (*Database, error) {
	body, err := json.Marshal(databaseConfig)
	if err != nil {
		err = fmt.Errorf("failed to serialize new database params: %v", err)
		return nil, err
	}

	data, err := b.apiClient.Call("POST", "db/", body)
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

	data, err := b.apiClient.Call("GET", path, nil)
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

	_, err = b.apiClient.Call("DELETE", path, nil)
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

	data, err := b.apiClient.Call("PATCH", path, body)
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

// CreateKey creates a new API key/database password with the same permissions as the requester.
func (b *BitDotIO) CreateKey() (*Credentials, error) {
	path := "api-key/"

	data, err := b.apiClient.Call("POST", path, nil)
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

// ListServiceAccounts lists metadata pertaining to service accounts the requester has created.
func (b *BitDotIO) ListServiceAccounts() ([]*ServiceAccount, error) {
	data, err := b.apiClient.Call("GET", "service-account/", nil)
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

	data, err := b.apiClient.Call("GET", path, nil)
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

	data, err := b.apiClient.Call("POST", path, nil)
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

	_, err = b.apiClient.Call("DELETE", path, nil)
	if err != nil {
		err = fmt.Errorf("failed to revoke service account keys: %v", err)
		return err
	}
	return err
}

// CreateImportJob creates a new import job. Client is responsible for closing
// any closable readers passed in as the File field of an *ImportJobConfig.
func (b *BitDotIO) CreateImportJob(fullDBName string, tableName string, config *ImportJobConfig) (*ImportJob, error) {
	// TODO: validate dbName
	if (config.FileURL == "") == (config.File == nil) {
		return nil, fmt.Errorf("Must provide File XOR FileURL")
	}

	path, err := url.JoinPath("db", fullDBName, "import/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	// Add non-file request parts
	fields := fieldParts{
		"table_name": strings.NewReader(tableName),
	}
	if v := config.SchemaName; v != "" {
		fields["schema_name"] = strings.NewReader(v)
	}
	if v := config.InferHeader; v != "" {
		if v != "auto" && v != "first_row" && v != "header" {
			return nil, fmt.Errorf("InferHeader options are 'auto', 'first_row', or 'header', got %s", v)
		}
		fields["infer_header"] = strings.NewReader(v)
	}
	if v := config.FileURL; v != "" {
		fields["schema_name"] = strings.NewReader(v)
	}

	// Add file request parts
	var files fileParts
	if f := config.File; f != nil {
		files = fileParts{"file": &formFile{tableName, f}}
	}

	data, err := b.apiClient.CallMultipart("POST", path, fields, files)
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

	data, err := b.apiClient.Call("GET", path, nil)
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
func (b *BitDotIO) CreateExportJob(fullDBName string, config *ExportJobConfig) (*ExportJob, error) {
	// TODO: validate dbName
	if (config.QueryString == "") == (config.TableName == "") {
		return nil, fmt.Errorf("Must provide QueryString XOR TableName")
	}

	// Explicit schema name is required by the API, but we can default to "public"
	// here if table_name is given.
	if config.TableName != "" && config.SchemaName == "" {
		config.SchemaName = "public"
	}

	path, err := url.JoinPath("db", fullDBName, "export/")
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	body, err := json.Marshal(config)
	if err != nil {
		err = fmt.Errorf("failed to marshal export job config: %v", err)
		return nil, err
	}

	data, err := b.apiClient.Call("POST", path, body)
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

	data, err := b.apiClient.Call("GET", path, nil)
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

// Query executes a query using the HTTP API and returns the reponse as JSON-serialized bytes.
func (b *BitDotIO) Query(fullDBName string, queryString string) (*QueryResult, error) {
	path := "query"

	query := &Query{DatabaseName: fullDBName, QueryString: queryString}
	body, err := json.Marshal(query)
	if err != nil {
		err = fmt.Errorf("failed to serialize query: %v", err)
		return nil, err
	}

	data, err := b.apiClient.Call("POST", path, body)
	if err != nil {
		err = fmt.Errorf("query request failed: %v", err)
		return nil, err
	}

	var queryResult QueryResult
	if err = json.Unmarshal(data, &queryResult); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &queryResult, err
}
