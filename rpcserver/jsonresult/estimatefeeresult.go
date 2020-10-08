package jsonresult

type EstimateFeeResult struct {
	EstimateFee          uint64
	EstimateFeeCoinPerKb uint64
	EstimateTxSizeInKb   uint64
}

func NewEstimateFeeResult(estimateFee uint64, estimateFeeCoinPerKb uint64, estimateTxSizeInKb uint64) *EstimateFeeResult {
	result := &EstimateFeeResult{
		EstimateFee:          estimateFee,
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
		EstimateTxSizeInKb:   estimateTxSizeInKb,
	}
	return result
}
