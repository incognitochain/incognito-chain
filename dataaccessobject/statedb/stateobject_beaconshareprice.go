package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type BeaconSharePrice struct {
	price uint64
}

func NewBeaconSharePrice() *BeaconSharePrice {
	return &BeaconSharePrice{}
}
func NewBeaconSharePriceWithValue(price uint64) *BeaconSharePrice {
	return &BeaconSharePrice{price}
}

func (s BeaconSharePrice) GetPrice() uint64 {
	return s.price
}

type BeaconSharePriceObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version         int
	beaconShareHash common.Hash
	shareInfo       *BeaconSharePrice
	objectType      int
	deleted         bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBeaconSharePriceObject(db *StateDB, hash common.Hash) *BeaconSharePriceObject {
	return &BeaconSharePriceObject{
		version:         defaultVersion,
		db:              db,
		beaconShareHash: hash,
		shareInfo:       &BeaconSharePrice{},
		objectType:      BeaconSharePriceType,
		deleted:         false,
	}
}

func newBeaconSharePriceWithValue(db *StateDB, key common.Hash, data interface{}) (*BeaconSharePriceObject, error) {
	var newSharePrice = NewBeaconSharePrice()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newSharePrice)
		if err != nil {
			return nil, err
		}
	} else {
		newSharePrice, ok = data.(*BeaconSharePrice)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidSharePriceType, reflect.TypeOf(data))
		}
	}
	return &BeaconSharePriceObject{
		version:         defaultVersion,
		beaconShareHash: key,
		shareInfo:       newSharePrice,
		db:              db,
		objectType:      BeaconSharePriceType,
		deleted:         false,
	}, nil
}

func (c BeaconSharePriceObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *BeaconSharePriceObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c BeaconSharePriceObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *BeaconSharePriceObject) SetValue(data interface{}) error {
	sharePrice, ok := data.(*BeaconSharePrice)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidSharePriceType, reflect.TypeOf(data))
	}

	c.shareInfo = sharePrice
	return nil
}

func (c BeaconSharePriceObject) GetValue() interface{} {
	return c.shareInfo
}

func (c BeaconSharePriceObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c BeaconSharePriceObject) GetHash() common.Hash {
	return c.beaconShareHash
}

func (c BeaconSharePriceObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *BeaconSharePriceObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *BeaconSharePriceObject) Reset() bool {
	c.shareInfo = NewBeaconSharePrice()
	return true
}

func (c BeaconSharePriceObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c BeaconSharePriceObject) IsEmpty() bool {
	temp := NewBeaconSharePrice()
	return reflect.DeepEqual(temp, c.shareInfo) || c.shareInfo == nil
}
