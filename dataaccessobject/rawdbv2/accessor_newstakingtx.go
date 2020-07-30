package rawdbv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
)

var NewStakingTXPrefix = []byte("newstakingtxstatus-")

type StakingTXInfo struct {
	MStakingTX map[int]map[string]string //shardID -> (committee->txid)
	Height     uint64
}

func StoreMapStakingTxNew(db incdb.KeyValueWriter, height uint64, mStakingTx map[int]map[string]string) error {
	key := NewStakingTXPrefix
	data, err := json.Marshal(StakingTXInfo{
		Height:     height,
		MStakingTX: mStakingTx,
	})
	if err != nil {
		return err
	}
	err = db.Put(key, data)
	return err
}

func GetMapStakingTxNew(db incdb.KeyValueReader) (*StakingTXInfo, error) {
	key := NewStakingTXPrefix
	data, err := db.Get(key)
	if err != nil {
		fmt.Println("GetShardStakingTxMapError", err)
		return nil, errors.New("GetShardStakingTxMapError")
	}
	value := &StakingTXInfo{}
	err = json.Unmarshal(data, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}
