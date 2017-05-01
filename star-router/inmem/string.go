package inmem

import (
  "github.com/danopia/stardust/star-router/base"
)

// Manages an in-memory Directory structure
// Directories are mutable by default
// Freeze() burns the writable fuse, then the contents are fixed forever
type String struct {
  name string
  writable bool
  value string
}
var _ base.String = (*String)(nil)

func NewString(name string, value string) *String {
  return &String{
    name: name,
    writable: true,
    value: value,
  }
}

// Prevents this entry from ever changing again
func (e *String) Freeze() {
  e.writable = false
}

func (e *String) Name() string {
  return e.name
}

func (e *String) Get() (value string, ok bool) {
  return e.value, true
}

func (e *String) Set(value string) (ok bool) {
  if e.writable {
    e.value = value
    ok = true
  }
  return
}
