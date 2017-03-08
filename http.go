package mailjet

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sync"
)

type httpClient struct {
	client        *http.Client
	apiKeyPublic  string
	apiKeyPrivate string
	headers       map[string]string
	request       *http.Request
	response      interface{}
	mutex         *sync.Mutex
}

func newHTTPClient(apiKeyPublic, apiKeyPrivate string) *httpClient {
	return &httpClient{
		client:        http.DefaultClient,
		apiKeyPublic:  apiKeyPublic,
		apiKeyPrivate: apiKeyPrivate,
		mutex:         new(sync.Mutex),
	}
}

// APIKeyPublic returns the public key.
func (c *httpClient) APIKeyPublic() string {
	return c.apiKeyPublic
}

// APIKeyPrivate returns the secret key.
func (c *httpClient) APIKeyPrivate() string {
	return c.apiKeyPrivate
}

func (c *httpClient) Client() *http.Client {
	return c.client
}

func (c *httpClient) SetClient(client *http.Client) {
	c.mutex.Lock()
	c.client = client
	c.mutex.Unlock()
}

func (c *httpClient) Send(req *http.Request) *httpClient {
	c.mutex.Lock()
	c.request = req
	c.mutex.Unlock()
	return c
}

func (c *httpClient) With(headers map[string]string) *httpClient {
	c.mutex.Lock()
	c.headers = headers
	c.mutex.Unlock()
	return c
}

func (c *httpClient) Read(response interface{}) *httpClient {
	c.mutex.Lock()
	c.response = response
	c.mutex.Unlock()
	return c
}

func (c *httpClient) Call() (count, total int, err error) {
	defer c.reset()
	for key, value := range c.headers {
		c.request.Header.Add(key, value)
	}

	resp, err := c.doRequest(c.request)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return count, total, err
	} else if resp == nil {
		return count, total, fmt.Errorf("empty response")
	}

	if c.response != nil {
		if resp.Header["Content-Type"] != nil {
			contentType := resp.Header["Content-Type"][0]
			if contentType == "application/json" {
				return readJSONResult(resp.Body, c.response)
			} else if contentType == "text/csv" {
				c.response, err = csv.NewReader(resp.Body).ReadAll()
			}
		}
	}

	return count, total, err
}

func (c *httpClient) reset() {
	c.headers = make(map[string]string)
	c.request = nil
	c.response = nil
}
