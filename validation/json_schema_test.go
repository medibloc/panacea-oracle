package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func Test(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	sl := gojsonschema.NewSchemaLoader()
	schemaLoader := gojsonschema.NewReferenceLoader(schemaURI)
	schema, err := sl.Compile(schemaLoader)
	require.NoError(t, err)

	result, err := validateSchema(schema, jsonInput)
	require.NoError(t, err)
	fmt.Println(result)

	result, err = validateSchema(schema, jsonInput)
	require.NoError(t, err)
	fmt.Println(result)

	result, err = validateSchema(schema, jsonInput)
	require.NoError(t, err)
	fmt.Println(result)
}

func validateSchema(schema *gojsonschema.Schema, jsonInput []byte) (*gojsonschema.Result, error) {
	result, err := schema.Validate(gojsonschema.NewBytesLoader(jsonInput))
	if err != nil {
		return nil, err
	}

	return result, nil
}
