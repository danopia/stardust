package inmem

import (
	"github.com/danopia/stardust/star-router/base"
)

// Manages an immutable in-memory UTF-8 string primitive
type String struct {
	name  string
	value string
}

var _ base.String = (*String)(nil)

func NewString(name string, value string) *String {
	return &String{
		name:  name,
		value: value,
	}
}

func (e *String) Name() string {
	return e.name
}

func (e *String) Get() (value string) {
	return e.value
}
