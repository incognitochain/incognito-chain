package jsonresult

import (
	"github.com/ninjadotorg/constant/privacy/zeroknowledge"
)

type TransactionDetail struct {
	BlockHash string `json:"BlockHash"`
	Index     uint64 `json:"index"`
	ShardID   byte   `json:"shardID"`
	Hash      string `json:"Hash"`
	Version   int8   `json:"Version"`
	Type      string `json:"Type"` // Transaction type
	LockTime  int64  `json:"LockTime"`
	Fee       uint64 `json:"Fee"` // Fee applies: always consant

	Proof     *zkp.PaymentProof `json:"Proof"`
	SigPubKey []byte            `json:"SigPubKey,omitempty"` // 64 bytes
	Sig       []byte            `json:"Sig,omitempty"`       // 64 bytes

	MetaData string `json:"MetaData"`
}
