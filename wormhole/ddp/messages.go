package ddp

import (
	"fmt"
	"log"
)

//var Chroots *Publication
//var Apps *Publication
//var Commands *Publication

var Methods map[string]func(c *Client, args ...interface{}) interface{}
var Publications map[string]*Publication

func init() {
	//Chroots = CreatePublication("chroots")
	//Apps = CreatePublication("fd-apps")
	//Commands = CreatePublication("commands")

	Methods = make(map[string]func(c *Client, args ...interface{}) interface{})
	Publications = make(map[string]*Publication)
}

func (c *Client) handleMessages() {
	for message := range c.Source {
		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(m *Message) {
	// log.Printf("<<< %+v", m)

	switch m.Type {

	case "connect":
		c.Sink <- &Message{
			Type:    "connected",
			Session: c.Session,
		}

	case "ping":
		c.Sink <- &Message{
			Type: "pong",
		}

	case "sub":
		if pub, ok := Publications[m.Name]; ok {
			pub.Subscribe(&ClientSub{
				Id:          m.Id,
				Client:      c,
				Publication: pub,
			})

		} else {
			log.Println("Client subbed nonexistant DDP publication", m.Name)

			c.Sink <- &Message{
				Type: "nosub",
				Id:   m.Id,
				Error: &ClientError{
					Code:    404,
					Reason:  fmt.Sprint("Publication '%s' not found", m.Name),
					Message: fmt.Sprint("Publication '%s' not found [404]", m.Name),
					Type:    "Meteor.Error",
				},
			}
		}

	case "unsub":
		if sub, ok := c.Subs[m.Id]; ok {
			sub.Publication.Unsubscribe(sub)
		}

	case "method":
		go c.runMethod(m)
	}
}

func (c *Client) runMethod(m *Message) {
	if handler, ok := Methods[m.Method]; ok {
		// log.Printf("Running method %s", m.Method)

		result := handler(c, m.Params...)
		c.Sink <- &Message{
			Type:   "result",
			Id:     m.Id,
			Result: result,
		}

	} else {
		log.Println("Client called nonexistant DDP method", m.Method)

		c.Sink <- &Message{
			Type: "result",
			Id:   m.Id,
			Error: &ClientError{
				Code:    404,
				Reason:  fmt.Sprint("Method '%s' not found", m.Method),
				Message: fmt.Sprint("Method '%s' not found [404]", m.Method),
				Type:    "Meteor.Error",
			},
		}
	}

	// Notify client that we're done
	c.Sink <- &Message{
		Type:    "updated",
		Methods: []string{m.Id},
	}
}
