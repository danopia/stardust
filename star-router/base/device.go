package base

// A virtual device which can access outside resources
type Device interface {
  // Name string
  //Properties map[string]interface{}

  Get(path string) (data interface{})
  Set(path string, data interface{})
  Map(path string, input interface{}) (output interface{})
  Observe(path string) (stream <-chan interface{})
}
