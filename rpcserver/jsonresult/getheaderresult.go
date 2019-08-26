package jsonresult

import "github.com/incognitochain/incognito-chain/blockchain"

type GetHeaderResult struct {
	BlockNum  int                    `json:"Blocknum"`
	ShardID   byte                   `json:"ShardID"`
	BlockHash string                 `json:"Blockhash"`
	Header    blockchain.ShardHeader `json:"Header"`
}
