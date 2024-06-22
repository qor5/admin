if ! command -v snippetgo &> /dev/null
then
    echo "snippetgo command not found. Installing..."

    # Install snippetgo using 'go install'
    go install github.com/sunfmin/snippetgo@v0.0.3

    # Check if installation was successful
    if command -v snippetgo &> /dev/null
    then
        echo "snippetgo successfully installed."
    else
        echo "Error: Failed to install snippetgo. Please check your Go environment and try again."
        exit 1
    fi
else
    echo "snippetgo is already installed."
fi

goModPath(){
    go list -m -json $1 | jq -r '.Dir'
}

snippetDirs=(
  $(goModPath github.com/qor5/web/v3)
  $(goModPath github.com/qor5/x/v3)
  $(goModPath github.com/qor5/admin/v3)
)
echo "${snippetDirs[@]}"
rm -rf ./generated/*
mkdir -p ./generated
gi=1
for d in "${snippetDirs[@]}"
do
  snippetgo -pkg=generated -dir=$d > ./generated/g${gi}.go
  gi=$((gi+1))
done

export DB_PARAMS="user=docs password=docs dbname=docs sslmode=disable host=localhost port=6532 TimeZone=Asia/Tokyo"
export ENV="development"
# export GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
go run ./build/main.go
rm ../docs/assets.go

function docsRestart() {
  echo "=================>"
  killall docgodocs
  go build -o /tmp/docgodocs ./server/main.go && /tmp/docgodocs
}

export -f docsRestart
find . -name "*.go" | entr -r bash -c "docsRestart"
