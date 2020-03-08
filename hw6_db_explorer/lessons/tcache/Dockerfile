# Builder
FROM golang:1.14.0-alpine3.11 as builder

RUN apk update && apk upgrade && \
    apk --update add git

WORKDIR /app

COPY main.go cache.go posts.go go.mod go.sum ./

RUN go build -o engine *.go

# Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app 

WORKDIR /app 

COPY --from=builder /app/engine /app

CMD /app/engine
