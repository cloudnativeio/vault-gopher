package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/trx35479/vault-gopher/secret-injector/apis"
	"github.com/trx35479/vault-gopher/secret-injector/utils"
)

const (
	// ServiceAccountTokenPath is the path were the token of service is mounted.
	// This is defined in the kubernetes deployment manifest
	// Namespace where the job will be run
	// This is normally mounted onto pod/job with service account configured
	ServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	// Service account that will be use to authenticate with vault should be different from the app service account
	VaultAuthenticationPath = "/etc/vault/secret/data"

	// We need to revoke the keys right after secrets have been provided
	// This path is a constant value since its the same path regardless of authentication method you use to authenticate to vault
	//VaultRevokeAuthPath = "auth/token/revoke-self"

	// Vault health endpoint
	// we will use this endpoint to check the status of vault before we send a request
	VaultHealthEndpoint = "sys/health"
)

var (
	// Aligned variables from vault configuration
	// APPROLE_NAME is the role in vault that has an attached policy to access specific secret
	// this should name be confused with approle authentication method
	vaultAddress   = os.Getenv("VAULT_ADDR")
	vaultNamespace = os.Getenv("VAULT_NAMESPACE")
	appRoleName    = os.Getenv("APPROLE_NAME")

	// This is the variables names the app will use and it's not related to vault
	// should not be confused with vault variables or terminology
	// these are variables used by the app in runtime
	vaultAuthPath         = os.Getenv("VAULT_AUTH_PATH")
	vaultSecretPath       = os.Getenv("VAULT_SECRET_PATH")
	kubernetesServiceHost = os.Getenv("KUBERNETES_SERVICE_HOST")
)

// This type gives us the ability to mutate the request url
// By creating a method that gets the absolute path of request url
type RequestUrl struct {
	// The base url of the vault
	BaseUrl string
	// Path of the object in vault
	Path string
}

