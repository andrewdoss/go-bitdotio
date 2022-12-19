package bitdotio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// APIClient provides an interface for potential mocking of an actual HTTP client
// TODO: Need to confirm this is the right organization/interface to enable
// testing of the BitDotIO methods. Will confirm when I write actual tests and
// get up to speed on Go HTTP mocking.
type APIClient interface {
	Call(method, path string, body []byte) ([]byte, error)
}

// DefaultAPIClient implements APIClient
type DefaultAPIClient struct {
	accessToken string
	HTTPClient  *http.Client
	Logger      Logger
}

// NewDefaultAPIClient constructs a default client for making API HTTP requests.
func NewDefaultAPIClient(accessToken string) *DefaultAPIClient {
	return &DefaultAPIClient{
		accessToken: accessToken,
		HTTPClient:  &http.Client{},
		Logger:      newDefaultLogger(),
	}
}

// Call creates and executes an authenticated HTTP request against bit.io APIs.
// TODO: Need to think more about the signature/interface — sending a nill argument
// for all requests without a body doens't seem quite right.
func (c *DefaultAPIClient) Call(method, path string, reqBody []byte) ([]byte, error) {
	req, err := c.NewRequest(method, path, reqBody)
	if err != nil {
		err = fmt.Errorf("failed to create a new request: %v", err)
		c.Logger.Errorf("%v", err)
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %v", err)
		c.Logger.Errorf("%v", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %v", err)
		c.Logger.Errorf("%v", err)
		return nil, err
	}

	if res.StatusCode >= 400 {
		err = fmt.Errorf("%s: %s", res.Status, string(body))
		c.Logger.Errorf("%v", err)
	}

	return body, err
}

// NewRequest constructs requests for bit.io APIs.
func (c *DefaultAPIClient) NewRequest(method, path string, body []byte) (*http.Request, error) {
	path, err := url.JoinPath(APIURL, APIVersion, path)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
		c.Logger.Errorf("%v", err)
	}

	// TODO: Find a cleaner way to handle potentially nil body
	var req *http.Request
	if body != nil {
		buf := bytes.NewBuffer(body)
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

	// TODO: Handle setting other headers with parms, if needed

	return req, nil
}
