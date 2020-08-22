package handler

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Kubernetes object metadata struct
type Meta struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Labels    map[string]interface{} `json:"labels"`
}

// Kubernetes secret object root struct
type Secret struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Data       map[string]interface{} `json:"data"`
	Type       string                 `json:"type"`
	Metadata   interface{}            `json:"metadata"`
}

// Construct the kubernetes manifest and return it as a byte
// the manifest will be in json format
func object(name, ns string, data map[string]interface{}) ([]byte, error) {
	// get the appName and inject it to metadata.labels
	var appName string

	// strings.Split returns a slice of byte [byte1 byte2 byte3]
	str := strings.Split(name, "-")
	// verify if the slice is not empty otherwise return an error
	if len(str) == 0 {
		return nil, fmt.Errorf("name of secrets does not satisfy the naming requirement")
	}
	// get the second byte from right
	env := str[len(str)-2]
	// current env name
	envName := strings.Split(ns, "-")[0]
	// get the last byte from right
	lastByte := str[len(str)-1]

	// check the name of object has the env string or not and return the app name only
	if env == envName {
		if lastByte == "secret" || lastByte == "secrets" {
			appName = strings.Join(str[:len(str)-2], "-")
		} else {
			appName = strings.Join(str[:len(str)-1], "-")
		}
	} else if lastByte == "secret" || lastByte == "secrets" || lastByte == envName {
		appName = strings.Join(str[:len(str)-1], "-")
	} else {
		appName = name
	}

	// get the component or the environment
	// namespace naming convention is in this format env + group/squad ex. sit-sre
	var component string

	// verify if the slice is not empty otherwise return an error
	namespace := strings.Split(ns, "-")
	if len(namespace) == 0 {
		return nil, fmt.Errorf("namespace does not satisfy the naming requirement")
	}
	// get the component or the environment
	component = strings.Join(namespace[1:], "-")

	// construct the manifest and return the values
	manifest := &Secret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		Data:       data,
		Metadata: &Meta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]interface{}{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/component":  component,
				"app.kubernetes.io/managed-by": "vault-gopher",
			},
		},
	}

	ret, _ := json.Marshal(manifest)
	return ret, nil
}

