package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/trx35479/vault-gopher/secret-injector/client"
)

func TestClient_GetClientToken(t *testing.T) {
	type fields struct {
		httpClient client.Client
	}
	type args struct {
		requestBody []byte
		url         string
		namespace   string
	}
	data, _ := json.Marshal(map[string]interface{}{
		"jwt":  "eyjt",
		"role": "sre",
	})
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "GetClientToken",
			fields:  fields{},
			args:    args{
				requestBody: data,
				url: "http://vault.com/v1/some/auth/path",
				namespace: "sre-ns",
			},
			want:    "token",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient: tt.fields.httpClient,
			}
			response := `{"auth":{"client_token": "token"}}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
				ns := r.Header.Get("X-Vault-Namespace")
				if ns != "sre-ns" {
					t.Errorf("Namespace is incorrect %s:", r.Header.Get("X-Vault-Namespace"))
				}
				var reqBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Error("Unexpected body in request")
				}
				if reqBody["jwt"] != "eyjt" || reqBody["role"] != "sre" {
					t.Errorf("Missing or incorrect jwt %s and role %s", reqBody["jwt"], reqBody["role"])
				}
			}))
			defer server.Close()
			got, err := c.GetClientToken(tt.args.requestBody, tt.args.url, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClientToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClientToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetData(t *testing.T) {
	type fields struct {
		httpClient client.Client
	}
	type args struct {
		token     string
		url       string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "GetData",
			fields:  fields{},
			args:    args{token: "token",
				url: "http://vault.com/v1/secret/data/tls",
				namespace: "sre-ns",
			},
			want:    map[string]interface{}{"tls": "data"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient: tt.fields.httpClient,
			}
			response := `{"data":{"data":{"tls": "data"}}}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
				ns := r.Header.Get("X-Vault-Namespace")
				if ns != "sre-ns" {
					t.Errorf("Namespace is incorrect %s:", r.Header.Get("X-Vault-Namespace"))
				}
				jwt := r.Header.Get("X-Vault-Token")
				if jwt != "token" {
					t.Errorf("Token is incorrect %s:", r.Header.Get("X-Vault-Token"))
				}
			}))
			defer server.Close()
			got, err := c.GetData(tt.args.token, tt.args.url, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetData() got = %v, want %v", got, tt.want)
			}
		})
	}
}