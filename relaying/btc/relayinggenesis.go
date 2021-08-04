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
	// Block 2063133 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000039e1af0389c1cdd005b57bc8bd9a753b6b261e6141ad692e38")
	prevBlkHash, _ := chainhash.NewHashFromStr("0000000000000038bacae6a99ac6264297689d4b3dda5974f1ff72ea9478d3ae")
	merkleRoot, _ := chainhash.NewHashFromStr("5f67482540307a8e0adc8f9e9d5dce129c866d5e47913dfec82eadefeeacab32")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1628065055, 0),
			Bits:       uint32(423816192),
			Nonce:      uint32(1143442369),
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
