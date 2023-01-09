package bitdotio

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// APIClient provides an interface for potential mocking of an actual HTTP client
type APIClient interface {
	Call(method, path string, body []byte) ([]byte, error)
	CallMultipart(method, path string, fields map[string]io.Reader, files fileParts) ([]byte, error)
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
func (c *DefaultAPIClient) Call(method, path string, data []byte) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}
	req, err := c.NewRequest(method, path, body)
	req.Header.Add("Accept", "application/json")

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
func (c *DefaultAPIClient) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	path, err := url.JoinPath(APIURL, APIVersion, path)
	if err != nil {
		err = fmt.Errorf("failed to construct request path: %v", err)
	}
	// This method is shared with requests with no body, so need to handle nil.
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("User-Agent", UserAgent)

	return req, nil
}

// formFile defines a file part for a multipart/form-data body
type formFile struct {
	filename string
	file     io.Reader
}

// fieldParts contains field value parts for a multipart/form-data body
type fieldParts map[string]io.Reader

// fileParts contains file parts for a multipart/form-data body
type fileParts map[string]*formFile

// Call creates and executes an authenticated HTTP request against bit.io APIs.
func (c *DefaultAPIClient) CallMultipart(method, path string, fields map[string]io.Reader, files fileParts) ([]byte, error) {
	var reqBody bytes.Buffer
	mpWriter := multipart.NewWriter(&reqBody)
	var err error
	// Write field value parts
	for key, fieldReader := range fields {
		var fieldWriter io.Writer
		if fieldWriter, err = mpWriter.CreateFormField(key); err != nil {
			return nil, err
		}
		if _, err := io.Copy(fieldWriter, fieldReader); err != nil {
			return nil, err
		}
	}
	// Write file parts
	for key, formFile := range files {
		var fileWriter io.Writer
		if fileWriter, err = mpWriter.CreateFormFile(key, formFile.filename); err != nil {
			return nil, err
		}
		// TODO: See if mpWriter materializes entire file in memory/ if so is
		// there a streaming way to handle the file
		if _, err := io.Copy(fileWriter, formFile.file); err != nil {
			return nil, err
		}
	}
	mpWriter.Close()

	req, err := c.NewRequest(method, path, &reqBody)
	if err != nil {
		err = fmt.Errorf("failed to create a new request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())
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
