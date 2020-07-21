package models

type Payload struct {
	RequestId string `json:"request_id"`
	LeaseId   string `json:"lease_id"`
	Renewable bool   `json:"renewable"`
	Data      struct {
		Data interface{} `json:"data"`
	} `json:"data"`
	WrapInfo string `json:"wrap_info"`
	Warnings string `json:"warnings"`
	Auth     struct {
		ClientToken   string   `json:"client_token"`
		Accessor      string   `json:"accessor"`
		Policies      []string `json:"policies"`
		TokenPolicies []string `json:"token_policies"`
		Metadata      struct {
			Role                     string `json:"role"`
			ServiceAccountName       string `json:"service_account_name"`
			ServiceAccountNamespace  string `json:"service_account_namespace"`
			ServiceAccountSecretName string `json:"service_account_secret_name"`
			ServiceAccountUid        string `json:"service_account_uid"`
		} `json:"metadata"`
		LeaseDuration int    `json:"lease_duration"`
		Renewable     bool   `json:"renewable"`
		EntityId      string `json:"entity_id"`
		TokenType     string `json:"token_type"`
		Orphan        bool   `json:"orphan"`
	} `json:"auth"`
	Errors []string `json:"errors,omitempty"`
}
