package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
)

func FetchBeaconBlockFromHeightV2(db incdb.Database, from uint64, to uint64) ([]*BeaconBlock, error) {
	beaconBlocks := []*BeaconBlock{}
	for i := from; i <= to; i++ {
		hash, err := rawdb.GetBeaconBlockHashByIndex(db, i)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(db, hash)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlock := BeaconBlock{}
		err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
		if err != nil {
			return beaconBlocks, NewBlockChainError(UnmashallJsonShardBlockError, err)
		}
		beaconBlocks = append(beaconBlocks, &beaconBlock)
	}
	return beaconBlocks, nil
}
