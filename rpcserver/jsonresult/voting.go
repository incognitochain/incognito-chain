package jsonresult

type GetEncryptionFlagResult struct {
	DCBFlag uint32 `json:"DCBFlag"`
	GOVFlag uint32 `json:"GOVFlag"`
}

type GetEncryptionLastBlockHeightResult struct {
	BlockHeight uint32 `json:"blockHeight"`
}
