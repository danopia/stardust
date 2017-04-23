package base

// This is basically a FID
type FileHandle interface {
  Qid() int64
  IoUnit() uint32

  Read(offset uint64, count uint32) (data []byte)
  Write(offset uint64, data []byte) (count uint32)
  Clunk()
}
