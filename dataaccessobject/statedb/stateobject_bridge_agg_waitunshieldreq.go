package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggWaitingUnshieldReq struct {
	unshieldID   common.Hash
	data         []WaitingUnshieldReqData
	beaconHeight uint64
}

type WaitingUnshieldReqData struct {
	IncTokenID    common.Hash `json:"IncTokenID"`
	BurningAmount uint64      `json:"BurningAmount"`
	RemoteAddress string      `json:"RemoteAddress"`
	Fee           uint64      `json:"Fee"`
}

func (w WaitingUnshieldReqData) Clone() WaitingUnshieldReqData {
	return WaitingUnshieldReqData{
		IncTokenID:    w.IncTokenID,
		BurningAmount: w.BurningAmount,
		RemoteAddress: w.RemoteAddress,
		Fee:           w.Fee,
	}
}

func (us *BridgeAggWaitingUnshieldReq) Clone() *BridgeAggWaitingUnshieldReq {
	clonedData := []WaitingUnshieldReqData{}
	clonedData = append(clonedData, us.data...)
	cloned := &BridgeAggWaitingUnshieldReq{
		unshieldID:   us.unshieldID,
		data:         clonedData,
		beaconHeight: us.beaconHeight,
	}
	return cloned
}

func (us *BridgeAggWaitingUnshieldReq) GetData() []WaitingUnshieldReqData {
	return us.data
}

func (us *BridgeAggWaitingUnshieldReq) SetData(data []WaitingUnshieldReqData) {
	us.data = data
}

func (us *BridgeAggWaitingUnshieldReq) GetUnshieldID() common.Hash {
	return us.unshieldID
}

func (us *BridgeAggWaitingUnshieldReq) SetUnshieldID(unshieldID common.Hash) {
	us.unshieldID = unshieldID
}

func (us *BridgeAggWaitingUnshieldReq) GetBeaconHeight() uint64 {
	return us.beaconHeight
}

func (us *BridgeAggWaitingUnshieldReq) SetBeaconHeight(beaconHeight uint64) {
	us.beaconHeight = beaconHeight
}

func (us BridgeAggWaitingUnshieldReq) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UnshieldID   common.Hash              `json:"UnshieldID"`
		Data         []WaitingUnshieldReqData `json:"Data"`
		BeaconHeight uint64                   `json:"BeaconHeight"`
	}{
		UnshieldID:   us.unshieldID,
		Data:         us.data,
		BeaconHeight: us.beaconHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (us *BridgeAggWaitingUnshieldReq) UnmarshalJSON(data []byte) error {
	temp := struct {
		UnshieldID   common.Hash              `json:"UnshieldID"`
		Data         []WaitingUnshieldReqData `json:"Data"`
		BeaconHeight uint64                   `json:"BeaconHeight"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	us.unshieldID = temp.UnshieldID
	us.data = temp.Data
	us.beaconHeight = temp.BeaconHeight
	return nil
}

func NewBridgeAggWaitingUnshieldReqStateWithValue(
	data []WaitingUnshieldReqData,
	unshieldID common.Hash,
	beaconHeight uint64,
) *BridgeAggWaitingUnshieldReq {
	return &BridgeAggWaitingUnshieldReq{
		unshieldID:   unshieldID,
		data:         data,
		beaconHeight: beaconHeight,
	}
}

func NewBridgeAggWaitingUnshieldReqState() *BridgeAggWaitingUnshieldReq {
	return &BridgeAggWaitingUnshieldReq{}
}

type BridgeAggWaitingUnshieldReqObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                         int
	BridgeAggWaitingUnshieldReqHash common.Hash
	BridgeAggWaitingUnshieldReq     *BridgeAggWaitingUnshieldReq
	objectType                      int
	deleted                         bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggWaitingUnshieldReqObject(db *StateDB, hash common.Hash) *BridgeAggWaitingUnshieldReqObject {
	return &BridgeAggWaitingUnshieldReqObject{
		version:                         defaultVersion,
		db:                              db,
		BridgeAggWaitingUnshieldReqHash: hash,
		BridgeAggWaitingUnshieldReq:     NewBridgeAggWaitingUnshieldReqState(),
		objectType:                      BridgeAggWaitingUnshieldReqObjectType,
		deleted:                         false,
	}
}

func newBridgeAggWaitingUnshieldReqObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeAggWaitingUnshieldReqObject, error) {
	var content = NewBridgeAggWaitingUnshieldReqState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, content)
		if err != nil {
			return nil, err
		}
	} else {
		content, ok = data.(*BridgeAggWaitingUnshieldReq)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggWaitingUnshieldReqType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggWaitingUnshieldReqObject{
		version:                         defaultVersion,
		BridgeAggWaitingUnshieldReqHash: key,
		BridgeAggWaitingUnshieldReq:     content,
		db:                              db,
		objectType:                      BridgeAggWaitingUnshieldReqObjectType,
		deleted:                         false,
	}, nil
}

func GenerateBridgeAggWaitingUnshieldReqObjectKey(unifiedTokenID common.Hash, unshieldID common.Hash) common.Hash {
	prefixHash := GetBridgeAggWaitingUnshieldReqPrefix(unifiedTokenID.Bytes())
	valueHash := common.HashH(unshieldID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeAggWaitingUnshieldReqObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeAggWaitingUnshieldReqObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeAggWaitingUnshieldReqObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeAggWaitingUnshieldReqObject) SetValue(data interface{}) error {
	waitingUnshieldReq, ok := data.(*BridgeAggWaitingUnshieldReq)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggWaitingUnshieldReqType, reflect.TypeOf(data))
	}
	t.BridgeAggWaitingUnshieldReq = waitingUnshieldReq
	return nil
}

func (t BridgeAggWaitingUnshieldReqObject) GetValue() interface{} {
	return t.BridgeAggWaitingUnshieldReq
}

func (t BridgeAggWaitingUnshieldReqObject) GetValueBytes() []byte {
	waitingUnshieldReq, ok := t.GetValue().(*BridgeAggWaitingUnshieldReq)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(waitingUnshieldReq)
	if err != nil {
		panic("failed to marshal redeem request")
	}
	return value
}

func (t BridgeAggWaitingUnshieldReqObject) GetHash() common.Hash {
	return t.BridgeAggWaitingUnshieldReqHash
}

func (t BridgeAggWaitingUnshieldReqObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeAggWaitingUnshieldReqObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeAggWaitingUnshieldReqObject) Reset() bool {
	t.BridgeAggWaitingUnshieldReq = NewBridgeAggWaitingUnshieldReqState()
	return true
}

func (t BridgeAggWaitingUnshieldReqObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeAggWaitingUnshieldReqObject) IsEmpty() bool {
	temp := NewBridgeAggWaitingUnshieldReqState()
	return reflect.DeepEqual(temp, t.BridgeAggWaitingUnshieldReq) || t.BridgeAggWaitingUnshieldReq == nil
}
