package jsonresult

type EstimateFeeResult struct {
	EstimateFeeCoinPerKb uint64
	EstimateTxSizeInKb   uint64
	IsFeePToken          bool
}

func NewEstimateFeeResult(estimateFeeCoinPerKb uint64, estimateTxSizeInKb uint64, isFeePToken bool) *EstimateFeeResult {
	result := &EstimateFeeResult{
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
		EstimateTxSizeInKb:   estimateTxSizeInKb,
		IsFeePToken:          isFeePToken,
	}
	return result
}
