package model

import (
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type Transaction struct {
	ShardHash 			 string `json:"ShardHash"`
	ShardHeight 		 uint64 `json:"ShardHeight"`
	TxSize				 uint64 `json:"TxSize"`
	Index                uint64 `json:"Index"`
	ShardId  			 byte `json:"ShardId"`
	Hash                 string `json:"Hash"`
	Version  			 int8   `json:"Version"`
	Type     			 string `json:"Type"` // Transaction type
	LockTime 			 string  `json:"LockTime"`
	Fee      			 uint64 `json:"Fee"` // Fee applies: always constant
	Image                string  `json:"Image"`
	IsPrivacy            bool `json:"IsPrivacy"`
	Proof           	 *string `json:"Proof,omitempty"`
	ProofDetail          jsonresult.ProofDetail       `json:"ProofDetail"`
	InputCoinPubKey 	 string            `json:"InputCoinPubKey"`
	SigPubKey            string `json:"SigPubKey,omitempty"` // 33 bytes
	Sig                  string `json:"Sig,omitempty"`       //
	PubKeyLastByteSender byte
	Metadata                      string     `json:"Metadata"`
	CustomTokenData               string      `json:"CustomTokenData"`
	PrivacyCustomTokenID          string      `json:"PrivacyCustomTokenID"`
	PrivacyCustomTokenName        string      `json:"PrivacyCustomTokenName"`
	PrivacyCustomTokenSymbol      string      `json:"PrivacyCustomTokenSymbol"`
	PrivacyCustomTokenData        string      `json:"PrivacyCustomTokenData"`
	PrivacyCustomTokenProofDetail jsonresult.ProofDetail `json:"PrivacyCustomTokenProofDetail"`
	PrivacyCustomTokenProof 		*string `json:"PrivacyCustomTokenProof"`
	PrivacyCustomTokenIsPrivacy   bool        `json:"PrivacyCustomTokenIsPrivacy"`
	PrivacyCustomTokenFee         uint64      `json:"PrivacyCustomTokenFee"`
	IsInMempool bool `json:"IsInMempool"`
	IsInBlock   bool `json:"IsInBlock"`
	Info     			string `json:"Info"` // 512 bytes
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
	PublicKey      string `json:"PublicKey"`
	CoinCommitment string `json:"CoinCommitment"`
	SNDerivator    string `json:"SNDerivator"`
	SerialNumber   string `json:"SerialNumber"`
	Randomness     string `json:"Randomness"`
	Value          uint64 `json:"Value"`
	Info           string `json:"Info"` //256 bytes
	TokenID		   string `json:"TokenID"`

}

type OutputCoin struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash     string `json:"TransactionHash"`
	PublicKey      		string `json:"PublicKey"`
	CoinCommitment 		string `json:"CoinCommitment"`
	SNDerivator    		string `json:"SNDerivator"`
	SerialNumber   		string `json:"SerialNumber"`
	Randomness     		string `json:"Randomness"`
	Value          		uint64 `json:"Value"`
	Info           		string `json:"Info"` //256 bytes
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
	LockTime 			 int64  `json:"LockTime"`
	CoinDetailsEncrypted string `json:"CoinDetailsEncrypted"`
	TransactionMemo       string `json:"TransactionMemo"`
}

type Commitment struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash string `json:"TransactionHash"`
	TokenID    string `json:"TokenID"`
	Commitment string `json:"Commitment"`
	Index      uint64 `json:"Index"`
}

type PublicKeyToTransactionHash struct {
	ShardId  			byte `json:"ShardId"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	TransactionHash     string `json:"TransactionHash"`
	PublicKey			string `json:"PublicKey"`
}