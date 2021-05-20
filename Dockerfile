FROM golang:1.14-alpine as builder

# Incognito work directory
WORKDIR $GOPATH/src/github.com/incognitochain/incognito-chain

RUN apk add --no-cache make gcc linux-headers git

ADD . .
RUN make build

# Bring Incognito bin file into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/geth /usr/local/bin/

#EXPOSE 8545 8546 30303 30303/udp
#ENTRYPOINT ["geth"]
