package config

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

//go:generate go run ../../cmd/zagreus schema --output ../../zagreus.schema.json

func JSONSchema() *jsonschema.Schema {
	durationType := reflect.TypeFor[Duration]()
	reflector := &jsonschema.Reflector{
		FieldNameTag: "yaml",
		Mapper: func(valueType reflect.Type) *jsonschema.Schema {
			if valueType == durationType {
				return &jsonschema.Schema{
					Type:        "string",
					Description: "A Go duration such as 1m, 15m, or 1h.",
					Examples:    []any{"1m", "15m", "1h"},
				}
			}
			return nil
		},
	}
	schema := reflector.Reflect(&Config{})
	schema.ID = jsonschema.ID("https://raw.githubusercontent.com/woodleighschool/zagreus/main/zagreus.schema.json")
	schema.Title = "Zagreus configuration"
	schema.Description = "External API connections, sync parameters"
	return schema
}

func JSONSchemaDocument() ([]byte, error) {
	document, err := json.MarshalIndent(JSONSchema(), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}
