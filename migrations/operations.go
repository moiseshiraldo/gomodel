package migrations

import (
	"encoding/json"
	"github.com/moiseshiraldo/gomodels"
	"reflect"
	"strings"
)

type Operation struct {
	CreateModel CreateModel
}

type CreateModel struct {
	Name   string
	Fields map[string]FieldDesc
}

type FieldDesc struct {
	Options gomodels.Field
}

func (f FieldDesc) MarshalJSON() ([]byte, error) {
	data := make(map[string]gomodels.Field)
	name := strings.Split(reflect.ValueOf(f.Options).Type().String(), ".")[1]
	data[name] = f.Options
	return json.Marshal(data)
}
