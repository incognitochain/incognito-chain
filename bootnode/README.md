# Bootnode service
## Standalone service provide for:
- Registering network node
- Get list alive network node

## How to Run
### Prerequisites
- Install Go >= 1.10
- Mac, Linux, Window OS
- Git clone source into $GOPATH/src/github.com/constant-money/constant-chain
- Run `go get -v`
### Build and RUN
- Run `cd ./bootnode`
- Run `sh ./build.sh`
- Run `constant-bootnode -p 9330`
- Run `constant-bootnode -h` to view helping
### Run directly
- Run `cd ./bootnode`
- Or run `sh ./run_bootnode.sh -p 9330`
