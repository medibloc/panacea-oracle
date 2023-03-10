package validation

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleSchema(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	schema := NewJSONSchema()
	err := schema.ValidateJSONSchema(jsonInput, schemaURI)
	require.NoError(t, err)

	require.Equal(t, 1, schema.cache.Size())
}

func TestTwoSchema(t *testing.T) {
	schemaURIs := []string{
		"https://json.schemastore.org/cdk.json",
		"https://json.schemastore.org/vsconfig.json",
	}

	schema := NewJSONSchema()
	jsonInput := []byte(`{
		"version":"1.0.0",
		"components": "",
		"versionReporting": true
	}`)
	err := schema.ValidateJSONSchemata(jsonInput, schemaURIs)
	require.ErrorContains(t, err, "components: Invalid type. Expected: array, given: string")

	require.Equal(t, 2, schema.cache.Size())
}
