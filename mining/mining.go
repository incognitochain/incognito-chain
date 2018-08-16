package mining

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/common"
	"time"
)

func NewBlkTmplGenerator(txSource TxSource, chain *blockchain.BlockChain) *BlkTmplGenerator {
	return &BlkTmplGenerator{
		txSource:    txSource,
		chain:       chain,
	}
}

func (g *BlkTmplGenerator) NewBlockTemplate() (*BlockTemplate, error) {
	prevBlockHash, _ := common.Hash{}.NewHash([]byte("1234567890123456789012"))
	txFees := make([]int64, 0, 1)
	var msgBlock blockchain.Block
	msgBlock.Header = blockchain.BlockHeader{
		Version:       0,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *prevBlockHash,
		Timestamp:     time.Now(),
		Difficulty:    0,
		Nonce:         0,
	}

	msgBlock.BlockHash = prevBlockHash

	return &BlockTemplate{
		Block:             &msgBlock,
		Fees:              txFees,
	}, nil

}