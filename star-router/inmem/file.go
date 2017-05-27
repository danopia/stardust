package inmem

import (
	"github.com/danopia/stardust/star-router/base"
)

// Manages an mutable-or-not in-memory byte array
type File struct {
	name     string
	writable bool
	data     []byte
}

var _ base.File = (*File)(nil)

func NewFile(name string, data []byte) *File {
	return &File{
		name:     name,
		writable: true,
		data:     data,
	}
}

func (e *File) Name() string {
	return e.name
}

// Prevents the binary data from ever changing again
// Chainable for NewFile(...).Freeze()
func (e *File) Freeze() *File {
	e.writable = false
	return e
}

func (e *File) GetSize() int64 {
	return int64(len(e.data))
}

func (e *File) Read(offset int64, numBytes int) (data []byte) {
	bytesRead := len(e.data) - int(offset)
	if bytesRead < 1 {
		// TODO: warn
		return nil
	}

	if numBytes < bytesRead {
		bytesRead = numBytes
	}

	data = make([]byte, bytesRead)
	copy(data, e.data[offset:])
	return
}

func (e *File) Write(offset int64, data []byte) (numBytes int) {
	// TODO
	return 0
}

func (e *File) Truncate() (ok bool) {
	// TODO
	return false
}
