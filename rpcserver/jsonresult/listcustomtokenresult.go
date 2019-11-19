package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/transaction"
)

type CustomToken struct {
	ID                 string   `json:"ID"`
	Name               string   `json:"Name"`
	Symbol             string   `json:"Symbol"`
	Image              string   `json:"Image"`
	Amount             uint64   `json:"Amount"`
	IsPrivacy          bool     `json:"IsPrivacy"`
	ListTxs            []string `json:"ListTxs"`
	InitiatorPublicKey string   `json:"InitiatorPublicKey"`
}

func NewPrivacyToken(obj transaction.TxCustomTokenPrivacy) *CustomToken {
	customToken := &CustomToken{}
	customToken.ID = obj.TxPrivacyTokenData.PropertyID.String()
	customToken.Symbol = obj.TxPrivacyTokenData.PropertySymbol
	customToken.Name = obj.TxPrivacyTokenData.PropertyName
	customToken.Amount = obj.TxPrivacyTokenData.Amount
	//customToken.Image = common.Render(obj.TxPrivacyTokenData.PropertyID[:])
	customToken.IsPrivacy = true
	return customToken
}

func NewPrivacyForCrossShard(obj blockchain.CrossShardTokenPrivacyMetaData) *CustomToken {
	customToken := &CustomToken{}
	customToken.ID = obj.TokenID.String()
	customToken.Symbol = obj.PropertySymbol
	customToken.Name = obj.PropertyName
	customToken.Amount = obj.Amount
	//customToken.Image = common.Render(obj.TokenID[:])
	customToken.IsPrivacy = true
	return customToken
}

type ListCustomToken struct {
	ListCustomToken []CustomToken `json:"ListCustomToken"`
}
