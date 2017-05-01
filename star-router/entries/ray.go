package entries

import (
	"github.com/danopia/stardust/star-router/base"
	"github.com/mattn/go-shellwords"
	"log"
	"strings"
	"github.com/danopia/stardust/star-router/inmem"
)

// Evaluation context for a ray
// Handles instructions one-by-one

type rayCtx struct {
	commands base.Queue
	output base.Queue
	result base.Queue
	scope base.Folder

	handle base.Handle
}

func newRayCtx() *rayCtx {
	return &rayCtx{
		commands: inmem.NewSyncQueue("commands"),
		output: inmem.NewBufferedQueue("output", 10),
		result: inmem.NewBufferedQueue("commands", 1),
		scope: inmem.NewFolder("scope"),
		handle: base.RootSpace.NewHandle(),
	}
}

func (c *rayCtx) getBundle() base.Folder {
	folder := inmem.NewFolder("ray-invocation")
	folder.Put("commands", commands)
	folder.Put("output", output)
	folder.Put("result", result)
	folder.Put("scope", scope)
	folder.Freeze()
	return folder
}

func (c *rayCtx) evalCommand(cmd string, args []string) (ok bool) {
	cmd = strings.ToLower(cmd)
	switch cmd {

	case "cd":
		if len(args) == 1 {
			ok = c.handle.Walk(args[0])
		}

	case "echo":
		text := strings.Join(args, " ")
		c.output.Push()
		log.Println(parts[1:])
		ok = true

	case "ls":
		var folder base.Folder
		if len(parts) == 2 {
			temp := handle.Clone()
			if ok = temp.Walk(parts[1]); !ok {
				break
			}

			if folder, ok = temp.GetFolder(); !ok {
				break
			}
		} else if len(parts) == 1 {
			if folder, ok = handle.GetFolder(); !ok {
				break
			}
		} else {
			break
		}

		// TODO: can't this fail?
		names := folder.Children()
		log.Println(strings.Join(names, "\t"))

	default:
		log.Println("No such command:", cmd)
	}

}

// Function that creates a new consul client when invoked
type rayFunc struct{}

var _ base.Function = (*rayFunc)(nil)

func (e *rayFunc) Name() string {
	return "ray"
}

func (e *rayFunc) Invoke(input base.Entry) (output base.Entry) {
	ctx := newRayCtx()

	go func() {

	}()
	return ctx.getBundle()

	str := input.(base.String)
	script, ok := str.Get()
	if !ok {
		panic("Ray couldn't get script contents")
	}


	lines := strings.Split(script, "\n")
	for _, raw := range lines {
		line := strings.Trim(raw, " \t")
		if len(line) == 0 {
			continue
		}

		log.Printf("+ %+v", line)
		ok := false

		parts, err := shellwords.Parse(line)
		if err != nil {
			log.Println("Ray failed to parse:", err)
			return nil
		}

		if !ok {
			log.Println("Ray failed at", line)
			return nil
		}

	}
	return nil
}
