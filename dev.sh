function exampleRestart() {
  echo "=================>"
  killall qor5example
  go build -o /tmp/qor5example example/main.go
  /tmp/qor5example
}

export -f exampleRestart

find . -name *.go | entr -r bash -c "exampleRestart"
