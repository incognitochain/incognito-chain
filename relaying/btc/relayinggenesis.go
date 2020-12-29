package btcrelaying

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func getHardcodedMainNetGenesisBlock() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 634140 from bitcoin mainnet
	genesisHash, _ := chainhash.NewHashFromStr("00000000000000000008d18906abd79b6f21ffb02a805d5c85f976efc6d76d6c")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000000000086affd69c8e851cf061c6f677ba585bd5dcdf86e050fc")
	merkleRoot, _ := chainhash.NewHashFromStr("43c7222b1016819809181ef15cc9d80502fa27c67232695a21f2e08626394419")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1591847533, 0),
			Bits:       uint32(387219253),
			Nonce:      uint32(124268688),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func getHardcodedTestNet3GenesisBlock() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 1896910 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("000000000000003b095b39f4048771e77cc8b2e0885228b6df12cf684242cdf1")
	prevBlkHash, _ := chainhash.NewHashFromStr("0000000000000034b04cb66e042432ef3114e5834abb4cf60706b5f4a1c33ea6")
	merkleRoot, _ := chainhash.NewHashFromStr("c3079548c42fb4f76b12cd3f28be6a7e80b3c854e33bc259c8353327a0cfda31")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(1073733632),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1607582969, 0),
			Bits:       uint32(424004321),
			Nonce:      uint32(2341547277),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func getHardcodedTestNet3GenesisBlockForInc2() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 1863675 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000087badd69a52c3f2afa787d7d073081041ccf12ae172e4432de")
	prevBlkHash, _ := chainhash.NewHashFromStr("00000000000000156ca9adbfe1873c5aada796fe31c8b3de32bdf9538a6ff4e6")
	merkleRoot, _ := chainhash.NewHashFromStr("8cf9e50d1c4f88e1b000e67e8d9b1405a3aa39a64cae93b913957d05c31468e1")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1603172228, 0),
			Bits:       uint32(436259028),
			Nonce:      uint32(857120045),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func putGenesisBlockIntoChainParams(
	genesisHash *chainhash.Hash,
	msgBlk *wire.MsgBlock,
	chainParams chaincfg.Params,
) *chaincfg.Params {
	chainParams.GenesisBlock = msgBlk
	chainParams.GenesisHash = genesisHash
	return &chainParams
}

func GetMainNetParams() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedMainNetGenesisBlock()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, chaincfg.MainNetParams)
}

func GetTestNet3Params() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedTestNet3GenesisBlock()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, chaincfg.TestNet3Params)
}

func GetTestNet3ParamsForInc2() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedTestNet3GenesisBlockForInc2()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, chaincfg.TestNet3Params)
}
