DIR=$(PWD)
cd $DIR/../ && \
snippetgo -pkg=docsrc > ./docsrc/examples-generated.go
cd $DIR

go run ./build/main.go

function docsRestart() {
  echo "=================>"
  killall qor5docs
  go build -o /tmp/qor5docs ./docsmain/main.go && /tmp/qor5docs
}

export -f docsRestart
find . -name "*.go" | entr -r bash -c "docsRestart"

