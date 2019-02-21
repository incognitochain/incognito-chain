package jsonresult

type GetEncryptionFlagResult struct {
	DCBFlag byte `json:"DCBFlag"`
	GOVFlag byte `json:"GOVFlag"`
}

type GetEncryptionLastBlockHeightResult struct {
	BlockHeight uint64 `json:"blockHeight"`
}
