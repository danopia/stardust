package entries

import (
	"github.com/danopia/stardust/star-router/inmem"
)

// Presents the root name
func NewRootEntry() *inmem.Folder {
	return inmem.NewFolderOf("/",
		newRomEntry(),
		newBootEntry(),
		inmem.NewFolder("n"),
		inmem.NewFolder("tmp"),
	)
}

// Presents a read-only name with compiled-in children
func newRomEntry() *inmem.Folder {
	drivers := inmem.NewFolderOf("drv",
		getOsDriver(),
		getAwsDriver(),
		getAwsNsDriver(),
		getConsulDriver(),
		getHueDriver(),
	).Freeze()

	bin := inmem.NewFolderOf("bin",
		getRayDriver(),
		getRaySshDriver(),
		getHttpdDriver(),
		getInitDriver(),
	).Freeze()

	return inmem.NewFolderOf("rom",
		drivers,
		bin,
		newShapesEntry(),
	).Freeze()
}

// Presents a read-only name with compiled-in children
func newBootEntry() *inmem.Folder {
	boot := inmem.NewFolder("boot")
	boot.Put("init", inmem.NewString("init", `
		invoke /rom/bin/ray-ssh /rom/bin/ray
  `))
	return boot
}
