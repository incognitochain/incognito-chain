package jsonresult

type EstimateFeeResult struct {
	EstimateFeeCoinPerKb uint64
	EstimateTxSizeInKb   uint64
	EstimateFee          uint64
	MinFeePerTx          uint64
}

func NewEstimateFeeResult(estimateFeeCoinPerKb, estimateTxSizeInKb, estimateFee, minFeePerTx uint64) *EstimateFeeResult {
	result := &EstimateFeeResult{
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
		EstimateTxSizeInKb:   estimateTxSizeInKb,
		EstimateFee:          estimateFee,
		MinFeePerTx:          minFeePerTx,
	}
	return result
}
