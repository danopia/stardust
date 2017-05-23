package base

import (
	"log"
	"path"
	//"strings"
)

//type ContextSystem struct {
//	rootCtx *context
//}

// An entity representing a running component of the system
// Maintains a namespace and provides useful helpers
type Context interface {
	Name() string
	//NewHandle() Handle
	Fork(name string) Context // TODO: bind map

	Put(path string, entry Entry) (ok bool)

	// All of these except GetLink() follow Links
	Get(path string) (entry Entry, ok bool)
	GetFolder(path string) (entry Folder, ok bool)
	GetFunction(path string) (entry Function, ok bool)
	GetShape(path string) (entry Shape, ok bool)
	GetList(path string) (entry List, ok bool)
	GetString(path string) (entry String, ok bool)
	GetLink(path string) (entry Link, ok bool)
	GetFile(path string) (entry File, ok bool)
	GetQueue(path string) (entry Queue, ok bool)
	GetLog(path string) (entry Log, ok bool)
}

// Only impl of Context
type context struct {
	name   string
	handle Handle
	parent Context
}

func NewRootContext(ns *Namespace) Context {
	return &context{
		name:   "root",
		handle: newRootHandle(ns),
	}
}

func (c *context) Name() string {
	return c.name
}

func (c *context) Fork(name string) Context {
	return &context{
		name:   name,
		handle: c.handle.Clone(),
		parent: c,
	}
}

func (c *context) Put(fullPath string, entry Entry) (ok bool) {
	parent, ok := c.GetFolder(path.Dir(fullPath))
	if !ok {
		log.Println("Parent path of", fullPath, "could not be selected")
		return
	}

	return parent.Put(path.Base(fullPath), entry)
}

// Cast helpers
func (c *context) Get(path string) (entry Entry, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	return h.Get(), true
}
func (c *context) GetFolder(path string) (entry Folder, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetFolder()
	return
}
func (c *context) GetFunction(path string) (entry Function, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetFunction()
	return
}
func (c *context) GetShape(path string) (entry Shape, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetShape()
	return
}
func (c *context) GetList(path string) (entry List, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetList()
	return
}
func (c *context) GetString(path string) (entry String, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetString()
	return
}
func (c *context) GetLink(path string) (entry Link, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetLink()
	return
}
func (c *context) GetFile(path string) (entry File, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetFile()
	return
}
func (c *context) GetQueue(path string) (entry Queue, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetQueue()
	return
}
func (c *context) GetLog(path string) (entry Log, ok bool) {
	h := c.handle.Clone()
	if !h.Walk(path) {
		return nil, false
	}

	entry, ok = h.GetLog()
	return
}
