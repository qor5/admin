function exampleRestart() {
  echo "=================>"
  killall qor5example
  source dev_env
#  export DEV_PRESETS=1
  go build -o /tmp/qor5example main.go && /tmp/qor5example
}

export -f exampleRestart

find . -name "*.go" | entr -r bash -c "exampleRestart"
