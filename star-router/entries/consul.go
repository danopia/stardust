package entries

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/inmem"
)

// Directory containing the clone function
func getConsulDriver() *inmem.Folder {
	return inmem.NewFolderOf("consul",
		inmem.NewFunction("invoke", startConsul),
	).Freeze()
}

// Function that creates a new Consul client when invoked
func startConsul(ctx base.Context, input base.Entry) (output base.Entry) {
	var config *api.Config

	if input != nil {
		inputStr := input.(base.String)
		uri, err := url.Parse(inputStr.Get())
		if err != nil {
			log.Println("URI parse error:", err)
			panic("Couldn't parse Consul endpoint")
		}

		port := uri.Port()
		if port == "" {
			if uri.Scheme == "https" {
				port = "443"
			} else {
				port = "8500"
			}
		}

		config = &api.Config{
			Address: fmt.Sprintf("%s:%s", uri.Hostname(), port),
			Scheme:  uri.Scheme,
		}

	} else {
		config = api.DefaultConfig()
	}

	log.Println("Connecting to", config)
	client, err := api.NewClient(config)
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
// This struct only represents a folder
// inmem strings are used to handle keys
type consulKV struct {
	root *consulRoot
	kv   *api.KV
	path string
}

var _ base.Folder = (*consulKV)(nil)

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
	// Calculate consul key prefix
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

// try as key first, then as folder, then for children presence, then fail
func (e *consulKV) Fetch(name string) (entry base.Entry, ok bool) {
	prefix := e.path
	if len(prefix) > 0 {
		prefix += "/"
	}

	// Try to get as normal string key
	pair, _, err := e.kv.Get(prefix+name, nil)
	if err != nil {
		return nil, false
	}
	if pair != nil {
		return inmem.NewString(name, string(pair.Value)), true
	}

	// Make a folder and see if it exists
	subDir := &consulKV{
		root: e.root,
		kv:   e.kv,
		path: prefix + name,
	}

	// Try to get as special folder key
	pair, _, err = e.kv.Get(prefix+name+"/", nil)
	if err != nil {
		return nil, false
	}
	if pair != nil {
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
func (e *consulKV) Put(name string, entry base.Entry) (ok bool) {
	prefix := e.path
	if len(prefix) > 0 {
		prefix += "/"
	}

	switch entry := entry.(type) {

	case base.String:
		// TODO: make sure not already a folder?
		_, err := e.kv.Put(&api.KVPair{
			Key:   prefix + name,
			Value: []byte(entry.Get()),
		}, nil)
		if err != nil {
			panic(err)
		}
		ok = true

	case base.Folder:
		// explicitly create the folder entry
		_, err := e.kv.Put(&api.KVPair{
			Key: prefix + name + "/",
		}, nil)
		if err != nil {
			panic(err)
		}

		// if the folder has contents, let's recursively store
		// TODO: clear existing children first??
		names := entry.Children()
		if len(names) > 0 {
			newEntry, dirOk := e.Fetch(name)
			newDir := newEntry.(base.Folder)
			if dirOk {
				ok = true
				for _, subName := range names {
					value, getOk := entry.Fetch(subName)
					if getOk {
						newDir.Put(subName, value)
					} else {
						ok = false
					}
				}
			}
		} else {
			ok = true
		}

	default:
		log.Println("consul: tried to store weird thing", entry, "as", name)

	}
	return
}
