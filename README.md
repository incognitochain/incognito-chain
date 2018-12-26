# Constant core
![Constant is digital money you can actually use.
](https://constant.money/static/images/block5.webp)
One prototype version for a new type of crypto curency

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

#### Environment for dev

- [Go](https://golang.org/) 1.11 or newer.
- Mac, Linux, Window OS

### Run node

- Clone Project
- Pre-install like "Prerequisites"
- Get node up
    ```bash
    $ ./constant --enablewallet --wallet "wallet" --walletpassphrase "12345678" --testnet --norpcauth
    ```

### Run with docker-compose
* To start dev container
    ```
    $ docker-compose -p cashdev -f dev-env/docker-compose.yaml up -d
    ```
* To stop dev container
    ```
    $ docker-compose -f dev-env/docker-compose.yaml down
    ```
* To start developing
    ```
    $ docker exec -it cash-prototype-dev sh
    ```
    ```
    $ glide install
    ```
    ```
    $ go build
    ```
    ```
    $ ./cash-prototype
    ```
* To start other nodes (these nodes will start will config file in dev-env/nodes-data/node-<NODE_NUMBER>)
    ```
    $ docker run -i -t --net cashdev_cash-net --mount type=bind,src=$PWD/cash-prototype,dst=/cash-prototype --mount type=bind,src=$PWD/dev-env/nodes-data/node<REPLACE THIS WILL NODE_NUMBER>,dst=/nodedata --expose 9333 alpine:3.7 /cash-prototype --configfile /nodedata/config.conf
    ```
## Config values
### How to use config
-   Refer to [config.go](https://github.com/ninjadotorg/constant/blob/master/config.go) or [sample-config.conf](https://github.com/ninjadotorg/constant/blob/master/sample-config.conf) in source code to get full explanation
-   Run node with config param in long or short format to change features of running node

## Other Utilities
-   Node wallet tool https://github.com/ninjadotorg/constant-wallet-extension
