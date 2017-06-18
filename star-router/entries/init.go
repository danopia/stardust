package entries

import (
	"log"
	"path"
	"sort"
	"time"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/extras"
	"github.com/stardustapp/core/inmem"
)

// Directory containing the clone function
func getInitDriver() base.Folder {
	return inmem.NewFolderOf("init",
		inmem.NewFunction("invoke", initFunc),
	).Freeze()
}

// Function that creates a new ray shell when invoked
func initFunc(ctx base.Context, input base.Entry) (output base.Entry) {
	log.Println("init: Bootstrapping...")

	s := &initSvc{
		ctx:      ctx,
		cfgDir:   input.(base.Folder), // TODO
		services: make(map[string]*service),
	}

	var names []string
	for _, name := range s.cfgDir.Children() {
		if folder, ok := s.cfgDir.Fetch(name); ok {
			s.services[name] = newService(folder.(base.Folder))
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })

	ctx.Put("/n/init", s)

	s.start(s.services["aws-ns"])
	s.start(s.services["redis-ns"])

	for _, name := range names {
		svc := s.services[name]
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
	ctx      base.Context
	cfgDir   base.Folder
	services map[string]*service
}

var _ base.Folder = (*initSvc)(nil)

func (s *initSvc) start(svc *service) {
	runPath, ok := extras.GetChildString(svc.cfgDir, "path")
	if !ok {
		return
	}

	// things about the invocation function
	var runFunc base.Function
	var inputShape base.Shape

	input, _ := s.ctx.Get(runPath)
	switch input := input.(type) {

	case base.Function:
		log.Println("init:", svc.cfgDir.Name(), "is a legacy Function, please update")
		runFunc = input

	case base.Folder:
		if runEntry, ok := input.Fetch("invoke"); ok {
			runFunc = runEntry.(base.Function)
		} else {
			return
		}

		inputShape, ok = s.ctx.GetShape(path.Join(runPath, "input-shape"))
		if !ok {
			log.Println("No input shape found for", runPath)
		}

	default:
		log.Println("init:", runPath, "is not executable")
		return

	}

	var inputEntry base.Entry
	inputPath, ok := extras.GetChildString(svc.cfgDir, "input-path")
	if ok {
		inputEntry, ok = s.ctx.Get(inputPath)
		if !ok {
			return
		}
	} else {
		inputEntry, _ = svc.cfgDir.Fetch("input")
	}

	if inputShape != nil {
		if ok := inputShape.Check(s.ctx, inputEntry); !ok {
			log.Println("Entry", inputPath, "doesn't match shape", inputShape, "for", runPath)
		}
	}

	output := runFunc.Invoke(s.ctx, inputEntry)

	mountPath, ok := extras.GetChildString(svc.cfgDir, "mount-path")
	if !ok {
		mountPath = ""
	}

	if mountPath != "" && output != nil {
		if ok = s.ctx.Put(mountPath, output); !ok {
			log.Println("Couldn't mount to", mountPath)
			return
		}
		log.Println("Mounted to", mountPath)
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
