package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

type Operation struct {
	CreateModel CreateModel
}

type CreateModel struct {
	Name   string
	Fields map[string]FieldOptions
}

type FieldOptions struct {
	Type    string
	Options gomodels.Field
}

func (f *FieldOptions) UnmarshalJSON(data []byte) error {
	obj := struct {
		Type    string
		Options json.RawMessage
	}{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	f.Type = obj.Type
	native, ok := gomodels.AvailableFields()[obj.Type]
	if !ok {
		return fmt.Errorf("invalid field type: %s", obj.Type)
	}
	f.Options, err = native.FromJson(obj.Options)
	if err != nil {
		return err
	}
	return nil
}
