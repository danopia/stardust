package entries

import (
	"github.com/danopia/stardust/star-router/inmem"
)

func newShapesEntry() *inmem.Folder {
	return inmem.NewFolderOf("shapes",
		awsConfigShape,
	).Freeze()
}

var awsConfigShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("aws-config",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("access_key_id", "String"),
			inmem.NewString("secret_access_key", "String"),
			inmem.NewFolderOf("session_token",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewString("region", "String"),
		),
	))
