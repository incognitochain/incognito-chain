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
	// Block 1860096 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("000000000000091e4663b65c1f14c72269e31e43b49057456f285c02ff991e7b")
	prevBlkHash, _ := chainhash.NewHashFromStr("000000000000009a861f511368483704df637fcdf0a7a1789442e9ae7fe34e61")
	merkleRoot, _ := chainhash.NewHashFromStr("5d1569f139880f8b951e5700ab0de4c3cfffe8c02bf692213634dcdf37a4f67a")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1602205306, 0),
			Bits:       uint32(436884112),
			Nonce:      uint32(164276898),
		},
		Transactions: []*wire.MsgTx{},
	}
	return &genesisBlock, genesisHash
}

func getHardcodedTestNet3GenesisBlockForInc2() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 1833130 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("00000000000000e658954c44586a2d80e775daa0bd2a6b7886a609a9511021ba")
	prevBlkHash, _ := chainhash.NewHashFromStr("00000000000001c1cb39bad79ea61a4adef11ff0cdd00da9dd9e7547152403cd")
	merkleRoot, _ := chainhash.NewHashFromStr("0d3a4e0a6d74cbec1f853b790a2933aeda80bb6a070b53a03ac1b05c083c4694")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(1073676288),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1599794353, 0),
			Bits:       uint32(436381186),
			Nonce:      uint32(3755060888),
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

func GetTestNet3ParamsForInc2() *chaincfg.Params {
	genesisBlock, genesisHash := getHardcodedTestNet3GenesisBlockForInc2()
	return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, &chaincfg.TestNet3Params)
}
