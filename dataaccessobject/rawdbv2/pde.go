package rawdbv2
type PDEContributionStore struct {
	Height        uint64
	Hash          string
	PDEContributionStatus []PDEContributionStatusInfo
}

type PDEContributionStatusInfo struct {
	Status             byte
	TokenID1Str        string
	Contributed1Amount uint64
	Returned1Amount    uint64
	TokenID2Str        string
	Contributed2Amount uint64
	Returned2Amount    uint64
	PDEContributionPairID string
}

type PDETradeStore struct {
	Height        uint64
	Hash          string
	PDETradeDetails []PDETradeInfo
}

type PDETradeInfo struct {
	TxReqId		string
	Status      byte
}

type PDECrossTradeStore struct {
	Height        uint64
	Hash          string
	PDECrossTradeDetails []PDECrossTradeInfo
}

type PDECrossTradeInfo struct {
	TxReqId		string
	Status      byte
}


type PDEWithdrawalStatusStore struct {
	Height        uint64
	Hash          string
	PDEWithdrawalStatusDetails []PDEWithdrawalStatusInfo
}

type PDEWithdrawalStatusInfo struct {
	TxReqId		string
	Status      byte
}

type PDEFeeWithdrawalStatusStore struct {
	Height        uint64
	Hash          string
	PDEFeeWithdrawalStatusDetails []PDEFeeWithdrawalStatusInfo
}

type PDEFeeWithdrawalStatusInfo struct {
	TxReqId		string
	Status      byte
}