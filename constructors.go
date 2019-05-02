package gomodels

type Value interface{}
type Values map[string]Value

type Constructor interface {
	Get(field string) Value
	Set(field string, val Value) error
}

type Instance struct {
	Model  *Model
	Values Values
}

func (ins Instance) Get(field string) Value {
	return ins.Values[field]
}

func (ins *Instance) Set(field string, val Value) error {
	ins.Values[field] = val
	return nil
}
