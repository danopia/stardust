package main

import (
	"flag"
	"log"
	"net/http"
	"fmt"

	"github.com/danopia/stardust/wormhole/ddp"
	//"github.com/danopia/stardust/wormhole/kernel"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/devices"
)

func main() {
	var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
  //var secret = flag.String("secret", "none", "Secret token, for id and auth")
	flag.Parse()

	http.HandleFunc("/sockjs/info", ddp.ServeSockJsInfo)
	http.HandleFunc("/sockjs", ddp.ServeSockJs)

	ns := base.NewNamespace("star://apt.danopia.net")
	base.RootSpace = ns

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

	//kernel.Start()
  log.Println("Kernel is started")

	//ns.Set("/n/irc/channels/#general/privmsg", "hello")

	host := fmt.Sprint("localhost:", *port)
	log.Printf("Listening on %s...", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
