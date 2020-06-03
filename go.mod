module github.com/incognitochain/incognito-chain

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/0xsirrush/color v1.7.0
	github.com/allegro/bigcache v1.2.1
	github.com/aristanetworks/goarista v0.0.0-20190704150520-f44d68189fd7 // indirect
	github.com/binance-chain/go-sdk v1.1.3
	github.com/blockcypher/gobcy v1.3.1
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/siphash v1.2.1 // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/dgryski/go-identicon v0.0.0-20140725220403-371855927d74
	github.com/ebfe/keccak v0.0.0-20150115210727-5cc570678d1b
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/elastic/gosigar v0.10.4 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/ethereum/go-ethereum v1.8.22-0.20190710074244-72029f0f88f6
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/hashicorp/golang-lru v0.5.1
	github.com/incognitochain/go-libp2p-grpc v0.0.0-20181024123959-d1f24bf49b50
	github.com/ipfs/go-cid v0.0.3 // indirect
	github.com/jbenet/goprocess v0.1.3
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/klauspost/compress v1.10.1
	github.com/libp2p/go-libp2p v0.3.1
	github.com/libp2p/go-libp2p-core v0.2.2
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-net v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.1.3
	github.com/libp2p/go-libp2p-protocol v0.1.0 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.1.1
	github.com/libp2p/go-libp2p-swarm v0.2.1
	github.com/libp2p/go-openssl v0.0.3 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multihash v0.0.8 // indirect
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/olivere/elastic v6.2.21+incompatible
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/tsdb v0.9.1 // indirect
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa // indirect
	github.com/stathat/consistent v1.0.0
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tendermint v0.32.0
	go.opencensus.io v0.22.0 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	google.golang.org/api v0.10.0
	google.golang.org/grpc v1.27.1
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	stathat.com/c/consistent v1.0.0

)

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
