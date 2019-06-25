# Bootnode service
## Standalone service provide for:
- Registering network node
- Get list alive network node

## How to Run
### Prerequisites
- Install Go >= 1.10
- Mac, Linux, Window OS
- Git clone source into $GOPATH/src/github.com/incognitochain/incognito-chain
- Run `go get -v`
### Build and RUN
- Run `cd ./bootnode`
- Run `sh ./build.sh`
- Run `incognito-bootnode -p 9330`
- Run `incognito-bootnode -h` to view helping
### Run directly
- Run `cd ./bootnode`
- Or run `sh ./run_bootnode.sh -p 9330`
