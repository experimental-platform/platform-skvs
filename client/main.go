package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

type skvsResponse struct {
	Key       string `json:"key"`
	Namespace bool   `json:"namespace"`
	Value     string `json:"value"`
}

func GetContainerIP(name string) (string, error) {
	defaultHeaders := map[string]string{"User-Agent": "protonet-skvs_cli"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		return "", err
	}

	listOptions := types.ContainerListOptions{Filter: filters.NewArgs()}
	listOptions.Filter.Add("name", name)

	containers, err := cli.ContainerList(context.Background(), listOptions)
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("Found no container named '%s'", name)
	}

	data, err := cli.ContainerInspect(context.Background(), containers[0].ID)
	if err != nil {
		return "", err
	}

	protonetNetworkData, ok := data.NetworkSettings.Networks["protonet"]
	if !ok {
		return "", errors.New("The SKVS container doesn't belong to the network 'protonet'.")
	}

	return protonetNetworkData.IPAddress, nil
}

// Get retrieves a value of an SKVS key
// It does not propely handle namespaces
func Get(key string) (string, error) {
	ip, err := GetContainerIP("skvs")
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf("http://%s/%s", ip, key)
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
func Set(key string, value string) error {
	ip, err := GetContainerIP("skvs")
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("http://%s/%s", ip, key)
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
func Delete(key string) error {
	ip, err := GetContainerIP("skvs")
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("http://%s/%s", ip, key)
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
