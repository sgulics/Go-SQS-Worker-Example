package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/sgulics/sqsworker/clients"
	"github.com/sgulics/sqsworker/sqsworker"
	"os"
	"os/signal"
	"sync"
)

func main() {
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials("not", "empty", ""),
		DisableSSL: aws.Bool(true),
		Region:      aws.String("us-east-1"),
		Endpoint: aws.String("http://localhost:4566"),

	}

	sqsClient := clients.CreateSqsClient(awsConfig)
	sampleClient := clients.NewSampleClient()
	ctx := context.Background()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt)

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(ctx)

	go startWorker(ctx, wg, sqsClient, sampleClient.HandleQueue1, &sqsworker.Config{
		QueueName:          "test_queue_1",
		MaxNumberOfMessage: 3,
		WaitTimeSecond:     15,
	})
	wg.Add(1)

	go startWorker(ctx, wg, sqsClient, sampleClient.HandleQueue2, &sqsworker.Config{
		QueueName:          "test_queue_2",
		MaxNumberOfMessage: 3,
		WaitTimeSecond:     5,
	})
	wg.Add(1)

	// wait for shutdown
	<-termChan
	fmt.Println("-----Shutdown Received-------")
	cancel()
	wg.Wait()
	fmt.Println("All done, shutting down")

}

func startWorker(ctx context.Context, wg *sync.WaitGroup, sqsClient sqsworker.QueueAPI, handlerFunc sqsworker.HandlerFunc, config *sqsworker.Config) {
	defer wg.Done()
	eventWorker := sqsworker.New(sqsClient, config)
	eventWorker.Start(ctx, handlerFunc)
	fmt.Println("Done in worker go routine for queue [%s]", config.QueueName)

}
