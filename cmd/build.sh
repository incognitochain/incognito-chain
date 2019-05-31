echo "Start build incognitoctl"

echo "go get"
go get -d

APP_NAME="incognitoctl"

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
mv ./$APP_NAME $GOPATH/bin/$APP_NAME

echo "Build incognitoctl success!"
