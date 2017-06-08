package inmem

import (
	"sync"

	"github.com/danopia/stardust/star-router/base"
)

// Manages an in-memory Directory structure
// Directories are mutable by default
// Freeze() burns the writable fuse, then the contents are fixed forever
type Folder struct {
	name       string
	writable   bool
	children   map[string]base.Entry
	childMutex sync.RWMutex
}

var _ base.Folder = (*Folder)(nil)

func NewFolder(name string) *Folder {
	return &Folder{
		name:     name,
		writable: true,
		children: make(map[string]base.Entry),
	}
}

func NewFolderOf(name string, children ...base.Entry) *Folder {
	ent := &Folder{
		name:     name,
		writable: true,
		children: make(map[string]base.Entry),
	}

	for _, child := range children {
		ent.Put(child.Name(), child)
	}
	return ent
}

// Prevents the directory of names in this folder from ever changing again
// Chainable for NewFolderOf(...).Freeze()
func (e *Folder) Freeze() *Folder {
	e.writable = false
	return e
}

func (e *Folder) Name() string {
	return e.name
}

func (e *Folder) Children() []string {
	e.childMutex.RLock()
	defer e.childMutex.RUnlock()

	names := make([]string, len(e.children))
	i := 0
	for k := range e.children {
		names[i] = k
		i++
	}
	return names
}

func (e *Folder) Fetch(name string) (entry base.Entry, ok bool) {
	e.childMutex.RLock()
	defer e.childMutex.RUnlock()

	entry, ok = e.children[name]
	return
}

func (e *Folder) Put(name string, entry base.Entry) (ok bool) {
	if e.writable {
		e.childMutex.Lock()
		defer e.childMutex.Unlock()

		if entry == nil {
			delete(e.children, name)
		} else {
			e.children[name] = entry
		}

		ok = true
	}
	return
}
