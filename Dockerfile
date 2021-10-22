# this image is what docker.io/golang:1.16.7-alpine3.14 on August 12 2021
FROM docker.io/golang@sha256:7e31a85c5b182e446c9e0e6fba57c522902f281a6a5a6cbd25afa17ac48a6b85 as builder
RUN mkdir /.cache && chown nobody:nogroup /.cache && touch -t 202101010000.00 /.cache

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download -modcacherw

COPY --chown=nobody:nogroup ./main.go ./main.go
COPY --chown=nobody:nogroup ./producer ./producer
COPY --chown=nobody:nogroup ./consumer ./consumer
COPY --chown=nobody:nogroup ./kafkaques ./kafkaques

RUN go build -ldflags="-X main.version=$(VERSION)" -o kafkaques .

# this image is what docker.io/alpine:3.14.1 on August 13 2021
FROM docker.io/alpine@sha256:be9bdc0ef8e96dbc428dc189b31e2e3b05523d96d12ed627c37aa2936653258c

USER nobody

COPY --chown=0:0 --from=builder /app/kafkaques /kafkaques

CMD ["/kafkaques"]
