package gomodels

type Value interface{}

type Constructor interface {
	Get(field string) (val Value, ok bool)
	Set(field string, val Value) (ok bool)
	New() Constructor
}

type Values map[string]Value

func (vals Values) Get(field string) (Value, bool) {
	val, ok := vals[field]
	return val, ok
}

func (vals Values) Set(field string, val Value) bool {
	vals[field] = val
	return true
}

func (vals Values) New() Constructor {
	return Values{}
}

type Instance struct {
	Constructor
	Model *Model
}
