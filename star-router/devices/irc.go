package devices

import (
  "log"
  "strings"
  "github.com/thoj/go-ircevent"
  "github.com/danopia/stardust/star-router/base"
)

// Manufactors virtual devices which can exist in more than one place at once
type IrcDevice struct {
  ns *base.Namespace
  irc *irc.Connection

  server string
  nickname string
  username string
  channelList []string

  configured bool
  connected bool
}

func NewIrcDevice(ns *base.Namespace) *IrcDevice {
  return &IrcDevice{
    ns: ns,
  }
}

func (d *IrcDevice) Get(path string) (data interface{}) {
  switch path {

  case "nickname":
    return d.nickname

  case "username":
    return d.username

  case "server":
    return d.server

  case "configured":
    return d.configured

  case "connected":
    return d.connected

  case "channel-list":
    return d.channelList

  default:
    //panic("404 getting irc " + path)
    return nil
  }
}

func (d *IrcDevice) Put(path string, data interface{}) {
  parts := strings.Split(path, "/")
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
      //panic("404 putting irc " + path)
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
    //panic("404 putting irc " + path)
  }
}

func (d *IrcDevice) gotWelcome(e *irc.Event) {
  d.connected = true

  for _, channel := range d.Get("channel-list").([]string) {
    d.irc.Join(channel)
  }
}

func (d *IrcDevice) gotMessage(e *irc.Event) {
  msg := e.Message()
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
  case "put":
    if len(parts) != 3 {
      return
    }
    d.ns.Put(parts[1], parts[2])
    d.irc.Privmsg(e.Arguments[0], "=> ok")

  case "get":
    val := d.ns.Get(parts[1])
    d.irc.Privmsgf(e.Arguments[0], "=> %v", val)
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
  d.irc.VerboseCallbackHandler = true
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
