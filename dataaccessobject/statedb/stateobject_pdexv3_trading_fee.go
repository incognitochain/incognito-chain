package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3TradingFeeState struct {
	tokenID common.Hash
	amount  uint64
}

func (pt *Pdexv3TradingFeeState) TokenID() common.Hash {
	return pt.tokenID
}

func (pt *Pdexv3TradingFeeState) Amount() uint64 {
	return pt.amount
}

func (pt *Pdexv3TradingFeeState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash `json:"TokenID"`
		Amount  uint64      `json:"Amount"`
	}{
		TokenID: pt.tokenID,
		Amount:  pt.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pt *Pdexv3TradingFeeState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash `json:"TokenID"`
		Amount  uint64      `json:"Amount"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pt.tokenID = temp.TokenID
	pt.amount = temp.Amount
	return nil
}

func (pt *Pdexv3TradingFeeState) Clone() *Pdexv3TradingFeeState {
	return &Pdexv3TradingFeeState{
		tokenID: pt.tokenID,
		amount:  pt.amount,
	}
}

func NewPdexv3TradingFeeState() *Pdexv3TradingFeeState {
	return &Pdexv3TradingFeeState{}
}

func NewPdexv3TradingFeeStateWithValue(
	tokenID common.Hash, amount uint64,
) *Pdexv3TradingFeeState {
	return &Pdexv3TradingFeeState{
		tokenID: tokenID,
		amount:  amount,
	}
}

type Pdexv3TradingFeeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3TradingFeeState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3TradingFeeObject(db *StateDB, hash common.Hash) *Pdexv3TradingFeeObject {
	return &Pdexv3TradingFeeObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3TradingFeeState(),
		objectType: Pdexv3TradingFeeObjectType,
		deleted:    false,
	}
}

func newPdexv3TradingFeeObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3TradingFeeObject, error,
) {
	var newPdexv3TradingFeeState = NewPdexv3TradingFeeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3TradingFeeState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3TradingFeeState, ok = data.(*Pdexv3TradingFeeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3TradingFeetateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3TradingFeeObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3TradingFeeState,
		db:         db,
		objectType: Pdexv3TradingFeeObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3TradingFeesObjectPrefix(poolPairID, nfctID string) []byte {
	str := string(GetPdexv3TradingFeesPrefix()) + "-" + poolPairID + "-" + nfctID
	temp := []byte(str)
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}

func GeneratePdexv3TradingFeesObjectKey(poolPairID, nfctID, tokenID string) common.Hash {
	prefixHash := generatePdexv3TradingFeesObjectPrefix(poolPairID, nfctID)
	valueHash := common.HashH([]byte(nfctID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (pt *Pdexv3TradingFeeObject) GetVersion() int {
	return pt.version
}

// setError remembers the first non-nil error it is called with.
func (pt *Pdexv3TradingFeeObject) SetError(err error) {
	if pt.dbErr == nil {
		pt.dbErr = err
	}
}

func (pt *Pdexv3TradingFeeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pt.trie
}

func (pt *Pdexv3TradingFeeObject) SetValue(data interface{}) error {
	newPdexv3TradingFeeState, ok := data.(*Pdexv3TradingFeeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3TradingFeetateType, reflect.TypeOf(data))
	}
	pt.state = newPdexv3TradingFeeState
	return nil
}

func (pt *Pdexv3TradingFeeObject) GetValue() interface{} {
	return pt.state
}

func (pt *Pdexv3TradingFeeObject) GetValueBytes() []byte {
	state, ok := pt.GetValue().(*Pdexv3TradingFeeObject)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 contribution state")
	}
	return value
}

func (pt *Pdexv3TradingFeeObject) GetHash() common.Hash {
	return pt.hash
}

func (pt *Pdexv3TradingFeeObject) GetType() int {
	return pt.objectType
}

// MarkDelete will delete an object in trie
func (pt *Pdexv3TradingFeeObject) MarkDelete() {
	pt.deleted = true
}

// reset all shard committee value into default value
func (pt *Pdexv3TradingFeeObject) Reset() bool {
	pt.state = NewPdexv3TradingFeeState()
	return true
}

func (pt *Pdexv3TradingFeeObject) IsDeleted() bool {
	return pt.deleted
}

// value is either default or nil
func (pt *Pdexv3TradingFeeObject) IsEmpty() bool {
	temp := NewPdexv3TradingFeeState()
	return reflect.DeepEqual(temp, pt.state) || pt.state == nil
}
