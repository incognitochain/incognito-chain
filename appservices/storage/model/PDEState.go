package model

type PDEContributionStatus struct {
	Height        uint64
	Hash          string
	Status             byte
	TokenID1Str        string
	Contributed1Amount uint64
	Returned1Amount    uint64
	TokenID2Str        string
	Contributed2Amount uint64
	Returned2Amount    uint64
	PDEContributionPairID string
}

type PDETrade struct {
	Height        uint64
	Hash          string
	TxReqId		string
	Status      byte
}

type PDECrossTrade struct {
	Height        uint64
	Hash          string
	TxReqId		string
	Status      byte
}

type PDEWithdrawalStatus struct {
	Height        uint64
	Hash          string
	TxReqId		string
	Status      byte
}

type PDEFeeWithdrawalStatus struct {
	Height        uint64
	Hash          string
	TxReqId		string
	Status      byte
}

