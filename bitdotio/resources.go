package bitdotio

// DatabaseList contains a list of Databases.
type DatabaseList struct {
	Databases []*Database `json:"databases"`
}

// Database contains metadata about a bit.io database.
// TODO: add actual UUID and timestamp types
type Database struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DateCreated       string `json:"date_created"`
	IsPrivate         bool   `json:"is_private"`
	Role              string `json:"role"`
	StorageLimitBytes int64  `json:"storage_limit_bytes"`
	StorageUsageBytes int64  `json:"storage_usage_bytes"`
	UsageCurrent      *Usage `json:"usage_current"`
	UsagePrevious     *Usage `json:"usage_previous"`
}

// Usage contains current rows queried for a bit.io database.
// TODO: add actual timestamp types.
type Usage struct {
	RowsQueried int64  `json:"rows_queried"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

// DatabaseConfig maps the Create/Update Database JSON body to a Go struct for marshalling
// TODO: probably not the right name and/or approach
type DatabaseConfig struct {
	Name              string `json:"name"`
	IsPrivate         bool   `json:"is_private"`
	StorageLimitBytes int64  `json:"storage_limit_bytes,omitempty"`
}
