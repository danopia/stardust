package helpers

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/inmem"
)

func GetChildString(folder base.Folder, name string) (value string, ok bool) {
	entry, ok := folder.Fetch(name)
	if !ok {
		log.Println("missed lookup for", name, "in", folder.Name())
		return "", false
	}
	value, ok = entry.(base.String).Get()
	return
}
