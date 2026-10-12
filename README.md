# Tulip worker

The worker polls the Tulip SQS queue, runs each build, and reports the result to the API. By default it expects the API and ElasticMQ to be running on the same host.

| Variable | Default | Purpose |
| --- | --- | --- |
| `API_URL` | `http://localhost:8080` | Tulip API base URL |
| `SQS_ENDPOINT` | AWS SDK default | Set to `http://localhost:9324` for local ElasticMQ |
| `QUEUE_URL` | `http://localhost:9324/tulip-queue` | Full SQS queue URL |
| `AWS_REGION` | `us-east-1` | AWS region |
| `BUILD_WORKSPACE` | `/tmp/builds` | Directory used for checked out repositories |

For the API, `SQS_ENDPOINT` configures the SQS endpoint, `SQS_QUEUE_URL` sets the full queue URL, and `SQS_STATIC_CREDENTIALS` defaults to `true` for local queue emulators. Set `SQS_STATIC_CREDENTIALS=false` to use the AWS SDK credential chain, such as an EC2 task or instance role.

The local ElasticMQ configuration moves messages that repeatedly fail JSON decoding to `tulip-dead-letter`. The worker leaves a build message on the queue when it cannot update the API; it acknowledges the message after a successful completion or failure update.
