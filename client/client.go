package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/experimental-platform/platform-utils/dockerutil"
)

type skvsResponse struct {
	Key       string `json:"key"`
	Namespace bool   `json:"namespace"`
	Value     string `json:"value"`
}

// Client is an object asociated with a server instance's access URL
type Client struct {
	url string
}

// NewFromDocker returns a client attached to an SKVS instance running on
// a local machine in a container named 'skvs'
func NewFromDocker() (*Client, error) {
	ip, err := dockerutil.GetContainerIP("skvs")
	if err != nil {
		return nil, err
	}

	return &Client{url: fmt.Sprintf("http://%s", ip)}, nil
}

// NewFromURL returns a client attached to an SKVS instance reachable
// through a specified URL
func NewFromURL(url string) *Client {
	return &Client{url: url}
}

func buildFullURL(baseURL, key string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, key)
	return u.String(), nil
}

// Get retrieves a value of an SKVS key
// It does not propely handle namespaces
func (c *Client) Get(key string) (string, error) {
	requestURL, err := buildFullURL(c.url, key)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(requestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("SKVS responded with %s", resp.Status)
	}

	responseBodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseStruct skvsResponse

	err = json.Unmarshal(responseBodyData, &responseStruct)
	if err != nil {
		return "", err
	}

	return responseStruct.Value, nil
}

// Set sets a value of a given SKVS key
func (c *Client) Set(key string, value string) error {
	requestURL, err := buildFullURL(c.url, key)
	if err != nil {
		return err
	}

	vals := url.Values{}
	vals.Set("value", value)
	resp, err := http.PostForm(requestURL, vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("SKVS responded with %s", resp.Status)
	}

	return nil
}

// Delete removes an SKVS entry and all its children
func (c *Client) Delete(key string) error {
	requestURL, err := buildFullURL(c.url, key)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("SKVS responded with %s", resp.Status)
	}

	return nil
}
