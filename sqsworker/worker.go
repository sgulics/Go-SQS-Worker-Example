package sqsworker

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
	"math"
	"strconv"
	"sync"
)

// HandlerFunc is used to define the Handler that is run on for each message
type HandlerFunc func(msg *sqs.Message) error

// HandleMessage wraps a function for handling sqs messages
func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

// Handler interface
type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

// InvalidEventError struct
type InvalidEventError struct {
	event string
	msg   string
}

func (e InvalidEventError) Error() string {
	return fmt.Sprintf("[Invalid Event: %s] %s", e.event, e.msg)
}

func getQueueURL(client QueueAPI, queueName string) (queueURL string) {
	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName), // Required
	}
	response, err := client.GetQueueUrl(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	queueURL = aws.StringValue(response.QueueUrl)

	return
}

// NewInvalidEventError creates InvalidEventError struct
func NewInvalidEventError(event, msg string) InvalidEventError {
	return InvalidEventError{event: event, msg: msg}
}

// QueueAPI interface is the minimum interface required from a queue implementation
type QueueAPI interface {
	DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	ChangeMessageVisibility(*sqs.ChangeMessageVisibilityInput) (*sqs.ChangeMessageVisibilityOutput, error)
}

// Worker struct
type Worker struct {
	Config    *Config
	Logger *logrus.Logger
	SqsClient QueueAPI
}

// Config struct
type Config struct {
	MaxNumberOfMessage int64
	QueueName          string
	QueueURL           string
	WaitTimeSecond     int64
}

// New sets up a new Worker
func New(client QueueAPI, config *Config) *Worker {
	//config.populateDefaultValues()
	config.QueueURL = getQueueURL(client, config.QueueName)
	logger := logrus.New()
	logger.Debug("QUEUE YRK")
	return &Worker{
		Config:    config,
		Logger: logger,
		SqsClient: client,
	}
}

// taken from DelayedJob
func retryIn(attempts int) int64 {
	return int64(5 + math.Pow(float64(attempts), 4))
}

// Start starts the polling and will continue polling till the application is forcibly stopped
func (worker *Worker) Start(ctx context.Context, h Handler) {
	fmt.Println("Starting poller for queue [%s]", worker.Config.QueueName)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker: Stopping polling for queue [%v] because a context kill signal was sent\n", worker.Config.QueueName)
			return
		default:
			fmt.Printf("worker: Start Polling for queue [%v]\n", worker.Config.QueueName)

			params := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(worker.Config.QueueURL), // Required
				MaxNumberOfMessages: aws.Int64(worker.Config.MaxNumberOfMessage),
				AttributeNames: []*string{
					aws.String("All"), // Required
				},
				WaitTimeSeconds: aws.Int64(worker.Config.WaitTimeSecond),
			}

			resp, err := worker.SqsClient.ReceiveMessage(params)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(resp.Messages) > 0 {
				worker.run(h, resp.Messages)
			}
		}
	}
}

// poll launches goroutine per received message and wait for all message to be processed
func (worker *Worker) run(handler Handler, messages []*sqs.Message) {
	numMessages := len(messages)
	fmt.Println(fmt.Sprintf("worker: Received %d messages for queue [%v]", numMessages, worker.Config.QueueName))

	var wg sync.WaitGroup
	wg.Add(numMessages)
	for i := range messages {
		go func(m *sqs.Message) {
			// launch goroutine
			defer wg.Done()
			if err := worker.handleMessage(m, handler); err != nil {
				fmt.Println(fmt.Sprintf("fatal error processing message [%v] in queue [%v]", m.MessageId, worker.Config.QueueName))
			}
		}(messages[i])
	}

	wg.Wait()
}

func (worker *Worker) handleMessage(m *sqs.Message, h Handler) error {
	var err error
	err = h.HandleMessage(m)
	if err != nil {
		fmt.Printf("error handling message [%v] on queue [%s]\n", *m.MessageId, worker.Config.QueueName)
		return worker.changeMessageVisibility(m)
	}

	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(worker.Config.QueueURL), // Required
		ReceiptHandle: m.ReceiptHandle,                    // Required
	}
	_, err = worker.SqsClient.DeleteMessage(params)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("worker: deleted message from queue: [%s]: %s", worker.Config.QueueName, aws.StringValue(m.ReceiptHandle)))

	return nil
}

func (worker *Worker) changeMessageVisibility(m *sqs.Message) error {
	i, err := strconv.Atoi(*m.Attributes["ApproximateReceiveCount"])
	if err != nil {
		return err
	}

	visibilityTimeout := retryIn(i)
	fmt.Println(fmt.Sprintf("Message ID [%v] Receive count [%v] new visibiliyTimeout [%v]", *m.MessageId, i, visibilityTimeout))
	_, err = worker.SqsClient.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(worker.Config.QueueURL),
		ReceiptHandle:     m.ReceiptHandle,
		VisibilityTimeout: aws.Int64(visibilityTimeout),
	})

	return err
}