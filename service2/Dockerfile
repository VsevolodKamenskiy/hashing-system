FROM golang:1.24 AS builder
WORKDIR /service2
COPY . /service2
RUN GO111MODULE=auto CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o main cmd/app/main.go

FROM alpine:latest
WORKDIR /service2
COPY --from=builder /service2/main .

ENTRYPOINT ["./main"]
