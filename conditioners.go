package gomodels

type Conditioner interface {
	Conditions() map[string]Value
	Root() (c Conditioner, isChain bool)
	Next() (c Conditioner, isOr bool, isNot bool)
	And(q Conditioner) Conditioner
	AndNot(q Conditioner) Conditioner
	Or(q Conditioner) Conditioner
	OrNot(q Conditioner) Conditioner
}

type condChain struct {
	root Conditioner
	next Conditioner
	or   bool
	not  bool
}

func (c condChain) Conditions() map[string]Value {
	if conditions, ok := c.root.(Q); ok {
		return conditions
	}
	return nil
}

func (c condChain) Root() (Conditioner, bool) {
	_, ok := c.root.(Q)
	return c.root, !ok
}

func (c condChain) Next() (Conditioner, bool, bool) {
	return c.next, c.or, c.not
}

func (c condChain) And(next Conditioner) Conditioner {
	return condChain{root: c, next: next}
}

func (c condChain) AndNot(next Conditioner) Conditioner {
	return condChain{root: c, next: next, not: true}
}

func (c condChain) Or(next Conditioner) Conditioner {
	return condChain{root: c, next: next, or: true}
}

func (c condChain) OrNot(next Conditioner) Conditioner {
	return condChain{root: c, next: next, or: true, not: true}
}

type Q map[string]Value

func (q Q) Conditions() map[string]Value {
	return q
}

func (q Q) Root() (Conditioner, bool) {
	return q, false
}

func (q Q) Next() (Conditioner, bool, bool) {
	return nil, false, false
}

func (q Q) And(next Conditioner) Conditioner {
	return condChain{root: q, next: next}
}

func (q Q) AndNot(next Conditioner) Conditioner {
	return condChain{root: q, next: next, not: true}
}

func (q Q) Or(next Conditioner) Conditioner {
	return condChain{root: q, next: next, or: true}
}

func (q Q) OrNot(next Conditioner) Conditioner {
	return condChain{root: q, next: next, or: true, not: true}
}