// Read the string json format from env and return it as a []byte
func getEnv(v string) []byte {
	env := os.Getenv(v)
	b := bytes.NewBufferString(env)
	data, err := ioutil.ReadAll(b)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// GetPath the absolute path, we have arbitrary path on api calls on vault and this method returns an clean path
func (s *RequestUrl) GetPath(p string) string {
	if s.Path == "" {
		log.Println("Path is need to get the absolute request url")
	}
	if s.BaseUrl == "" {
		log.Println("Url is need to get the absolute request url")
	}
	return strings.Trim(s.BaseUrl, "/") + "/v1" + "/" + strings.Trim(s.Path, "/") + "/" + strings.Trim(p, "/")
}

// Main handler that perform the api calls to vault and kubernetes
// this is called from the main function and returns data structure depending on the result of api calls
func CreateObject(objectName string) error {
	var client apis.Client

	// Read the mounted token so we can use it by adding it the vault-token header in http request
	vaultToken, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", VaultAuthenticationPath, "token"))
	if err != nil {
		return fmt.Errorf("cannot read vault token error %v", err)
	}

	// We get the client token to be used to get the secrets
	// Check the authentication method used
	// authPath is in the format of auth/"method", where method is either kubernetes or approle
	// ensure that authPath is free of leading and trailing "/" so we can get the right slice
	authPath := strings.Split(strings.Trim(vaultAuthPath, "/"), "/")
	if len(authPath) != 2 {
		return fmt.Errorf("incorrect number of slice found in auth path: %d", len(authPath))
	}
	method := authPath[1]

	// Initialise a request body variable with the type of []byte
	var requestBody []byte

	// Ensure the token in a slice is not null before doing the switch statement
	// if the token in slice is null, return an error
	if method != "" {
		switch method {
		case "kubernetes":
			kubernetesData, err := json.Marshal(map[string]string{
				"jwt":  string(vaultToken),
				"role": appRoleName,
			})
			if err != nil {
				return fmt.Errorf("failed to construct json payload for auth method: %s", method)
			}
			requestBody = kubernetesData
		case "approle":
			approleData, err := json.Marshal(map[string]string{
				"secret_id": string(vaultToken),
				"role_id":   appRoleName,
			})
			if err != nil {
				return fmt.Errorf("failed to construct json payload for auth method: %s", method)
			}
			requestBody = approleData
		default:
			return fmt.Errorf("no matching auth method found ")
		}
	} else {
		return fmt.Errorf("variable VAULT_AUTH_PATH was not set")
	}

	// Additional check the endpoint of the vault
	// ATLS-618 Add poll of vault endpoint/sleep in gopher startup
	err = client.GetStatus(vaultAddress, VaultHealthEndpoint)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	// Initialise a struct to get the full path of the authentication url in vault
	// we use then the mounted token / service account token
	loginUrl := &RequestUrl{
		BaseUrl: vaultAddress,
		Path:    vaultAuthPath,
	}
	loginAuthPath := loginUrl.GetPath("login")
	clientToken, err := client.GetClientToken(requestBody, loginAuthPath, vaultNamespace)
	if err != nil {
		return fmt.Errorf("error encountered while authenticating to vault: %s", err)
	}

	// ATLS-627 support for secret segregation
	cm := getEnv("SECRET_OBJECT")

	var vars map[string][]string

	if err := json.Unmarshal([]byte(cm), &vars); err != nil {
		return fmt.Errorf("error processing the map env: %s", err)
	}

	for key, values := range vars {
		// We use the temporary token that vault server provided to access the secret
		// Client token has ttl equals to 900second
		dataUrl := &RequestUrl{
			BaseUrl: vaultAddress,
			Path:    vaultSecretPath,
		}
		// Instantiate a map[string]interface{} type
		// Placeholder of the kv secret we fetch from the vault
		data := make(map[string]interface{})

		for _, value := range values {
			secretPath := dataUrl.GetPath(strings.TrimSpace(string(value)))
			payload, err := client.GetData(clientToken.(string), secretPath, vaultNamespace)
			if err != nil {
				return fmt.Errorf("encountered error while fetching secrets from vault: %s", err)
			}
			// We safeguard the runtime here
			// Sometimes's a call to secret returns an empty object
			if len(payload) != 0 {
				for key, secret := range payload {
					data[key] = secret
				}
			}
		}
		if len(data) != 0 {
			if err := create(data, objectName, strings.TrimSpace(key)); err != nil {
				return fmt.Errorf("kubernetes secret cannot be created error: %s", err)
			}
		}
	}
	return nil
}

// Handler to create the object
// ATLS-627 creating multiple object
func create(m map[string]interface{}, objectName, secretObjectName string) error {
	var client apis.Client
	// Read the token from the mount volume and parse it
	serviceAcctToken, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", ServiceAccountPath, "token"))
	if err != nil {
		return fmt.Errorf("cannot read kubernetes token error: %v", err)
	}
	// Read the token from the mount volume and parse it
	namespace, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", ServiceAccountPath, "namespace"))
	if err != nil {
		return fmt.Errorf("cannot read kuberneres namespace error: %v", err)
	}
	// Read the the token and return a byte
	cacrt, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", ServiceAccountPath, "ca.crt"))
	if err != nil {
		return fmt.Errorf("cannot read ca certificate error %v", err)
	}

	if len(m) != 0 {
		// We get that secrets payload and feed it to Object() function and return the json formatted secret object manifest for kubernetes api
		object, err := Object(secretObjectName, string(namespace), utils.EncodeValue(m))
		if err != nil {
			return fmt.Errorf("encountered error while constructing kubernetes object: %s", err)
		}
		// This call the api that checks the object in kubernetes api
		// Depending on the return values, the api call to create the object will switch between POST and PUT method
		status, err := client.Get(string(serviceAcctToken), kubernetesServiceHost,
			string(namespace), objectName, secretObjectName, cacrt)
		if err != nil {
			return fmt.Errorf("encountered error while verifying secret object in kubernetes: %s", err)
		}
		// Create the object to kubernetes api
		// Object would be created if it's not present or updated if exist, the status variable will define how the object will be created
		resp, err := client.Create(string(serviceAcctToken), kubernetesServiceHost,
			string(namespace), objectName, secretObjectName, status.(int), cacrt, object)
		if err != nil {
			return fmt.Errorf("encountered error while creating the kubernetes secret object: %s", err)
		}
		// Handle the resp coming from kubernetes api
		// loop into the struct and check if the "code" key is present in the struct
		// pretty much not a good way to handle it but kubernetes returns a nested data structure
		for key, value := range resp {
			if key == "code" {
				return fmt.Errorf("error creating secret with repond code: %v\nErrorMessage: %v", value, resp["message"])
			}
		}
	}
	return nil
}

