package base

type Namespace struct {
  Root    Channel
  //Binds   map[int64]Channel
}

/*
func NewNamespace() (n *Namespace) {
  root := channels.NewVolatileFolder(1)

  fd := root.

  n = &Namespace{
    Root: root,
  }
  return
}
*/

/*
func (c *Channel) walk(names []string) (newChan Channel, qids []int64) {
  qids = []int64{}
  for name := range names {

  }
}
*/
