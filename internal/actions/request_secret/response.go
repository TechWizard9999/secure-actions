package request_secret

type Response struct {
	SecretName string `json:"secretName" jsonschema:"Name of the stored secret"`
	Stored     bool   `json:"stored" jsonschema:"Whether the secret was stored"`
	Message    string `json:"message" jsonschema:"Status message"`
}