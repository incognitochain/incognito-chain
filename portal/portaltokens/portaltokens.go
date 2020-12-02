package portaltokens

import (
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

//TODO: add more functions
type PortalTokenProcessor interface {
	ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc bMeta.ChainRetriever) (bool, error)
	ParseAndVerifyProofForRedeem(proof string, redeemReq *statedb.RedeemRequest, bc bMeta.ChainRetriever, matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error)
	IsValidRemoteAddress(address string) (bool, error)
	GetChainID() (string)
}

type PortalToken struct {
	ChainID string
}