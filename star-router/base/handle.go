package base

import (
	"log"
	"strings"
)

// A stateful handle into the name space.
// Remembers how it got where it is
type Handle interface {
	Stack() []Entry
	Get() Entry
	Path() string

	// All of these except GetLink() follow Links
	GetFolder() (entry Folder, ok bool)
	GetFunction() (entry Function, ok bool)
	GetShape() (entry Shape, ok bool)
	GetList() (entry List, ok bool)
	GetString() (entry String, ok bool)
	GetLink() (entry Link, ok bool)
	GetFile() (entry File, ok bool)
	GetQueue() (entry Queue, ok bool)
	GetLog() (entry Log, ok bool)

	Clone() Handle
	SelectName(name string) (ok bool)
	Walk(path string) (ok bool)
}

// Only impl of Handle
type handle struct {
	stack []Entry
	names []string
}

func newRootHandle(ns *Namespace) Handle {
	return &handle{
		stack: []Entry{ns.Root},
		names: []string{ns.BaseUri},
	}
}

func NewDetachedHandle(root Entry) Handle {
	return &handle{
		stack: []Entry{root},
		names: []string{""},
	}
}

func (h *handle) Clone() Handle {
	newH := &handle{
		stack: make([]Entry, len(h.stack)),
		names: make([]string, len(h.names)),
	}
	copy(newH.stack, h.stack)
	copy(newH.names, h.names)
	return newH
}

func (h *handle) Stack() (stack []Entry) {
	stack = make([]Entry, len(h.stack))
	copy(stack, h.stack)
	return
}

func (h *handle) Get() Entry {
	entry := h.stack[len(h.stack)-1]
	for {
		link, ok := entry.(Link)
		if !ok {
			break
		}

		ok = h.Walk(link.Target())
		if !ok {
			log.Println("FAILED to walk Link", entry.Name, "to", link.Target())
			return nil
		}

		entry = h.Get()
	}
	return entry
}

// Cast helpers
func (h *handle) GetFolder() (entry Folder, ok bool) {
	entry, ok = h.Get().(Folder)
	return
}
func (h *handle) GetFunction() (entry Function, ok bool) {
	entry, ok = h.Get().(Function)
	return
}
func (h *handle) GetShape() (entry Shape, ok bool) {
	entry, ok = h.Get().(Shape)
	return
}
func (h *handle) GetList() (entry List, ok bool) {
	entry, ok = h.Get().(List)
	return
}
func (h *handle) GetString() (entry String, ok bool) {
	entry, ok = h.Get().(String)
	return
}

// Specialcased: Doesn't follow Links (duh)
func (h *handle) GetLink() (entry Link, ok bool) {
	rawE := h.stack[len(h.stack)-1]
	entry, ok = rawE.(Link)
	return
}
func (h *handle) GetFile() (entry File, ok bool) {
	entry, ok = h.Get().(File)
	return
}
func (h *handle) GetQueue() (entry Queue, ok bool) {
	entry, ok = h.Get().(Queue)
	return
}
func (h *handle) GetLog() (entry Log, ok bool) {
	entry, ok = h.Get().(Log)
	return
}

func (h *handle) Path() string {
	return strings.Join(h.names, "/")
}

func (h *handle) SelectName(name string) (ok bool) {
	//log.Println("Selecting name", name, "from within", h.Path())
	switch name {

	case ".":
		ok = true

	case "..":
		if len(h.stack) > 1 {
			h.stack = h.stack[:len(h.stack)-1]
			h.names = h.names[:len(h.names)-1]
		}
		ok = true

	case "/":
		h.stack = []Entry{h.stack[0]}
		h.names = []string{h.names[0]}
		ok = true

	default:
		if dir, getOk := h.GetFolder(); getOk {
			if child, fetchOk := dir.Fetch(name); fetchOk {
				h.stack = append(h.stack, child)
				h.names = append(h.names, name)
				ok = true
			}
		} else {
			log.Println("Cannot select into a non-Folder entry")
		}
	}

	if !ok {
		log.Println("Failed to select name", name)
	}
	return
}

// Select a sequence of names (e.g. a path)
func (h *handle) Walk(path string) (ok bool) {
	if strings.HasPrefix(path, "/") {
		// Special case for absolute walks
		// Select / then walk back down
		h.SelectName("/")

		if len(path) == 1 {
			// even specialer case: Walk("/")
			return true
		} else {
			return h.Walk(path[1:])
		}
	}

	for _, name := range strings.Split(path, "/") {
		ok = h.SelectName(name)
		if !ok {
			return
		}
	}

	return true
}
