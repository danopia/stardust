package entries

import (
	//"fmt"
	"log"
	"os"
	//"sort"
	"path"
	"strings"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

// Directory containing the clone function
func getOsDriver() *inmem.Folder {
	return inmem.NewFolderOf("host-os",
		inmem.NewFolderOf("fs",
			inmem.NewFunction("invoke", startOsFs),
		).Freeze(),
	).Freeze()
}

// Function that creates a new Osfs client when invoked
func startOsFs(ctx base.Context, input base.Entry) (output base.Entry) {
	// TODO: accept root path and acl info

	hostPath := "."
	if folder, ok := input.(base.Folder); ok {
		hostPath, _ = helpers.GetChildString(folder, "host-path")
	}

	//if _, err := os.Open("/"); err == nil {
	return &osFolder{
		root: hostPath,
		path: "",
		//file: file,
	}
}

// Directory/String tree backed by host operating system
// This struct only represents a folder
// inmem files are used to handle keys?
type osFolder struct {
	root string
	path string
}

var _ base.Folder = (*osFolder)(nil)

func (e *osFolder) Name() string {
	return path.Base(e.path)
}

func (e *osFolder) Children() []string {
	f, err := os.Open(path.Join(e.root, e.path))
	if err != nil {
		log.Println("couldn't open", e.root, e.path, "to readdir")
		return nil // TODO
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil // TODO
	}
	//sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })

	children := make([]string, len(list))
	for i, f := range list {
		children[i] = f.Name()
	}
	return children
}

func (e *osFolder) Fetch(name string) (entry base.Entry, ok bool) {
	p := path.Clean(path.Join(e.path, name))
	// TODO: or maybe stardust will block '..' later?
	if !strings.HasPrefix(p, e.path) {
		log.Println("osfs: preventing breakout from", e.path, "to", p, "using", name)
		return
	}

	fi, err := os.Open(path.Join(e.root, p))
	if err != nil {
		return
	}
	stat, err := fi.Stat()
	if err != nil {
		return
	}

	switch mode := stat.Mode(); {
	case mode.IsRegular():
		return &osFile{e.root, p, fi}, true
	case mode.IsDir():
		return &osFolder{e.root, p}, true
	// case mode&os.ModeSymlink != 0:
	// case mode&os.ModeNamedPipe != 0:
	default:
		return
	}
}

func (e *osFolder) Put(name string, entry base.Entry) (ok bool) {
	/*
		prefix := e.path
		if len(prefix) > 0 {
			prefix += "/"
		}

		switch entry := entry.(type) {

		case base.String:
			// TODO: make sure not already a folder?
			_, err := e.kv.Put(&api.KVPair{
				Key:   prefix + name,
				Value: []byte(entry.Get()),
			}, nil)
			if err != nil {
				panic(err)
			}
			ok = true

		case base.Folder:
			// explicitly create the folder entry
			_, err := e.kv.Put(&api.KVPair{
				Key: prefix + name + "/",
			}, nil)
			if err != nil {
				panic(err)
			}

			// if the folder has contents, let's recursively store
			// TODO: clear existing children first??
			names := entry.Children()
			if len(names) > 0 {
				newEntry, dirOk := e.Fetch(name)
				newDir := newEntry.(base.Folder)
				if dirOk {
					ok = true
					for _, subName := range names {
						value, getOk := entry.Fetch(subName)
						if getOk {
							newDir.Put(subName, value)
						} else {
							ok = false
						}
					}
				}
			} else {
				ok = true
			}

		default:
			log.Println("consul: tried to store weird thing", entry, "as", name)

		}*/
	return
}

// Byte buffer backed by host operating system
type osFile struct {
	root string
	path string
	file *os.File
}

var _ base.File = (*osFile)(nil)

func (e *osFile) Name() string {
	return path.Base(e.path)
}

func (e *osFile) GetSize() int64 {
	if stat, err := e.file.Stat(); err == nil {
		return stat.Size()
	}
	return -1
}

func (e *osFile) Read(offset int64, numBytes int) (data []byte) {
	buf := make([]byte, numBytes)
	if n, err := e.file.ReadAt(buf, offset); numBytes > 0 {
		data = buf[0:n]
	} else {
		log.Println("read error on", e.path, "offset", offset, "err", err)
	}
	return
}

func (e *osFile) Write(offset int64, data []byte) (numBytes int) {
	if n, err := e.file.WriteAt(data, offset); err != nil {
		log.Println("write error on", e.path, "offset", offset, "err", err)
		return -1
	} else {
		return n
	}
}

func (e *osFile) Truncate() (ok bool) {
	err := e.file.Truncate(0)
	return err == nil
}

// TODO: close
