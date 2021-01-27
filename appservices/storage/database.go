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
	GetBeaconStateRepository () repository.BeaconStateRepository
	GetShardStateRepository () repository.ShardStateRepository
	GetPDEStateRepository () repository.PDEStateRepository
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
