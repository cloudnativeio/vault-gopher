package apis


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/trx35479/vault-gopher/pkg/client"
	"github.com/trx35479/vault-gopher/pkg/log"
	"github.com/trx35479/vault-gopher/pkg/models"
)

type Client struct {
	httpClient client.Client
}

var logger = log.NewLogger()

// GetClientToken function to get the needed token before data can be provided by vault
// Note that we will use the kubernetes auth on vault
func (c *Client) GetClientToken(requestBody []byte, url, namespace string) (interface{}, error) {
	client := c.httpClient.Http(false)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating the request for url: %s", url)
	}
	// Add namespace header before sent to the vault server
	req.Header.Add("X-Vault-Namespace", namespace)
	// Set the user-agent so it will be identifiable in the logs
	// Initially set to the name of the application, add the version of it in the future
	req.Header.Set("User-Agent", "vault-gopher")
	// Send the request to Vault
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to vault api for url: %s", url)
	}

	logger.LogGopher(resp, req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body")
	}
	// Additional check if the payload return by the vault has an error
	if checkError(body) {
		return nil, fmt.Errorf("error found in the payload")
	}
	var token *models.Payload
	err = json.Unmarshal([]byte(body), &token)
	if err != nil {
		return nil, fmt.Errorf("error handling the payload")
	}

	var clientToken string
	clientToken = token.Auth.ClientToken

	return clientToken, nil
}

// GetData function to get the secret data from vault
// This should be executed after the login is successful
func (c *Client) GetData(token, url, namespace string) (map[string]interface{}, error) {
	client := c.httpClient.Http(false)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the request for url: %s", url)
	}
	// Add the header X-Vault-Token in the http request
	req.Header.Add("X-Vault-Token", token)
	// Add namespace header before sent to the vault server
	req.Header.Add("X-Vault-Namespace", namespace)
	// Set the user-agent so it will be identifiable in the logs
	// Initially set to the name of the application, add the version of it in the future
	req.Header.Set("User-Agent", "vault-gopher")
	// Send the request to Vault
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to Vault API for url: %s", url)
	}

	logger.LogGopher(resp, req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body")
	}
	// Additional check if the payload return by the vault has an error
	if checkError(body) {
		return nil, fmt.Errorf("error found in the payload")
	}

	// This is to protect runtime error rather return a nil value to handler
	// There's always a posibility that the secret object is empty
	if len(body) != 0 {
		var data *models.Payload
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			return nil, fmt.Errorf("error handling the payload")
		}

		rData := data.Data.Data.(map[string]interface{})

		return rData, nil
	}
	return nil, nil
}

// RevokeToken function revoke self token so vault won't have to keep the token alive for 900s
func (c *Client) RevokeToken(vaultAddress, path, token, namespace string) (ok bool, err error) {
	client := c.httpClient.Http(false)
	requestUrl := fmt.Sprintf("%s/v1/%s", vaultAddress, path)
	req, err := http.NewRequest(http.MethodPost, requestUrl, nil)
	if err != nil {
		return false, fmt.Errorf("error creating the request for url: %s", requestUrl)
	}
	// Add the header X-Vault-Token in the http request
	req.Header.Add("X-Vault-Token", token)
	// Add namespace header before sent to the vault server
	req.Header.Add("X-Vault-Namespace", namespace)
	// Set the user-agent so it will be identifiable in the logs
	// Initially set to the name of the application, add the version of it in the future
	req.Header.Set("User-Agent", "vault-gopher")
	// Send the request to Vault
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request to vault api for url: %s", requestUrl)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error response fromvault api for url: %s", requestUrl)
	}
	return true, nil
}

// GetStatus is fix to query the status of vault endpoint
// This is especially if you are using istio service mesh in kubernetes cluster
func (c *Client) GetStatus(address, path string) error {
	client := c.httpClient.Http(false)

	// Loop and send the request in an 1 sec interval
	for i := 0; ; i++ {
		url := fmt.Sprintf("%s/v1/%s", strings.Trim(address, "/"), strings.Trim(path, "/"))
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("error creating the request for url: %s", url)
		}
		req.Header.Set("User-Agent", "vault-gopher")

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		logger.LogGopher(resp, req)

		if resp.StatusCode == 200 || resp.StatusCode == 429 {
			break
		}
	}
	return nil
}

// checkError function to find error key in the slice return bool
func checkError(data []byte) bool {
	var payload map[string]interface{}

	err := json.Unmarshal([]byte(data), &payload)
	if err != nil {
		logger.Fatal(err)
	}
	for k := range payload {
		if k == "errors" {
			return true
		}
	}
	return false
}
