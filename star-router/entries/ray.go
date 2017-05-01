package entries

import (
	"github.com/danopia/stardust/star-router/base"
	"github.com/mattn/go-shellwords"
	"log"
	"strings"
	//"github.com/danopia/stardust/star-router/inmem"
)

// Function that creates a new consul client when invoked
type rayFunc struct{}

var _ base.Function = (*rayFunc)(nil)

func (e *rayFunc) Name() string {
	return "ray"
}

func (e *rayFunc) Invoke(input base.Entry) (output base.Entry) {
	str := input.(base.String)
	script, ok := str.Get()
	if !ok {
		panic("Ray couldn't get script contents")
	}

	handle := base.RootSpace.NewHandle()

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

		cmd := strings.ToLower(parts[0])

		switch cmd {

		case "cd":
			if len(parts) == 2 {
				ok = handle.Walk(parts[1])
			}

		case "echo":
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

		if !ok {
			log.Println("Ray failed at", line)
			return nil
		}

	}
	return nil
}
