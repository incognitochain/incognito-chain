package jsonresult

type EstimateFeeResult struct {
	FeeRate map[string]uint64 `json:"FeeRate"`
}
