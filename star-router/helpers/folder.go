package helpers

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/inmem"
)

func GetChildString(folder base.Folder, name string) (value string, ok bool) {
	entry, ok := folder.Fetch(name)
	if !ok {
		if name != "optional" {
			log.Println("missed lookup for", name, "in", folder.Name())
		}
		return "", false
	}

	str, ok := entry.(base.String)
	if !ok {
		log.Println("wanted string, got something else from", name)
		return "", false
	}

	return str.Get(), true
}
