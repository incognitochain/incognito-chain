FROM golang:1.14-alpine

ENV GOROOT="/usr/local/go"
ENV GOPATH="$HOME/go"
ENV WORKDIR = $GOPATH/src/github.com/incognitochain/incognito-chain

# Create app directory
WORKDIR $WORKDIR

RUN apk add --no-cache curl

RUN apk update
RUN apk add git
RUN apk install build-essential

COPY . .

RUN go get -d
RUN make build

