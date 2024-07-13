# syntax=docker/dockerfile:1

# docker image build takes around 210 seconds due to the kafka go client dependency...
FROM golang:1.22.5-alpine3.20 AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=1

# dependencies needed due to the kafka go client - https://github.com/confluentinc/confluent-kafka-go?tab=readme-ov-file#static-builds-on-linux
RUN apk add --no-progress --no-cache gcc musl-dev
WORKDIR /build
COPY . .
RUN go mod download

# ldflags and tags needed due to the kafka go client - https://github.com/confluentinc/confluent-kafka-go?tab=readme-ov-file#static-builds-on-linux
RUN go build --ldflags '-extldflags "-static"' -tags musl -o /build/user-service

FROM scratch
WORKDIR /app
COPY --from=builder /build/user-service .
ENTRYPOINT ["/app/user-service"]