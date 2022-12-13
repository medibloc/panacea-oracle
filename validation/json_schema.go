package validation

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateJSONSchemata gets data from desiredSchemaURI list one by one and calls ValidateJSONSchema.
func ValidateJSONSchemata(jsonInput []byte, desiredSchemaURIs []string) error {
	for _, desiredSchemaURI := range desiredSchemaURIs {
		err := ValidateJSONSchema(jsonInput, desiredSchemaURI)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateJSONSchema performs the JSON Schema validation: https://json-schema.org/
// This fetches the schema definition via http(s) or local filesystem.
// If jsonInput is not a valid JSON or if jsonInput doesn't conform to the desired JSON schema, an error is returned.
//
// TODO: accept io.Reader instead of []byte
func ValidateJSONSchema(jsonInput []byte, desiredSchemaURI string) error {
	schemaLoader := gojsonschema.NewReferenceLoader(desiredSchemaURI)
	docLoader := gojsonschema.NewBytesLoader(jsonInput)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return fmt.Errorf("failed to validate JSON schema: %w", err)
	}

	if !result.Valid() {
		var sb strings.Builder
		for _, err := range result.Errors() {
			sb.WriteString("\n\t")
			sb.WriteString(err.String())
		}
		return fmt.Errorf("JSON doc doesn't conform to the desired JSON schema: %s", sb.String())
	}

	return nil
}
