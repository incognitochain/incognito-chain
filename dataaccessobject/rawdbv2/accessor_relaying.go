package rawdbv2

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/tendermint/tendermint/types"
)

// key prefix
var (
	RelayingBNBHeaderStatePrefix = []byte("relayingbnbheaderstate-")
	RelayingBNBHeaderChainPrefix = []byte("relayingbnbheaderchain-")
)

func NewBNBHeaderRelayingStateKey(beaconHeight uint64) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	key := append(RelayingBNBHeaderStatePrefix, beaconHeightBytes...)
	return string(key) //prefix + beaconHeight
}

func NewRelayingBNBHeaderChainKey(blockHeight uint64) string {
	blockHeightBytes := []byte(fmt.Sprintf("%d", blockHeight))
	key := append(RelayingBNBHeaderChainPrefix, blockHeightBytes...)
	return string(key) //prefix + blockHeight
}

func StoreRelayingBNBHeaderChain(db incdb.Database, blockHeight uint64, header []byte) error {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	err := db.Put([]byte(key), header)
	if err != nil {
		return NewRawdbError(StoreRelayingBNBHeaderError, err)
	}

	return nil
}

func GetRelayingBNBHeaderChain(db incdb.Database, blockHeight uint64) ([]byte, error) {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	data, err := db.Get([]byte(key))
	if err != nil {
		return nil, NewRawdbError(GetRelayingBNBHeaderError, err)
	}

	return data, nil
}

func GetBNBDataHashByBlockHeight(db incdb.Database, blockHeight uint64) ([]byte, error) {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	data, err := db.Get([]byte(key))
	if err != nil {
		return nil, NewRawdbError(GetRelayingBNBHeaderError, err)
	}

	var bnbBlock types.Block
	err = json.Unmarshal(data, &bnbBlock)
	if err != nil {
		return nil, NewRawdbError(GetRelayingBNBHeaderError, err)
	}

	return bnbBlock.DataHash, nil
}