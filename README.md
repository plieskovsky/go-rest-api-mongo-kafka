# User service

Service that provides user management via REST API. The service publishes events to a kafka topic upon User creation, update or deletion.

## REST API Documentation

The Users REST API Documentation is [here](docs/users_rest_api_docs.md). The service also exposes a `/metrics` and `/health` endpoint
to monitor its behaviour and state.

## Service configuration

Service can be configured via environment variables. If not provided, defaults are used.

| Variable Name                  | Description                                                  | Type     | Default                                  |
|--------------------------------|--------------------------------------------------------------|----------|------------------------------------------|
| HTTP_PORT                      | port on which the service HTTP server will listen            | int      | 8080                                     |
| MONGO_URL                      | url of the Mongo server                                      | string   | mongodb://user:password@localhost:27017/ |
| MONGO_DB_NAME                  | Name of the DB to use                                        | string   | demo                                     |
| MONGO_OPERATION_TIMEOUT        | timeout of the mongo DB calls                                | duration | 3s                                       |
| KAFKA_SERVER                   | url of the kafka server                                      | string   | localhost:9092                           |
| EVENTS_TOPIC_NAME              | name of the kafka topic name to which to publish user events | string   | UserEvents                               |
| HTTP_GRACEFUL_SHUTDOWN_PERIOD  | duration of the graceful HTTP server shutdown                | duration | 5s                                       |
| MONGO_GRACEFUL_SHUTDOWN_PERIOD | duration of the graceful Mongo connection shutdown           | duration | 5s                                       |
| KAFKA_GRACEFUL_SHUTDOWN_PERIOD | duration of the graceful Kafka producer shutdown             | duration | 5s                                       |


## Homework Notes/Improvements:
- wrote some unit and some e2e tests to showcase how I would write and structure them - did not cover all the functionality to not spend too much time on them
- did not write documentation at each function/variable etc. as their names are self-explanatory. Added docs only in case it's needed or helpful.
- mongo and kafka connections are created in a way, so they can be reused if needed by other mongo collections/kafka topic producers
- chose kafka as async communication because i wanted to try running it locally in a dockerized env, but it came with couple challenges...
  - no straightaway health check support in go client lib 
  - etc/hosts needs `kafka 127.0.0.1` entry so the publisher can connect to it when running the app locally (not via docker compose though)
  - docker image build takes around 210 seconds due to the special requirements needed by the kafka go client
  - next time I would go with NATS for the async communication which should be easier to setup and manage..

## Development

### Known issues

- When running the user-service locally with `go run .` it failed to connect to kafka with error `Failed to resolve 'kafka:9092': nodename nor servname provided, or not known`.
  This can be fixed by adding `127.0.0.1 kafka` to the local etc/hosts file.
- When building the docker image on apple M1 chip it fails with errors related to gcc and kafka lib. To fix it set the default docker platform to linux `export DOCKER_DEFAULT_PLATFORM=linux/amd64` -
  https://github.com/confluentinc/confluent-kafka-go/issues/675#issuecomment-985329526

### Tests

- `make test-unit` to run unit tests
- `make test-e2e` to run e2e tests (spin up the local services with `make run-in-docker` before running e2e tests)

### Running the service locally

- `make run-in-docker` spins up all the needed services + the user-service by using the docker compose.

To run the service locally outside of docker for a faster development you can:
1. comment out the `services:user-service` section in the docker-compose.yml
2. run `make run-in-docker` to spin up the cluster without the user-service 
3. run `go run .` to start the service

### Cheat sheet
#### Kafka

Listen to all the messages on the UserEvents topic
```bash
docker-compose exec kafka kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic UserEvents --from-beginning
```

#### REST API

Create multiple users in DB
```bash 
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"first_name":"john","last_name":"wick","nickname":"johnnywicky","password":"securepwd","email":"johnnywicky@gmail.com","country":"UK"}' \
  localhost:8080/v1/users -v
  
  curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"first_name":"andrey","last_name":"anakonda","nickname":"anakonda","password":"anakondapwd","email":"anakonda@gmail.com","country":"AK"}' \
  localhost:8080/v1/users -v
  
    curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"first_name":"cecilia","last_name":"ceckata","nickname":"ceckana","password":"cecikypwd","email":"cuza@gmail.com","country":"CZ"}' \
  localhost:8080/v1/users -v
  
      curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"first_name":"dupond","last_name":"dereke","nickname":"duduko","password":"didipwd","email":"drak@gmail.com","country":"CZ"}' \
  localhost:8080/v1/users -v
```

Update user in DB
```bash
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"id":"b79d4ce5-40f9-11ef-a3eb-0242ac170004", "first_name":"Johnn","last_name":"Wickk","nickname":"johnnywickyy","password":"securepwdd","email":"johnnywicky@gmail.comm","country":"UKK"}' \
  localhost:8080/v1/users/b79d4ce5-40f9-11ef-a3eb-0242ac170004 -v
```

Get user from DB
```bash
curl  --request GET localhost:8080/v1/users/b79d4ce5-40f9-11ef-a3eb-0242ac170004 -v
```

Get users from DB
```bash
curl  --request GET -v "localhost:8080/v1/users?pageSize=2&page=1&sortBy=first_name.asc&country=UK"
```

Delete user in DB
```bash
curl  --request DELETE localhost:8080/v1/users/b79d4ce5-40f9-11ef-a3eb-0242ac170004 -v
```