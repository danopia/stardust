package entries

import (
	"fmt"
	"log"
	"strings"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/inmem"
)

// Directory containing the clone function
func getFakeConsulDriver() *inmem.Folder {
	return inmem.NewFolderOf("rom",
		inmem.NewFunction("invoke", startFakeConsul),
	).Freeze()
}

// Function that creates a new Consul client when invoked
func startFakeConsul(ctx base.Context, input base.Entry) (output base.Entry) {
	return &romRoot{}
}

// Main Entity representing a consul client
// Presents k/v tree as a child
type romRoot struct {
}

var _ base.Folder = (*romRoot)(nil)

func (e *romRoot) Name() string {
	return "rom"
}
func (e *romRoot) Children() []string {
	return []string{"kv"}
}
func (e *romRoot) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "kv":
		return &romKV{
			root: e,
		}, true

	default:
		return
	}
}
func (e *romRoot) Put(name string, entry base.Entry) (ok bool) {
	return false
}

// Directory/String tree backed by consul kv
// This struct only represents a folder
// inmem strings are used to handle keys
type romKV struct {
	root *romRoot
	path string
}

var _ base.Folder = (*romKV)(nil)

func (e *romKV) Name() string {
	if e.path == "" {
		return "kv"
	} else {
		if idx := strings.LastIndex(e.path, "/"); idx != -1 {
			return e.path[idx+1:]
		} else {
			return e.path
		}
	}
}

func (e *romKV) Children() []string {
	// Calculate consul key prefix
	prefix := ""
	if len(e.path) > 0 {
		prefix = fmt.Sprintf("%s/", e.path)
	}

	// Get list of basenames matching the prefix
	children := make([]string, 0)
	for key, _ := range romData {
		if strings.HasPrefix(key, prefix) {
			baseName := strings.TrimRight(key[len(prefix):], "/")
			if len(baseName) > 0 && !strings.Contains(baseName, "/") {
				children = append(children, baseName)
			}
		}
	}
	return children
}

// try as key first, then as folder, then for children presence, then fail
func (e *romKV) Fetch(name string) (entry base.Entry, ok bool) {
	prefix := e.path
	if len(prefix) > 0 {
		prefix += "/"
	}

	// Try to get as normal string key
	if value, ok := romData[prefix+name]; ok {
		return inmem.NewString(name, value), true
	}

	// Make a folder and see if it exists
	subDir := &romKV{
		root: e.root,
		path: prefix + name,
	}

	// Try to get as special folder key
	if _, ok = romData[prefix+name+"/"]; ok {
		return subDir, true
	}

	// Check if it has any children
	if len(subDir.Children()) > 0 {
		return subDir, true
	}
	return nil, false
}

// this never works because you can't bind foreign nodes into the consul tree
// you have to get a node then set its value
func (e *romKV) Put(name string, entry base.Entry) (ok bool) {
	prefix := e.path
	if len(prefix) > 0 {
		prefix += "/"
	}

	log.Println("rom: dropping put to %s", prefix+name)
	return false
}
