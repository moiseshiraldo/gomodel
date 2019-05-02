package gomodels

import (
	"fmt"
	"strings"
)

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Values) (Constructor, error) {
	cols := make([]string, 0, len(values))
	vals := make([]interface{}, 0, len(values))
	phs := make([]string, 0, len(values))
	index := 1
	for col, val := range values {
		cols = append(cols, fmt.Sprintf("'%s'", col))
		vals = append(vals, val)
		phs = append(phs, fmt.Sprintf("$%d", index))
		index += 1
	}
	colStr := strings.Join(cols, ", ")
	phStr := strings.Join(phs, ", ")
	query := fmt.Sprintf(
		"INSERT INTO '%s' (%s) VALUES (%s)", m.Model.Table(), colStr, phStr,
	)
	db := Databases["default"]
	result, err := db.Exec(query, vals...)
	if err != nil {
		return nil, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	pk, err := result.LastInsertId()
	if err != nil {
		return nil, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	instance := Instance{m.Model, Values{m.Model.pk: pk}}
	for name, val := range values {
		instance.Values[name] = val
	}
	return &instance, nil
}
