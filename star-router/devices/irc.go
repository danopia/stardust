package devices

import (
	"github.com/danopia/stardust/wormhole/common"
	"github.com/stardustapp/core/base"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

// Manufactors virtual devices which can exist in more than one place at once
type IrcDevice struct {
	ns  *base.Namespace
	irc *irc.Connection

	server      string
	nickname    string
	username    string
	channelList []string

	configured bool
	connected  bool

	privmsgObservers  map[string][]chan<- interface{}
	observerDispenser <-chan int64
}

func NewIrcDevice() base.Device {
	return &IrcDevice{
		ns:                base.RootSpace,
		privmsgObservers:  make(map[string][]chan<- interface{}),
		observerDispenser: common.NewInt64Dispenser().C,
	}
}

func (d *IrcDevice) Get(path string) (data interface{}) {
	switch path {

	case "/nickname":
		return d.nickname

	case "/username":
		return d.username

	case "/server":
		return d.server

	case "/configured":
		return d.configured

	case "/connected":
		return d.connected

	case "/channel-list":
		return d.channelList

	default:
		//panic("404 getting irc " + path)
		return nil
	}
}

func (d *IrcDevice) Set(path string, data interface{}) {
	parts := strings.Split(path[1:], "/")
	switch len(parts) {

	case 1:
		switch parts[0] {

		case "nickname":
			d.nickname = data.(string)
			if d.irc == nil {
				d.configure()
			} else {
				d.irc.Nick(d.nickname)
			}

		case "username":
			d.username = data.(string)
			d.configure()

		case "server":
			d.server = data.(string)
			d.configure()

		case "channel-list":
			d.channelList = data.([]string)

		default:
			//panic("404 Setting irc " + path)
		}

	case 3:
		switch parts[0] {

		case "channels", "queries":
			channel := parts[1]
			switch parts[2] {
			case "privmsg":
				d.irc.Privmsg(channel, data.(string))
			}

		}

	default:
		//panic("404 Setting irc " + path)
	}
}

func (d *IrcDevice) gotWelcome(e *irc.Event) {
	d.connected = true

	for _, channel := range d.Get("/channel-list").([]string) {
		d.irc.Join(channel)
	}
}

func (d *IrcDevice) gotMessage(e *irc.Event) {
	channel := e.Arguments[0]
	msg := e.Message()

	// Broadcast to anyone who cares
	for _, C := range d.privmsgObservers[channel] {
		C <- map[string]string{
			"nickname": e.Nick,
			"hostname": e.Host,
			"username": e.User,
			"source":   e.Source,
			"message":  msg,
		}
	}

	if !strings.HasPrefix(msg, "!") {
		return
	}
	msg = strings.TrimLeft(msg, "!")

	parts := strings.SplitN(msg, " ", 3)
	if len(parts) < 2 {
		return
	}
	cmd := strings.ToLower(parts[0])

	switch cmd {
	/*
	  case "set":
	    if len(parts) != 3 {
	      return
	    }
	    d.ns.Set(parts[1], parts[2])
	    d.irc.Privmsg(channel, "=> ok")

	  case "get":
	    val := d.ns.Get(parts[1])
	    d.irc.Privmsgf(channel, "=> %v", val)

	  case "cp":
	    if len(parts) != 3 {
	      return
	    }
	    d.ns.Copy(parts[1], parts[2])
	    d.irc.Privmsg(channel, "=> ok")

	  case "map":
	    if len(parts) != 3 {
	      return
	    }
	    val := d.ns.Map(parts[1], parts[2])
	    d.irc.Privmsgf(channel, "=> %v", val)

	  case "subscribe":
	    handle := <-d.observerDispenser
	    if stream := d.ns.Observe(parts[1]); stream != nil {
	      d.irc.Privmsgf(channel, "=> Name Handle %v subscribed to %s", handle, parts[1])

	      go func(handle int64, channel string, stream <-chan interface{}) {
	        for x := range stream {
	          d.irc.Privmsgf(channel, "%v => %v", handle, x)
	        }
	        d.irc.Privmsgf(channel, "%v stream closed", handle)
	      }(handle, channel, stream)

	    } else {
	      d.irc.Privmsgf(channel, "!! Name %s refused the subscription", parts[1])
	    }
	*/
	}

	//event.Message() contains the message
	//event.Nick Contains the sender
	//event.Arguments[0] Contains the channel
}

func (d *IrcDevice) configure() {
	if d.configured {
		return
	}
	if d.nickname == "" || d.username == "" || d.server == "" {
		return
	}
	log.Println("Configuring IRC device")
	d.configured = true

	d.irc = irc.IRC(d.nickname, d.username)
	//d.irc.VerboseCallbackHandler = true
	d.irc.Debug = true
	// d.irc.UseTLS = true
	// d.irc.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	d.irc.AddCallback("001", d.gotWelcome)
	//d.irc.AddCallback("366", func(e *irc.Event) {  })
	d.irc.AddCallback("PRIVMSG", d.gotMessage)

	err := d.irc.Connect(d.server)
	if err != nil {
		panic(err)
	}

	go d.irc.Loop()
}

func (d *IrcDevice) Map(path string, input interface{}) (output interface{}) {
	return nil
}

func (d *IrcDevice) Observe(path string) (stream <-chan interface{}) {
	parts := strings.Split(path[1:], "/")
	switch len(parts) {

	case 1:
		switch parts[0] {

		case "connected":
			// TODO
			return make(chan interface{})

		default:
			return nil
		}

	case 3:
		switch parts[0] {

		case "channels", "queries":
			C := make(chan interface{})
			channel := parts[1]
			switch parts[2] {
			case "privmsg":
				d.privmsgObservers[channel] = append(d.privmsgObservers[channel], C)
			}
			return C

		default:
			return nil
		}

	default:
		return nil
	}
}
