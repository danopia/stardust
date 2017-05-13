package inmem

import (
	"github.com/danopia/stardust/star-router/base"
)

// Manages an in-memory Function entry
// No attempt is made to bring logic, so you must provide an implementation
type Function struct {
	name string
	impl func(input base.Entry) (output base.Entry)
}

var _ base.Function = (*Function)(nil)

func NewFunction(name string, impl func(input base.Entry) (output base.Entry)) *Function {
	return &Function{name, impl}
}

func (e *Function) Name() string {
	return e.name
}

func (e *Function) Invoke(input base.Entry) (output base.Entry) {
	return e.impl(input)
}
