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
	// Block 1764110 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("0000000000004791be42e055e8fe33e89d404723380da9434ee565ad9d89a235")
	prevBlkHash, _ := chainhash.NewHashFromStr("00000000a0107ee1b423f7ba814b6ceeeaa3e9a49767ec4582c57f270759f20b")
	merkleRoot, _ := chainhash.NewHashFromStr("0ed8b4025e655cf635ff4203d7dfb628527daa6ef91a12d587f7c3fb6ec8ff51")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1591668850, 0),
			Bits:       uint32(453050367),
			Nonce:      uint32(2572181929),
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
