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
	// Block 1,720,530 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000221df44a7733debbdb24625e9f40c1481bd738936174b08c4c8")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000002f032a62e7c01d58a0dbf5c0dfb9a1fe1068fa99d4315a10cf")
	merkleRoot, _ := chainhash.NewHashFromStr("9fad800db0242ed1a8341fd4e0c4abd4f94fad03d0c723b0ccd0b07e8e259db8")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(545259520),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1587337625, 0),
			Bits:       uint32(436381688),
			Nonce:      uint32(3436621438),
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
