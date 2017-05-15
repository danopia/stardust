package entries

import (
	"log"
	"path"
	"time"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

// Directory containing the clone function
func getInitDriver() base.Folder {
	return inmem.NewFolderOf("init",
		inmem.NewFunction("invoke", initFunc),
	).Freeze()
}

// Function that creates a new ray shell when invoked
func initFunc(input base.Entry) (output base.Entry) {
	log.Println("init: Bootstrapping...")

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

	temp := s.handle.Clone()
	if ok := temp.Walk("/n"); ok {
		outputParent := temp.Get().(base.Folder)
		outputParent.Put("init", s)
	}

	for name, svc := range s.services {
		if !svc.running {
			log.Println("init: Starting service", name)
			s.start(svc)
			if !svc.running {
				log.Println("init: Failed to start", name)
			}
		}
	}
	log.Println("init: All services started")

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

var _ base.Folder = (*initSvc)(nil)

func (s *initSvc) start(svc *service) {
	runPath, ok := helpers.GetChildString(svc.cfgDir, "path")
	if !ok {
		return
	}

	// things about the invocation function
	var runFunc base.Function
	var inputShape base.Shape

	temp := s.handle.Clone()
	temp.Walk(runPath)
	switch input := temp.Get().(type) {

	case base.Function:
		log.Println("init:", svc.cfgDir.Name(), "is a legacy Function, please update")
		runFunc = input

	case base.Folder:
		if runEntry, ok := input.Fetch("invoke"); ok {
			runFunc = runEntry.(base.Function)
		} else {
			return
		}

		temp2 := temp.Clone()
		temp2.Walk("input-shape")
		inputShape, ok = temp2.GetShape()
		if !ok {
			log.Println("No input shape found for", runPath)
		}

	default:
		log.Println("init:", runPath, "is not executable")

	}

	inputPath, ok := helpers.GetChildString(svc.cfgDir, "input-path")
	if !ok {
		return
	}

	temp = s.handle.Clone()
	temp.Walk(inputPath)
	inputEntry := temp.Get()
	if !ok {
		return
	}

	if inputShape != nil {
		if ok := inputShape.Check(inputEntry); !ok {
			log.Println("Entry", inputPath, "doesn't match shape", inputShape, "for", runPath)
		}
	}

	output := runFunc.Invoke(inputEntry)

	mountPath, ok := helpers.GetChildString(svc.cfgDir, "mount-path")
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

	svc.running = true
}

func (e *initSvc) Name() string {
	return "init"
}

func (e *initSvc) Children() []string {
	names := make([]string, len(e.services))
	i := 0
	for k := range e.services {
		names[i] = k
		i++
	}
	return names
}

func (e *initSvc) Fetch(name string) (entry base.Entry, ok bool) {
	entry, ok = e.services[name]
	return
}

func (e *initSvc) Put(name string, entry base.Entry) (ok bool) {
	return false
}

type service struct {
	cfgDir  base.Folder
	running bool
}

var _ base.Folder = (*service)(nil)

func newService(cfg base.Folder) *service {
	return &service{
		cfgDir: cfg,
	}
}

func (e *service) Name() string {
	return e.cfgDir.Name()
}

func (e *service) Children() []string {
	names := e.cfgDir.Children()
	names = append(names, "running")
	return names
}

func (e *service) Fetch(name string) (entry base.Entry, ok bool) {
	if name == "running" {
		running := "no"
		if e.running {
			running = "yes"
		}
		return inmem.NewString("running", running), true
	}

	// fallback to config dir
	entry, ok = e.cfgDir.Fetch(name)
	return
}

func (e *service) Put(name string, entry base.Entry) (ok bool) {
	// TODO: support writing to "running"
	return false
}
