package entries

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/inmem"
)

// Directory containing the clone function
func getAwsDriver() *inmem.Folder {
	return inmem.NewFolderOf("aws", &awsClone{}).Freeze()
}

// Function that creates a new AWS client when invoked
type awsClone struct{}

var _ base.Function = (*awsClone)(nil)

func (e *awsClone) Name() string {
	return "clone"
}

func (e *awsClone) Invoke(input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)

	accessKeyEntry, ok := inputFolder.Fetch("access_key_id")
	if !ok {
		log.Println("no aws creds")
		return nil
	}
	accessKey, ok := accessKeyEntry.(base.String).Get()

	secretKeyEntry, ok := inputFolder.Fetch("secret_access_key")
	if !ok {
		log.Println("no aws secret creds")
		return nil
	}
	secretKey, ok := secretKeyEntry.(base.String).Get()

	regionEntry, ok := inputFolder.Fetch("region")
	if !ok {
		log.Println("no aws region")
		return nil
	}
	region, ok := regionEntry.(base.String).Get()

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")

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
		return &awsSqsReceiveMessage{e.sqs}, true

	default:
		return
	}
}

func (e *awsSqs) Put(name string, entry base.Entry) (ok bool) {
	return false
}

// Function that creates a new AWS client when invoked
type awsSqsReceiveMessage struct {
	svc *sqs.SQS
}

var _ base.Function = (*awsSqsReceiveMessage)(nil)

func (e *awsSqsReceiveMessage) Name() string {
	return "receive-message"
}

func (e *awsSqsReceiveMessage) Invoke(input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)

	queueUrlEntry, ok := inputFolder.Fetch("queue-url")
	if !ok {
		log.Println("no queue url")
		return nil
	}
	queueUrl, ok := queueUrlEntry.(base.String).Get()

	params := &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(queueUrl),
	}

	resp, err := e.svc.ReceiveMessage(params)
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
}
