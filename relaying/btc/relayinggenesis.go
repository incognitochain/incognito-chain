package btcrelaying

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func getHardcodedMainNetGenesisBlock() (*wire.MsgBlock, *chainhash.Hash) {
	// Block 697298 from bitcoin mainnet
	genesisHash, _ := chainhash.NewHashFromStr("0000000000000000001128af8d34e168792e04d80109d5bd568aee079fa312d3")
	prevBlkHash, _ := chainhash.NewHashFromStr("0000000000000000000694bf9cafd23dc163ae42c583169dc047599c02e7eccc")
	merkleRoot, _ := chainhash.NewHashFromStr("439630ecdc993a273dce80e7d3b8d15003fb15b75a79861a8cff8afc5317436e")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(805298180),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1629770313, 0),
			Bits:       uint32(387061771),
			Nonce:      uint32(3047844147),
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
	// Block 2064989 from bitcoin testnet3
	genesisHash, _ := chainhash.NewHashFromStr("00000000a51a6c208820e26b20eed4197dfb22c4851558286af4b19ea5dd6fc9")
	prevBlkHash, _ := chainhash.NewHashFromStr("0000000021b46e28a781c075ab0f72dcd38316953a440517596105957249e6fc")
	merkleRoot, _ := chainhash.NewHashFromStr("a98808832a23a706ba6f5fd0d2988a31e2fe87f904fbcf99fcd20c2931133bd7")
	var genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(536870912),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  time.Unix(1629088229, 0),
			Bits:       uint32(486604799),
			Nonce:      uint32(1465578260),
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
