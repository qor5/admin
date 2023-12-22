FROM golang:1.21.5-alpine as builder
RUN apk update && apk add git gcc libc-dev sqlite sqlite-dev && rm -rf /var/cache/apk/*
WORKDIR /qor5
COPY . .
RUN set -x && go get -d -v ./...
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./example/cmd/publisher/

FROM alpine:3.16
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
COPY --from=builder /app/entry  /bin/publisher
CMD /bin/publisher
