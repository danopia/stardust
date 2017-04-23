package client

import (
	"log"
  "strconv"

	"github.com/danopia/stardust/wormhole/ddp"
)

type Invocation struct {
  method string
  id int64
  result chan interface{}
	// callback func(result interface{})
}

// called sync from source pump
func (c *SockJsClient) onResult(msg ddp.Message) {
  msgId, err := strconv.ParseInt(msg.Id, 10, 64)
  if err != nil {
    log.Println("Failed to parse result context id", msg.Id, err)
    return
  }

  if ivk, ok := c.invocations[msgId]; ok {
    delete(c.invocations, msgId)
    ivk.result <- msg.Result
    close(ivk.result)

  } else {
    log.Println("Received result for missing method invocation", msgId)
  }
}

// wait for result
//func (c *SockJsClient) Invoke(method string, arg interface{}, cb func(result interface{})) {
func (c *SockJsClient) Invoke(method string, arg interface{}) (result interface{}) {
  ivk := &Invocation{
    method: method,
    id: <-c.idGen,
    result: make(chan interface{}),
    ///callback: cb,
  }
  c.invocations[ivk.id] = ivk

  c.Sink <- ddp.Message{
    Type:   "method",
    Method: ivk.method,
    Id:     strconv.FormatInt(ivk.id, 10),
    Params: []interface{}{arg},
  }

  // wait for result
  result = <- ivk.result
  return
}
