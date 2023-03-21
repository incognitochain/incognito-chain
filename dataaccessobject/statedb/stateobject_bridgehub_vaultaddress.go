package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubNetworkVaultState struct {
	vaultAddress string
	networkId    int
}

func (b BridgeHubNetworkVaultState) NetworkId() int {
	return b.networkId
}

func (b *BridgeHubNetworkVaultState) SetNetworkId(networkId int) {
	b.networkId = networkId
}

func (b BridgeHubNetworkVaultState) VaultAddress() string {
	return b.vaultAddress
}

func (b *BridgeHubNetworkVaultState) SetVaultAddress(vaultAddress string) {
	b.vaultAddress = vaultAddress
}

func (b BridgeHubNetworkVaultState) Clone() *BridgeHubNetworkVaultState {
	return &BridgeHubNetworkVaultState{
		vaultAddress: b.vaultAddress,
		networkId:    b.networkId,
	}
}

func (b *BridgeHubNetworkVaultState) IsDiff(compareParam *BridgeHubNetworkVaultState) bool {
	if compareParam == nil {
		return true
	}
	return b.vaultAddress != compareParam.vaultAddress || b.networkId != compareParam.networkId
}

func (b BridgeHubNetworkVaultState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		VaultAddress string
		NetworkId    int
	}{
		VaultAddress: b.vaultAddress,
		NetworkId:    b.networkId,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubNetworkVaultState) UnmarshalJSON(data []byte) error {
	temp := struct {
		VaultAddress string
		NetworkId    int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.vaultAddress = temp.VaultAddress
	b.networkId = temp.NetworkId
	return nil
}

func NewBridgeHubNetworkVaultState() *BridgeHubNetworkVaultState {
	return &BridgeHubNetworkVaultState{}
}

func NewBridgeHubNetworkVaultStateWithValue(vaultAddress string, networkId int) *BridgeHubNetworkVaultState {
	return &BridgeHubNetworkVaultState{
		vaultAddress: vaultAddress,
		networkId:    networkId,
	}
}

type BridgeHubNetworkVaultObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubNetworkVaultState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubNetworkVaultState(db *StateDB, hash common.Hash) *BridgeHubNetworkVaultObject {
	return &BridgeHubNetworkVaultObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubNetworkVaultState(),
		objectType: BridgeHubVaultObjectType,
		deleted:    false,
	}
}

func newBridgeHubNetworkVaultObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubNetworkVaultObject, error) {
	var newBridgePToken = NewBridgeHubNetworkVaultState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePToken)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgePToken, ok = data.(*BridgeHubNetworkVaultState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubVaultStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubNetworkVaultObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgePToken,
		db:         db,
		objectType: BridgeHubVaultObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubVaultObjectKey(bridgeID string, networkId int) common.Hash {
	prefixHash := GetBridgeHubVaultPrefix([]byte(bridgeID))
	valueHash := common.HashH(common.IntToBytes(networkId))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubNetworkVaultObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubNetworkVaultObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubNetworkVaultObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubNetworkVaultObject) SetValue(data interface{}) error {
	newBridgeHubPToken, ok := data.(*BridgeHubNetworkVaultState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubVaultStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubPToken
	return nil
}

func (t BridgeHubNetworkVaultObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubNetworkVaultObject) GetValueBytes() []byte {
	bridgeHubPTokenState, ok := t.GetValue().(*BridgeHubNetworkVaultState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubPTokenState)
	if err != nil {
		panic("failed to marshal BridgeHubPTokenState")
	}
	return value
}

func (t BridgeHubNetworkVaultObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubNetworkVaultObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubNetworkVaultObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubNetworkVaultObject) Reset() bool {
	t.state = NewBridgeHubNetworkVaultState()
	return true
}

func (t BridgeHubNetworkVaultObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubNetworkVaultObject) IsEmpty() bool {
	return t.state == nil
}
