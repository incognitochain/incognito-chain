package model

type PDEShare struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	PDEShareInfo []PDEShareInfo `json:"PDEShareInfo"`
}

type PDEShareInfo struct {
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount         string // init amount store as string because mongodb don't support uint64
}

type PDEPoolForPair struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	PDEPoolForPairInfo  []PDEPoolForPairInfo `json:"PDEPoolForPairInfo"`
}
type PDEPoolForPairInfo struct {
	Token1ID        string
	Token1PoolValue string
	Token2ID        string
	Token2PoolValue string
}

type PDETradingFee struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	PDETradingFeeInfo  []PDETradingFeeInfo `json:"PDETradingFeeInfo"`
}

type PDETradingFeeInfo struct {
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount         string // init amount store as string because mongodb don't support uint64
}

type WaitingPDEContribution struct {
	BeaconBlockHash    string					 	`json:"BeaconBlockHash"`
	BeaconEpoch									uint64					 					`json:"BeaconEpoch"`
	BeaconHeight									uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	WaitingPDEContributionInfo []WaitingPDEContributionInfo `json:"WaitingPDEContributionInfo"`
}

type WaitingPDEContributionInfo struct {
	PairID             string
	ContributorAddress string
	TokenID            string
	Amount         string // init amount store as string because mongodb don't support uint64
	TXReqID            string

}

type Custodian struct {
	BeaconBlockHash     string					 	`json:"BeaconBlockHash"`
	BeaconEpoch			uint64					 					`json:"BeaconEpoch"`
	BeaconHeight		uint64					 					`json:"BeaconHeight"`
	BeaconTime			int64 `json:"BeaconTime"`
	CustodianInfo       []CustodianInfo  `json:"CustodianInfo"`
}

type CustodianInfo struct {
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	HoldingPubTokens       map[string]string // tokenID : amount
	LockedAmountCollateral map[string]string // tokenID : amount
	RemoteAddresses        map[string]string // tokenID : remote address
	RewardAmount           map[string]string // tokenID : amount
}

type WaitingPortingRequest struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	WaitingPortingRequestInfo []WaitingPortingRequestInfo   `json:"WaitingPortingRequestInfo"`
}

type MatchingPortingCustodianDetail struct {
	IncAddress             string
	RemoteAddress          string
	Amount                 string
	LockedAmountCollateral string
}

type WaitingPortingRequestInfo struct {
	UniquePortingID 		string
	TokenID         		string
	PorterAddress   		string
	Amount          		string
	Custodians      		[]MatchingPortingCustodianDetail
	PortingFee      		uint64
	WaitingBeaconHeight    	uint64
	TXReqID         		string
}

type FinalExchangeRate struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 			`json:"BeaconTime"`
	FinalExchangeRateInfo    []FinalExchangeRateInfo `json:"FinalExchangeRateInfo"`
}

type FinalExchangeRateInfo struct {
	Amount					string
	TokenID					string
}

type RedeemRequest struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	RedeemRequestInfo []RedeemRequestInfo   `json:"RedeemRequestInfo"`
}

type MatchingRedeemCustodianDetail struct {
	IncAddress    string
	RemoteAddress string
	Amount        string
}


type RedeemRequestInfo struct {
	UniqueRedeemID        string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          string
	Custodians            []MatchingRedeemCustodianDetail
	RedeemFee             uint64
	RedeemBeaconHeight    uint64
	TXReqID               string
}

type LockedCollateral struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	LockedCollateralInfo     []LockedCollateralInfo  `json:"LockedCollateralInfo"`

}

type LockedCollateralInfo struct {
	TotalLockedCollateralForRewards uint64
	CustodianAddress		string
	Amount                  string
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
	Amount         string // init amount store as string because mongodb don't support uint64
	Info           []byte
	InitTx         string
	Txs            []string
}

type BridgeTokenState struct {
	BeaconBlockHash    		string					 	`json:"BeaconBlockHash"`
	BeaconEpoch				uint64					 					`json:"BeaconEpoch"`
	BeaconHeight			uint64					 					`json:"BeaconHeight"`
	BeaconTime				int64 `json:"BeaconTime"`
	BridgeTokenInfo []BridgeTokenInfo  `json:"BridgeTokenInfo"`
}

type BridgeTokenInfo struct {
	TokenID         string `json:"tokenId"`
	Amount          string // init amount store as string because mongodb don't support uint64
	ExternalTokenID []byte       `json:"externalTokenId"`
	Network         string       `json:"network"`
	IsCentralized   bool         `json:"isCentralized"`

}

type CommitteeRewardState struct {
	ShardID		   byte         `json:"ShardID"`
	ShardHash 			string `json:"ShardHash"`
	ShardHeight 		uint64 `json:"ShardHeight"`
	CommitteeReward     []CommitteeReward `json:"CommitteeReward"`
}

type CommitteeReward struct {
	Address      string `json:"Address"`
	TokenId      string `json:"TokenId"`
	Amount         string // init amount store as string because mongodb don't support uint64
}