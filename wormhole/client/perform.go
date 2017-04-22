package client

import (
	"log"

	"github.com/danopia/stardust/wormhole/ddp"
)

func (c *SockJsClient) perform() {
	c.Sink <- ddp.Message{
		Type:    "connect",
		Version: "1",
		Support: []string{"1", "pre2", "pre1"},
	}

	welcome := <-c.Source
	if welcome.ServerId == "" {
		log.Println("Expected server welcome, received %+v instead", welcome)
		return
	}

	go c.processSource()
}

func (c *SockJsClient) processSource() {
	for msg := range c.Source {
		switch msg.Type {

		default:
			log.Printf("!UNKNOWN! %+v\n", msg)
		}
	}
}
