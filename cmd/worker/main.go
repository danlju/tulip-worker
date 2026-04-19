package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/danlju/tulip-worker/internal/api"
	"github.com/danlju/tulip-worker/internal/queue"
	"github.com/danlju/tulip-worker/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	queueURL := "/tulip-queue" //os.Getenv("QUEUE_URL")
	apiURL := os.Getenv("API_URL")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL: "http://localhost:9324",
					}, nil
				},
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	queueClient := queue.NewClient(sqsClient, queueURL)

	apiClient := api.NewClient(apiURL)
	handler := worker.NewBuildHandler(apiClient)

	pool := worker.NewPool(5, handler)

	log.Println("tulip-worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down...")
			pool.Shutdown()
			return

		default:
			if pool.Available() == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			msgs, err := queueClient.Receive(ctx, 5)
			if err != nil {
				log.Println("receive error:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, msg := range msgs {
				m := msg
				pool.Submit(worker.Job{Msg: m})

				go queueClient.Delete(context.Background(), m.ReceiptHandle)
			}
		}
	}
}
