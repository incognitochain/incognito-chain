package model

import (
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type PDEShare struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount             uint64
}

type PDEPoolForPair struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	Token1ID        string
	Token1PoolValue uint64
	Token2ID        string
	Token2PoolValue uint64
}

type PDETradingFee struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount             uint64
}

type WaitingPDEContribution struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	PairID             string
	ContributorAddress string
	TokenID            string
	Amount             uint64
	TXReqID            string
}


type Custodian struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	HoldingPubTokens       map[string]uint64 // tokenID : amount
	LockedAmountCollateral map[string]uint64 // tokenID : amount
	RemoteAddresses        map[string]string // tokenID : remote address
	RewardAmount           map[string]uint64 // tokenID : amount
}

type WaitingPortingRequest struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	UniquePortingID 		string
	TokenID         		string
	PorterAddress   		string
	Amount          		uint64
	Custodians      		[]statedb.MatchingPortingCustodianDetail
	PortingFee      		uint64
	WaitingBeaconHeight    	uint64
	TXReqID         		string
}

type FinalExchangeRate struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	Amount					uint64
	TokenID					string
}

type RedeemRequest struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	UniqueRedeemID        string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	Custodians            []data.MatchingRedeemCustodianDetail
	RedeemFee             uint64
	RedeemBeaconHeight    uint64
	TXReqID               string
}

type LockedCollateral struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	TotalLockedCollateralForRewards uint64
	CustodianAddress		string
	Amount                  uint64
	//LockedCollateralDetail          map[string]uint64 // custodianAddress : amount
}


type TokenState struct {
	ShardID		   byte         `json:"ShardID"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	Token   []TokenInformation `json:"Token"`
}

type TokenInformation struct {
	TokenID        string
	PropertyName   string
	PropertySymbol string
	TokenType      int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
	Info           []byte
	InitTx         string
	Txs            []string
}

type CommitteeRewardState struct {
	ShardID		   byte         `json:"ShardID"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	CommitteeRewardState map[string]map[string]uint64 `json:"CommitteeRewardState"` //address->tokenId->amount
}