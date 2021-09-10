function exampleRestart() {
  echo "=================>"
  killall qor5example
  source example/dev_env
#  export DEV_PRESETS=1
  go build -o /tmp/qor5example example/main.go && /tmp/qor5example
}

export -f exampleRestart

find . -name "*.go" | entr -r bash -c "exampleRestart"
