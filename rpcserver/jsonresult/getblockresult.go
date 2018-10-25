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
	temp := &GetBlockResult{
		BlockProducerSign: block.ChainLeaderSig,
		BlockProducer:     block.ChainLeader,
		Hash:              block.Hash().String(),
		PreviousBlockHash: block.Header.PrevBlockHash.String(),
		Version:           block.Header.Version,
		Height:            block.Height,
		Time:              block.Header.Timestamp,
		ChainID:           block.Header.ChainID,
		MerkleRoot:        block.Header.MerkleRoot.String(),
	}
	for _, tx := range block.Transactions {
		temp.TxHashes = append(temp.TxHashes, tx.Hash().String())
	}
	self = temp
}
