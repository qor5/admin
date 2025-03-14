FROM golang:1.23.0-bullseye

ADD https://raw.githubusercontent.com/iBakuman/remote-debug/refs/heads/master/go/dlv.sh /dlv.sh

RUN apt-get update \
    && apt-get install -y --no-install-recommends inotify-tools psmisc \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && chmod +x /dlv.sh

ENV DELVE_PORT=40000

EXPOSE $DELVE_PORT

ENTRYPOINT ["/dlv.sh"]