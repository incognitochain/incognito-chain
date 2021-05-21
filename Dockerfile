FROM golang:1.14-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /incognitochain
RUN cd /incognitochain && make build

# Bring Incognito bin file into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates make
WORKDIR /incognitochain
COPY --from=builder /incognitochain/incognito /usr/local/bin/
COPY --from=builder /incognitochain/priv2.json /incognitochain/
COPY --from=builder /incognitochain/whitelist.json /incognitochain/
COPY --from=builder /incognitochain/config/local/ /incognitochain/config/local/
COPY --from=builder /incognitochain/config/testnet2/ /incognitochain/config/testnet2/
COPY --from=builder /incognitochain/config/mainnet/ /incognitochain/config/mainnet/