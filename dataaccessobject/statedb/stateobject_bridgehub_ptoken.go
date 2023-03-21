package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubNetworkState struct {
	pTokenID     common.Hash
	pTokenAmount uint64 // pTokenID : amount
	vaultAddress string
	networkId    int
}

func (b BridgeHubNetworkState) NetworkId() int {
	return b.networkId
}

func (b *BridgeHubNetworkState) SetNetworkId(networkId int) {
	b.networkId = networkId
}

func (b BridgeHubNetworkState) PTokenAmount() uint64 {
	return b.pTokenAmount
}

func (b *BridgeHubNetworkState) SetPTokenAmount(pTokenAmount uint64) {
	b.pTokenAmount = pTokenAmount
}

func (b BridgeHubNetworkState) PTokenID() common.Hash {
	return b.pTokenID
}

func (b *BridgeHubNetworkState) SetPTokenID(pTokenID common.Hash) {
	b.pTokenID = pTokenID
}

func (b BridgeHubNetworkState) VaultAddress() string {
	return b.vaultAddress
}

func (b *BridgeHubNetworkState) SetVaultAddress(vaultAddress string) {
	b.vaultAddress = vaultAddress
}

func (b BridgeHubNetworkState) Clone() *BridgeHubNetworkState {
	return &BridgeHubNetworkState{
		pTokenID:     b.pTokenID,
		pTokenAmount: b.pTokenAmount,
		vaultAddress: b.vaultAddress,
		networkId:    b.networkId,
	}
}

func (b *BridgeHubNetworkState) IsDiff(compareParam *BridgeHubNetworkState) bool {
	if compareParam == nil {
		return true
	}
	return b.pTokenAmount != compareParam.pTokenAmount || b.pTokenID != compareParam.pTokenID || b.vaultAddress != compareParam.vaultAddress || b.networkId != compareParam.networkId
}

func (b BridgeHubNetworkState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PTokenID     common.Hash
		PTokenAmount uint64
		VaultAddress string
		NetworkId    int
	}{
		PTokenID:     b.pTokenID,
		PTokenAmount: b.pTokenAmount,
		VaultAddress: b.vaultAddress,
		NetworkId:    b.networkId,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubNetworkState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PTokenID     common.Hash
		PTokenAmount uint64
		VaultAddress string
		NetworkId    int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.pTokenID = temp.PTokenID
	b.pTokenAmount = temp.PTokenAmount
	b.vaultAddress = temp.VaultAddress
	b.networkId = temp.NetworkId
	return nil
}

func NewBridgeHubNetworkState() *BridgeHubNetworkState {
	return &BridgeHubNetworkState{}
}

func NewBridgeHubNetworkStateWithValue(pTokenAmount uint64, pTokenID common.Hash, vaultAddress string, networkId int) *BridgeHubNetworkState {
	return &BridgeHubNetworkState{
		pTokenID:     pTokenID,
		pTokenAmount: pTokenAmount,
		vaultAddress: vaultAddress,
		networkId:    networkId,
	}
}

type BridgeHubNetworkObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubNetworkState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubNetworkObject(db *StateDB, hash common.Hash) *BridgeHubNetworkObject {
	return &BridgeHubNetworkObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubNetworkState(),
		objectType: BridgeHubBridgeInfoNetworkObjectType,
		deleted:    false,
	}
}

func newBridgeHubNetworkObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubNetworkObject, error) {
	var newBridgePToken = NewBridgeHubNetworkState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePToken)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgePToken, ok = data.(*BridgeHubNetworkState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubNetworkObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgePToken,
		db:         db,
		objectType: BridgeHubBridgeInfoNetworkObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubPTokenObjectKey(bridgeID string, networkId int) common.Hash {
	prefixHash := GetBridgeHubPTokenPrefix([]byte(bridgeID))
	valueHash := common.HashH(common.IntToBytes(networkId))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubNetworkObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubNetworkObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubNetworkObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubNetworkObject) SetValue(data interface{}) error {
	newBridgeHubPToken, ok := data.(*BridgeHubNetworkState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubPToken
	return nil
}

func (t BridgeHubNetworkObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubNetworkObject) GetValueBytes() []byte {
	bridgeHubPTokenState, ok := t.GetValue().(*BridgeHubNetworkState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubPTokenState)
	if err != nil {
		panic("failed to marshal BridgeHubPTokenState")
	}
	return value
}

func (t BridgeHubNetworkObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubNetworkObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubNetworkObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubNetworkObject) Reset() bool {
	t.state = NewBridgeHubNetworkState()
	return true
}

func (t BridgeHubNetworkObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubNetworkObject) IsEmpty() bool {
	return t.state == nil
}
