package portaltokens

import (
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
)

//TODO: add more functions
type PortalTokenProcessor interface {
	IsValidRemoteAddress(address string) (bool, error)
	GetChainID() string

	GetExpectedMemoForPorting(portingID string) string
	GetExpectedMemoForRedeem(redeemID string, custodianIncAddress string) string
	ParseAndVerifyProof(
		proof string, bc bMeta.ChainRetriever, expectedMemo string, expectedPaymentInfos map[string]uint64) (bool, error)
}

type PortalToken struct {
	ChainID string
}