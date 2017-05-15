package entries

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

// Directory containing the clone function
func getAwsDriver() *inmem.Folder {
	return inmem.NewFolderOf("aws",
		inmem.NewFunction("invoke", startAws),
		inmem.NewString("input-shape-path", "/rom/shapes/aws-config"),
	).Freeze()
}

// Function that creates a new AWS client when invoked
func startAws(input base.Entry) (output base.Entry) {
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

	return &awsRoot{creds, sess}
}

// Main Entity representing a consul client
// Presents k/v tree as a child
type awsRoot struct {
	creds *credentials.Credentials
	sess  *session.Session
}

var _ base.Folder = (*awsRoot)(nil)

func (e *awsRoot) Name() string {
	return "aws"
}
func (e *awsRoot) Children() []string {
	return []string{"sqs"}
}
func (e *awsRoot) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "sqs":
		return &awsSqs{
			root: e,
			sqs:  sqs.New(e.sess),
		}, true

	default:
		return
	}
}
func (e *awsRoot) Put(name string, entry base.Entry) (ok bool) {
	return false
}

// Folder of AWS SQS API calls
type awsSqs struct {
	root *awsRoot
	sqs  *sqs.SQS
}

var _ base.Folder = (*awsSqs)(nil)

func (e *awsSqs) Name() string {
	return "sqs"
}
func (e *awsSqs) Children() []string {
	return []string{"receive-message"}
}

func (e *awsSqs) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "receive-message":
		return inmem.NewFunction("receive-message", func(input base.Entry) (output base.Entry) {
			inputFolder := input.(base.Folder)
			queueUrl, _ := helpers.GetChildString(inputFolder, "queue-url")

			params := &sqs.ReceiveMessageInput{
				QueueUrl: aws.String(queueUrl),
			}

			resp, err := e.sqs.ReceiveMessage(params)
			if err == nil {
				if len(resp.Messages) == 0 {
					return nil
				}

				msg := resp.Messages[0]
				return inmem.NewFolderOf("received-message",
					inmem.NewString("msg-id", *msg.MessageId),
					inmem.NewString("body", *msg.Body),
					inmem.NewString("handle", *msg.ReceiptHandle),
				).Freeze()
			}
			return nil
		}), true

	default:
		return
	}
}

func (e *awsSqs) Put(name string, entry base.Entry) (ok bool) {
	return false
}
