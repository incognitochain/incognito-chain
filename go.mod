module github.com/incognitochain/incognito-chain

go 1.13

require (
	cloud.google.com/go v0.104.0
	github.com/0xBahamoot/go-bigcompressor v1.0.5
	github.com/0xsirrush/color v1.7.0
	github.com/allegro/bigcache v1.2.1
	github.com/aristanetworks/goarista v0.0.0-20190704150520-f44d68189fd7 // indirect
	github.com/binance-chain/go-sdk v1.1.3
	github.com/blockcypher/gobcy v1.3.1
	github.com/bnb-chain/go-sdk v1.3.0 // indirect
	github.com/bnb-chain/node v0.10.6 // indirect
	github.com/btcsuite/btcd v0.23.0
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cosmos/cosmos-sdk v0.46.9 // indirect
	github.com/cweill/gotests v1.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/siphash v1.2.1 // indirect
	github.com/dgryski/go-identicon v0.0.0-20140725220403-371855927d74
	github.com/ebfe/keccak v0.0.0-20150115210727-5cc570678d1b
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/elastic/gosigar v0.10.4 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/eteu-technologies/near-api-go v0.0.2-0.20220525104145-c042eac16f21
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/incognitochain/go-libp2p-grpc v0.0.0-20181024123959-d1f24bf49b50
	github.com/incognitochain/go-libp2p-pubsub v0.2.7-0.20210126072501-9870234752e4
	github.com/jbenet/goprocess v0.1.4
	github.com/jessevdk/go-flags v1.5.0
	github.com/jrick/logrotate v1.0.0
	github.com/klauspost/compress v1.15.11
	github.com/libp2p/go-libp2p v0.11.0
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-net v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.2.6
	github.com/libp2p/go-libp2p-protocol v0.1.0 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.3.5
	github.com/libp2p/go-libp2p-swarm v0.2.8
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/olivere/elastic v6.2.21+incompatible
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.9.1 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa // indirect
	github.com/spf13/viper v1.13.0
	github.com/stathat/consistent v1.0.0 // indirect
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.8.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.26
	github.com/tendermint/tm-db v0.6.7 // indirect
	golang.org/x/crypto v0.5.0
	golang.org/x/sync v0.1.0
	google.golang.org/api v0.97.0
	google.golang.org/grpc v1.50.1
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v2 v2.4.0
	stathat.com/c/consistent v1.0.0
)

replace (
	// follow cosmos
	// use cosmos fork of keyring
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// github.com/btcsuite/btcd/btcec => github.com/btcsuite/btcd/btcec/v2 v2.3.2
	// dgrijalva/jwt-go is deprecated and doesn't receive security updates.
	// TODO: remove it: https://github.com/cosmos/cosmos-sdk/issues/13134
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2
	// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	github.com/jhump/protoreflect => github.com/jhump/protoreflect v1.9.0
	github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
	// use informal system fork of tendermint
	github.com/tendermint/tendermint => github.com/informalsystems/tendermint v0.34.26

	github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.46.9
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.5.0
	// github.com/tendermint/btcd => 
)
