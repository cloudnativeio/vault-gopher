package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/trx35479/vault-gopher/secret-injector/client"
)

func TestClient_Get(t *testing.T) {
	type fields struct {
		httpClient client.Client
	}
	type args struct {
		token      string
		host       string
		ns         string
		objectName string
		secretName string
		ca         []byte
	}
	caCrt, err := ioutil.ReadFile("../../testdata/test.crt")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "TestGet",
			fields:  fields{},
			args:    args{
				token: "token",
				host: "127.0.0.1",
				ns: "ops-sre",
				objectName: "secrets",
				secretName: "test-secret",
				ca: caCrt,
			},
			want:    200,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient: tt.fields.httpClient,
			}
			response := 200
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(response)
				accept := r.Header.Get("Accept")
				if accept != "application/json" {
					t.Errorf("Accept header is incorrect %s:", r.Header.Get("Accept"))
				}
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Content-Type header is incorrect %s:", r.Header.Get("Content-Type"))
				}
				const BEARER_SCHEMA = "Bearer "
				req := r.Header.Get("Authorization")
				jwt := req[len(BEARER_SCHEMA):]
				if jwt != "token" {
					t.Errorf("Token is incorrect %s:", r.Header.Get("Authorization"))
				}
				userAgent := r.Header.Get("User-Agent")
				if userAgent != "vault-gopher" {
					t.Errorf("User-Agent header is set incorrect %s:", r.Header.Get("User-Agent"))
				}
			}))

			go server.Config.ListenAndServeTLS("../../testdata/root.crt", "../../testdata/key.key")
			server.StartTLS()
			defer server.Close()
			time.Sleep(100 * time.Millisecond)
			got, err := c.Get(tt.args.token, tt.args.host, tt.args.ns, tt.args.objectName, tt.args.secretName, tt.args.ca)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	type fields struct {
		httpClient client.Client
	}
	type args struct {
		token      string
		host       string
		ns         string
		objectName string
		secretName string
		status     int
		ca         []byte
		payload    []byte
	}
	caCrt, err := ioutil.ReadFile("../../testdata/test.crt")
	if err != nil {
		t.Error(err)
	}
	data := map[string]interface{}{
		"test1": "test1",
		"test2": "test2",
	}
	// Kubernetes object metadata struct
	type Meta struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	// Kubernetes secret object root struct
	type Secret struct {
		ApiVersion string                 `json:"apiVersion"`
		Kind       string                 `json:"kind"`
		Data       map[string]interface{} `json:"data"`
		Type       string                 `json:"type"`
		Metadata   Meta                   `json:"metadata"`
	}
	// Instantiate a manifest struct
	manifest := &Secret{ApiVersion: "v1",
		Kind: "Secret",
		Type: "Opaque",
		Metadata: Meta{
			Name:      "test",
			Namespace: "ops-sre",
		},
		Data: data,
	}
	// This object will be sent as a payload and apis return the same payload if success
	object, _ := json.Marshal(manifest)

	// Expected return
	var ret map[string]interface{}
	err = json.Unmarshal([]byte(object), &ret)
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "TestCreate",
			fields:  fields{},
			args:    args{
				token: "token",
				host: "127.0.0.1",
				ns: "ops-sre",
				objectName: "secrets",
				secretName: "tls",
				status: 200,
				ca: caCrt,
				payload: object,
			},
			want:    ret,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient: tt.fields.httpClient,
			}
			response := string(object)
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
				accept := r.Header.Get("Accept")
				if accept != "application/json" {
					t.Errorf("Accept header is incorrect %s:", r.Header.Get("Accept"))
				}
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Content-Type header is incorrect %s:", r.Header.Get("Content-Type"))
				}
				const BEARER_SCHEMA = "Bearer "
				reqtoken := r.Header.Get("Authorization")
				jwt := reqtoken[len(BEARER_SCHEMA):]
				if jwt != "token" {
					t.Errorf("Token is incorrect %s:", r.Header.Get("Authorization"))
				}
				userAgent := r.Header.Get("User-Agent")
				if userAgent != "ob-vault-gopher" {
					t.Errorf("User-Agent header is set incorrect %s:", r.Header.Get("User-Agent"))
				}
			}))
			go server.Config.ListenAndServeTLS("../../testdata/test.pem", "../../testdata/key.pem")
			server.StartTLS()
			defer server.Close()
			time.Sleep(100 * time.Millisecond)
			got, err := c.Create(tt.args.token, tt.args.host, tt.args.ns, tt.args.objectName, tt.args.secretName, tt.args.status, tt.args.ca, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}