package common

type Int64Dispenser struct {
  lastInt int64
  C <-chan int64
}

func NewInt64Dispenser() (d *Int64Dispenser) {
  C := make(chan int64)

  d = &Int64Dispenser{
    lastInt: 0,
    C: C,
  }

  go func() {
    for {
      nextInt := d.lastInt + 1
      C <- nextInt
      d.lastInt = nextInt
    }
  }()

  return d
}
