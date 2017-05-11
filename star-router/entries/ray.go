package entries

import (
	"fmt"
	"log"
	"path"
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
	output   base.Log
	//result   base.Queue
	environ base.Folder
	cwd     base.String

	handle base.Handle
}

func newRayCtx() *rayCtx {
	log.Println("Starting new Ray")
	ctx := &rayCtx{
		commands: inmem.NewSyncQueue("commands"),
		output:   inmem.NewLog("output"),
		//result:   inmem.NewBufferedQueue("result", 1),
		environ: inmem.NewFolder("environ"),
		handle:  base.RootSpace.NewHandle(),
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
	return inmem.NewFolderOf("ray-invocation",
		c.commands,
		c.output,
		//c.result,
		c.environ,
		c.cwd,
	).Freeze()
}

func (c *rayCtx) writeOut(label, line string) {
	c.output.Append(inmem.NewString(label, line))
}

func (c *rayCtx) evalCommand(cmd string, args []string) (ok bool) {
	cmd = strings.ToLower(cmd)
	switch cmd {

	case "help":
		c.writeOut(cmd, "Available commands:")
		cmdList := []string{"help", "cat", "cd", "echo", "ls", "invoke"}
		for _, cmd := range cmdList {
			c.writeOut(cmd, fmt.Sprintf("  - %s", cmd))
		}
		c.writeOut(cmd, "")
		c.writeOut(cmd, "The shell will exit on any error.")
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
				c.writeOut(cmd, value)

			default:
				c.writeOut(cmd, "Name wasn't a type that you can cat")
				ok = false
				return
			}
		}

	case "invoke":
		if len(args) == 0 {
			c.writeOut(cmd, "Not enough args for invoke")
			return
		}

		temp := c.handle.Clone()
		if ok = temp.Walk(args[0]); !ok {
			c.writeOut(cmd, fmt.Sprintf("Couldn't find function named %s", args[0]))
			return
		}
		function := temp.Get().(base.Function)

		var input base.Entry
		if len(args) >= 2 {
			temp := c.handle.Clone()
			if ok = temp.Walk(args[1]); !ok {
				c.writeOut(cmd, fmt.Sprintf("Couldn't find input named %s", args[1]))
				return
			}
			input = temp.Get()
		}

		output := function.Invoke(input)

		if len(args) >= 3 && output != nil {
			parentDir := path.Dir(args[2])
			temp := c.handle.Clone()
			if ok = temp.Walk(parentDir); !ok {
				c.writeOut(cmd, fmt.Sprintf("Couldn't find output parent named %s", parentDir))
				return
			}
			outputParent := temp.Get().(base.Folder)
			outputParent.Put(path.Base(args[2]), output)
			c.writeOut(cmd, fmt.Sprintf("Wrote result to %s", args[2]))
		}

	case "cd":
		if len(args) == 1 {
			ok = c.handle.Walk(args[0])
		} else if len(args) == 0 {
			ok = c.handle.Walk("/")
		}

	case "echo":
		text := strings.Join(args, " ")
		c.writeOut(cmd, text)
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
		c.writeOut(cmd, text)

	default:
		text := fmt.Sprintf("No such command: %v", cmd)
		c.writeOut(cmd, text)
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
	ctx := newRayCtx()

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

	//case base.Queue:
	//	ctx.commands = input // TODO: do this earlier

	default:
		if input != nil {
			log.Println("Ray can't deal with input", input)
			panic("Ray got weird input")
		}
	}

	go ctx.writeOutputToStdout() // TODO
	go ctx.pumpCommands()
	return ctx.getBundle()
}

func (c *rayCtx) writeOutputToStdout() {
	sub := c.output.Subscribe(nil)
	for {
		entry, ok := sub.Next()
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
