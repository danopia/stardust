package main

import (
	"flag"
	"log"
	"net/http"
	"fmt"

	//"github.com/danopia/stardust/wormhole/ddp"
	//"github.com/danopia/stardust/wormhole/kernel"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/devices"
)

func main() {
	var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
  //var secret = flag.String("secret", "none", "Secret token, for id and auth")
	flag.Parse()

	//http.HandleFunc("/sockjs/info", ddp.ServeSockJsInfo)
	//http.HandleFunc("/sockjs", ddp.ServeSockJs)

	ns := base.NewNamespace("star://apt.danopia.net")
	ns.AddDevice("/mnt/consul", devices.NewConsulDevice())
	ns.Put("/mnt/consul/key", "hello world")

	ns.AddDevice("/mnt/irc/freenode", devices.NewIrcDevice())
	ns.Put("/mnt/irc/freenode/nickname", "star-router")
	ns.Put("/mnt/irc/freenode/username", "stardust")
	ns.Put("/mnt/irc/freenode/server", "chat.freenode.net:6667")
	ns.Put("/mnt/irc/freenode/channel-list", []string{ "##stardust" })

	//kernel.Start()
  log.Println("Kernel is started")



	host := fmt.Sprint("localhost:", *port)
	log.Printf("Listening on %s...", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
