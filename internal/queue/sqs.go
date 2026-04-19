package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/danlju/tulip-worker/internal/model"
)

type Message struct {
	ReceiptHandle *string
	Payload       model.BuildRequest
}

type Client struct {
	sqs      *sqs.Client
	queueURL string
}

func NewClient(sqsClient *sqs.Client, queueURL string) *Client {
	return &Client{sqs: sqsClient, queueURL: queueURL}
}

func (c *Client) Receive(ctx context.Context, max int) ([]Message, error) {
	out, err := c.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: int32(max),
		WaitTimeSeconds:     10,
	})
	if err != nil {
		return nil, err
	}

	var messages []Message

	for _, m := range out.Messages {
		var payload model.BuildRequest
		if err := json.Unmarshal([]byte(*m.Body), &payload); err != nil {
			log.Println("invalid message:", err)
			continue
		}

		messages = append(messages, Message{
			ReceiptHandle: m.ReceiptHandle,
			Payload:       payload,
		})
	}

	return messages, nil
}

func (c *Client) Delete(ctx context.Context, receiptHandle *string) {
	_, err := c.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: receiptHandle,
	})

	if err != nil {
		log.Println("failed to delete message:", err)
	}
}
