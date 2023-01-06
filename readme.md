# SQS golang

don't forget to add aws token on `~/.aws/config`
```shell
[default]
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=ap-southeast-1
```

## command

available `action` :
- create (create new queue)
- depth (check depth)
- purge (remove queue)
- send (send message to queue)
- receive (consume message)

### use AWS SQS 

```shell
go run main.go -action send -queue test-dev.fifo -local=false -url https://sqs.AWS_ZONE.amazonaws.com/AWS_ID
```

### use local elasticMQ
```shell
go run main.go -action send -queue test-dev.fifo -local=true
```