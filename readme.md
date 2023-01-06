# SQS golang

## command

### use AWS SQS 

```shell
go run main.go -action send -queue test-dev.fifo -local=false -url https://sqs.AWS_ZONE.amazonaws.com/AWS_ID
```

### use local elasticMQ
```shell
go run main.go -action send -queue test-dev.fifo -local=true
```