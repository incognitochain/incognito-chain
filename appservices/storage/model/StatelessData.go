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
	TransactionCustomToken *TransactionCustomToken
}

type TransactionCustomToken struct {
	Tx			   Transaction          // used for privacy functionality
	PropertyID     string// = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string
	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
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
	TokenID		   string `json:"TokenID"`

}

type OutputCoin struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash     string `json:"TransactionHash"`
	PublicKey      		[]byte `json:"PublicKey"`
	CoinCommitment 		[]byte `json:"CoinCommitment"`
	SNDerivator    		[]byte `json:"SNDerivator"`
	SerialNumber   		[]byte `json:"SerialNumber"`
	Randomness     		[]byte `json:"Randomness"`
	Value          		uint64 `json:"Value"`
	Info           		[]byte `json:"Info"` //256 bytes
	TokenID		   		string  `json:"TokenID"`
	FromShardID      	byte `json:"FromShardID"`
	ToShardID        	byte `json:"ToShardID"`
	FromCrossShard   	bool `json:"FromCrossShard"`
	CrossBlockHash   	string `json:"CrossBlockHash"`
	CrossBlockHeight 	uint64 `json:"CrossBlockHeight"`
	PropertyName     	string `json:"PropertyName"`
	PropertySymbol   	string `json:"PropertySymbol"`
	Type             	int `json:"Type"`   // action type
	Mintable         	bool `json:"Mintable"`  // default false
	Amount           	uint64 `json:"Amount"` // init amount

}

type Commitment struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash string `json:"TransactionHash"`
	TokenID    string `json:"TokenID"`
	Commitment []byte `json:"Commitment"`
	Index      uint64 `json:"Index"`
}