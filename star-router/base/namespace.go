package base

import (
//"log"
//"strings"
)

var RootSpace *Namespace

type Namespace struct {
	BaseUri string
	Root    Entry
}

func NewNamespace(baseUri string, root Entry) *Namespace {
	return &Namespace{
		BaseUri: baseUri,
		Root:    root,
	}
}

func (n *Namespace) NewHandle() Handle {
	return newRootHandle(n)
}
