package base

type Channel interface {
  Qid() int64

  Walk(names []string) (newChan Channel, qids []int64)
  Open(mode byte) FileHandle
  Create(name string, perm uint32, mode byte) FileHandle
}
