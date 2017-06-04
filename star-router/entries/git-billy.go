package entries

// Implements go-billy to support the go-git library.
// Basically exposes a Stardust File/Folder tree to go-git cleanly.

import (
	"log"
  "strings"
  "os"
  "io"
  "time"
	"path"

  "gopkg.in/src-d/go-billy.v2"
	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

type billyAdapter struct {
	ctx base.Context
  prefix string
}

func (a *billyAdapter) Create(filename string) (billy.File, error) {
  log.Println("[billy] create", filename)
	ok := a.ctx.Put(a.prefix + "/" + filename, inmem.NewFile(path.Base(filename), nil))
	if !ok {
		log.Println("billy create:", filename, "couldn't be made")
		return nil, os.ErrNotExist
	}
	return a.Open(filename)
}
func (a *billyAdapter) Open(filename string) (billy.File, error) {
  log.Println("[billy] open", filename)
  if entry, ok := a.ctx.GetFile(a.prefix + "/" + filename); ok {
    return &billyFile{
			ctx: a.ctx,
			path: a.prefix + "/" + filename,
      file: entry,
    }, nil
  } else {
    return nil, os.ErrNotExist
  }
}
func (a *billyAdapter) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
  log.Println("[billy] openfile", filename, flag, perm)
  return nil, billy.ErrNotSupported
}
func (a *billyAdapter) Stat(filename string) (billy.FileInfo, error) {
  log.Println("[billy] stat", filename)
  if entry, ok := a.ctx.Get(a.prefix + "/" + filename); ok {
    return &billyFileInfo{entry}, nil
  } else {
    return nil, os.ErrNotExist
  }
}
func (a *billyAdapter) ReadDir(path string) ([]billy.FileInfo, error) {
  log.Println("[billy] readdir", path)
  if folder, ok := a.ctx.GetFolder(a.prefix + "/" + path); ok {
    children := folder.Children()
		stats := make([]billy.FileInfo, len(children))
		for i, name := range children {
			stats[i], _ = a.Stat(path + "/" + name)
		}
		return stats, nil
  } else {
    return nil, os.ErrNotExist
  }
}
func (a *billyAdapter) TempFile(dir, prefix string) (billy.File, error) {
  log.Println("[billy] tempfile", dir, prefix)
  return nil, billy.ErrNotSupported
}
func (a *billyAdapter) Rename(from, to string) error {
  log.Println("[billy] rename", from, to)
  return billy.ErrNotSupported
}
func (a *billyAdapter) Remove(filename string) error {
  log.Println("[billy] remove", filename)
  return billy.ErrNotSupported
}
func (a *billyAdapter) MkdirAll(filename string, perm os.FileMode) error {
  log.Println("[billy] mkdirall", filename, perm)
  // TODO: use perm
  parts := strings.Split(filename, "/")
  var partial string
  for _, part := range parts {
    if len(partial) > 0 {
      partial += "/"
    }
    partial += part

    stat, err := a.Stat(partial)
    if err == os.ErrNotExist {
      ok := a.ctx.Put(a.prefix + "/" + partial, inmem.NewFolder(part))
      if !ok {
        log.Println("billy mkdirall:", partial, "couldn't be made")
        return billy.ErrNotSupported
      }
    } else if err != nil {
      return err
    } else if !stat.IsDir() {
      log.Println("billy mkdirall:", partial, "already exists as non-dir")
      return billy.ErrNotSupported
    }
  }
  return nil
}
func (a *billyAdapter) Join(elem ...string) string {
  log.Println("[billy] join", elem)
  return strings.Join(elem, "/")
}
func (a *billyAdapter) Dir(path string) billy.Filesystem {
  log.Println("[billy] dir", path)
  return nil
}
func (a *billyAdapter) Base() string {
  log.Println("[billy] base")
  return ""
}

type billyFileInfo struct {
  entry base.Entry
}

func (f *billyFileInfo) Name() string {
  log.Println("[billy fileinfo] name", f.entry.Name())
  return f.entry.Name()
}
func (f *billyFileInfo) Size() int64 {
  log.Println("[billy fileinfo] size", f.entry.Name())
  if file, ok := f.entry.(base.File); ok {
    return file.GetSize()
  }
  return -1
}
func (f *billyFileInfo) Mode() os.FileMode {
  var mode os.FileMode = 0777

  if _, ok := f.entry.(base.Folder); ok {
    mode = mode | os.ModeDir
  }

  if _, ok := f.entry.(base.Link); ok {
    mode = mode | os.ModeSymlink
  }

  // TODO: https://golang.org/pkg/os/#FileMode
  log.Println("[billy fileinfo] mode", f.entry.Name(), mode)
  return mode
}
func (f *billyFileInfo) ModTime() time.Time {
  log.Println("[billy fileinfo] modtime", f.entry.Name())
  return time.Unix(0, 0)
}
func (f *billyFileInfo) IsDir() bool {
  log.Println("[billy fileinfo] isdir", f.entry.Name())
  _, ok := f.entry.(base.Folder)
  return ok
}
func (f *billyFileInfo) Sys() interface{} {
  log.Println("[billy fileinfo] sys", f.entry.Name())
  return nil // Sys is optional
}



type billyFile struct {
	ctx base.Context
	path string
  file base.File
  offset int64
}

func (f *billyFile) Filename() string {
  log.Println("[billy file] filename", f.file.Name())
  return f.file.Name()
}

func (f *billyFile) IsClosed() bool {
  log.Println("[billy file] isclosed", f.file.Name())
  return false // TODO
}

func (f *billyFile) Write(p []byte) (n int, err error) {
  log.Println("[billy file] write", f.file.Name(), len(p))
	n = f.file.Write(f.offset, p)
	if n < len(p) {
		log.Println("write failed:", n, len(p))
		err = io.EOF // TODO
	}
	f.offset += int64(n)
	return
}

func (f *billyFile) Read(p []byte) (n int, err error) {
  log.Println("[billy file] read", f.file.Name(), len(p))
	bytes := f.file.Read(f.offset, len(p))
	copy(p, bytes)
	n = len(bytes)
	if n < len(p) {
		err = io.EOF
	}
	f.offset += int64(n)
	return
}

func (f *billyFile) Seek(offset int64, whence int) (n int64, err error) {
  log.Println("[billy file] seek", f.file.Name(), offset, whence)
	size := f.file.GetSize()
	switch whence {

	case io.SeekStart:
		f.offset = offset

	case io.SeekCurrent:
		f.offset = f.offset + offset

	case io.SeekEnd:
		f.offset = size + offset
	}

	if f.offset < 0 {
		err = io.EOF
	}
	n = f.offset
	return
}

// commits the file to the store
// TODO: support close more widely
func (f *billyFile) Close() error {
  log.Println("[billy file] close", f.file.Name())
	if ok := f.ctx.Put(f.path, f.file); ok {
		return nil
	} else {
		return billy.ErrNotSupported
	}
}
