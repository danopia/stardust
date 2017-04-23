package channels

import (
  //"log"
  //"github.com/danopia/stardust/wormhole/kernel/base"
)

type VolatileHandle struct {
  file *Volatile
}

func (v *VolatileHandle) Qid() int64 {
  return v.file.qid
}

func (v *VolatileHandle) IoUnit() uint32 {
  return 256
}

func (v *VolatileHandle) Read(offset uint64, count uint32) (data []byte) {
  return v.file.data
}

func (v *VolatileHandle) Write(offset uint64, data []byte) (count uint32) {
  v.file.data = data
  return uint32(len(data))
}

func (v *VolatileHandle) Clunk() {
  // TODO
}
