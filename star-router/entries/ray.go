package entries

import (
	"fmt"
	"log"
	"strings"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/inmem"
	"github.com/mattn/go-shellwords"
)

// Evaluation context for a ray
// Handles instructions one-by-one

type rayCtx struct {
	commands base.Queue
	output   base.Queue
	result   base.Queue
	scope    base.Folder

	handle base.Handle
}

func newRayCtx() *rayCtx {
	return &rayCtx{
		commands: inmem.NewSyncQueue("commands"),
		output:   inmem.NewBufferedQueue("output", 10),
		result:   inmem.NewBufferedQueue("commands", 1),
		scope:    inmem.NewFolder("scope"),
		handle:   base.RootSpace.NewHandle(),
	}
}

func (c *rayCtx) pumpCommands() {
	defer c.output.Close()
	for {
		entry, ok := c.commands.Next()
		if !ok {
			return
		}
		line, ok := entry.(base.String).Get()
		if !ok {
			log.Println("Ray failed to get string from", entry)
			return
		}

		log.Printf("+ %+v", line)

		parts, err := shellwords.Parse(line)
		if err != nil {
			log.Println("Ray failed to parse:", err)
			return
		}

		ok = c.evalCommand(parts[0], parts[1:])
		if !ok {
			log.Println("Ray failed at", line)
			return
		}

	}
}

func (c *rayCtx) getBundle() base.Folder {
	folder := inmem.NewFolder("ray-invocation")
	folder.Put("commands", c.commands)
	folder.Put("output", c.output)
	folder.Put("result", c.result)
	folder.Put("scope", c.scope)
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
		c.output.Push(inmem.NewString(cmd, text))
		ok = true

	case "ls":
		var folder base.Folder
		if len(args) == 1 {
			temp := c.handle.Clone()
			if ok = temp.Walk(args[0]); !ok {
				return
			}

			if folder, ok = temp.GetFolder(); !ok {
				return
			}
		} else if len(args) == 0 {
			if folder, ok = c.handle.GetFolder(); !ok {
				return
			}
		} else {
			return
		}

		// TODO: can't this fail?
		names := folder.Children()
		text := strings.Join(names, "\t")
		c.output.Push(inmem.NewString(cmd, text))

	default:
		text := fmt.Sprintf("No such command: %v", cmd)
		c.output.Push(inmem.NewString(cmd, text))
	}
	return
}

// Function that creates a new consul client when invoked
type rayFunc struct{}

var _ base.Function = (*rayFunc)(nil)

func (e *rayFunc) Name() string {
	return "ray"
}

func (e *rayFunc) Invoke(input base.Entry) (output base.Entry) {
	// Start a new command evaluator
	ctx := newRayCtx()

	go func(ctx *rayCtx) {
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

			ctx.commands.Push(inmem.NewString("ray-input", line))
		}
		ctx.commands.Close()
	}(ctx)

	go ctx.pumpCommands()
	ctx.writeOutputToStdout()
	return ctx.getBundle()
}

func (c *rayCtx) writeOutputToStdout() {
	for {
		entry, ok := c.output.Next()
		if !ok {
			return
		}
		line, ok := entry.(base.String).Get()
		if !ok {
			log.Println("Ray failed to get string from output", entry)
			return
		}

		log.Printf("> %+v", line)
	}
}
