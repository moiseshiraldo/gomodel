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

func sqlCreateQuery(table string, values Values) (string, []interface{}) {
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
		"INSERT INTO '%s' (%s) VALUES (%s)", table, colStr, phStr,
	)
	return query, vals
}
