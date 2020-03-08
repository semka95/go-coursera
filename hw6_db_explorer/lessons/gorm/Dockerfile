# Builder
FROM golang:1.14.0-alpine3.11 as builder

RUN apk update && apk upgrade && \
    apk --update add git

WORKDIR /app

COPY main.go go.mod go.sum ./

RUN go build -o engine main.go

# Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app 

WORKDIR /app 

EXPOSE 8080

COPY --from=builder /app/engine /app

ADD ./gorm_templates /app/gorm_templates/

CMD /app/engine
