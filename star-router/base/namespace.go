package base

import (
  "log"
  "strings"
)

var RootSpace *Namespace

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
  log.Println("mounting device at", path)
  n.Mounts[path] = device
}

func (n *Namespace) Get(path string) (data interface{}) {
  partialLen := 0

  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      var subPath string
      if partialLen == len(path) {
        subPath = "/"
      } else {
        subPath = path[partialLen:]
      }

      log.Println("asking", partialPath, "device for", subPath)
      return device.Get(subPath)
    }
  }

  //panic("404 getting " + path)
  return nil
}

func (n *Namespace) Set(path string, data interface{}) {
  if device, ok := data.(Device); ok {
    n.AddDevice(path, device)
    return
  }

  partialLen := 0
  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      log.Println("giving", partialPath, "device data for", path[partialLen:])
      device.Set(path[partialLen:], data)
      return
    }
  }

  //panic("404 setting " + path)
}

func (n *Namespace) Map(path string, input interface{}) (output interface{}) {
  partialLen := 0

  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      var subPath string
      if partialLen == len(path) {
        subPath = "/"
      } else {
        subPath = path[partialLen:]
      }

      log.Println("asking", partialPath, "device to map", subPath)
      return device.Map(subPath, input)
    }
  }

  //panic("404 getting " + path)
  return nil
}

func (n *Namespace) Observe(path string) (stream <-chan interface{}) {
  partialLen := 0

  for _, name := range strings.Split(path[1:], "/") {
    partialLen += len(name) + 1
    partialPath := path[:partialLen]

    if device, ok := n.Mounts[partialPath]; ok {
      var subPath string
      if partialLen == len(path) {
        subPath = "/"
      } else {
        subPath = path[partialLen:]
      }

      log.Println("asking", partialPath, "device to observe", subPath)
      return device.Observe(subPath)
    }
  }

  //panic("404 getting " + path)
  return nil
}

func (n *Namespace) Copy(source, target string) {
  n.Set(target, n.Get(source))
}
