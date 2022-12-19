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

// NewBitDotIO constructs a new BitDotIO client
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
func (b *BitDotIO) CreateDatabase(name string, isPrivate bool) (*Database, error) {
	databaseConfig := DatabaseConfig{Name: name, IsPrivate: isPrivate}
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

// DeleteDatabase gets metadata about a single database.
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
// TODO: Need to overhaul to add StorageLimitBytes and so  StorageLimitBytes,
// setDBName, and isPrivate can be set independently, rather than always both
// being passed in the request.
func (b *BitDotIO) UpdateDatabase(username, dbName, setDBName string, isPrivate bool) (*Database, error) {
	path, err := url.JoinPath("db/", username, dbName)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		return nil, err
	}

	databaseConfig := DatabaseConfig{Name: setDBName, IsPrivate: isPrivate}
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
