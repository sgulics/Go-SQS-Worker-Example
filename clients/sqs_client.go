package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

func CreateSqsClient(awsConfigs ...*aws.Config) sqsiface.SQSAPI {
	awsSession := session.Must(session.NewSession())

	return sqs.New(awsSession, awsConfigs...)
}

