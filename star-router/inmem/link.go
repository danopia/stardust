package inmem

import (
	"github.com/danopia/stardust/star-router/base"
)

// Manages an in-memory symlink with immutable destination
type Link struct {
	name   string
	target string
}

var _ base.Link = (*Link)(nil)

func NewLink(name string, target string) *Link {
	return &Link{
		name:   name,
		target: target,
	}
}

func (e *Link) Name() string {
	return e.name
}

func (e *Link) Target() (value string) {
	return e.target
}
