FROM golang:1.11rc1-alpine

ENV REFLEXGLIDE=https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz
ENV WORK=/go/src/github.com/ninjadotorg/constant
ENV GOROOT="/usr/local/go"
ENV GOPATH="/go"

# Create app directory
WORKDIR $WORK

RUN apk add --no-cache curl

RUN apk update
RUN apk add git

RUN mkdir glide
RUN curl -Lk $REFLEXGLIDE -o glide.tar.gz
RUN tar -xzf glide.tar.gz -C glide/
RUN mv glide/linux-amd64/glide /bin/glide

COPY glide.* ./


RUN pwd && ls -lah && glide install && glide update

COPY . .

RUN go get -d

WORKDIR $WORK

RUN go build

EXPOSE 9333:9333

CMD [ "./constant" ]
