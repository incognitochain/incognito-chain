package wire

import "github.com/internet-cash/prototype/blockchain"

type MessageBlock struct {
	Header       blockchain.BlockHeader
	Transactions []*MessageTransaction
}
