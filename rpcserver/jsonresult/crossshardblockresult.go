package jsonresult

type CrossShardDataResult struct {
	HasCrossShard                       bool                              `json:"HasCrossShard"`
	CrossShardConstantResultList        []CrossShardConstantResult        `json:"CrossShardConstantResult"`
	CrossShardConstantPrivacyResultList []CrossShardConstantPrivacyResult `json:"CrossShardConstantPrivacyResult"`
	CrossShardCSTokenResultList         []CrossShardCSTokenResult         `json:"CrossShardCSTokenResult"`
}
type CrossShardConstantResult struct {
	PublicKey string `json:"PublicKey"`
	Value     uint64 `json:"Value"`
}

type CrossShardConstantPrivacyResult struct {
	PublicKey string `json:"PaymentAddress"`
}
type CrossShardCSTokenResult struct {
	Name                               string                           `json:"Name"`
	Symbol                             string                           `json:"Symbol"`
	Amount                             uint64                           `json:"Amount"`
	TokenID                            string                           `json:"TokenID"`
	TokenImage                         string                           `json:"TokenImage"`
	IsPrivacy                          bool                             `json:"IsPrivacy"`
	CrossShardCSTokenBalanceResultList []CrossShardCSTokenBalanceResult `json:"CrossShardCSTokenBalanceResultList"`
	CrossShardPrivacyCSTokenResultList []CrossShardPrivacyCSTokenResult `json:"CrossShardPrivacyCSTokenResult"`
}

type CrossShardCSTokenBalanceResult struct {
	PaymentAddress string `json:"PaymentAddress"`
	Value          uint64 `json:"Value"`
}
type CrossShardPrivacyCSTokenResult struct {
	PublicKey string `json:"PublicKey"`
}
