echo "Start build bootnode"

echo "go get"
go get -d

APP_NAME="constant-bootnode"

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./constant-bootnode $GOPATH/bin/$APP_NAME"
mv ./constant-bootnode $GOPATH/bin/$APP_NAME

echo "Build bootnode success!"
