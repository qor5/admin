FROM golang:1.22.5-alpine as builder
RUN apk update && apk add git gcc libc-dev && rm -rf /var/cache/apk/*
ARG GITHUB_TOKEN
WORKDIR /qor5
COPY . .
RUN set -x && go get -d -v ./docs/docsrc/...
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./docs/docsrc/server/

FROM alpine:3.18.9
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
COPY --from=builder /app/entry /bin/docsmain
CMD /bin/docsmain
