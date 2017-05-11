package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type HTTPClient struct {
	httpClient http.Client
	endpoint   string
	apikey     string
	username   string
	password   string
	id         string
	csrf       string
	conf       HTTPClientConfig
}

type HTTPClientConfig struct {
	URLPrefix           string
	HeaderAPIKeyName    string
	Apikey              string
	HeaderClientKeyName string
	CsrfDisable         bool
}

// Inspired by syncthing/cmd/cli

const insecure = false

// HTTPNewClient creates a new HTTP client to deal with Syncthing
func HTTPNewClient(baseURL string, cfg HTTPClientConfig) (*HTTPClient, error) {

	// Create w new Http client
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
	client := HTTPClient{
		httpClient: httpClient,
		endpoint:   baseURL,
		apikey:     cfg.Apikey,
		conf:       cfg,
		/* TODO - add user + pwd support
		username:   c.GlobalString("username"),
		password:   c.GlobalString("password"),
		*/
	}

	if client.apikey == "" {
		if err := client.getCidAndCsrf(); err != nil {
			return nil, err
		}
	}
	return &client, nil
}

// Send request to retrieve Client id and/or CSRF token
func (c *HTTPClient) getCidAndCsrf() error {
	request, err := http.NewRequest("GET", c.endpoint, nil)
	if err != nil {
		return err
	}
	if _, err := c.handleRequest(request); err != nil {
		return err
	}
	if c.id == "" {
		return errors.New("Failed to get device ID")
	}
	if !c.conf.CsrfDisable && c.csrf == "" {
		return errors.New("Failed to get CSRF token")
	}
	return nil
}

// GetClientID returns the id
func (c *HTTPClient) GetClientID() string {
	return c.id
}

// formatURL Build full url by concatenating all parts
func (c *HTTPClient) formatURL(endURL string) string {
	url := c.endpoint
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += strings.TrimLeft(c.conf.URLPrefix, "/")
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url + strings.TrimLeft(endURL, "/")
}

// HTTPGet Send a Get request to client and return an error object
func (c *HTTPClient) HTTPGet(url string, data *[]byte) error {
	_, err := c.HTTPGetWithRes(url, data)
	return err
}

// HTTPGetWithRes Send a Get request to client and return both response and error
func (c *HTTPClient) HTTPGetWithRes(url string, data *[]byte) (*http.Response, error) {
	request, err := http.NewRequest("GET", c.formatURL(url), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.handleRequest(request)
	if err != nil {
		return res, err
	}
	if res.StatusCode != 200 {
		return res, errors.New(res.Status)
	}

	*data = c.responseToBArray(res)

	return res, nil
}

// HTTPPost Send a POST request to client and return an error object
func (c *HTTPClient) HTTPPost(url string, body string) error {
	_, err := c.HTTPPostWithRes(url, body)
	return err
}

// HTTPPostWithRes Send a POST request to client and return both response and error
func (c *HTTPClient) HTTPPostWithRes(url string, body string) (*http.Response, error) {
	request, err := http.NewRequest("POST", c.formatURL(url), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	res, err := c.handleRequest(request)
	if err != nil {
		return res, err
	}
	if res.StatusCode != 200 {
		return res, errors.New(res.Status)
	}
	return res, nil
}

func (c *HTTPClient) responseToBArray(response *http.Response) []byte {
	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// TODO improved error reporting
		fmt.Println("ERROR: " + err.Error())
	}
	return bytes
}

func (c *HTTPClient) handleRequest(request *http.Request) (*http.Response, error) {
	if c.conf.HeaderAPIKeyName != "" && c.apikey != "" {
		request.Header.Set(c.conf.HeaderAPIKeyName, c.apikey)
	}
	if c.conf.HeaderClientKeyName != "" && c.id != "" {
		request.Header.Set(c.conf.HeaderClientKeyName, c.id)
	}
	if c.username != "" || c.password != "" {
		request.SetBasicAuth(c.username, c.password)
	}
	if c.csrf != "" {
		request.Header.Set("X-CSRF-Token-"+c.id[:5], c.csrf)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	// Detect client ID change
	cid := response.Header.Get(c.conf.HeaderClientKeyName)
	if cid != "" && c.id != cid {
		c.id = cid
	}

	// Detect CSR token change
	for _, item := range response.Cookies() {
		if item.Name == "CSRF-Token-"+c.id[:5] {
			c.csrf = item.Value
			goto csrffound
		}
	}
	// OK CSRF found
csrffound:

	if response.StatusCode == 404 {
		return nil, errors.New("Invalid endpoint or API call")
	} else if response.StatusCode == 401 {
		return nil, errors.New("Invalid username or password")
	} else if response.StatusCode == 403 {
		if c.apikey == "" {
			// Request a new Csrf for next requests
			c.getCidAndCsrf()
			return nil, errors.New("Invalid CSRF token")
		}
		return nil, errors.New("Invalid API key")
	} else if response.StatusCode != 200 {
		data := make(map[string]interface{})
		// Try to decode error field of APIError struct
		json.Unmarshal(c.responseToBArray(response), &data)
		if err, found := data["error"]; found {
			return nil, fmt.Errorf(err.(string))
		} else {
			body := strings.TrimSpace(string(c.responseToBArray(response)))
			if body != "" {
				return nil, fmt.Errorf(body)
			}
		}
		return nil, errors.New("Unknown HTTP status returned: " + response.Status)
	}
	return response, nil
}
