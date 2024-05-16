package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// HTTPClient is a struct that represents an HTTP client.
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient is a constructor function that creates a new instance of HTTPClient.
func NewHTTPClient(connTimeout time.Duration, maxIdleConn int) *HTTPClient {
	if maxIdleConn == 0 {
		maxIdleConn = 10
	}

	if connTimeout == 0 {
		connTimeout = time.Second * 30

	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConn,
			IdleConnTimeout:     connTimeout,
			DisableCompression:  true,
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: time.Second * 10,
		},
	}
	return &HTTPClient{
		client: client,
	}
}

// Get performs an HTTP GET request and returns the response body, status code, and error if any.
func (c *HTTPClient) Get(url string) ([]byte, int, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}

// Post performs an HTTP POST request and returns the response body, status code, and error if any.
func (c *HTTPClient) Post(url string, object interface{}) ([]byte, int, error) {
	by, err := json.Marshal(object)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(by))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return responseBody, resp.StatusCode, nil
}

// Delete performs an HTTP DELETE request and returns the response body, status code, and error if any.
func (c *HTTPClient) Delete(url string) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}
