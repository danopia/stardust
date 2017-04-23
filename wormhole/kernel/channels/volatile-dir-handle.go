package channels

import (
  //"log"
  //"github.com/danopia/stardust/wormhole/kernel/base"
)

type VolatileDirHandle struct {
  file *Volatile
}

func (v *VolatileDirHandle) Qid() int64 {
  return v.file.qid
}

func (v *VolatileDirHandle) IoUnit() uint32 {
  return 0
}

func (v *VolatileDirHandle) Read(offset uint64, count uint32) (data []byte) {
  return []byte("yes this is dir")
}

func (v *VolatileDirHandle) Write(offset uint64, data []byte) (count uint32) {
  return 0
}

func (v *VolatileDirHandle) Clunk() {
  // TODO
}
