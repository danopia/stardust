package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/danopia/stardust/wormhole/ddp"
	"github.com/danopia/stardust/wormhole/kernel"
)

func main() {
	var port = flag.Int("port", 9234, "TCP port that the wormhole should be available on")
	//var secret = flag.String("secret", "none", "Secret token, for id and auth")
	flag.Parse()

	//log.SetPrefix("[chronos] ")

	http.HandleFunc("/sockjs/info", ddp.ServeSockJsInfo)
	http.HandleFunc("/sockjs", ddp.ServeSockJs)

	//head.WaitForShutdown()
	//defer head.ShutdownLeaves()

	// Prepare the app
	//app.RefreshChroots()

	kernel.Start()

	host := fmt.Sprint("localhost:", *port)
	log.Printf("Listening on %s...", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
