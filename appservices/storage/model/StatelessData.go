package model

import "github.com/incognitochain/incognito-chain/metadata"

type Transaction struct {
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant
	Info     []byte // 512 bytes
	SigPubKey            []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig                  []byte `json:"Sig, omitempty"`       //
	Metadata metadata.Metadata


}

type Instruction struct {

}

type InputCoin struct {

}

type OutputCoin struct {

}