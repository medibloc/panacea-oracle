package validation

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type JSONSchema struct {
	cache *schemaCache
}

func NewJSONSchema() *JSONSchema {
	return &JSONSchema{
		newSchemaCache(),
	}
}

// ValidateJSONSchemata gets data from desiredSchemaURI list one by one and calls ValidateJSONSchema.
func (s *JSONSchema) ValidateJSONSchemata(jsonInput []byte, desiredSchemaURIs []string) error {
	for _, desiredSchemaURI := range desiredSchemaURIs {
		if err := s.ValidateJSONSchema(jsonInput, desiredSchemaURI); err != nil {
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
func (s *JSONSchema) ValidateJSONSchema(jsonInput []byte, desiredSchemaURI string) error {
	schema, err := s.cache.Get(desiredSchemaURI)
	if err != nil {
		return fmt.Errorf("failed to get schema. %w", err)
	}

	docLoader := gojsonschema.NewBytesLoader(jsonInput)

	result, err := schema.Validate(docLoader)
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

// newReferenceSchema creates the corresponding JSON Schema of the URI
func newReferenceSchema(schemaURI string) (*gojsonschema.Schema, error) {
	jsonLoader := gojsonschema.NewReferenceLoader(schemaURI)
	return gojsonschema.NewSchema(jsonLoader)
}
