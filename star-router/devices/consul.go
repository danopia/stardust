package devices

import (
  "github.com/hashicorp/consul/api"
)

// Manufactors virtual devices which can exist in more than one place at once
type ConsulDevice struct {
  client *api.Client
  kv *api.KV
  // Name string
  //Properties map[string]interface{}
}

func NewConsulDevice() *ConsulDevice {
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
  pair, _, err := c.kv.Get(path, nil)
  if err != nil {
    panic(err)
  }

  if pair == nil {
    return nil
  }
  return string(pair.Value) // []byte
}

func (c *ConsulDevice) Put(path string, data interface{}) {
  p := &api.KVPair{Key: path, Value: []byte(data.(string))}
  _, err := c.kv.Put(p, nil)
  if err != nil {
      panic(err)
  }
}
