FROM golang:1.23-alpine as builder
RUN apk update && apk add gcc git libc-dev sqlite sqlite-dev && rm -rf /var/cache/apk/*
WORKDIR /qor5
COPY . .
RUN set -x && go get -d -v ./...
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./example/cmd/publisher/

FROM alpine:3.21.3
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
COPY --from=builder /app/entry  /bin/publisher
CMD /bin/publisher
