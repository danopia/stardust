package entries

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/inmem"
	"github.com/mattn/go-shellwords"
)

// Directory containing the clone function
func getRayDriver() base.Folder {
	return inmem.NewFolderOf("ray",
		inmem.NewFunction("invoke", rayFunc),
		inmem.NewLink("input-shape", "/rom/shapes/ray-opts"),
	).Freeze()
}

// Function that returns the given ray's location
type cwdProvider struct {
	ray *rayCtx
}

var _ base.Function = (*cwdProvider)(nil)

func (e *cwdProvider) Name() string {
	return "cwd"
}

func (e *cwdProvider) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	return inmem.NewString("cwd", e.ray.cwd)
}

// Evaluation context for a ray
// Handles instructions one-by-one

type rayCtx struct {
	ctx      base.Context
	commands base.Queue
	output   base.Log
	//result   base.Queue
	environ base.Folder
	cwd     string
}

func newRayCtx(cctx base.Context) *rayCtx {
	log.Println("Starting new Ray")
	ctx := &rayCtx{
		ctx:      cctx,
		commands: inmem.NewSyncQueue("commands"),
		output:   inmem.NewLog("output"),
		//result:   inmem.NewBufferedQueue("result", 1),
		environ: inmem.NewFolder("environ"),
	}
	return ctx
}

func (c *rayCtx) pumpCommands() {
	defer c.output.Close()
	for {
		entry, ok := c.commands.Next()
		if !ok {
			return
		}
		line, ok := entry.(base.String)
		if !ok {
			log.Println("Ray failed to get string from", entry)
			return
		}

		log.Printf("+ %+v", line.Get())

		parts, err := shellwords.Parse(line.Get())
		if err != nil {
			log.Println("Ray failed to parse:", err)
			return
		}

		ok = c.evalCommand(parts[0], parts[1:])
		if !ok {
			log.Println("Ray failed at", line.Get())
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
		&cwdProvider{c},
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
		cmdList := []string{"help", "cat", "cd", "echo", "ls", "invoke", ":"}
		for _, cmd := range cmdList {
			c.writeOut(cmd, fmt.Sprintf("  - %s", cmd))
			time.Sleep(10 * time.Millisecond)
		}
		c.writeOut(cmd, "")
		c.writeOut(cmd, "The shell will exit on any error.")
		ok = true

	case ":":
		ok = true

	case "cat":
		ok = true
		for _, subPath := range args {
			entry, ok := c.ctx.Get(path.Join(c.cwd, subPath))
			if !ok {
				return ok
			}

			switch entry := entry.(type) {
			case base.String:
				var value string
				value = entry.Get()
				c.writeOut(cmd, value)

			default:
				c.writeOut(cmd, "Name wasn't a type that you can cat")
				return false
			}
		}

	case "invoke":
		if len(args) == 0 {
			c.writeOut(cmd, "Not enough args for invoke")
			return
		}

		functionE, ok := c.ctx.Get(path.Join(c.cwd, args[0]))
		if !ok {
			c.writeOut(cmd, fmt.Sprintf("Couldn't find function named %s", args[0]))
			return ok
		}
		function := functionE.(base.Function)

		var input base.Entry
		if len(args) >= 2 && args[1] != "/dev/null" {
			input, ok = c.ctx.Get(path.Join(c.cwd, args[1]))
			if !ok {
				c.writeOut(cmd, fmt.Sprintf("Couldn't find input named %s", args[1]))
				return ok
			}
		}

		output := function.Invoke(c.ctx, input)

		if len(args) >= 3 && args[2] != "/dev/null" && output != nil {
			ok = c.ctx.Put(path.Join(c.cwd, args[2]), output)
			if !ok {
				c.writeOut(cmd, fmt.Sprintf("Couldn't write output to %s", output))
				return ok
			}
			c.writeOut(cmd, fmt.Sprintf("Wrote result to %s", args[2]))
		}

	case "cd":
		if len(args) == 1 {
			_, ok = c.ctx.GetFolder(path.Join(c.cwd, args[0]))
			if ok {
				c.cwd = path.Join(c.cwd, args[0])
			}
		} else if len(args) == 0 {
			c.cwd = ""
			ok = true
		}

	case "echo":
		text := strings.Join(args, " ")
		c.writeOut(cmd, text)
		ok = true

	case "ls":
		var folder base.Folder
		if len(args) == 1 {
			if folder, ok = c.ctx.GetFolder(path.Join(c.cwd, args[0])); !ok {
				return
			}
		} else if len(args) == 0 {
			if folder, ok = c.ctx.GetFolder(c.cwd); !ok {
				return
			}
		} else {
			return
		}

		// TODO: can't this fail?
		names := folder.Children()
		text := strings.Join(names, "\t")
		c.writeOut(cmd, text)

	case "ll":
		var folder base.Folder
		if len(args) == 1 {
			if folder, ok = c.ctx.GetFolder(path.Join(c.cwd, args[0])); !ok {
				return
			}
		} else if len(args) == 0 {
			if folder, ok = c.ctx.GetFolder(c.cwd); !ok {
				return
			}
		} else {
			return
		}

		// TODO: can't this fail?
		names := folder.Children()
		for _, name := range names {
			entry, _ := folder.Fetch(name)
			var extra string

			switch entry := entry.(type) {
			case base.String:
				extra = "string"
				if txt, err := json.Marshal(entry.Get()); err == nil {
					extra = fmt.Sprintf("string %v", string(txt))
				}

			case base.Folder:
				extra = fmt.Sprintf("folder (%d children)", len(entry.Children()))

			case base.Log:
				extra = "log"
			case base.Queue:
				extra = "queue"
			case base.Function:
				extra = "function"
			case base.List:
				extra = "list"
			case base.File:
				extra = "file"
			}

			c.writeOut(cmd, fmt.Sprintf("%s\t %s", name, extra))
			time.Sleep(10 * time.Millisecond)
		}
		ok = true

	default:
		text := fmt.Sprintf("No such command: %v", cmd)
		c.writeOut(cmd, text)
	}
	return
}

// Function that creates a new ray shell when invoked
func rayFunc(cctx base.Context, input base.Entry) (output base.Entry) {
	// Start a new command evaluator
	ctx := newRayCtx(cctx)

	switch input := input.(type) {
	case base.String:

		go func(ctx *rayCtx) {
			script := input.Get()
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
		line, ok := entry.(base.String)
		if !ok {
			log.Println("Ray failed to get string from output", entry)
			return
		}

		log.Printf("> %+v", line.Get())
	}
}
