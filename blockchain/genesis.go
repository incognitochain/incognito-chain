package blockchain

import (
	"github.com/internet-cash/prototype/transaction"
	"github.com/internet-cash/prototype/common"
	"time"
)

type GenesisBlockGenerator struct {
}

func (self GenesisBlockGenerator) CalcMerkleRoot(txns []*transaction.Tx) common.Hash {
	if len(txns) == 0 {
		return common.Hash{}
	}

	utilTxns := make([]*transaction.Tx, 0, len(txns))
	for _, tx := range txns {
		utilTxns = append(utilTxns, tx)
	}
	merkles := Merkle{}.BuildMerkleTreeStore(utilTxns)
	return *merkles[len(merkles)-1]
}

func (self GenesisBlockGenerator) CreateGenesisBlock(time time.Time, nonce int, difficulty int, version int, genesisReward int) *Block {
	genesisBlock := Block{}
	// update default genesis block
	genesisBlock.Header.Timestamp = time
	genesisBlock.Header.Nonce = nonce
	genesisBlock.Header.Difficulty = difficulty
	genesisBlock.Header.Version = version

	tx := transaction.Tx{
		Version: 1,
		TxIn: []transaction.TxIn{
			{
				Sequence:        0xffffffff,
				SignatureScript: []byte{},
				PreviousOutPoint: transaction.OutPoint{
					Hash: common.Hash{},
				},
			},
		},
		TxOut: []transaction.TxOut{{
			Value:    genesisReward,
			PkScript: []byte(GENESIS_BLOCK_PUBKEY),
		}},
	}
	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)
	genesisBlock.Transactions = append(genesisBlock.Transactions, &tx)
	return &genesisBlock
}
