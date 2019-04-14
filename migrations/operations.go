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
	switch obj.Type {
	case "IntegerField":
		field := gomodels.IntegerField{}
		err = json.Unmarshal(obj.Options, &field)
		f.Options = field
	case "CharField":
		field := gomodels.CharField{}
		err = json.Unmarshal(obj.Options, &field)
		f.Options = field
	case "BooleanField":
		field := gomodels.BooleanField{}
		err = json.Unmarshal(obj.Options, &field)
		f.Options = field
	case "AutoField":
		field := gomodels.AutoField{}
		err = json.Unmarshal(obj.Options, &field)
		f.Options = field
	default:
		return fmt.Errorf("invalid field type: %s", obj.Type)
	}
	if err != nil {
		return err
	}
	fmt.Printf("%T\n", f.Options)
	return nil
}
