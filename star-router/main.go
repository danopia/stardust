package main

import (
	//"flag"
	//"fmt"
	"log"
	//"net/http"

	//"github.com/danopia/stardust/wormhole/ddp"
	//"github.com/danopia/stardust/wormhole/kernel"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/entries"
)

func main() {
	//var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
	//var secret = flag.String("secret", "none", "Secret token, for id and auth")
	//flag.Parse()

	//http.HandleFunc("/sockjs/info", ddp.ServeSockJsInfo)
	//http.HandleFunc("/sockjs", ddp.ServeSockJs)

	root := entries.NewRootEntry()
	ns := base.NewNamespace("star://apt.danopia.net", root)
	ctx := base.NewRootContext(ns)

	//base.RootSpace = ns

	// Get the Init executable from /rom/bin
	init, ok := ctx.GetFunction("/rom/bin/init/invoke") // TODO
	if !ok {
		panic("Init executable not found. That shouldn't happen.")
	}

	// Get the Consul driver in /rom/drv
	consulClone, ok := ctx.GetFunction("/rom/drv/consul/invoke") // TODO
	if !ok {
		panic("Consul Driver not found. That shouldn't happen.")
	}

	// Mount the Consul driver at /n/consul
	ctx.Put("/n/consul", consulClone.Invoke(ctx, nil))

	// Bind consul keyval tree to /boot/cfg
	kv, ok := ctx.GetFolder("/n/consul/kv")
	if !ok {
		panic("Consul KV not found. That shouldn't happen.")
	}
	ctx.Put("/boot/cfg", kv)

	// Get the init config
	services, ok := ctx.GetFolder("/boot/cfg/services")
	if !ok {
		panic("/boot/cfg/services wasn't a Folder, provide services and try again")
	}

	// Run init
	log.Println("Bootstrapped kernel. Handing control to initsys")
	init.Invoke(ctx, services)
}
