package btcrelaying

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func getHardcodedMainNetGenesisBlock() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 623600 from bitcoin mainnet
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000000000f65159e17a1f0fa398dcd1a42665ec56c892d94d2512d")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000000000136638c38399f3455519c2fdce9e1603c1ce9407659c0b")
	merkleRoot, _ := chainhash.NewHashFromStr("17580d6e5a14340ad0ed3af4fc071f00a6b1e3a6fc87a87b72bf0cf8cc426ca6")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(541065216),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1585564707, 0),
			Bits:       uint32(387201857),
			Nonce:      uint32(1406428975),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func getHardcodedTestNet3GenesisBlock() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 1692037 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000eab0041219ba2ae1270edcae0303ad44443fce04cd3f7a2d4cf")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000070ca79878b1f120522330324679399da74c7177d3a2800b4640")
	merkleRoot, _ := chainhash.NewHashFromStr("9d5d8bb6b2afde75e2ea7042e494fd2c0d2c843957cce2e5ea1acc9d1e03609a")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1585566428, 0),
			Bits:       uint32(437256176),
			Nonce:      uint32(716756173),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func putGenesisBlockIntoChainParams(
	genesisHash *chainhash.Hash,
	msgBlk *wire.MsgBlock,
	chainParams *chaincfg.Params,
) *chaincfg.Params {
	chainParams.GenesisBlock = msgBlk
	chainParams.GenesisHash = genesisHash
	return chainParams
}

func GetMainNetParams() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedMainNetGenesisBlock()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, &chaincfg.MainNetParams)
}

func GetTestNet3Params() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedTestNet3GenesisBlock()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, &chaincfg.TestNet3Params)
}
