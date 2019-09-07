package gomodel

// The Conditioner interface defines the methods necessary to construct SQL
// predicates from the data types implementing the interface, and combine them
// to create composite predicates.
type Conditioner interface {
	// Root returns the first part of a composite predicate, that can be another
	// composite one if isChain is true or a simple one otherwise.
	//
	// For example, let's take the Conditioner representing this predicate:
	//  (username = 'alice' OR username = 'bob') AND isAdmin = FALSE
	//
	// The method should return isChain as true and a Conditioner representing:
	//  username = 'alice' OR username = 'bob'
	//
	// Whose Root method should return isChain as false and the Conditioner:
	//  username = 'alice'
	Root() (conditioner Conditioner, isChain bool)
	// Conditions returns a map of values representing a predicate consisting
	// of simple conditions joined by the AND operator, where the key is the
	// column and operator part of the condition.
	//
	// The = operator can be omitted. For example:
	//
	//   active = True AND id >= 10
	//
	// Would return: map[string]Value{"active": true, "id >=": 10}
	Conditions() map[string]Value
	// Next returns the next conditioner and the operator joining them:
	//  AND: isOr is false, isNot is false
	//  AND NOT: isOr is false, isNot is true
	//  OR: isOr is true, isNot is false
	//  OR NOT: isOr is true, isNot is true
	Next() (conditioner Conditioner, isOr bool, isNot bool)
	// And joins the given conditioner and the current one by the AND operator.
	And(conditioner Conditioner) Conditioner
	// AndNot joins the given conditioner and the current one by the AND NOT
	// operator.
	AndNot(conditioner Conditioner) Conditioner
	// Or joins the given conditioner and the current one by the OR operator.
	Or(conditioner Conditioner) Conditioner
	// OrNot joins the given conditioner and the current one by the OR NOT
	// operator.
	OrNot(conditioner Conditioner) Conditioner
}

// condChain implements the Conditioner interface for composite predicates.
type condChain struct {
	root Conditioner
	next Conditioner
	or   bool
	not  bool
}

// Conditions implements the Conditioner interface.
func (c condChain) Conditions() map[string]Value {
	if conditions, ok := c.root.(Q); ok {
		return conditions
	}
	return nil
}

// Root implements the Conditioner interface.
func (c condChain) Root() (Conditioner, bool) {
	_, ok := c.root.(Q)
	return c.root, !ok
}

// Next implements the Conditioner interface.
func (c condChain) Next() (Conditioner, bool, bool) {
	return c.next, c.or, c.not
}

// And implements the Conditioner interface.
func (c condChain) And(next Conditioner) Conditioner {
	return condChain{root: c, next: next}
}

// AndNot implements the Conditioner interface.
func (c condChain) AndNot(next Conditioner) Conditioner {
	return condChain{root: c, next: next, not: true}
}

// Or implements the Conditioner interface.
func (c condChain) Or(next Conditioner) Conditioner {
	return condChain{root: c, next: next, or: true}
}

// OrNot implements the Conditioner interface.
func (c condChain) OrNot(next Conditioner) Conditioner {
	return condChain{root: c, next: next, or: true, not: true}
}

// Q is a map of values that implements the Conditioner interface for predicates
// consisting of simple conditions joined by the AND operator.
type Q map[string]Value

// Conditions implements Conditioner interface.
func (q Q) Conditions() map[string]Value {
	return q
}

// Root implements Conditioner interface.
func (q Q) Root() (Conditioner, bool) {
	return q, false
}

// Next implements Conditioner interface.
func (q Q) Next() (Conditioner, bool, bool) {
	return nil, false, false
}

// And implements Conditioner interface.
func (q Q) And(next Conditioner) Conditioner {
	return condChain{root: q, next: next}
}

// AndNot implements Conditioner interface.
func (q Q) AndNot(next Conditioner) Conditioner {
	return condChain{root: q, next: next, not: true}
}

// Or implements Conditioner interface.
func (q Q) Or(next Conditioner) Conditioner {
	return condChain{root: q, next: next, or: true}
}

// OrNot implements Conditioner interface.
func (q Q) OrNot(next Conditioner) Conditioner {
	return condChain{root: q, next: next, or: true, not: true}
}
