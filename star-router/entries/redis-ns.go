package entries

import (
	"log"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/extras"
	"github.com/stardustapp/core/inmem"

	"github.com/go-redis/redis"
)

// Performant Namesystem using Redis as a storage driver
// Kept seperate from the Redis API client, as this is very special-cased
// Assumes literally everything except credentials

// Directory containing the clone function
func getRedisNsDriver() *inmem.Folder {
	return inmem.NewFolderOf("redis-ns",
		inmem.NewFunction("invoke", startRedisNs),
		redisConnInfoShape,
	).Freeze()
}

var redisConnInfoShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewFolderOf("address",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewFolderOf("password",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewFolderOf("prefix",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewFolderOf("database",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
		),
	))

// Function that creates a new AWS client when invoked
func startRedisNs(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	address, _ := extras.GetChildString(inputFolder, "address")
	password, _ := extras.GetChildString(inputFolder, "password")
	prefix, _ := extras.GetChildString(inputFolder, "prefix")
	//database, _ := extras.GetChildString(inputFolder, "database")

	if address == "" {
		address = "localhost:6379"
	}
	if prefix == "" {
		prefix = "sdns:"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // use default DB
	})

	ns := &redisNs{
		svc:    client,
		prefix: prefix,
	}
	return ns.getRoot()
}

// Main Entity representing a set of entries in a Redis server
type redisNs struct {
	svc    *redis.Client
	prefix string
}

func (ns *redisNs) getRoot() base.Entry {
	rootNid := ns.svc.Get(ns.prefix + "root").Val()
	if rootNid == "" {
		log.Println("Initializing redisns root")
		rootNid = ns.newNode("root", "Folder")
		ns.svc.Set(ns.prefix+"root", rootNid, 0)
	}
	return ns.getEntry(rootNid)
}

func (ns *redisNs) newNode(name, typeStr string) string {
	nid := extras.GenerateId()
	if ns.svc.SetNX(ns.prefixFor(nid, "type"), typeStr, 0).Val() == false {
		panic("Redis node " + nid + " already exists, can't make new " + name + " " + typeStr)
	}
	ns.svc.Set(ns.prefixFor(nid, "name"), name, 0)
	log.Println("Created redisns node", nid, "named", name, "type", typeStr)
	return nid
}

func (ns *redisNs) prefixFor(nid, key string) string {
	return ns.prefix + "nodes/" + nid + ":" + key
}

func (ns *redisNs) nameOf(nid string) string {
	return ns.svc.Get(ns.prefixFor(nid, "name")).Val()
}
func (ns *redisNs) typeOf(nid string) string {
	return ns.svc.Get(ns.prefixFor(nid, "type")).Val()
}

func (ns *redisNs) getEntry(nid string) base.Entry {
	name := ns.nameOf(nid)
	prefix := ns.prefixFor(nid, "")
	switch ns.typeOf(nid) {

	case "String":
		value := ns.svc.Get(prefix + "value").Val()
		return inmem.NewString(name, value)

	case "Link":
		value := ns.svc.Get(prefix + "target").Val()
		return inmem.NewLink(name, value)

	case "File":
		// TODO: writable file struct!
		data, _ := ns.svc.Get(prefix + "raw-data").Bytes()
		return inmem.NewFile(name, data)

	case "Folder":
		return &redisNsFolder{
			ns:     ns,
			nid:    nid,
			prefix: prefix,
		}

	default:
		log.Println("redisns key", nid, name, "has unknown type", ns.typeOf(nid))
		return nil
	}
}

// Persists as a Folder from an redisNs instance
// Presents as a dynamic name tree
type redisNsFolder struct {
	ns     *redisNs
	prefix string
	nid    string
}

var _ base.Folder = (*redisNsFolder)(nil)

func (e *redisNsFolder) Name() string {
	return e.ns.nameOf(e.nid)
}

func (e *redisNsFolder) Children() []string {
	names, _ := e.ns.svc.HKeys(e.prefix + "children").Result()
	return names
}

func (e *redisNsFolder) Fetch(name string) (entry base.Entry, ok bool) {
	nid := e.ns.svc.HGet(e.prefix+"children", name).Val()
	if nid == "" {
		return nil, false
	}

	entry = e.ns.getEntry(nid)
	ok = entry != nil
	return
}

// replaces whatever node reference was already there w/ a new node
func (e *redisNsFolder) Put(name string, entry base.Entry) (ok bool) {
	if entry == nil {
		// unlink a child, leaves it around tho
		// TODO: garbage collection!
		e.ns.svc.HDel(e.prefix+"children", name)
		return true
	}

	var nid string
	switch entry := entry.(type) {

	case *redisNsFolder:
		// the folder already exists in redis, make a reference
		nid = entry.nid

	case base.Folder:
		nid = e.ns.newNode(entry.Name(), "Folder")
		dest := e.ns.getEntry(nid).(base.Folder)

		// recursively copy entire folder to redis
		for _, child := range entry.Children() {
			if childEnt, ok := entry.Fetch(child); ok {
				dest.Put(child, childEnt)
			} else {
				log.Println("redisns: Failed to get child", child, "of", name)
			}
		}

	case base.String:
		nid = e.ns.newNode(entry.Name(), "String")
		e.ns.svc.Set(e.ns.prefixFor(nid, "value"), entry.Get(), 0)

	case base.Link:
		nid = e.ns.newNode(entry.Name(), "Link")
		e.ns.svc.Set(e.ns.prefixFor(nid, "target"), entry.Target(), 0)

	case base.File:
		nid = e.ns.newNode(entry.Name(), "File")

		size := entry.GetSize()
		data := entry.Read(0, int(size))
		e.ns.svc.Set(e.ns.prefixFor(nid, "raw-data"), data, 0)

	}

	if nid == "" {
		log.Println("redisns put failed for", name, "on node", e.nid)
		return false
	} else {
		e.ns.svc.HSet(e.prefix+"children", name, nid)
		return true
	}
}
