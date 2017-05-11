package inmem

import (
	"sync"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/wormhole/common"
)

// Manages an in-memory Log structure
// Logs never end (yet) and don't expose their contents
// To read from a Log, subscribe a function or queue
// TODO: A subscription is a Folder with control functions
type Log struct {
	name      string
	writable  bool
	feed      chan base.Entry
	ents      map[int64]base.Entry
	dispenser <-chan int64
	latest    int64
	subs      []*LogSub
}

var _ base.Log = (*Log)(nil)

func NewLog(name string) *Log {
	log := &Log{
		name:      name,
		writable:  true,
		feed:      make(chan base.Entry),
		ents:      make(map[int64]base.Entry),
		dispenser: common.NewInt64Dispenser().C,
		subs:      []*LogSub{},
	}

	go log.pumpFeed()
	return log
}

func (e *Log) pumpFeed() {
	for input := range e.feed {
		seq := <-e.dispenser
		e.ents[seq] = input
		e.latest = seq

		// notify any waiting clients about it
		for _, sub := range e.subs {
			select {
			case sub.waitChan <- seq:
			}
		}
	}
}

func (e *Log) Name() string {
	return e.name
}

// Prevents this log from ever receiving new entries
func (e *Log) Close() {
	if e.writable {
		close(e.feed)
	}
	e.writable = false
}

func (e *Log) Append(value base.Entry) (ok bool) {
	if e.writable {
		e.feed <- value
		ok = true
	}
	return
}

// always start at idx 0
// TODO: let you start at latest, probably.
func (e *Log) Subscribe(opts base.Entry) (queue base.Queue) {
	sub := &LogSub{
		log:      e,
		waitChan: make(chan int64),
	}
	e.subs = append(e.subs, sub)
	return sub
}

// Represents one subscription to a Log
type LogSub struct {
	log      *Log
	lastIdx  int64
	waitChan chan int64
	mutex    sync.Mutex
}

var _ base.Queue = (*LogSub)(nil)

func (e *LogSub) Name() string {
	return "subscription"
}

// won't accept outside writes
func (e *LogSub) Close() {
}

// won't accept outside writes
func (e *LogSub) Push(value base.Entry) (ok bool) {
	return false
}

// get the next entry, waiting for more data if needed
func (e *LogSub) Next() (value base.Entry, ok bool) {
	value, ok = e.TryNext()
	if !ok {
		<-e.waitChan
		value, ok = e.TryNext()
	}
	return
}

// if the log has more data, grab it asap
func (e *LogSub) TryNext() (value base.Entry, ok bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.lastIdx < e.log.latest {
		e.lastIdx += 1
		return e.log.ents[e.lastIdx], true
	}
	return
}
