// Package bitdotio provides a Go SDK for bit.io database connections and developer APIs.
package bitdotio

import (
	"encoding/json"
	"fmt"
	"net/url"
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

	// DBURL is the url for database connections.
	DBURL string = "db.bit.io"

	// DBPort is the port for database connections.
	DBPort string = "5432"

	UserAgent string = "go-bitdotio-sdk/" + ClientVersion
)

type BitDotIO struct {
	AccessToken string
	APIClient   APIClient
}

// NewBitDotIO constructs a new BitDotIO client.
func NewBitDotIO(accessToken string) *BitDotIO {
	return &BitDotIO{
		AccessToken: accessToken,
		APIClient:   NewDefaultAPIClient(accessToken),
	}
}

// ListDatabases lists metadata for all databases that you own or are a collaborator on.
func (b *BitDotIO) ListDatabases() (*DatabaseList, error) {
	data, err := b.APIClient.Call("GET", "db/", nil)
	if err != nil {
		err = fmt.Errorf("failed to get list of databases: %v", err)
		return nil, err
	}
	var databaseList DatabaseList
	if err = json.Unmarshal(data, &databaseList); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &databaseList, err
}

// CreateDatabase creates a new database.
// TODO: currently doesn't support variable storageLimitBytes
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
func (b *BitDotIO) ListServiceAccounts() (*ServiceAccountList, error) {
	data, err := b.APIClient.Call("GET", "service-account/", nil)
	if err != nil {
		err = fmt.Errorf("failed to get a list of service accounts: %v", err)
		return nil, err
	}
	var serviceAccountList ServiceAccountList
	if err = json.Unmarshal(data, &serviceAccountList); err != nil {
		err = fmt.Errorf("JSON unmarshaling failed: %s", err)
	}
	return &serviceAccountList, err
}
