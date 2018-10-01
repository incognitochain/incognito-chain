package jsonrpc

import "github.com/ninjadotorg/cash-prototype/blockchain"

type GetHeaderResult struct {
	BlockNum  int                    `json:"blocknum"`
	ChainID   byte                   `json:"chainid"`
	BlockHash string                 `json:"blockhash"`
	Header    blockchain.BlockHeader `json:"header"`
}
