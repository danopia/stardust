package base

import (
  "log"
  "strings"
)

type Namespace struct {
	BaseUri string
	Mounts map[string]Device
}

func NewNamespace(baseUri string) *Namespace {
  return &Namespace{
		BaseUri: baseUri,
		Mounts: make(map[string]Device),
	}
}

func (n *Namespace) AddDevice(path string, device Device) {
  n.Mounts[path] = device
}

func (n *Namespace) Get(path string) (data interface{}) {
  partialLen := 0
  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      log.Println("asking", partialPath, "device for", path[partialLen+1:])
      return device.Get(path[partialLen+1:])
    }
  }

  panic("404 getting " + path)
}

func (n *Namespace) Put(path string, data interface{}) {
  partialLen := 0
  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      log.Println("giving", partialPath, "device data for", path[partialLen+1:])
      device.Put(path[partialLen+1:], data)
      return
    }
  }

  panic("404 putting " + path)
}
