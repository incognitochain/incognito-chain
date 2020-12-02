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
	GetPDEShareStorer () repository.PDEShareStorer

	GetShardStorer () repository.ShardStateStorer
	GetTransactionStorer() repository.TransactionStorer

	GetInputCoinStorer() repository.InputCoinStorer
	GetOutputCoinStorer() repository.OutputCoinStorer
	GetCommitmentStorer() repository.CommitmentStorer
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
