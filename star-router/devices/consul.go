package devices

import (
  "github.com/hashicorp/consul/api"
  "github.com/danopia/stardust/star-router/base"
)

// Manufactors virtual devices which can exist in more than one place at once
type ConsulDevice struct {
  client *api.Client
  kv *api.KV
  // Name string
  //Properties map[string]interface{}
}

func NewConsulDevice() base.Device {
  // Get a new client
  client, err := api.NewClient(api.DefaultConfig())
  if err != nil {
      panic(err)
  }

  return &ConsulDevice{
    client: client,
    kv:     client.KV(),
  }
}

func (c *ConsulDevice) Get(path string) (data interface{}) {
  if path == "/" {
    return []string{"kv"}
  }

  pair, _, err := c.kv.Get(path[1:], nil)
  if err != nil {
    panic(err)
  }

  if pair == nil {
    return nil
  }
  return string(pair.Value) // []byte
}

func (d *ConsulDevice) Set(path string, data interface{}) {
  p := &api.KVPair{Key: path[1:], Value: []byte(data.(string))}
  _, err := d.kv.Put(p, nil)
  if err != nil {
      panic(err)
  }
}

func (d *ConsulDevice) Map(path string, input interface{}) (output interface{}) {
  return nil
}

func (d *ConsulDevice) Observe(path string) (stream <-chan interface{}) {
  return nil
}
