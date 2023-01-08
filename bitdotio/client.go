package bitdotio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// APIClient provides an interface for potential mocking of an actual HTTP client
type APIClient interface {
	Call(method, path string, body []byte) ([]byte, error)
}

// DefaultAPIClient implements APIClient
type DefaultAPIClient struct {
	accessToken string
	HTTPClient  *http.Client
}

// NewDefaultAPIClient constructs a default client for making API HTTP requests.
func NewDefaultAPIClient(accessToken string) *DefaultAPIClient {
	return &DefaultAPIClient{
		accessToken: accessToken,
		HTTPClient:  &http.Client{},
	}
}

// Call creates and executes an authenticated HTTP request against bit.io APIs.
func (c *DefaultAPIClient) Call(method, path string, reqBody []byte) ([]byte, error) {
	req, err := c.NewRequest(method, path, reqBody)
	if err != nil {
		err = fmt.Errorf("failed to create a new request: %v", err)
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)

	var resBody []byte
	if err == nil {
		resBody, err = io.ReadAll(res.Body)
		res.Body.Close()
	}

	if err != nil {
		err = fmt.Errorf("request failed with error: %v", err)
	} else if res.StatusCode >= 400 {
		err = c.HandleErrorResponse(res, resBody)
	}

	return resBody, err
}

// HandleErrorResponse converts an Error API response to an Error.
// TODO: Possibly should provide further unmarshalling of error body.
func (s *DefaultAPIClient) HandleErrorResponse(res *http.Response, resBody []byte) error {
	return &APIError{Status: res.StatusCode, Body: string(resBody)}
}

// NewRequest constructs requests for bit.io APIs.
func (c *DefaultAPIClient) NewRequest(method, path string, body []byte) (*http.Request, error) {
	path, err := url.JoinPath(APIURL, APIVersion, path)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
	}

	// This method is shared with requests with no body, so need to handle nil.
	var req *http.Request
	if body != nil {
		buf := bytes.NewReader(body)
		req, err = http.NewRequest(method, path, buf)
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("User-Agent", UserAgent)

	return req, nil
}
