FROM ubuntu:18.04

RUN apt update
RUN apt -y upgrade

RUN apt install -y wget git

RUN wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
RUN tar -xvf go1.11.5.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /root/go/src/github.com/ninjadotorg
RUN cd /root/go/src/github.com/ninjadotorg && \
    git clone https://github.com/ninjadotorg/constant -b master
RUN cd /root/go/src/github.com/ninjadotorg/constant && \
    dep ensure -v
RUN cd /root/go/src/github.com/ninjadotorg/constant && \
    /usr/local/go/bin/go build -o constant
RUN cd /root/go/src/github.com/ninjadotorg/constant/bootnode && \
    /usr/local/go/bin/go build -o bootnode *.go

WORKDIR /root/go/src/github.com/ninjadotorg/constant

CMD ["/bin/bash"]
