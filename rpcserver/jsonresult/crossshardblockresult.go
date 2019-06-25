package jsonresult

type CrossShardDataResult struct {
	HasCrossShard                  bool                         `json:"HasCrossShard"`
	CrossShardPRVResultList        []CrossShardPRVResult        `json:"CrossShardPRVResult"`
	CrossShardPRVPrivacyResultList []CrossShardPRVPrivacyResult `json:"CrossShardPRVPrivacyResult"`
	CrossShardCSTokenResultList    []CrossShardCSTokenResult    `json:"CrossShardCSTokenResult"`
}
type CrossShardPRVResult struct {
	PublicKey string `json:"PublicKey"`
	Value     uint64 `json:"Value"`
}

type CrossShardPRVPrivacyResult struct {
	PublicKey string `json:"PublicKey"`
}
type CrossShardCSTokenResult struct {
	Name                               string                           `json:"Name"`
	Symbol                             string                           `json:"Symbol"`
	Amount                             uint64                           `json:"Amount"`
	TokenID                            string                           `json:"TokenID"`
	TokenImage                         string                           `json:"TokenImage"`
	IsPrivacy                          bool                             `json:"IsPrivacy"`
	CrossShardCSTokenBalanceResultList []CrossShardCSTokenBalanceResult `json:"CrossShardCSTokenBalanceResultList"`
	CrossShardPrivacyCSTokenResultList []CrossShardPrivacyCSTokenResult `json:"CrossShardPrivacyCSTokenResultList"`
}

type CrossShardCSTokenBalanceResult struct {
	PaymentAddress string `json:"PaymentAddress"`
	Value          uint64 `json:"Value"`
}
type CrossShardPrivacyCSTokenResult struct {
	PublicKey string `json:"PublicKey"`
}
