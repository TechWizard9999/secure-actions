package request_secret

type Request struct {
	Name        string `json:"name" jsonschema:"Name used to identify the secret"`
	Prompt      string `json:"prompt,omitempty" jsonschema:"Prompt shown to the user"`
	Description string `json:"description,omitempty" jsonschema:"Optional description"`
}