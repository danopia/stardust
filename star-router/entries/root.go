package entries

import (
	"github.com/danopia/stardust/star-router/inmem"
)

// Presents the root name
func NewRootEntry() *inmem.Folder {
	root := inmem.NewFolder("/")
	root.Put("rom", newRomEntry())
	root.Put("boot", newBootEntry())
	root.Put("n", inmem.NewFolder("n"))
	root.Put("tmp", inmem.NewFolder("tmp"))
	return root
}

// Presents a read-only name with compiled-in children
func newRomEntry() *inmem.Folder {
	drivers := inmem.NewFolder("drv")
	drivers.Put("aws", getAwsDriver())
	drivers.Put("consul", getConsulDriver())
	drivers.Freeze()

	bin := inmem.NewFolder("bin")
	bin.Put("ray", &rayFunc{})
	bin.Put("ray-ssh", &raySshFunc{})
	bin.Freeze()

	rom := inmem.NewFolder("rom")
	rom.Put("drv", drivers)
	rom.Put("bin", bin)
	rom.Freeze()
	return rom
}

// Presents a read-only name with compiled-in children
func newBootEntry() *inmem.Folder {
	boot := inmem.NewFolder("boot")
	boot.Put("init", inmem.NewString("init", `
		invoke /rom/bin/ray-ssh /rom/bin/ray
  `))
	return boot
}
