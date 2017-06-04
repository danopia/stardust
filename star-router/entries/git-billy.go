package entries

// Implements go-billy to support the go-git library.
// Basically exposes a Stardust File/Folder tree to go-git cleanly.

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/danopia/stardust/star-router/base"
	"gopkg.in/src-d/go-billy.v2"
	//"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

type billyAdapter struct {
	ctx       base.Context
	prefix    string
	tempCount int
	tempFiles map[string]base.File
}

func newBillyAdapter(ctx base.Context, prefix string) *billyAdapter {
	return &billyAdapter{
		ctx:       ctx,
		prefix:    prefix,
		tempFiles: make(map[string]base.File),
	}
}

func (a *billyAdapter) Create(filename string) (billy.File, error) {
	log.Println("[billy] create", filename)

	// TODO: include dirmode?
	if err := a.MkdirAll(path.Dir(filename), 0666); err != nil {
		return nil, err
	}

	ok := a.ctx.Put(a.prefix+"/"+filename, inmem.NewFile(path.Base(filename), nil))
	if !ok {
		log.Println("billy create:", filename, "couldn't be made")
		return nil, os.ErrNotExist
	}
	return a.Open(filename)
}
func (a *billyAdapter) Open(filename string) (billy.File, error) {
	log.Println("[billy] open", filename)
	if tempFile, ok := a.tempFiles[filename]; ok {
		return &billyFile{
			ctx:      a.ctx,
			path:     "", // not server-backed
			filename: filename,
			file:     tempFile,
		}, nil
	} else if entry, ok := a.ctx.GetFile(a.prefix + "/" + filename); ok {
		return &billyFile{
			ctx:      a.ctx,
			path:     a.prefix + "/" + filename,
			filename: filename,
			file:     entry,
		}, nil
	} else {
		return nil, os.ErrNotExist
	}
}
func (a *billyAdapter) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	log.Println("[billy] openfile", filename, flag, perm)
	if flag == 577 { // writeonly, truncate, create
		// TODO: include dirmode?
		if err := a.MkdirAll(path.Dir(filename), perm); err != nil {
			return nil, err
		}

		return &billyFile{
			ctx:      a.ctx,
			path:     a.prefix + "/" + filename,
			filename: filename,
			file:     inmem.NewFile(path.Base(filename), nil),
		}, nil
	} else {
		return nil, billy.ErrNotSupported
	}
}
func (a *billyAdapter) Stat(filename string) (billy.FileInfo, error) {
	log.Println("[billy] stat", filename)
	if entry, ok := a.ctx.Get(a.prefix + "/" + strings.TrimPrefix(filename, "/")); ok {
		return &billyFileInfo{entry}, nil
	} else {
		return nil, os.ErrNotExist
	}
}
func (a *billyAdapter) ReadDir(path string) ([]billy.FileInfo, error) {
	log.Println("[billy] readdir", path)
	dirPath := a.prefix
	if len(path) > 0 {
		dirPath += "/" + path
	}

	if folder, ok := a.ctx.GetFolder(dirPath); ok {
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

// tempfiles never commit to the server and are cross-open interactable
func (a *billyAdapter) TempFile(dir, prefix string) (billy.File, error) {
	a.tempCount++
	filename := fmt.Sprintf("%s_%d_%d", prefix, a.tempCount, time.Now().UnixNano())
	log.Println("[billy] tempfile", dir, prefix, filename)

	fullName := a.Join(dir, filename)
	a.tempFiles[fullName] = inmem.NewFile(filename, nil)
	return a.Open(fullName)
}

func (a *billyAdapter) Rename(from, to string) error {
	log.Println("[billy] rename", from, to)
	if tmp, ok := a.tempFiles[from]; ok {
		// commits the tempfile then deletes from ram
		if ok := a.ctx.Put(a.prefix+"/"+to, tmp); ok {
			delete(a.tempFiles, from)
			return nil
		} else {
			return billy.ErrNotSupported
		}

	} else {
		// TODO: files are easy, what about folders?
		return billy.ErrNotSupported
	}
}

func (a *billyAdapter) Remove(filename string) error {
	log.Println("[billy] remove", filename)

	if _, ok := a.tempFiles[filename]; ok {
		// temp files aren't committed, just free from RAM
		delete(a.tempFiles, filename)

	} else if ok := a.ctx.Put(a.prefix+"/"+filename, nil); !ok {
		log.Println("billy remove:", filename, "couldn't be deleted")
		return billy.ErrNotSupported
	}
	return nil
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
			ok := a.ctx.Put(a.prefix+"/"+partial, inmem.NewFolder(part))
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
	if file, ok := f.entry.(base.File); ok {
		size := file.GetSize()
		log.Println("[billy fileinfo] size", f.entry.Name(), size)
		return size
	}
	return -1
}
func (f *billyFileInfo) Mode() os.FileMode {
	var mode os.FileMode = 0664

	if _, ok := f.entry.(base.Folder); ok {
		mode = mode | os.ModeDir | 0111
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
	_, ok := f.entry.(base.Folder)
	log.Println("[billy fileinfo] isdir", f.entry.Name(), ok)
	return ok
}
func (f *billyFileInfo) Sys() interface{} {
	log.Println("[billy fileinfo] sys", f.entry.Name())
	return nil // Sys is optional
}

type billyFile struct {
	ctx      base.Context
	path     string // rooted in ctx
	filename string // rooted in billy root
	file     base.File
	offset   int64
	dirty    bool
}

func (f *billyFile) Filename() string {
	log.Println("[billy file] filename", f.file.Name())
	return f.filename
}

func (f *billyFile) IsClosed() bool {
	log.Println("[billy file] isclosed", f.file.Name())
	return false // TODO
}

func (f *billyFile) Write(p []byte) (n int, err error) {
	//log.Println("[billy file] write", f.file.Name(), len(p))
	n = f.file.Write(f.offset, p)
	f.dirty = true
	if n < len(p) {
		log.Println("write failed:", n, len(p))
		err = io.EOF // TODO
	}
	f.offset += int64(n)
	return
}

func (f *billyFile) Read(p []byte) (n int, err error) {
	bytes := f.file.Read(f.offset, len(p))
	copy(p, bytes)
	n = len(bytes)
	//log.Println("[billy file] read", f.file.Name(), f.offset, len(p), n)
	if n < len(p) {
		err = io.EOF
	}
	f.offset += int64(n)
	return
}

func (f *billyFile) Seek(offset int64, whence int) (n int64, err error) {
	//log.Println("[billy file] seek", f.file.Name(), offset, whence)
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
	if f.path == "" {
		// tempfile; not really stored
		return nil
	} else if !f.dirty {
		// no changes were made
		return nil
	} else if ok := f.ctx.Put(f.path, f.file); ok {
		log.Println("[billy file] close: committed", f.file.Name())
		return nil
	} else {
		return billy.ErrNotSupported
	}
}
