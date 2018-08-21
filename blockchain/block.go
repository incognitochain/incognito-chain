package blockchain

import (
	"github.com/ninjadotorg/money-prototype/transaction"
	"github.com/ninjadotorg/money-prototype/common"
	"strconv"
)

const (
	// Default length of list tx in block
	defaultTransactionAlloc = 2048
)

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
	BlockHash    *common.Hash
}

func (self Block) AddTransaction(tx transaction.Tx) error {
	self.Transactions = append(self.Transactions, tx)
	return nil
}

func (self Block) ClearTransactions() {
	self.Transactions = make([]transaction.Tx, 0, defaultTransactionAlloc)
}

func (self Block) Hash() (*common.Hash) {
	if self.BlockHash != nil {
		return self.BlockHash
	}
	record := strconv.Itoa(self.Header.Version) + self.Header.MerkleRoot.String() + self.Header.Timestamp.String() + self.Header.PrevBlockHash.String() + strconv.Itoa(self.Header.Nonce) + strconv.Itoa(len(self.Transactions))
	hash := common.DoubleHashH([]byte(record))
	self.BlockHash = &hash
	return self.BlockHash
}
