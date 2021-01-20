package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type PDEStateRepository interface {
	StoreLatestPDEBestState(ctx context.Context , pdeContributionStore *rawdbv2.PDEContributionStore, pdeTradeStore *rawdbv2.PDETradeStore, pdeCrossTradeStore *rawdbv2.PDECrossTradeStore,
		pdeWithdrawalStatusStore *rawdbv2.PDEWithdrawalStatusStore, pdeFeeWithdrawalStatusStore *rawdbv2.PDEFeeWithdrawalStatusStore) error
}