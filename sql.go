package gomodels

import (
	"fmt"
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

func sqlCreateQuery(
	model *Model, values Container, driver string,
) (string, []interface{}) {
	cols := make([]string, 0, len(model.fields))
	vals := make([]interface{}, 0, len(model.fields))
	placeholders := make([]string, 0, len(model.fields))
	index := 1
	for name := range model.fields {
		if getter, ok := values.(Getter); ok {
			if val, ok := getter.Get(name); ok {
				cols = append(cols, fmt.Sprintf("\"%s\"", name))
				vals = append(vals, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", index))
				index += 1
			}
		} else if val, ok := getStructField(values, name); ok {
			cols = append(cols, fmt.Sprintf("\"%s\"", name))
			vals = append(vals, val)
			placeholders = append(placeholders, fmt.Sprintf("$%d", index))
			index += 1
		}
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	if driver == "postgres" {
		query = fmt.Sprintf("%s RETURNING \"%s\"", query, model.pk)
	}
	return query, vals
}

func sqlInsertQuery(
	i Instance, fields []string, driver string,
) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.model.fields))
	cols := make([]string, 0, len(i.model.fields))
	placeholders := make([]string, 0, len(i.model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.model.fields {
			if name == i.model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("\"%s\"", name))
				vals = append(vals, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", index))
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("\"%s\"", name))
				vals = append(vals, val)
				placeholders = append(
					placeholders, fmt.Sprintf("$%d", index+1),
				)
			}
		}
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		i.model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	if driver == "postgres" {
		query = fmt.Sprintf("%s RETURNING \"%s\"", query, i.model.pk)
	}
	return query, vals
}

func sqlUpdateQuery(i Instance, fields []string) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.model.fields))
	cols := make([]string, 0, len(i.model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.model.fields {
			if name == i.model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("\"%s\" = $%d", name, index))
				vals = append(vals, val)
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("\"%s\" = $%d", name, index+1))
				vals = append(vals, val)
			}
		}
	}
	vals = append(vals, i.Get(i.model.pk))
	query := fmt.Sprintf(
		"UPDATE \"%s\" SET %s WHERE \"%s\" = $%d",
		i.model.Table(),
		strings.Join(cols, ", "),
		i.model.pk,
		len(cols)+1,
	)
	return query, vals
}
