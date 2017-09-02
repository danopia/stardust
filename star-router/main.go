package main

import (
	//"flag"
	//"fmt"
	"log"

	//"github.com/danopia/stardust/wormhole/ddp"
	//"github.com/danopia/stardust/wormhole/kernel"
	"github.com/danopia/stardust/star-router/entries"
	"github.com/stardustapp/core/base"
	//"github.com/stardustapp/core/inmem"
)

func main() {
	//var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
	//var consulUri = flag.String("consul-uri", "http://127.0.0.1:8500", "Base URI for Consul (serves as nvram)")
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

	// Get the ROM driver in /rom/drv
	romClone, ok := ctx.GetFunction("/rom/drv/rom/invoke") // TODO
	if !ok {
		panic("ROM Driver not found. That shouldn't happen.")
	}

	// Mount the ROM driver at /n/rom
	ctx.Put("/n/rom", romClone.Invoke(ctx, nil))

	// Bind rom keyval tree to /boot/cfg
	kv, ok := ctx.GetFolder("/n/rom/kv")
	if !ok {
		panic("ROM KV not found. That shouldn't happen.")
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
