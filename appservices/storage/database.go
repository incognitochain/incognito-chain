package storage

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/appservices/storage/repository"
)

type KindDB int

const (
	MONGODB = iota
)

type DatabaseDriver interface {
	//StoreFullBeaconState(beacon data.Beacon) error //TODO: will use this function for atomic/bulk insert.
	GetBeaconStorer () repository.BeaconStateStorer
	GetShardStorer () repository.ShardStateStorer

	GetPDEShareStorer () repository.PDEShareStorer
	GetPDEPoolForPairStateStorer() repository.PDEPoolForPairStateStorer
	GetPDETradingFeeStorer() repository.PDETradingFeeStorer
	GetWaitingPDEContributionStorer() repository.WaitingPDEContributionStorer

	GetCustodianStorer() repository.CustodianStorer
	GetWaitingPortingRequestStorer() repository.WaitingPortingRequestStorer
	GetFinalExchangeRatesStorer() repository.FinalExchangeRatesStorer
	GetWaitingRedeemRequestStorer() repository.WaitingRedeemRequestStorer
	GetMatchedRedeemRequestStorer() repository.MatchedRedeemRequestStorer
	GetLockedCollateralStorer() repository.LockedCollateralStorer




	GetTransactionStorer() repository.TransactionStorer
	GetInputCoinStorer() repository.InputCoinStorer
	GetOutputCoinStorer() repository.OutputCoinStorer
	GetCommitmentStorer() repository.CommitmentStorer

	GetTokenStateStorer() repository.TokenStateStorer


}

var dbDriver = make(map[KindDB]DatabaseDriver)

func AddDBDriver (kind KindDB, driver DatabaseDriver) error {
	if  _ , ok := dbDriver[kind]; ok  {
		return fmt.Errorf("DBDriver is existing")
	}
	dbDriver[kind] = driver
	return nil
}

func GetDBDriver(kind KindDB) DatabaseDriver {
	return dbDriver[kind]
}
