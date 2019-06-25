# Incognito chain

![Incognito is digital money you can actually use.](https://avatars2.githubusercontent.com/u/45321944?s=200&v=4)

A new type of crypto-currency

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

#### Environment for dev

- [Go](https://golang.org/) 1.10 or newer.
- Mac, Linux, Windows OS

#### Installation
##### Linux/BSD/macOS/Windows - Build from Source
- Install Go according to the installation instructions here: [http://golang.org/doc/install](http://golang.org/doc/install)
- Ensure Go was installed properly and is a supported version:
```$ go version```
```$ go env GOROOT GOPATH```
- Clone Project to GOPATH
NOTE: The GOROOT and GOPATH above must not be the same path. It is recommended that GOPATH is set to a directory in your home directory such as ~/goprojects to avoid write permission issues. It is also recommended to add $GOPATH/bin to your PATH at this point.
- Install go lib with `go get -v`
- Run `go build -o incognito` in project dir or run `sh build.sh` to copy into go bin dir

### Run node

- Pre-install like "Prerequisites"
- Get node up with `go build`
    `
    $ ./incognito --enablewallet --wallet "wallet" --walletpassphrase "12345678" --testnet --norpcauth
    `
- Get node up with `sh build.sh`
    `
    $ incognito --enablewallet --wallet "wallet" --walletpassphrase "12345678" --testnet --norpcauth
    `

### Run with docker-compose
* To start dev container
    `
    $ docker-compose -p incognitodev -f dev-env/docker-compose.yaml up -d
    `
* To stop dev container
    `
    $ docker-compose -f dev-env/docker-compose.yaml down
    `
* To start developing
    `
    $ docker exec -it incognito-prototype-dev sh
    `
    `
    $ glide install
    `
    `
    $ go build
    `
    `
    $ ./incognito
    `
* To start other nodes (these nodes will start will config file in dev-env/nodes-data/node-<NODE_NUMBER>)
    `
    $ docker run -i -t --net incognitodev_incognito-net --mount type=bind,src=$PWD/incognito,dst=/constant --mount type=bind,src=$PWD/dev-env/nodes-data/node<REPLACE THIS WILL NODE_NUMBER>,dst=/nodedata --expose 9333 alpine:3.7 /constant --configfile /nodedata/config.conf
    `
## Config values
### How to use config
-   Refer to [config.go](https://github.com/incognitochain/incognito-chain/blob/master/config.go) or [sample-config.conf](https://github.com/incognitochain/incognito-chain/blob/master/sample-config.conf) in source code to get full explanation
-   Run node with config param in long or short format to change features of running node

## Other Utilities
-   Wallet client js https://github.com/incognitochain/incognito-chain-wallet-client-2
