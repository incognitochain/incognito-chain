package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

func StoreBeaconBlockHashByIndex(stateDB *StateDB, height uint64, hash common.Hash) error {
	key := common.HashH([]byte(fmt.Sprintf("beaconblockindex-%v", height)))
	err := stateDB.SetStateObject(BlockHashObjectType, key, &hash)
	if err != nil {
		return NewStatedbError(StoreBlockHashError, err)
	}
	return nil
}

func GetBeaconBlockHashByIndex(stateDB *StateDB, height uint64) (common.Hash, error) {
	key := common.HashH([]byte(fmt.Sprintf("beaconblockindex-%v", height)))
	stateObj, err := stateDB.getStateObject(BlockHashObjectType, key)
	if err != nil {
		return common.Hash{}, NewStatedbError(GetBlockHashError, err)
	}
	if stateObj == nil {
		return common.Hash{}, NewStatedbError(GetBlockHashError, err)
	}
	return *stateObj.GetValue().(*common.Hash), nil
}

func StoreShardBlockHashByIndex(stateDB *StateDB, shardID byte, height uint64, hash common.Hash) error {
	key := common.HashH([]byte(fmt.Sprintf("shardblockindex-%v-%v", shardID, height)))
	err := stateDB.SetStateObject(BlockHashObjectType, key, &hash)
	if err != nil {
		return NewStatedbError(StoreBlockHashError, err)
	}
	return nil
}

func GetShardBlockHashByIndex(stateDB *StateDB, shardID byte, height uint64) (common.Hash, error) {
	key := common.HashH([]byte(fmt.Sprintf("shardblockindex-%v-%v", shardID, height)))
	stateObj, err := stateDB.getStateObject(BlockHashObjectType, key)
	if err != nil {
		return common.Hash{}, NewStatedbError(GetBlockHashError, err)
	}
	if stateObj == nil {
		return common.Hash{}, NewStatedbError(GetBlockHashError, err)
	}

	return *stateObj.GetValue().(*common.Hash), nil
}
