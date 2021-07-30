package jsonresult

import "github.com/incognitochain/incognito-chain/blockchain/pdex"

type Pdexv3State struct {
	BeaconTimeStamp int64       `json:"BeaconTimeStamp"`
	Params          pdex.Params `json:"Params"`
}
