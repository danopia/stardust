package kernel

import (
  "log"
  "github.com/danopia/stardust/wormhole/ddp"
  "github.com/danopia/stardust/wormhole/kernel/base"
  "github.com/danopia/stardust/wormhole/kernel/channels"
)

func Start() {
  namespaces := make(map[string]*base.Namespace)

  ddp.Methods["boot"] = func(c *ddp.Client, args ...interface{}) interface{} {
    log.Println("Client is booting")

    // TODO: check if already booted
    root := channels.NewVolatileFolder(1)

    fd := root.Create("motd", 0666, 0x10 | 0x1) // truncate & write
    fd.Write(0, []byte("Hello World!"))
    fd.Clunk()

    namespaces[c.Session] = &base.Namespace{
      Root: root,
    }

    return 5
  }

}
