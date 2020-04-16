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
	genesisHash, _ := chainhash.NewHashFromStr("00000000000000e4d31ce62f7fc156edb9c663e40acec7cacac4083287f57423")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000185b26ef5794c0001f51c1f35db2a989a8f9c2dafcf5ac83a178")
	merkleRoot, _ := chainhash.NewHashFromStr("5b97f801047a1e693119827b895956caba096d394a00229675ece7c8066eb684")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1586251423, 0),
			Bits:       uint32(436273151),
			Nonce:      uint32(2481162887),
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
