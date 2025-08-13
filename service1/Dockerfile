FROM golang:1.24 AS builder
WORKDIR /service1
COPY . /service1
RUN go mod tidy && go mod download
RUN GO111MODULE=auto CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o main cmd/app/main.go

FROM alpine:latest
WORKDIR /service1
COPY --from=builder /service1/main .

ENTRYPOINT ["./main"]
