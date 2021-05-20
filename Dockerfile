FROM golang:1.14-alpine

# Create app directory
WORKDIR $GOPATH/src/github.com/incognitochain/incognito-chain

RUN apk add --no-cache make gcc musl-dev linux-headers git

COPY . .

#RUN go get -d

#RUN go build -v -o ./bin/incognito

RUN make build
