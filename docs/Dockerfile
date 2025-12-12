FROM golang:1.24-alpine AS builder
RUN apk update && apk add gcc git libc-dev && rm -rf /var/cache/apk/*
WORKDIR /qor5
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./docs/docsrc/server/

FROM alpine:3.21
RUN apk --no-cache --update upgrade && \
    apk add --no-cache ca-certificates && \
    apk add --no-cache tzdata
COPY --from=builder /app/entry /bin/docsmain

RUN apk add --no-cache shadow && \
    useradd -U -u 1000 appuser && \
    chown -R 1000:1000 /bin/docsmain
USER 1000

CMD /bin/docsmain
