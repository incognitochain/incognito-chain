package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type GetHeaderResult struct {
	BlockNum  int                   `json:"blocknum"`
	ShardID   byte                  `json:"shardID"`
	BlockHash string                `json:"blockhash"`
	Header    blockchain.ShardBlock `json:"header"`
}
