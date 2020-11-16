# Go SQS Worker Example

Testing out how to start up multiple SQS workers that also support graceful shutdown, and exponential backoff.

The SQS worker code is borrowed from https://github.com/h2ik/go-sqs-poller 

## Localstack

This example uses localstack

Docker Hub
https://hub.docker.com/r/localstack/localstack

Github:
https://github.com/localstack/localstack

` docker run --rm -p 4566:4566 --name localstack localstack/localstack`

https://gist.github.com/lobster1234/57e803ebca47c3c263a9d53ccd1f1783

### Create Queues
`aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name test_queue_1`
`aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name test_queue_2`

### List Queues
`aws --endpoint-url=http://localhost:4566 sqs list-queues`

### Send a message to queue
`aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4576/123456789012/test_queue --message-body 'Test Message!'`

### Change Max Receive Count
`aws sqs set-queue-attributes --endpoint-url=http://localhost:4566 --queue-url="http://localhost:4566/000000000000/test_queue_1" --attributes 'RedrivePolicy.maxReceiveCount=10'`

### Get queue properties
`aws sqs get-queue-attributes --endpoint-url=http://localhost:4566 --queue-url="http://localhost:4566/000000000000/test_queue_1"`

## Notes

### Graceful Exit

https://guzalexander.com/2017/05/31/gracefully-exit-server-in-go.html

https://medium.com/@matryer/make-ctrl-c-cancel-the-context-context-bd006a8ad6ff
https://callistaenterprise.se/blogg/teknik/2019/10/05/go-worker-cancellation/

### SQS Workers

https://github.com/h2ik/go-sqs-poller

https://questhenkart.medium.com/sqs-consumer-design-achieving-high-scalability-while-managing-concurrency-in-go-d5a8504ea754

https://medium.com/@marcioghiraldelli/elegant-use-of-golang-channels-with-aws-sqs-dad20cd59f34

### SQS Publisher

https://gist.github.com/jsokel/22dc8cd07ca9ee1c8c4af46f8db78ca8