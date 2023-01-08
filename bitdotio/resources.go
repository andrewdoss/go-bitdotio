package bitdotio

import "time"

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

// DatabaseConfig maps the Create/Update Database JSON body to a Go struct for marshalling
type DatabaseConfig struct {
	Name string `json:"name,omitempty"`
	// TODO: This field seems like a potential footgun, as the zero-value is valid and makes a db public.
	IsPrivate         bool  `json:"is_private"`
	StorageLimitBytes int64 `json:"storage_limit_bytes,omitempty"`
}

// Credentials contains credentials for a personal or service account
type Credentials struct {
	Username string `json:"username"`
	APIKEY   string `json:"api_key"`
}

// ServiceAccountList contains a list of service accounts.
type ServiceAccountList struct {
	ServiceAccounts []*ServiceAccount `json:"service_accounts"`
}

// ServiceAccount contains metadata about a bit.io service account
type ServiceAccount struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	DateCreated      time.Time     `json:"date_created"`
	Role             string        `json:"role"`
	Databases        []*DatabaseID `json:"databases"`
	TokenCount       int64         `json:"token_count"`
	ActiveTokenCount int64         `json:"active_token_count"`
}
