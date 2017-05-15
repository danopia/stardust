package entries

import (
	"github.com/danopia/stardust/star-router/inmem"
)

func newShapesEntry() *inmem.Folder {
	return inmem.NewFolderOf("shapes",
		getAwsConfigShape(),
	).Freeze()
}

func getAwsConfigShape() *inmem.Shape {
	cfg := inmem.NewFolderOf("aws-config",
    inmem.NewString("type", "Folder"),
    inmem.NewFolderOf("props",
      inmem.NewString("access_key_id", "String"),
      inmem.NewString("secret_access_key", "String"),
      inmem.NewString("session_token", "String"),
      inmem.NewString("region", "String"),
    ),
  )
	return inmem.NewShape(cfg)
}
