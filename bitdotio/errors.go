package bitdotio

import "encoding/json"

type APIError struct {
	Status int    `json:"status,omitempty"`
	Body   string `body:"body,omitempty"`
}

func (e *APIError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}
