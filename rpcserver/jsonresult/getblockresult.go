package jsonresult

import "github.com/ninjadotorg/cash/blockchain"

type GetBlockResult struct {
	Data              string             `json:"Data"`
	Hash              string             `json:"Hash"`
	Confirmations     int64              `json:"confirmations"`
	Height            int32              `json:"Height"`
	Version           int                `json:"Version"`
	MerkleRoot        string             `json:"MerkleRoot"`
	Time              int64              `json:"Time"`
	ChainID           byte               `json:"ChainID"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	NextBlockHash     string             `json:"NextBlockHash"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []GetBlockTxResult `json:"Txs"`
	BlockProducerSign string             `json:"BlockProducerSign"`
	BlockProducer     string             `json:"BlockProducer"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}

func (self *GetBlockResult) Init(block *blockchain.Block) {
	self.BlockProducerSign = block.ChainLeaderSig
	self.BlockProducer = block.ChainLeader
	self.Hash = block.Hash().String()
	self.PreviousBlockHash = block.Header.PrevBlockHash.String()
	self.Version = block.Header.Version
	self.Height = block.Height
	self.Time = block.Header.Timestamp
	self.ChainID = block.Header.ChainID
	self.MerkleRoot = block.Header.MerkleRoot.String()
	self.TxHashes = make([]string, 0)
	for _, tx := range block.Transactions {
		self.TxHashes = append(self.TxHashes, tx.Hash().String())
	}
}
