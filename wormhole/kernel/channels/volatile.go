package channels

import (
  "log"
  "github.com/danopia/stardust/wormhole/kernel/base"
)

type Volatile struct {
  qid int64
  names map[string]*Volatile
  data []byte
}

func (v *Volatile) Qid() int64 {
  return v.qid
}

func (v *Volatile) isFolder() bool {
  return v.names != nil
}

func (v *Volatile) Walk(names []string) (newChan base.Channel, qids []int64) {
  var ok bool
  currentChan := v

  qids = []int64{}

  for _, name := range names {
    if !currentChan.isFolder() {
      log.Println("Can't walk non-directory to", name)
      return
    }

    currentChan, ok = currentChan.names[name]
    if !ok {
      log.Println("Can't walk to nonexistant name", name)
      return
    }

    qids = append(qids, currentChan.Qid())
  }

  log.Println("Done walking", names)
  newChan = currentChan
  return
}


func (v *Volatile) Open(mode byte) base.FileHandle {
  if v.isFolder() {
    return &VolatileDirHandle{
      file: v,
    }

  } else {
    return &VolatileHandle{
      file: v,
    }
  }
}

func (v *Volatile) Create(name string, perm uint32, mode byte) base.FileHandle {
  if !v.isFolder() {
    log.Println("Can't create name", name, "under a file")
    return nil
  }

  file := &Volatile{
    qid: 2,
    data: []byte{},
  }
  v.names[name] = file

  channel, _ := v.Walk([]string{name})
  return channel.Open(mode)
}

func NewVolatileFolder(qid int64) *Volatile {
  return &Volatile{
    qid: qid,
    names: make(map[string]*Volatile),
  }
}
