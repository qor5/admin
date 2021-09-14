FROM golang:alpine as builder
RUN apk update && apk add git gcc libc-dev sqlite sqlite-dev && rm -rf /var/cache/apk/*
ARG GITHUB_TOKEN
WORKDIR /qor5
COPY . .
RUN set -x && go get -d -v ./...
RUN GOOS=linux GOARCH=amd64 go build -o /app/entry ./example/

FROM alpine
COPY --from=builder /app/entry  /bin/example
CMD /bin/example
