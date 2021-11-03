FROM golang:1.17 as builder
RUN mkdir /.cache && chown nobody:nogroup /.cache && touch -t 202101010000.00 /.cache

ARG VERSION
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download -modcacherw

COPY --chown=nobody:nogroup ./main.go ./main.go
COPY --chown=nobody:nogroup ./producer ./producer
COPY --chown=nobody:nogroup ./consumer ./consumer
COPY --chown=nobody:nogroup ./kafkaques ./kafkaques

RUN mkdir bin
RUN go build -trimpath -ldflags='-linkmode external -w -extldflags "-static" -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=kakkoyun' -a -o ./bin/kafkaques .

FROM alpine:3.14

USER nobody

COPY --chown=0:0 --from=builder /app/bin/kafkaques /bin/kafkaques

CMD ["kafkaques"]
