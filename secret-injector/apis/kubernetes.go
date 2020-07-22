package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Get check if the secret object is already configured
func (c *Client) Get(token, host, ns, objectName, secretName string, ca []byte) (interface{}, error) {
	client := c.httpClient.Https(ca)

	requestUrl := fmt.Sprintf("https://%s/api/v1/namespaces/%s/%s/%s", host, ns, objectName, secretName)
	// Instantiate an http request
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request to kubernetes api: %s", requestUrl)
	}
	// Set the accepted content type in request
	req.Header.Set("Accept", "application/json")
	// Set the content type in http request
	req.Header.Set("Content-Type", "application/json")
	// Set the authorization bearer adding the token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// Set the user-agent so it will be identifiable in the logs
	// Initially set to the name of the application, add the version of it in the future
	req.Header.Set("User-Agent", "vault-gopher")
	// Send the actual request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to kubernetes api: %s", err)
	}
	defer resp.Body.Close()

	logger.LogGopher(resp, req)

	return resp.StatusCode, nil
}

// Create creates a secret object in Kubernetes
func (c *Client) Create(token, host, ns, objectName, secretName string, status int, ca, payload []byte) (map[string]interface{}, error) {
	client := c.httpClient.Https(ca)

	// We switch method depending on the status, otherwise error will occur if object exist
	var method string
	var requestUrl string
	if status == 200 {
		method = http.MethodPut
		requestUrl = fmt.Sprintf("https://%s/api/v1/namespaces/%s/%s/%s", host, ns, objectName, secretName)
	} else {
		// else use the POST method
		method = http.MethodPost
		requestUrl = fmt.Sprintf("https://%s/api/v1/namespaces/%s/%s", host, ns, objectName)
	}
	// Instantiate an http request
	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request to kubernetes api: %s", requestUrl)
	}
	// Set the accepted content type in request
	req.Header.Set("Accept", "application/json")
	// Set the content type in http request
	req.Header.Set("Content-Type", "application/json")
	// Set the authorization bearer adding the token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// Set the user-agent so it will be identifiable in the logs
	// Initially set to the name of the application, add the version of it in the future
	req.Header.Set("User-Agent", "vault-gopher")
	// Send the actual request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to kubernetes api: %s", err)
	}

	logger.LogGopher(resp, req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body")
	}
	var ret map[string]interface{}
	err = json.Unmarshal([]byte(body), &ret)
	if err != nil {
		return nil, fmt.Errorf("error handling the payload")
	}
	return ret, nil
}

