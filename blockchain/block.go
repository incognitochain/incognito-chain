package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/common"
	"strconv"
)

const (
	// Default length of list tx in block
	defaultTransactionAlloc = 2048
)

type AgentDataPoint struct {
	AgentID string
	AgentSig string
	NumOfCoins float64
	NumOfBonds float64
	Tax float64
	EligibleAgentIDs []string
	LockTime int64
}

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
	blockHash    *common.Hash
	AgentDataPoints map[string]*AgentDataPoint
}

func (self *Block) AddTransaction(tx transaction.Transaction) error {
	self.Transactions = append(self.Transactions, tx)
	return nil
}

func (self *Block) ClearTransactions() {
	self.Transactions = make([]transaction.Transaction, 0, defaultTransactionAlloc)
}

func (self *Block) Hash() (*common.Hash) {
	//if self.blockHash != nil {
	//	return self.blockHash
	//}
	record := strconv.Itoa(self.Header.Version) + self.Header.MerkleRoot.String() + self.Header.Timestamp.String() + self.Header.PrevBlockHash.String() + strconv.Itoa(self.Header.Nonce) + strconv.Itoa(len(self.Transactions))
	hash := common.DoubleHashH([]byte(record))
	//self.blockHash = &hash
	//return self.blockHash
	return &hash
}
