package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/danopia/stardust/wormhole/client"
	"github.com/danopia/stardust/wormhole/ddp"
)

func main() {

	var host = flag.String("host", "127.0.0.1", "Wormhole server hostname")
	var port = flag.Int("port", 9234, "TCP port to connect to")
	//var secret = flag.String("secret", "none", "Secret token, for id and auth")
	flag.Parse()

	shell := ishell.New()
	shell.SetHomeHistoryPath(".wormshell_history")

	shell.AddCmd(&ishell.Cmd{
		Name: "mount",
		Help: "source, target, fs, flags, data",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 3 {
				c.Println("Usage: mount <source> <target> <filesystem-type> [mount-flags] [data]")
				c.Err(errors.New("Wrong number of args"))
				return
			}

			var flags int
			if len(c.Args) > 3 {
				flags, _ = strconv.Atoi(c.Args[3])
			}

			var data string
			if len(c.Args) > 4 {
				data = c.Args[4]
			}

			c.Println("Mounting", strings.Join(c.Args, " "))
			client.Sink <- ddp.Message{
				Type:   "method",
				Method: "mount",
				Id:     "1",
				Params: []interface{}{map[string]interface{}{
					"source":         c.Args[0],
					"target":         c.Args[1],
					"filesystemType": c.Args[2],
					"mountFlags":     flags,
					"data":           data,
				}},
			}

		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "tmpfs",
		Help: "target, flags, data",
		Func: func(c *ishell.Context) {

			if len(c.Args) < 1 {
				c.Println("Usage: tmpfs <target> [mount-flags] [data]")
				c.Err(errors.New("Wrong number of args"))
				return
			}

			var flags int
			if len(c.Args) > 1 {
				flags, _ = strconv.Atoi(c.Args[1])
			}

			var data string
			if len(c.Args) > 2 {
				data = c.Args[2]
			}

			c.Println("Mounting tmpfs to", c.Args[0])
			client.Sink <- ddp.Message{
				Type:   "method",
				Method: "mount",
				Id:     "1",
				Params: []interface{}{map[string]interface{}{
					"source":         "tmpfs",
					"target":         c.Args[0],
					"filesystemType": "tmpfs",
					"mountFlags":     flags,
					"data":           data,
				}},
			}

		},
	})

	url := fmt.Sprint("ws://", *host, ":", *port, "/sockjs")
	shell.Printf("Connecting to %s... ", url)
	client := client.Connect(url)
	shell.Println(" ready!")

	shell.Println("Welcome to the server")
	shell.ShowPrompt(true)
	shell.Start()
}
