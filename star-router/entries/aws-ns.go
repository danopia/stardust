package entries

import (
	"log"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

// Persistent Namesystem using DynamoDB as a storage driver
// Kept seperate from the AWS API client, as this is very special-cased
// Assumes literally everything except credentials
// A table can contain multiple NSes, the default is named "root"
// Though the implementation only lets you use "root" atm (TODO)

// Table info:
//   Named: Stardust_NameSystem
//   Provisioning: I'm starting low @ 1 write, 2 read
//   Hash/partition: ParentId (string)
//   Range/sort: Name (string)

// Directory containing the clone function
func getAwsNsDriver() *inmem.Folder {
	return inmem.NewFolderOf("aws-ns",
		inmem.NewFunction("invoke", startAwsNs),
		inmem.NewLink("input-shape", "/rom/shapes/aws-config"),
	).Freeze()
}

// Function that creates a new AWS client when invoked
func startAwsNs(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	accessKey, _ := helpers.GetChildString(inputFolder, "access_key_id")
	secretKey, _ := helpers.GetChildString(inputFolder, "secret_access_key")
	sessionToken, _ := helpers.GetChildString(inputFolder, "session_token")
	region, _ := helpers.GetChildString(inputFolder, "region")

	creds := credentials.NewStaticCredentials(accessKey, secretKey, sessionToken)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(region),
			Credentials: creds,
		},
	}))

	dynamo := dynamodb.New(sess)

	ns := &awsNs{
		creds: creds,
		sess:  sess,
		svc:   dynamo,
		root:  "root",
	}

	// TODO: there is no actual entry for the root entry

	return &awsNsFolder{
		ns: ns,
		id: ns.root,
	}
}

// Main Entity representing a consul client
type awsNs struct {
	creds *credentials.Credentials
	sess  *session.Session
	svc   *dynamodb.DynamoDB
	root  string
}

// Persists as a Folder from an awsNs instance
// Presents as a dynamic name tree
type awsNsFolder struct {
	ns *awsNs
	id string // aka path
}

var _ base.Folder = (*awsNsFolder)(nil)

func (e *awsNsFolder) Name() string {
	return path.Base(e.id)
}

func (e *awsNsFolder) Children() []string {
	params := &dynamodb.QueryInput{
		TableName: aws.String("Stardust_NameSystem"),

		// Match children of this entry
		KeyConditionExpression: aws.String("ParentId = :pid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pid": {S: aws.String(e.id)},
		},

		// Only get their names
		ProjectionExpression: aws.String("#n"),
		ExpressionAttributeNames: map[string]*string{
			"#n": aws.String("Name"),
		},
	}

	// TODO: pagination, just like everywhere else in SD lol.
	result, err := e.ns.svc.Query(params)
	if err != nil {
		log.Println("dynamodb children error:", err)
		return nil
	}

	kids := make([]string, *result.Count)
	for i, d := range result.Items {
		kids[i] = *d["Name"].S
	}
	return kids
}

func (e *awsNsFolder) Fetch(name string) (entry base.Entry, ok bool) {
	params := &dynamodb.GetItemInput{
		TableName: aws.String("Stardust_NameSystem"),

		// Get the child directly
		Key: map[string]*dynamodb.AttributeValue{
			"ParentId": {S: aws.String(e.id)},
			"Name":     {S: aws.String(name)},
		},
	}

	result, err := e.ns.svc.GetItem(params)
	if err != nil {
		log.Println("dynamodb fetch error:", err)
		return nil, false
	}

	if result.Item == nil {
		// log.Println("dynamodb ns key", e.id, name, "didn't exist")
		return nil, false
	}

	typeStr := *result.Item["Type"].S
	switch typeStr {

	case "String":
		var value string
		if att, ok := result.Item["Value"]; ok {
			value = *att.S
		}
		return inmem.NewString(name, value), true

	case "File":
		var data []byte
		if att, ok := result.Item["Value"]; ok {
			data = att.B
		}
		return inmem.NewFile(name, data), true

	case "Folder":
		return &awsNsFolder{
			ns: e.ns,
			id: path.Join(e.id, name),
		}, true

	default:
		log.Println("dynamodb ns key", e.id, name, "has unknown type", typeStr)
		return nil, false
	}
}

func (e *awsNsFolder) Put(name string, entry base.Entry) (ok bool) {
	if entry == nil {
		// oh boy this is a delete now
		target, ok := e.Fetch(name)
		if !ok { // already doesn't exist i guess
			return true
		}

		if folder, ok := target.(base.Folder); ok {
			// let's get recursive
			for _, child := range folder.Children() {
				folder.Put(child, nil)
			}
		}

		// finally delete actual target
		params := &dynamodb.DeleteItemInput{
			TableName: aws.String("Stardust_NameSystem"),

			// Get the child directly
			Key: map[string]*dynamodb.AttributeValue{
				"ParentId": {S: aws.String(e.id)},
				"Name":     {S: aws.String(name)},
			},
		}

		_, err := e.ns.svc.DeleteItem(params)
		if err != nil {
			log.Println("dynamodb put/delete error:", err)
			return false
		}
		return true
	}

	// TODO: check if it's already a name of same type??
	params := &dynamodb.PutItemInput{
		TableName: aws.String("Stardust_NameSystem"),

		// Get the child directly
		Item: map[string]*dynamodb.AttributeValue{
			"ParentId": {S: aws.String(e.id)},
			"Name":     {S: aws.String(name)},
		},
	}

	switch entry := entry.(type) {

	case base.Folder:
		// easy enough
		params.Item["Type"] = &dynamodb.AttributeValue{S: aws.String("Folder")}

	case base.String:
		params.Item["Type"] = &dynamodb.AttributeValue{S: aws.String("String")}
		if value := entry.Get(); len(value) > 0 {
			params.Item["Value"] = &dynamodb.AttributeValue{S: aws.String(value)}
		}

	case base.File:
		// Get the byte buffer first
		size := entry.GetSize()
		data := entry.Read(0, int(size))

		params.Item["Type"] = &dynamodb.AttributeValue{S: aws.String("File")}
		if len(data) > 0 {
			params.Item["Value"] = &dynamodb.AttributeValue{B: data}
		}

	default:
		return false
	}

	_, err := e.ns.svc.PutItem(params)
	if err != nil {
		log.Println("dynamodb put error:", err)
		return false
	}
	return true
}
