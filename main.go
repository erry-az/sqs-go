package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	action  = flag.String("action", "send", "Action to perform (create,depth, purge, send (test), receive")
	queue   = flag.String("queue", "test", "Queue to work with")
	local   = flag.Bool("local", true, "Local (ElasticMQ) or Remote (AmazonSQS")
	passURL = flag.String("url", "", "iam id for amazon path")
)

var (
	queueName string
	svc       *sqs.Client
	fullURL   string
	attrib    string
)

func init() {
	flag.Parse()

	if *passURL == "" && !*local {
		log.Println("please use flag -iam-id for aws")
	}

	awsCfg := aws.Config{}

	if *local {
		awsCfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               "http://localhost:9324",
				HostnameImmutable: true,
				PartitionID:       "aws",
				SigningRegion:     "us-west-2",
			}, nil
		})
		fullURL = "http://localhost:9324/queue/" + *queue
	} else {
		awsCfgDef, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		awsCfg = awsCfgDef
		awsCfg.Region = "ap-southeast-1"
		fullURL = *passURL + "/" + *queue
	}

	log.Println(fullURL)

	svc = sqs.NewFromConfig(awsCfg)
	queueName = *queue
}

/*
 *
 *	Program entry point creates connection to SQS and Mongo then pool SQS for messages
 *
 */
func main() {
	switch *action {
	case "create":
		createSQSQueue()
		break
	case "depth":
		getSQSQueueDepth()
		break
	case "purge":
		purgeQueue()
		break
	case "send":
		sendMessage()
		break
	case "receive":
		receiveMessage()
		break
	default:
		fmt.Println("Unrecognized action - try again!")
	}
}

func createSQSQueue() {
	params := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName), // Required
	}
	resp, err := svc.CreateQueue(context.Background(), params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%+v", resp)
}

func getSQSQueueDepth() {
	attrib = "ApproximateNumberOfMessages"

	sendParams := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(fullURL), // Required
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeName(attrib), // Required
		},
	}

	resp, sendErr := svc.GetQueueAttributes(context.Background(), sendParams)
	if sendErr != nil {
		fmt.Println("Depth: " + sendErr.Error())
		return
	}

	fmt.Println(resp)
}

func sendMessage() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	//params := &sqs.SendMessageInput{
	//	MessageBody:            aws.String(fmt.Sprintf("testing %d", randNum)), // Required
	//	QueueUrl:               aws.String(fullURL),                            // Required
	//	MessageGroupId:         aws.String("666"),
	//	MessageDeduplicationId: aws.String(strconv.FormatUint(randNum, 10)),
	//}
	//resp, err := svc.SendMessage(context.Background(), params)

	resp, err := svc.SendMessageBatch(context.Background(), &sqs.SendMessageBatchInput{
		Entries: []types.SendMessageBatchRequestEntry{
			{
				Id:                     aws.String(fmt.Sprintf("testing-1-%d", r1.Uint64())),
				MessageBody:            aws.String(fmt.Sprintf("testing 1 %d", r1.Uint64())), // Required
				MessageGroupId:         aws.String("666"),
				MessageDeduplicationId: aws.String(strconv.FormatUint(r1.Uint64(), 10)),
			},
			{
				Id:                     aws.String(fmt.Sprintf("testing-2-%d", r1.Uint64())),
				MessageBody:            aws.String(fmt.Sprintf("testing 2 %d", r1.Uint64())), // Required
				MessageGroupId:         aws.String("666"),
				MessageDeduplicationId: aws.String(strconv.FormatUint(r1.Uint64(), 10)),
			},
		},
		QueueUrl: aws.String(fullURL),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resp)

}

func receiveMessage() {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(fullURL), // Required
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5,
		AttributeNames:      []types.QueueAttributeName{types.QueueAttributeNameAll},
		VisibilityTimeout:   120,
	}

	for {
		resp, err := svc.ReceiveMessage(context.Background(), params)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		messages := make([]types.DeleteMessageBatchRequestEntry, 0, len(resp.Messages))

		for _, message := range resp.Messages {
			fmt.Println(*message.Body)
			fmt.Println(*message.ReceiptHandle)
			fmt.Println(message.Attributes)

			messages = append(messages, types.DeleteMessageBatchRequestEntry{
				Id:            message.MessageId,
				ReceiptHandle: message.ReceiptHandle,
			})
		}

		if len(messages) > 0 {
			out, err := svc.DeleteMessageBatch(context.Background(), &sqs.DeleteMessageBatchInput{
				Entries:  messages,
				QueueUrl: aws.String(fullURL),
			})
			if err != nil {
				log.Fatalln(err)
			}

			log.Println("success", out.Successful)
			log.Println("err", out.Failed)
		}
	}
}

func purgeQueue() {
	params := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(fullURL), // Required
	}
	resp, err := svc.PurgeQueue(context.Background(), params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resp)
}
