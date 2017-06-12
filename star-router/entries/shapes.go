package entries

import (
	"github.com/stardustapp/core/inmem"
)

func newShapesEntry() *inmem.Folder {
	return inmem.NewFolderOf("shapes",
		functionShape,
		awsConfigShape,
		hueBridgeConfigShape,
		awsSqsReceiveMessageInputShape,
	).Freeze()
}

var functionShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("function",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("invoke", "Function"),
			inmem.NewFolderOf("input-shape",
				inmem.NewString("type", "Shape"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewFolderOf("output-shape",
				inmem.NewString("type", "Shape"),
				inmem.NewString("optional", "yes"),
			),
		),
	))

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

var awsSqsReceiveMessageInputShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("sqs-receive-message-input",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("queue-url", "String"),
			inmem.NewString("max-number-of-messages", "String"),
			inmem.NewString("wait-time-seconds", "String"),
			inmem.NewString("visibility-timeout", "String"),
		),
	))

var hueBridgeConfigShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("hue-bridge-config",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewFolderOf("bridge-id",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			inmem.NewFolderOf("lan-ip-address",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "no"),
			),
			inmem.NewFolderOf("username",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "no"),
			),
		),
	))
