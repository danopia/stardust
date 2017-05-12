package entries

import (
	"log"
	"path"
	"time"

	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/inmem"
)

// Function that creates a new ray shell when invoked
type initFunc struct{}

var _ base.Function = (*initFunc)(nil)

func (e *initFunc) Name() string {
	return "init"
}

func (e *initFunc) Invoke(input base.Entry) (output base.Entry) {
	log.Println("Loading init process...")

	s := &initSvc{
		cfgDir:   input.(base.Folder), // TODO
		services: make(map[string]*service),
		handle:   base.RootSpace.NewHandle(),
	}

	for _, name := range s.cfgDir.Children() {
		if folder, ok := s.cfgDir.Fetch(name); ok {
			s.services[name] = newService(folder.(base.Folder))
		}
	}

	for name, svc := range s.services {
		if !svc.running {
			log.Println("Starting service", name)
			s.start(svc)
		}
	}
	log.Println("All services started")

	for {
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

// Context for a long-term init processer
type initSvc struct {
	cfgDir   base.Folder
	services map[string]*service
	handle   base.Handle
}

//var _ base.Folder = (*initSvc)(nil)

func (s *initSvc) start(svc *service) {
	svc.running = true

	pathEntry, ok := svc.cfgDir.Fetch("path")
	if !ok {
		return
	}
	runPath, ok := pathEntry.(base.String).Get()
	if !ok {
		return
	}

	temp := s.handle.Clone()
	temp.Walk(runPath)
	runFunc, ok := temp.GetFunction()
	if !ok {
		return
	}

	inputPathEntry, ok := svc.cfgDir.Fetch("input-path")
	if !ok {
		return
	}
	inputPath, ok := inputPathEntry.(base.String).Get()
	if !ok {
		return
	}

	temp = s.handle.Clone()
	temp.Walk(inputPath)
	inputEntry := temp.Get()
	if !ok {
		return
	}

	output := runFunc.Invoke(inputEntry)

	mountPathEntry, ok := svc.cfgDir.Fetch("mount-path")
	if !ok {
		return
	}
	mountPath, ok := mountPathEntry.(base.String).Get()
	if !ok {
		mountPath = ""
	}

	if mountPath != "" && output != nil {
		parentDir := path.Dir(mountPath)
		temp := s.handle.Clone()
		if ok = temp.Walk(parentDir); !ok {
			//c.writeOut(cmd, fmt.Sprintf("Couldn't find output parent named %s", parentDir))
			return
		}
		outputParent := temp.Get().(base.Folder)
		outputParent.Put(path.Base(mountPath), output)
		//c.writeOut(cmd, fmt.Sprintf("Wrote result to %s", args[2]))
	}
}

/*
func (s *initSvc) getBundle() base.Folder {
	return inmem.NewFolderOf("ray-invocation",
		c.commands,
		c.output,
		//c.result,
		c.environ,
		c.cwd,
	).Freeze()
}
*/

type service struct {
	cfgDir  base.Folder
	running bool
}

func newService(cfg base.Folder) *service {
	return &service{
		cfgDir: cfg,
	}
}
