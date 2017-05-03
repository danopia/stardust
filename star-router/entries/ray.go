package entries

import (
	"fmt"
	"log"
	"strings"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/inmem"
	"github.com/mattn/go-shellwords"
)

// Read-only string impl that returns the given handle's location
type handlePathString struct {
	handle base.Handle
}

var _ base.String = (*handlePathString)(nil)

func (e *handlePathString) Name() string {
	return "cwd"
}

func (e *handlePathString) Get() (value string, ok bool) {
	return e.handle.Path(), true
}

func (e *handlePathString) Set(value string) (ok bool) {
	return false
}

// Evaluation context for a ray
// Handles instructions one-by-one

type rayCtx struct {
	commands base.Queue
	output   base.Queue
	result   base.Queue
	environ  base.Folder
	cwd      base.String

	handle base.Handle
}

func newRayCtx(cmdQueue base.Queue) *rayCtx {
	ctx := &rayCtx{
		commands: cmdQueue,
		output:   inmem.NewBufferedQueue("output", 10),
		result:   inmem.NewBufferedQueue("result", 1),
		environ:  inmem.NewFolder("environ"),
		handle:   base.RootSpace.NewHandle(),
	}
	ctx.cwd = &handlePathString{ctx.handle}
	return ctx
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
	folder.Put("environ", c.environ)
	folder.Put("cwd", c.cwd)
	folder.Freeze()
	return folder
}

func (c *rayCtx) evalCommand(cmd string, args []string) (ok bool) {
	cmd = strings.ToLower(cmd)
	switch cmd {

	case "help":
		c.output.Push(inmem.NewString(cmd, "Available commands:"))
		cmdList := []string{"help", "cat", "cd", "echo", "ls"}
		for _, cmd := range cmdList {
			c.output.Push(inmem.NewString(cmd, fmt.Sprintf("  - %s", cmd)))
		}
		c.output.Push(inmem.NewString(cmd, ""))
		c.output.Push(inmem.NewString(cmd, "The shell will exit on any error."))
		ok = true

	case "cat":
		ok = true
		for _, path := range args {
			temp := c.handle.Clone()
			if ok = temp.Walk(path); !ok {
				return
			}
			entry := temp.Get()

			switch entry := entry.(type) {
			case base.String:
				var value string
				value, ok = entry.Get()
				if !ok {
					return
				}
				c.output.Push(inmem.NewString(cmd, value))

			default:
				ok = false
				return
			}
		}

	case "cd":
		if len(args) == 1 {
			ok = c.handle.Walk(args[0])
		} else if len(args) == 0 {
			ok = c.handle.Walk("/")
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

// Function that creates a new ray shell when invoked
type rayFunc struct{}

var _ base.Function = (*rayFunc)(nil)

func (e *rayFunc) Name() string {
	return "ray"
}

func (e *rayFunc) Invoke(input base.Entry) (output base.Entry) {
	// Start a new command evaluator
	ctx := newRayCtx(inmem.NewSyncQueue("commands"))

	switch input := input.(type) {
	case base.String:

		go func(ctx *rayCtx) {
			script, ok := input.Get()
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

	case base.Queue:
		ctx.commands = input // TODO: do this earlier

	default:
		log.Println("Ray can't deal with input", input)
		panic("Ray got weird input")
	}

	go ctx.pumpCommands()
	//ctx.writeOutputToStdout()
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
