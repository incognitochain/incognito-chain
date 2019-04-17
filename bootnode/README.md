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
### Build
- Run `cd ./bootnode`
- Run `sh ./build.sh`
### Run
- Build successfully
- Run `constant-bootnode -p 9330`
- Or run `sh ./run.sh`
