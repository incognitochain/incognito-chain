package portaltokens

import (
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
)

type PortalTokenProcessor interface {
	IsValidRemoteAddress(address string, bcr bMeta.ChainRetriever) (bool, error)
	GetChainID() string
	GetMinTokenAmount() uint64

	GetExpectedMemoForPorting(portingID string) string
	GetExpectedMemoForRedeem(redeemID string, custodianIncAddress string) string
	ParseAndVerifyProof(
		proof string, bc bMeta.ChainRetriever, expectedMemo string, expectedPaymentInfos map[string]uint64) (bool, error)
}


// set MinTokenAmount to avoid attacking with amount is less than smallest unit of cryptocurrency
// such as satoshi in BTC
type PortalToken struct {
	ChainID string
	MinTokenAmount uint64 			// minimum amount for porting/redeem
}