#!/usr/bin/env bash

apt update
apt -y upgrade

echo "Install wget git"
apt install -y wget
apt install -y git

echo "Install golang..."
wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
tar -xvf go1.11.5.linux-amd64.tar.gz
mv go /usr/local

mkdir ~/go/bin -p

if !(grep -q "GOROOT" ~/.bashrc); then
    echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bashrc
fi

export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

echo "Install dep..."
go get -u github.com/golang/dep/cmd/dep

echo "Clone constant..."
mkdir ~/go/src/github.com/ninjadotorg -p
cd ~/go/src/github.com/ninjadotorg
git clone https://github.com/ninjadotorg/constant -b master

echo "Install constant packages..."
cd constant
dep ensure -v
