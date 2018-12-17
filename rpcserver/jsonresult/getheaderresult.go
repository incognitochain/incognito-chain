package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type GetHeaderResult struct {
	BlockNum  int                      `json:"blocknum"`
	shardID   byte                     `json:"shardID"`
	BlockHash string                   `json:"blockhash"`
	Header    blockchain.BlockHeaderV2 `json:"header"`
}
