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
	"github.com/danlju/tulip-worker/internal/executor"
	"github.com/danlju/tulip-worker/internal/queue"
	"github.com/danlju/tulip-worker/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	queueURL := envOrDefault("QUEUE_URL", "http://localhost:9324/tulip-queue")
	apiURL := envOrDefault("API_URL", "http://localhost:8080")
	sqsEndpoint := os.Getenv("SQS_ENDPOINT")

	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(envOrDefault("AWS_REGION", "us-east-1")),
	}
	if sqsEndpoint != "" {
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: sqsEndpoint}, nil
			}),
		))
	}
	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		log.Fatal(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	queueClient := queue.NewClient(sqsClient, queueURL)

	apiClient := api.NewClient(apiURL)

	exec := executor.NewExecutor(envOrDefault("BUILD_WORKSPACE", "/tmp/builds"))

	handler := worker.NewBuildHandler(apiClient, exec)

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

			msgs, err := queueClient.Receive(ctx, pool.Available())
			if err != nil {
				log.Println("receive error:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, msg := range msgs {
				m := msg
				pool.Submit(worker.Job{
					Msg: m,
					Ack: func(ackCtx context.Context) error { return queueClient.Delete(ackCtx, m.ReceiptHandle) },
					Heartbeat: func(heartbeatCtx context.Context) error {
						return queueClient.ExtendVisibility(heartbeatCtx, m.ReceiptHandle, 120)
					},
				})
			}
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
