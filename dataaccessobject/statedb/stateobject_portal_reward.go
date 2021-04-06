package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type PortalRewardInfo struct {
	rewards map[string]uint64 // tokenID : amount
}

func (p PortalRewardInfo) GetRewards() map[string]uint64 {
	return p.rewards
}

func (p *PortalRewardInfo) SetRewards(rewards map[string]uint64) {
	p.rewards = rewards
}

func (p *PortalRewardInfo) AddPortalRewardInfo(tokenID string, amount uint64) {
	if p.rewards == nil {
		p.rewards = make(map[string]uint64)
		p.rewards[tokenID] = amount
		return
	}
	p.rewards[tokenID] += amount
}

func (p PortalRewardInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Rewards map[string]uint64
	}{
		Rewards: p.rewards,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (p *PortalRewardInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		Rewards map[string]uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	p.rewards = temp.Rewards
	return nil
}

func NewPortalRewardInfo() *PortalRewardInfo {
	return &PortalRewardInfo{}
}

func NewPortalRewardInfoWithValue(
	rewards map[string]uint64) *PortalRewardInfo {

	return &PortalRewardInfo{
		rewards: rewards,
	}
}

type PortalRewardInfoObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version              int
	portalRewardInfoHash common.Hash
	portalRewardInfo     *PortalRewardInfo
	objectType           int
	deleted              bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalRewardInfoObject(db *StateDB, hash common.Hash) *PortalRewardInfoObject {
	return &PortalRewardInfoObject{
		version:              defaultVersion,
		db:                   db,
		portalRewardInfoHash: hash,
		portalRewardInfo:     NewPortalRewardInfo(),
		objectType:           PortalRewardInfoObjectType,
		deleted:              false,
	}
}

func newPortalRewardInfoObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PortalRewardInfoObject, error) {
	var portalRewardInfo = NewPortalRewardInfo()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, portalRewardInfo)
		if err != nil {
			return nil, err
		}
	} else {
		portalRewardInfo, ok = data.(*PortalRewardInfo)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalRewardInfoStateType, reflect.TypeOf(data))
		}
	}
	return &PortalRewardInfoObject{
		version:              defaultVersion,
		portalRewardInfoHash: key,
		portalRewardInfo:     portalRewardInfo,
		db:                   db,
		objectType:           PortalRewardInfoObjectType,
		deleted:              false,
	}, nil
}

func GeneratePortalRewardInfoObjectKey(beaconHeight uint64, custodianIncognitoAddress string) common.Hash {
	prefixHash := GetPortalRewardInfoStatePrefix(beaconHeight)
	valueHash := common.HashH([]byte(custodianIncognitoAddress))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PortalRewardInfoObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PortalRewardInfoObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PortalRewardInfoObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PortalRewardInfoObject) SetValue(data interface{}) error {
	portalRewardInfo, ok := data.(*PortalRewardInfo)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalRewardInfoStateType, reflect.TypeOf(data))
	}
	t.portalRewardInfo = portalRewardInfo
	return nil
}

func (t PortalRewardInfoObject) GetValue() interface{} {
	return t.portalRewardInfo
}

func (t PortalRewardInfoObject) GetValueBytes() []byte {
	portalRewardInfo, ok := t.GetValue().(*PortalRewardInfo)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(portalRewardInfo)
	if err != nil {
		panic("failed to marshal portal reward info")
	}
	return value
}

func (t PortalRewardInfoObject) GetHash() common.Hash {
	return t.portalRewardInfoHash
}

func (t PortalRewardInfoObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PortalRewardInfoObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PortalRewardInfoObject) Reset() bool {
	t.portalRewardInfo = NewPortalRewardInfo()
	return true
}

func (t PortalRewardInfoObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PortalRewardInfoObject) IsEmpty() bool {
	temp := NewPortalRewardInfo()
	return reflect.DeepEqual(temp, t.portalRewardInfo) || t.portalRewardInfo == nil
}
