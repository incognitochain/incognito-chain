FROM golang:1.14-alpine

# Create app directory
WORKDIR $GOPATH/src/github.com/incognitochain/incognito-chain

RUN apk add --no-cache make gcc musl-dev linux-headers git

COPY . .

#RUN go get -d

#RUN go build -v -o ./bin/incognito

RUN make build

## Build Geth in a stock Go builder container
#FROM golang:1.16-alpine as builder

#RUN apk add --no-cache make gcc musl-dev linux-headers git

#ADD . /go-ethereum
#RUN cd /go-ethereum && make geth

## Pull Geth into a second stage deploy alpine container
#FROM alpine:latest

#RUN apk add --no-cache ca-certificates
#COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

#EXPOSE 8545 8546 30303 30303/udp
#ENTRYPOINT ["geth"]
