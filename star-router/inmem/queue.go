package inmem

import (
	"github.com/danopia/stardust/star-router/base"
)

// Manages an in-memory Queue structure
// Queues get Closed instead of Frozen
// A closed queue will not accept any more entries
// TODO: allow restricting queue contents by type/etc?
type Queue struct {
	name     string
	writable bool
	channel  chan base.Entry
}

var _ base.Queue = (*Queue)(nil)

func NewSyncQueue(name string) *Queue {
	return &Queue{
		name:     name,
		writable: true,
		channel:  make(chan base.Entry),
	}
}

func NewBufferedQueue(name string, buffer int) *Queue {
	return &Queue{
		name:     name,
		writable: true,
		channel:  make(chan base.Entry, buffer),
	}
}

// Prevents this queue from ever receiving new entries
func (e *Queue) Close() {
	e.writable = false
}

func (e *Queue) Name() string {
	return e.name
}

func (e *Queue) Push(value base.Entry) (ok bool) {
	e.channel <- value
	return true
}

func (e *Queue) Next() (value base.Entry, ok bool) {
	value, ok = <-e.channel
	return
}

func (e *Queue) TryNext() (value base.Entry, ok bool) {
	select {
	case value, ok = <-e.channel:
		return
	default:
		return nil, false
	}
}
