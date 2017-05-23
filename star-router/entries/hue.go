package entries

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"

	"github.com/heatxsink/go-hue/configuration"
	"github.com/heatxsink/go-hue/lights"
	"github.com/heatxsink/go-hue/portal"
)

// Directory containing the clone function
func getHueDriver() *inmem.Folder {
	return inmem.NewFolderOf("hue",
		inmem.NewFunction("discover", discoverHue),
		inmem.NewFolderOf("init-bridge",
			inmem.NewFunction("invoke", startHue),
			inmem.NewLink("input-shape", "/rom/shapes/hue-bridge-config"),
		).Freeze(),
	).Freeze()
}

func discoverHue(ctx base.Context, input base.Entry) (output base.Entry) {
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

func curryPair(bridge base.Entry) func(ctx base.Context, input base.Entry) (output base.Entry) {
	return func(ctx base.Context, input base.Entry) (output base.Entry) {
		return pairHue(ctx, bridge, input)
	}
}

func pairHue(ctx base.Context, bridge, input base.Entry) (output base.Entry) {
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
func startHue(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	ipAddress, _ := helpers.GetChildString(inputFolder, "lan-ip-address")
	secret, _ := helpers.GetChildString(inputFolder, "username")

	return inmem.NewFolderOf(input.Name(),
		&hueLightDir{lights.New(ipAddress, secret)},
	).Freeze()
}

// Name representing the set of lights from a Hue bridge
type hueLightDir struct {
	*lights.Lights
}

var _ base.Folder = (*hueLightDir)(nil)

func (e *hueLightDir) Name() string {
	return "lights"
}

func (e *hueLightDir) Children() []string {
	lights, err := e.GetAllLights()
	if err != nil {
		log.Println("lights.GetAllLights() ERROR:", err)
		return nil
	}

	names := make([]string, len(lights))
	for idx, light := range lights {
		names[idx] = light.Name
	}
	return names
}

func (e *hueLightDir) Fetch(name string) (entry base.Entry, ok bool) {
	return
}
func (e *hueLightDir) Put(name string, entry base.Entry) (ok bool) {
	return false
}

/*
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
