FROM golang:1.22.2-alpine as builder
RUN apk update && apk add git gcc libc-dev sqlite sqlite-dev && rm -rf /var/cache/apk/*
ARG GITHUB_TOKEN
WORKDIR /go/src/github.com/qor5/admin/docs
COPY . .
RUN set -x && go get -d -v ./docsrc/examples/server/...
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./docsrc/examples/server/

FROM alpine
RUN apk update && apk add sqlite sqlite-dev && rm -rf /var/cache/apk/*
COPY --from=builder /app/entry  /bin/docsmain
CMD /bin/docsmain
