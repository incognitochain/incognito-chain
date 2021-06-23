package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type CurrentPDEState struct {
	WaitingPDEContributions        map[string]*rawdbv2.PDEContribution
	DeletedWaitingPDEContributions map[string]*rawdbv2.PDEContribution
	PDEPoolPairs                   map[string]*rawdbv2.PDEPoolForPair
	PDEShares                      map[string]uint64
	PDETradingFees                 map[string]uint64
}
