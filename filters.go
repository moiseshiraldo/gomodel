package gomodels

import (
	"fmt"
	"strings"
)

type Filterer interface {
	Query(values ...interface{}) (placeholder string, val []interface{})
	And(f Filterer) Filterer
}

type Filter struct {
	sibs []Filterer
}

func (f Filter) Query(values ...interface{}) (string, []interface{}) {
	query := ""
	for _, sib := range f.sibs {
		q, v := sib.Query(values...)
		if query != "" {
			query += " AND "
		}
		query += q
		values = v
	}
	return query, values
}

func (f Filter) And(sib Filterer) Filterer {
	f.sibs = append(f.sibs, sib)
	return f
}

type Q map[string]Value

func (q Q) Query(values ...interface{}) (string, []interface{}) {
	filters := make([]string, 0, len(q))
	for column, value := range q {
		values = append(values, value)
		filters = append(
			filters, fmt.Sprintf("\"%s\" = $%d", column, len(values)),
		)
	}
	query := fmt.Sprintf("(%s)", strings.Join(filters, " AND "))
	return query, values
}

func (q Q) And(sib Filterer) Filterer {
	return Filter{sibs: []Filterer{q, sib}}
}
