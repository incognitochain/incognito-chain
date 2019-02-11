#!/usr/bin/env bash

apt update
apt -y upgrade

echo "Install wget git"
apt install wget
apt install git

echo "Install golang..."
wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
tar -xvf go1.11.5.linux-amd64.tar.gz
mv go /usr/local

echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bashrc

source ~/.bashrc

echo "Install dep..."
go get -u github.com/golang/dep/cmd/dep

mkdir ~/go/bin -p
mkdir ~/go/src/github.com/ninjadotorg -p

echo "Clone constant..."
cd ~/go/src/github.com/ninjadotorg
git clone https://github.com/ninjadotorg/constant -b master

echo "Install constant packages..."
cd constant
dep ensure -v
