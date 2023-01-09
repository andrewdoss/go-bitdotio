package bitdotio

import (
	"io"
	"time"
)

// DatabaseList contains a list of Databases.
type DatabaseList struct {
	Databases []*Database `json:"databases"`
}

// DatabaseID contains identifying information for a bit.io database.
type DatabaseID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Database contains metadata about a bit.io database.
type Database struct {
	DatabaseID
	DateCreated       time.Time `json:"date_created"`
	IsPrivate         bool      `json:"is_private"`
	Role              string    `json:"role"`
	StorageLimitBytes int64     `json:"storage_limit_bytes"`
	StorageUsageBytes int64     `json:"storage_usage_bytes"`
	UsageCurrent      *Usage    `json:"usage_current"`
	UsagePrevious     *Usage    `json:"usage_previous"`
}

// Usage contains current rows queried for a bit.io database.
// TODO: Possibly parse out the Dates as time.Time type
type Usage struct {
	RowsQueried int64  `json:"rows_queried"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

// DatabaseConfig maps the Create/Update Database JSON body to a Go struct for marshalling.
type DatabaseConfig struct {
	Name string `json:"name,omitempty"`
	// TODO: This field seems like a potential footgun, as the zero-value is valid and makes a db public.
	IsPrivate         bool  `json:"is_private"`
	StorageLimitBytes int64 `json:"storage_limit_bytes,omitempty"`
}

// Credentials contains credentials for a personal or service account.
type Credentials struct {
	Username string `json:"username"`
	APIKEY   string `json:"api_key"`
}

// ServiceAccountList contains a list of service accounts.
type ServiceAccountList struct {
	ServiceAccounts []*ServiceAccount `json:"service_accounts"`
}

// ServiceAccount contains metadata about a bit.io service account.
type ServiceAccount struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	DateCreated      time.Time     `json:"date_created"`
	Role             string        `json:"role"`
	Databases        []*DatabaseID `json:"databases"`
	TokenCount       int64         `json:"token_count"`
	ActiveTokenCount int64         `json:"active_token_count"`
}

// TransferJob contains metadata about an import or export job.
type TransferJob struct {
	ID           string    `json:"id"`
	DateCreated  time.Time `json:"date_created"`
	DateFinished time.Time `json:"date_finished"`
	State        string    `json:"state"`
	Retries      int64     `json:"retries"`
	ErrorType    string    `json:"error_type"`
	ErrorID      string    `json:"error_id"`
	StatusURL    string    `json:"status_url"`
}

// ExportJob contains metadata about an export job.
type ExportJob struct {
	TransferJob
	ExportFormat string `json:"export_format"`
	FileName     string `json:"file_name"`
	DownloadURL  string `json:"download_url"`
}

// ImportJob contains metadata about an import job.
// TODO: Possibly handle "error_details" differently
type ImportJob struct {
	TransferJob
}

// ImportJobConfig contains configuration parameters for a new import job.
type ImportJobConfig struct {
	SchemaName  string    `json:"schema_name,omitempty"`
	InferHeader string    `json:"infer_header,omitempty"` // "auto", "first_row", or "header"
	FileURL     string    `json:"file_url,omitempty"`
	File        io.Reader `json:"-"`
}
