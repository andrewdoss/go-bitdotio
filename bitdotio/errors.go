package bitdotio

import "encoding/json"

// APIError indicates a completed API response with an error status.
type APIError struct {
	Status int    `json:"status,omitempty"`
	Body   string `body:"body,omitempty"`
}

func (e *APIError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}
