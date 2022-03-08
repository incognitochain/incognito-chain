package metadata

type SolTokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UIAmount       float64 `json:"uiAmount"`
	UIAmountString string  `json:"uiAmountString"`
}
type SolTokenAccountData struct {
	IsNative    bool           `json:"isNative"`
	Mint        string         `json:"mint"`
	Owner       string         `json:"owner"`
	State       string         `json:"state"`
	TokenAmount SolTokenAmount `json:"tokenAmount"`
}
type AccountInfo struct {
	Info SolTokenAccountData `json:"info"`
}
type SolAccountData struct {
	Parsed AccountInfo `json:"parsed"`
	Type   string      `json:"type"`
}

type ShieldInfo struct {
	Amount              uint64 `json:"amount"`
	ReceivingIncAddrStr string `json:"receiverAddrStr"`
	ExternalTokenID     []byte `json:"externalTokenIDStr"`
}

const SolPubKeyLen = 32
const SolShieldInstLen = 157 // 1 + 8 + 148
const SolShieldInstTag = 0
