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
	// Block 1,746,476 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("000000003eaf514e03ec578414626423cb95e8b1ec555966f321571889da7b31")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000002ebf3221c2f61fffb3d572d9aa746308cfb411d4f5ba155e442fc10e")
	merkleRoot, _ := chainhash.NewHashFromStr("83b930ab57ac2703ecc36e1c9d130fbefb1fe99a4fc8ecb825b53fd84144d9b8")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1590376637, 0),
			Bits:       uint32(486604799),
			Nonce:      uint32(2608682919),
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
