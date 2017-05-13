package entries

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
	"github.com/heatxsink/go-hue/configuration"
	"github.com/heatxsink/go-hue/portal"
)

// Directory containing the clone function
func getHueDriver() *inmem.Folder {
	return inmem.NewFolderOf("hue",
		inmem.NewFunction("discover", discoverHue),
		inmem.NewFunction("clone", startHue),
	).Freeze()
}

func discoverHue(input base.Entry) (output base.Entry) {
	pp, err := portal.GetPortal()
	if err != nil {
		log.Println("Hue Bridge Discovery error:", err)
		return nil
	}

	folder := inmem.NewFolder("discovered-bridges")
	for _, b := range pp {
		entry := inmem.NewFolderOf(b.ID,
			inmem.NewString("lan-ip-address", b.InternalIPAddress),
			inmem.NewString("mac-address", b.MacAddress),
		)
		entry.Put("pair", inmem.NewFunction("pair", curryPair(entry)))
		folder.Put(b.ID, entry.Freeze())
	}

	return folder.Freeze()
}

func curryPair(bridge base.Entry) func(input base.Entry) (output base.Entry) {
	return func(input base.Entry) (output base.Entry) {
		return pairHue(bridge, input)
	}
}

func pairHue(bridge, input base.Entry) (output base.Entry) {
	bridgeFolder := bridge.(base.Folder)
	ipAddress, _ := helpers.GetChildString(bridgeFolder, "lan-ip-address")
	macAddress, _ := helpers.GetChildString(bridgeFolder, "mac-address")

	var appName, deviceType string
	if input != nil {
		inputFolder := input.(base.Folder)
		appName, _ = helpers.GetChildString(inputFolder, "app-name")
		deviceType, _ = helpers.GetChildString(inputFolder, "device-type")
	}
	if appName == "" {
		appName = "apt.danopia.net"
	}
	if deviceType == "" {
		deviceType = "stardust hue"
	}

	c := configuration.New(ipAddress)
	response, err := c.CreateUser(appName, deviceType)
	if err != nil {
		log.Println("Error: ", err)
		return nil
	}
	secret := response[0].Success["username"].(string)

	return inmem.NewFolderOf(bridge.Name(),
		inmem.NewString("lan-ip-address", ipAddress),
		inmem.NewString("mac-address", macAddress),
		inmem.NewString("app-name", appName),
		inmem.NewString("device-type", deviceType),
		inmem.NewString("username", secret),
	).Freeze()
}

// Function that creates a new Hue client when invoked
func startHue(input base.Entry) (output base.Entry) {
	/*client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	return &consulRoot{client}*/
	return nil
}

/*
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
func (e *consulKV) Get() (value string, ok bool) {
	if e.path == "" {
		return "", false
	}

	pair, _, err := e.kv.Get(e.path, nil)
	if err != nil {
		panic(err)
	}

	if pair == nil {
		return "", false
	}
	return string(pair.Value), true // []byte
}

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
