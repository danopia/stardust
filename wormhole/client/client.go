package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/danopia/stardust/wormhole/ddp"
	"github.com/gorilla/websocket"
)

func ReadByte(r io.Reader) (b byte, err error) {
	var p [1]byte
	_, err = r.Read(p[:])
	if err == nil {
		b = p[0]
	}
	return
}

type SockJsClient struct {
	Conn   *websocket.Conn
	Source chan ddp.Message
	Sink   chan ddp.Message

	Invocations map[string]Invocation
}

type Invocation struct {
	Callback func(c *SockJsClient, result interface{})
}

func (c *SockJsClient) pumpSink() {
	var p [1]string
	for message := range c.Sink {
		pp, err := json.Marshal(message)
		if err != nil {
			log.Println("Failed to marshal message %+v for sockjs server", message)
			close(c.Sink)
		} else {
			p[0] = string(pp)
			c.Conn.WriteJSON(p)
		}
	}
}

func (c *SockJsClient) startFrame() (prefix byte, r io.Reader, err error) {
	mType, r, err := c.Conn.NextReader()
	if err != nil {
		return
	}

	if mType != 1 {
		err = errors.New(fmt.Sprintf("Weird websocket message type %v", mType))
	}

	prefix, err = ReadByte(r)
	return
}

func (c *SockJsClient) pumpSource() {
	for { // each incoming

		prefix, r, err := c.startFrame()
		if err != nil {
			log.Println("Error starting sockjs frame: ", err)
			close(c.Source)
			return
		}

		switch prefix {
		case 'a':
			var p []string
			if err := json.NewDecoder(r).Decode(&p); err != nil {
				log.Println("Error unmarshalling JSON frame: ", err)
				close(c.Source)
				return
			}

			var message ddp.Message
			if err := json.Unmarshal([]byte(p[0]), &message); err == nil {
				c.Source <- message
			} else {
				log.Println("Error unmarshalling JSON message: ", err)
				close(c.Source)
				return
			}

		case 'h':
			// log.Println("Received server heartbeat")

		default:
			log.Println("Unknown sockjs frame", prefix)
			close(c.Source)
			return
		}
	}
}

func Connect(url string) (c SockJsClient) {
	var cstDialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, _, err := cstDialer.Dial(url, nil)
	if err != nil {
		fmt.Printf("Dial: %v", err)
	}

	c = SockJsClient{
		Conn:   conn,
		Source: make(chan ddp.Message),
		Sink:   make(chan ddp.Message),

		Invocations: make(map[string]Invocation),
	}

	// Wait for an initial 'o' frame
	prefix, _, err := c.startFrame()
	if err != nil {
		log.Println("Error starting initial sockjs frame: ", err)
		return
	}
	if prefix != 111 {
		log.Println("SockJS server started with a", prefix, "prefix")
		return
	}

	go c.pumpSink()
	go c.pumpSource()

	// Send a welcome packet for auth
	// s.Sink <- common.Packet{
	// 	Cmd:     "auth",
	// 	Context: secret,
	// }

	// log.Printf("Connection established to sockjs server")
	return
}
