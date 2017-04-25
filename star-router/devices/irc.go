package devices

import (
  "github.com/thoj/go-ircevent"
)

// Manufactors virtual devices which can exist in more than one place at once
type IrcDevice struct {
  irc *irc.Connection

  server string
  nickname string
  username string
  channelList []string

  configured bool
  connected bool
}

func NewIrcDevice() *IrcDevice {
  return &IrcDevice{}
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

  }
  panic("404 getting irc " + path)
}

func (d *IrcDevice) Put(path string, data interface{}) {
  switch path {

  case "nickname":
    d.nickname = data.(string)
    d.configure()

  case "username":
    d.username = data.(string)
    d.configure()

  case "server":
    d.server = data.(string)
    d.configure()

  case "channel-list":
    d.channelList = data.([]string)

  default:
    panic("404 putting irc " + path)
  }
}

func (d *IrcDevice) configure() {
  if d.configured {
    return
  }
  if d.nickname == "" || d.username == "" || d.server == "" {
    return
  }
  d.configured = true


  d.irc = irc.IRC(d.nickname, d.username)
  d.irc.VerboseCallbackHandler = true
  d.irc.Debug = true
  // d.irc.UseTLS = true
  // d.irc.TLSConfig = &tls.Config{InsecureSkipVerify: true}

  d.irc.AddCallback("001", func(e *irc.Event) {
    d.connected = true

    for _, channel := range d.Get("channel-list").([]string) {
      d.irc.Join(channel)
    }
  })
  //d.irc.AddCallback("366", func(e *irc.Event) {  })

  err := d.irc.Connect(d.server)
  if err != nil {
    panic(err)
  }

  go d.irc.Loop()
}
