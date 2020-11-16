package clients

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strconv"
)

type sampleClient struct {

}

// Client would have your business logic
type Client interface {
	HandleQueue1(msg *sqs.Message) error
	HandleQueue2(msg *sqs.Message) error
}

func NewSampleClient() Client {
	return &sampleClient{}

}

func (r *sampleClient) HandleQueue1(msg *sqs.Message) error {
	fmt.Printf("Queue [Queue1] Message ID [%v] Message [%v]\n", aws.StringValue(msg.MessageId), aws.StringValue(msg.Body))
	//time.Sleep(3 * time.Second)
	i, err := strconv.Atoi(*msg.Attributes["ApproximateReceiveCount"])
	if err != nil {
		return err
	}

	if i == 4 {
		fmt.Printf("done processing Message ID [%v]\n", aws.StringValue(msg.MessageId))
		return nil
	}
	fmt.Printf("Message ID [%v] is going to fail\n", aws.StringValue(msg.MessageId))
	return errors.New("Oops")
}

func (r *sampleClient) HandleQueue2(msg *sqs.Message) error {
	fmt.Printf("Queue [Queue2] Message ID [%v] Message [%v]\n", aws.StringValue(msg.MessageId), aws.StringValue(msg.Body))
	fmt.Printf("done processing Message ID [%v]\n", aws.StringValue(msg.MessageId))
	return nil
}