package gomodels

import (
	"fmt"
	"strings"
)

type Conditioner interface {
	Predicate(driver string, placeholder int) (pred string, vals []interface{})
	And(q Conditioner) Conditioner
	AndNot(q Conditioner) Conditioner
	Or(q Conditioner) Conditioner
	OrNot(q Conditioner) Conditioner
}

type Filter struct {
	root Conditioner
	next Conditioner
	or   bool
	not  bool
}

func (f Filter) Predicate(driver string, pHolder int) (string, []interface{}) {
	pred := ""
	rootPred, values := f.root.Predicate(driver, pHolder)
	if f.next != nil {
		operator := "AND"
		if f.or {
			operator = "OR"
		}
		if f.not {
			operator += " NOT"
		}
		nextPred, nextValues := f.next.Predicate(driver, pHolder+len(values))
		pred = fmt.Sprintf("(%s) %s (%s)", rootPred, operator, nextPred)
		values = append(values, nextValues...)
	} else {
		pred = rootPred
	}
	return pred, values
}

func (f Filter) And(next Conditioner) Conditioner {
	f.next = next
	return f
}

func (f Filter) AndNot(next Conditioner) Conditioner {
	f.next = next
	f.not = true
	return f
}

func (f Filter) Or(next Conditioner) Conditioner {
	f.next = next
	f.or = true
	return f
}

func (f Filter) OrNot(next Conditioner) Conditioner {
	f.next = next
	f.or = true
	f.not = true
	return f
}

type Q map[string]Value

func (q Q) Predicate(driver string, pHolder int) (string, []interface{}) {
	conditions := make([]string, 0, len(q))
	values := make([]interface{}, 0, len(q))
	for column, value := range q {
		values = append(values, value)
		if driver == "postgres" {
			conditions = append(
				conditions, fmt.Sprintf("\"%s\" = $%d", column, pHolder),
			)
			pHolder += 1
		} else {
			conditions = append(conditions, fmt.Sprintf("\"%s\" = ?", column))
		}
	}
	pred := fmt.Sprintf("%s", strings.Join(conditions, " AND "))
	return pred, values
}

func (q Q) And(next Conditioner) Conditioner {
	return Filter{root: q, next: next}
}

func (q Q) AndNot(next Conditioner) Conditioner {
	return Filter{root: q, next: next, not: true}
}

func (q Q) Or(next Conditioner) Conditioner {
	return Filter{root: q, next: next, or: true}
}

func (q Q) OrNot(next Conditioner) Conditioner {
	return Filter{root: q, next: next, or: true, not: true}
}
