package entries

import (
	"fmt"
	"strings"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/inmem"
	"github.com/hashicorp/consul/api"
)

// Directory containing the clone function
func getConsulDriver() *inmem.Folder {
	return inmem.NewFolderOf("consul",
		inmem.NewFunction("clone", startConsul),
	).Freeze()
}

// Function that creates a new Consul client when invoked
func startConsul(input base.Entry) (output base.Entry) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	return &consulRoot{client}
}

// Main Entity representing a consul client
// Presents k/v tree as a child
type consulRoot struct {
	client *api.Client
}

var _ base.Folder = (*consulRoot)(nil)

func (e *consulRoot) Name() string {
	return "consul"
}
func (e *consulRoot) Children() []string {
	return []string{"kv"}
}
func (e *consulRoot) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "kv":
		return &consulKV{
			root: e,
			kv:   e.client.KV(),
		}, true

	default:
		return
	}
}
func (e *consulRoot) Put(name string, entry base.Entry) (ok bool) {
	return false
}

// Directory/String tree backed by consul kv
type consulKV struct {
	root *consulRoot
	kv   *api.KV
	path string
}

var _ base.Folder = (*consulKV)(nil)
var _ base.String = (*consulKV)(nil)

func (e *consulKV) Name() string {
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
func (e *consulKV) Children() []string {
	// Calculate consule key prefix
	prefix := ""
	if len(e.path) > 0 {
		prefix = fmt.Sprintf("%s/", e.path)
	}

	// Get list of keys matching the prefix
	keys, _, err := e.kv.Keys(prefix, "/", nil)
	if err != nil {
		panic(err)
	}

	// Filter down to basenames
	children := make([]string, 0, len(keys))
	for _, key := range keys {
		baseName := key[len(prefix):]
		if len(baseName) > 0 {
			children = append(children, strings.TrimRight(baseName, "/"))
		}
	}
	return children
}

// this always works
func (e *consulKV) Fetch(name string) (entry base.Entry, ok bool) {
	prefix := e.path
	if len(prefix) > 0 {
		prefix += "/"
	}

	return &consulKV{
		root: e.root,
		kv:   e.kv,
		path: prefix + name,
	}, true
}

// this never works because you can't bind foreign nodes into the consul tree
// you have to get a node then set its value
func (e *consulKV) Put(name string, entry base.Entry) (ok bool) {
	return false
}

// TODO: refactor this to use new immutable string design
func (e *consulKV) Get() (value string) {
	if e.path == "" {
		return ""
	}

	pair, _, err := e.kv.Get(e.path, nil)
	if err != nil {
		panic(err)
	}

	if pair == nil {
		return ""
	}
	return string(pair.Value) // []byte
}

/*
func (e *consulKV) Set(value string) (ok bool) {
	p := &api.KVPair{
		Key:   e.path,
		Value: []byte(value),
	}

	_, err := e.kv.Put(p, nil)
	if err != nil {
		panic(err)
	}
	return true
}
*/
