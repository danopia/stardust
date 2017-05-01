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
  return root
}

// Presents a read-only name with compiled-in children
func newRomEntry() *inmem.Folder {
  drivers := inmem.NewFolder("drv")
  drivers.Put("consul", getConsulDriver())
  drivers.Freeze()

  bin := inmem.NewFolder("bin")
  bin.Put("ray", &rayFunc{})
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
    cd /
    ls
    cd rom
    ls
    echo "Hello World!"
  `))
  return boot
}
