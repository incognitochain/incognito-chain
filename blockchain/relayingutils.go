package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/pkg/errors"
)

type RelayingHeaderChainState struct{
	BNBHeaderChain *relaying.LatestHeaderChain
	BTCHeaderChain interface{}
}

func InitRelayingHeaderChainStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*RelayingHeaderChainState, error) {
	bnbHeaderChainState, err := getBNBHeaderChainState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	btcHeaderChainState, err := getBTCHeaderChainState(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &RelayingHeaderChainState{
		BNBHeaderChain: bnbHeaderChainState,
		BTCHeaderChain: btcHeaderChainState,
	}, nil
}


// getBNBHeaderChainState gets bnb header chain state at beaconHeight
func getBNBHeaderChainState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*relaying.LatestHeaderChain, error) {
	relayingStateKey := lvdb.NewRelayingStateKey(beaconHeight)
	relayingStateValueBytes, err := db.GetItemByKey([]byte(relayingStateKey))
	if err != nil {
		return nil, err
	}

	var hc relaying.LatestHeaderChain
	err = json.Unmarshal(relayingStateValueBytes, &hc)
	if err != nil {
		return nil, err
	}

	return &hc, nil
}

// todo
// getBTCHeaderChainState gets btc header chain state at beaconHeight
func getBTCHeaderChainState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (interface{}, error) {
	return nil, nil
}


// storeBNBHeaderChainState stores bnb header chain state at beaconHeight
func storeBNBHeaderChainState(db database.DatabaseInterface,
	beaconHeight uint64,
	bnbHeaderRelaying *relaying.LatestHeaderChain) error {
	key := lvdb.NewRelayingStateKey(beaconHeight)
	value, err := json.Marshal(bnbHeaderRelaying)
	if err != nil {
		return err
	}
	err = db.Put([]byte(key), value)
	if err != nil {
		return database.NewDatabaseError(database.StoreCustodianDepositStateError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

//todo
func storeRelayingHeaderStateToDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
	relayingHeaderState *RelayingHeaderChainState,
) error {
	err := storeBNBHeaderChainState(db, beaconHeight, relayingHeaderState.BNBHeaderChain)
	if err != nil {
		return err
	}
	return nil
}


