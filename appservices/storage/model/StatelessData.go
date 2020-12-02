package model

import (
	"github.com/incognitochain/incognito-chain/metadata"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
)

type Transaction struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	Hash                string `json:"Hash"`
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant
	Info     []byte // 512 bytes
	SigPubKey            []byte `json:"SigPubKey,omitempty"` // 33 bytes
	Sig                  []byte `json:"Sig,omitempty"`       //
	Proof *zkp.PaymentProof  `json:"Proof,omitempty"`
	PubKeyLastByteSender byte
	Metadata metadata.Metadata
}

type Instruction struct {

}

type InputCoin struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash                string `json:"TransactionHash"`
	PublicKey      []byte `json:"PublicKey"`
	CoinCommitment []byte `json:"CoinCommitment"`
	SNDerivator    []byte `json:"SNDerivator"`
	SerialNumber   []byte `json:"SerialNumber"`
	Randomness     []byte `json:"Randomness"`
	Value          uint64 `json:"Value"`
	Info           []byte `json:"Info"` //256 bytes
}

type OutputCoin struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash                string `json:"TransactionHash"`
	PublicKey      []byte `json:"PublicKey"`
	CoinCommitment []byte `json:"CoinCommitment"`
	SNDerivator    []byte `json:"SNDerivator"`
	SerialNumber   []byte `json:"SerialNumber"`
	Randomness     []byte `json:"Randomness"`
	Value          uint64 `json:"Value"`
	Info           []byte `json:"Info"` //256 bytes
}

type Commitment struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash string
	TokenID    string
	Commitment []byte
	Index      uint64
}