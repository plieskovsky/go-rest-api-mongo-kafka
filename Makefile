BINARY_NAME=user-service
UNAME := $(shell uname -m)

build:
	go build -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}

clean:
	rm ${BINARY_NAME}

test-unit:
	go clean -testcache && go test ./internal/... -v

test-e2e:
	go clean -testcache && go test ./e2e_test/... -v

run-in-docker: stop-in-docker
	# if on mac m1 chip this can fail on user-service image build. Run `export DOCKER_DEFAULT_PLATFORM=linux/amd64` to fix
	# https://github.com/confluentinc/confluent-kafka-go/issues/675#issuecomment-985329526
	docker compose up -d

stop-in-docker:
	docker compose down