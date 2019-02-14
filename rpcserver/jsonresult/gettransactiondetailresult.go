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
	LockTime  string `json:"LockTime"`
	Fee       uint64 `json:"Fee"` // Fee applies: always consant
	Image     string `json:"Image"`

	Proof          *zkp.PaymentProof `json:"Proof"`
	InputCoinsJson string            `json:"InputCoinsJson"`
	SigPubKey      []byte            `json:"SigPubKey,omitempty"` // 64 bytes
	Sig            []byte            `json:"Sig,omitempty"`       // 64 bytes

	Metadata               string `json:"Metadata"`
	CustomTokenData        string `json:"CustomTokenData"`
	PrivacyCustomTokenData string `json:"PrivacyCustomTokenData"`
}
