package base

// A virtual device which can access outside resources
type Device interface {
  // Name string
  //Properties map[string]interface{}

  Get(path string) (data interface{})
  Put(path string, data interface{})
}
