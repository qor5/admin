FROM golang:1.23-alpine AS builder
RUN apk update && apk add gcc git libc-dev && rm -rf /var/cache/apk/*
WORKDIR /qor5
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./example/

FROM alpine:3.21
RUN apk --update upgrade && \
    apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/entry /bin/example

RUN apk add --no-cache shadow && \
    useradd -U -u 1000 appuser && \
    chown -R 1000:1000 /bin/example
USER 1000

CMD /bin/example
