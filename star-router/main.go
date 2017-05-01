package main

import (
	"flag"
	"log"
	"net/http"
	//"fmt"

	"github.com/danopia/stardust/wormhole/ddp"
	//"github.com/danopia/stardust/wormhole/kernel"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/entries"
)

func main() {
	//var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
  //var secret = flag.String("secret", "none", "Secret token, for id and auth")
	flag.Parse()

	http.HandleFunc("/sockjs/info", ddp.ServeSockJsInfo)
	http.HandleFunc("/sockjs", ddp.ServeSockJs)

	ns := base.NewNamespace("star://apt.danopia.net", entries.NewRootEntry())
	base.RootSpace = ns
	handle := ns.NewHandle()

	// Get the Ray executable from /rom/bin
	handle.Walk("/rom/bin/ray")
	ray, ok := handle.GetFunction()
	if !ok {
		panic("Ray executable not found. That shouldn't happen.")
	}

	// Get the Consul driver in /rom/drv
	handle.Walk("/rom/drv/consul/clone")
	consulClone, ok := handle.GetFunction()
	if !ok {
		panic("Consul Driver not found. That shouldn't happen.")
	}

	// Get /boot for bootstrapping
	handle.Walk("/boot")
	boot, ok := handle.GetFolder()
	if !ok {
		panic("/boot not found. That shouldn't happen.")
	}

	// Mount the Consul driver at /n/consul
	handle.Walk("/n")
	names, ok := handle.GetFolder()
	if !ok {
		panic("/n not found. That shouldn't happen.")
	}
	names.Put("consul", consulClone.Invoke(nil))

  // Bind consul keyval tree to /boot/cfg
	handle.Walk("consul/kv")
	boot.Put("cfg", handle.Get())

	// Get a reference to the config
	handle.Walk("/boot/cfg")
	config, ok := handle.GetFolder()
	if !ok {
		panic("/boot/cfg wasn't a Directory, make sure Consul driver is working")
	}

	// Start using the config
	log.Println("boot keys:", boot.Children())
	log.Println("config keys:", config.Children())

	// Run the init script
	handle.Walk("/boot/init")
	init, ok := handle.GetString()
	if !ok {
		panic("/boot/init wasn't a String, provide an init script and try again")
	}

	ray.Invoke(init)

	/*
	ns.AddDevice("/rom/drv/irc", devices.NewDriverDevice(devices.NewIrcDevice))
	ns.AddDevice("/rom/drv/consul", devices.NewDriverDevice(devices.NewConsulDevice))

	ns.Copy("/rom/drv/consul/clone", "/cfg")

	ns.Copy("/rom/drv/irc/clone", "/n/irc")
	ns.Set("/n/irc/nickname", "star-router")
	ns.Set("/n/irc/username", "stardust")
	ns.Set("/n/irc/server", "irc.stardustapp.run:6667")
	ns.Set("/n/irc/channel-list", []string{ "#general" })

	ns.Copy("/rom/drv/irc/clone", "/n/freenode")
	//ns.Set("/n/freenode/nickname", "star-router")
	ns.Set("/n/freenode/username", "stardust")
  ns.Set("/n/freenode/server", "chat.freenode.net:6667")
	ns.Set("/n/freenode/channel-list", []string{ "##stardust" })
	*/

	//kernel.Start()
  log.Println("Kernel booted successfully")

	//ns.Set("/n/irc/channels/#general/privmsg", "hello")

	/*
	host := fmt.Sprint("localhost:", *port)
	log.Printf("Listening on %s...", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	*/

}
