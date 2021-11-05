package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3Infos struct {
	liquidityMintedEpochs uint64
}

func (pi Pdexv3Infos) LiquidityMintedEpochs() uint64 {
	return pi.liquidityMintedEpochs
}

func (pi Pdexv3Infos) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		LiquidityMintedEpochs uint64
	}{
		LiquidityMintedEpochs: pi.liquidityMintedEpochs,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pi *Pdexv3Infos) UnmarshalJSON(data []byte) error {
	temp := struct {
		LiquidityMintedEpochs uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pi.liquidityMintedEpochs = temp.LiquidityMintedEpochs
	return nil
}

func NewPdexv3Infos() *Pdexv3Infos {
	return &Pdexv3Infos{}
}

func NewPdexv3InfosWithValue(
	liquidityMintedEpochs uint64,
) *Pdexv3Infos {
	return &Pdexv3Infos{
		liquidityMintedEpochs: liquidityMintedEpochs,
	}
}

type Pdexv3InfosObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version         int
	pdexv3InfosHash common.Hash
	Pdexv3Infos     *Pdexv3Infos
	objectType      int
	deleted         bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3InfosObject(db *StateDB, hash common.Hash) *Pdexv3InfosObject {
	return &Pdexv3InfosObject{
		version:         defaultVersion,
		db:              db,
		pdexv3InfosHash: hash,
		Pdexv3Infos:     NewPdexv3Infos(),
		objectType:      Pdexv3InfosObjectType,
		deleted:         false,
	}
}

func newPdexv3InfosObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3InfosObject, error) {
	var newPdexv3Infos = NewPdexv3Infos()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3Infos)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3Infos, ok = data.(*Pdexv3Infos)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3InfosStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3InfosObject{
		version:         defaultVersion,
		pdexv3InfosHash: key,
		Pdexv3Infos:     newPdexv3Infos,
		db:              db,
		objectType:      Pdexv3InfosObjectType,
		deleted:         false,
	}, nil
}

func GeneratePdexv3InfosObjectKey() common.Hash {
	prefixHash := GetPdexv3InfosPrefix()
	return common.HashH(prefixHash)
}

func (t Pdexv3InfosObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *Pdexv3InfosObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t Pdexv3InfosObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *Pdexv3InfosObject) SetValue(data interface{}) error {
	newPdexv3Infos, ok := data.(*Pdexv3Infos)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3InfosStateType, reflect.TypeOf(data))
	}
	t.Pdexv3Infos = newPdexv3Infos
	return nil
}

func (t Pdexv3InfosObject) GetValue() interface{} {
	return t.Pdexv3Infos
}

func (t Pdexv3InfosObject) GetValueBytes() []byte {
	Pdexv3Infos, ok := t.GetValue().(*Pdexv3Infos)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(Pdexv3Infos)
	if err != nil {
		panic("failed to marshal pdex v3 infos state")
	}
	return value
}

func (t Pdexv3InfosObject) GetHash() common.Hash {
	return t.pdexv3InfosHash
}

func (t Pdexv3InfosObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *Pdexv3InfosObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *Pdexv3InfosObject) Reset() bool {
	t.Pdexv3Infos = NewPdexv3Infos()
	return true
}

func (t Pdexv3InfosObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t Pdexv3InfosObject) IsEmpty() bool {
	temp := NewPdexv3Infos()
	return reflect.DeepEqual(temp, t.Pdexv3Infos) || t.Pdexv3Infos == nil
}
