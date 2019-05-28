package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

func sqlColumnOptions(null bool, pk bool, unique bool) string {
	options := ""
	if null {
		options += " NULL"
	} else {
		options += " NOT NULL"
	}
	if pk {
		options += " PRIMARY KEY"
	} else if unique {
		options += " UNIQUE"
	}
	return options
}

func sqlCreateQuery(table string, values Values) (string, []interface{}) {
	cols := make([]string, 0, len(values))
	vals := make([]interface{}, 0, len(values))
	placeholders := make([]string, 0, len(values))
	index := 1
	for col, val := range values {
		cols = append(cols, fmt.Sprintf("'%s'", col))
		vals = append(vals, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index))
		index += 1
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
	)
	return query, vals
}

func sqlInsertQuery(i Instance, fields []string) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.Model.fields))
	cols := make([]string, 0, len(i.Model.fields))
	placeholders := make([]string, 0, len(i.Model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.Model.fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s'", name))
				vals = append(vals, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", index))
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s'", name))
				vals = append(vals, val)
				placeholders = append(
					placeholders, fmt.Sprintf("$%d", index+1),
				)
			}
		}
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		i.Model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, vals
}

func sqlUpdateQuery(i Instance, fields []string) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.Model.fields))
	cols := make([]string, 0, len(i.Model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.Model.fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s' = $%d", name, index))
				vals = append(vals, val)
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s' = $%d", name, index+1))
				vals = append(vals, val)
			}
		}
	}
	vals = append(vals, i.Get(i.Model.pk))
	query := fmt.Sprintf(
		"UPDATE \"%s\" SET %s WHERE \"%s\" = $%d",
		i.Model.Table(),
		strings.Join(cols, ", "),
		i.Model.pk,
		len(cols)+1,
	)
	return query, vals
}

func getContainerType(container Container) (string, error) {
	switch container.(type) {
	case Values:
		return containers.Map, nil
	default:
		if _, ok := container.(Builder); ok {
			return containers.Builder, nil
		} else {
			ct := reflect.TypeOf(container)
			if ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}
			if ct.Kind() == reflect.Struct {
				return containers.Struct, nil
			}
		}
		return "", fmt.Errorf("invlid container")
	}
}

func getRecipients(qs QuerySet, conType string) (Container, []interface{}) {
	container := qs.Container()
	recipients := make([]interface{}, 0, len(qs.Columns()))
	switch conType {
	case containers.Map:
		for _, name := range qs.Columns() {
			val := qs.Model().fields[name].NativeVal()
			recipients = append(recipients, &val)
		}
	case containers.Builder:
		recipients = container.(Builder).Recipients(qs.Columns())
	default:
		cv := reflect.Indirect(reflect.ValueOf(container))
		for _, name := range qs.Columns() {
			f := cv.FieldByName(strings.Title(name))
			if f.IsValid() && f.CanAddr() {
				recipients = append(recipients, f.Addr().Interface())
			}
		}
	}
	return container, recipients
}
