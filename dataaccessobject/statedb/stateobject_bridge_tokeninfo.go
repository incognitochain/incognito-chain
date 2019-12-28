package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeTokenInfoState struct {
	incTokenID      common.Hash
	externalTokenID []byte
	amount          uint64
	network         string
	isCentralized   bool
}

func (ethtx BridgeTokenInfoState) IncTokenID() common.Hash {
	return ethtx.incTokenID
}

func (ethtx *BridgeTokenInfoState) SetIncTokenID(incTokenID common.Hash) {
	ethtx.incTokenID = incTokenID
}

func (ethtx BridgeTokenInfoState) ExternalTokenID() []byte {
	return ethtx.externalTokenID
}

func (ethtx *BridgeTokenInfoState) SetExternalTokenID(externalTokenID []byte) {
	ethtx.externalTokenID = externalTokenID
}

func (ethtx BridgeTokenInfoState) Amount() uint64 {
	return ethtx.amount
}

func (ethtx *BridgeTokenInfoState) SetAmount(amount uint64) {
	ethtx.amount = amount
}

func (ethtx BridgeTokenInfoState) Network() string {
	return ethtx.network
}

func (ethtx *BridgeTokenInfoState) SetNetwork(network string) {
	ethtx.network = network
}

func (ethtx BridgeTokenInfoState) IsCentralized() bool {
	return ethtx.isCentralized
}

func (ethtx *BridgeTokenInfoState) SetIsCentralized(isCentralized bool) {
	ethtx.isCentralized = isCentralized
}

func (ethtx BridgeTokenInfoState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		IncTokenID      common.Hash
		ExternalTokenID []byte
		Amount          uint64
		Network         string
		IsCentralized   bool
	}{
		IncTokenID:      ethtx.incTokenID,
		ExternalTokenID: ethtx.externalTokenID,
		Amount:          ethtx.amount,
		Network:         ethtx.network,
		IsCentralized:   ethtx.isCentralized,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ethtx *BridgeTokenInfoState) UnmarshalJSON(data []byte) error {
	temp := struct {
		IncTokenID      common.Hash
		ExternalTokenID []byte
		Amount          uint64
		Network         string
		IsCentralized   bool
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ethtx.incTokenID = temp.IncTokenID
	ethtx.externalTokenID = temp.ExternalTokenID
	ethtx.amount = temp.Amount
	ethtx.network = temp.Network
	ethtx.isCentralized = temp.IsCentralized
	return nil
}

func NewBridgeTokenInfoState() *BridgeTokenInfoState {
	return &BridgeTokenInfoState{}
}

func NewBridgeTokenInfoStateWithValue(incTokenID common.Hash, externalTokenID []byte, amount uint64, network string, isCentralized bool) *BridgeTokenInfoState {
	return &BridgeTokenInfoState{incTokenID: incTokenID, externalTokenID: externalTokenID, amount: amount, network: network, isCentralized: isCentralized}
}

type BridgeTokenInfoObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version              int
	bridgeTokenInfoHash  common.Hash
	bridgeTokenInfoState *BridgeTokenInfoState
	objectType           int
	deleted              bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeTokenInfoObject(db *StateDB, hash common.Hash) *BridgeTokenInfoObject {
	return &BridgeTokenInfoObject{
		version:              defaultVersion,
		db:                   db,
		bridgeTokenInfoHash:  hash,
		bridgeTokenInfoState: NewBridgeTokenInfoState(),
		objectType:           BridgeTokenInfoObjectType,
		deleted:              false,
	}
}
func newBridgeTokenInfoObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeTokenInfoObject, error) {
	var newBridgeTokenInfoState = NewBridgeTokenInfoState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeTokenInfoState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeTokenInfoState, ok = data.(*BridgeTokenInfoState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeTokenInfoStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeTokenInfoObject{
		version:              defaultVersion,
		bridgeTokenInfoHash:  key,
		bridgeTokenInfoState: newBridgeTokenInfoState,
		db:                   db,
		objectType:           BridgeTokenInfoObjectType,
		deleted:              false,
	}, nil
}

func GenerateBridgeTokenInfoObjectKey(isCentralized bool, incTokenID common.Hash) common.Hash {
	prefixHash := GetBridgeTokenInfoPrefix(isCentralized)
	valueHash := common.HashH(incTokenID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ethtx BridgeTokenInfoObject) GetVersion() int {
	return ethtx.version
}

// setError remembers the first non-nil error it is called with.
func (ethtx *BridgeTokenInfoObject) SetError(err error) {
	if ethtx.dbErr == nil {
		ethtx.dbErr = err
	}
}

func (ethtx BridgeTokenInfoObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ethtx.trie
}

func (ethtx *BridgeTokenInfoObject) SetValue(data interface{}) error {
	var newBridgeTokenInfoState = NewBridgeTokenInfoState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeTokenInfoState)
		if err != nil {
			return err
		}
	} else {
		newBridgeTokenInfoState, ok = data.(*BridgeTokenInfoState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeTokenInfoStateType, reflect.TypeOf(data))
		}
	}
	ethtx.bridgeTokenInfoState = newBridgeTokenInfoState
	return nil
}

func (ethtx BridgeTokenInfoObject) GetValue() interface{} {
	return ethtx.bridgeTokenInfoState
}

func (ethtx BridgeTokenInfoObject) GetValueBytes() []byte {
	data := ethtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge token info state")
	}
	return []byte(value)
}

func (ethtx BridgeTokenInfoObject) GetHash() common.Hash {
	return ethtx.bridgeTokenInfoHash
}

func (ethtx BridgeTokenInfoObject) GetType() int {
	return ethtx.objectType
}

// MarkDelete will delete an object in trie
func (ethtx *BridgeTokenInfoObject) MarkDelete() {
	ethtx.deleted = true
}

func (ethtx *BridgeTokenInfoObject) Reset() bool {
	ethtx.bridgeTokenInfoState = NewBridgeTokenInfoState()
	return true
}

func (ethtx BridgeTokenInfoObject) IsDeleted() bool {
	return ethtx.deleted
}

// value is either default or nil
func (ethtx BridgeTokenInfoObject) IsEmpty() bool {
	temp := NewBridgeTokenInfoState()
	return reflect.DeepEqual(temp, ethtx.bridgeTokenInfoState) || ethtx.bridgeTokenInfoState == nil
}
