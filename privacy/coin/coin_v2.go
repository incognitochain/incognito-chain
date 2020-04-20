package coin

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

// PlainCoin is the struct that we use when we process, it contains explicitly what a coin should have
type PlainCoinV2 struct {
	version    uint8
	shardID    uint8
	index      uint8
	value      uint64
	publicKey  *operation.Point
	randomness *operation.Scalar // r (not rG)
	commitment *operation.Point
	keyImage   *operation.Point
	info       []byte
}

// Init (Coin) initializes a coin
func (c *PlainCoinV2) Init() *PlainCoinV2 {
	if c == nil {
		c = new(PlainCoinV1)
	}
	c.version = 2
	c.shardID = 0
	c.index = 0
	c.value = 0
	c.publicKey = new(operation.Point).Identity()
	c.randomness = new(operation.Scalar)
	c.commitment = new(operation.Point).Identity()
	c.keyImage = new(operation.Point).Identity()
	c.info = []byte{}
	return c
}

func (c PlainCoinV2) GetVersion() uint8               { return 2 }
func (c PlainCoinV2) GetShardID() uint8               { return c.shardId }
func (c PlainCoinV2) GetIndex() uint8                 { return c.index }
func (c PlainCoinV2) GetValue() uint64                { return c.value }
func (c PlainCoinV2) GetPublicKey() *operation.Point  { return c.publicKey }
func (c PlainCoinV2) GetCommitment() *operation.Point { return c.commitment }
func (c PlainCoinV2) GetInfo() []byte                 { return c.info }
func (c PlainCoinV2) GetKeyImage() *operation.Point   { return c.keyImage }

func (c *PlainCoinV2) SetVersion(version uint8)                  { c.version = 2 }
func (c *PlainCoinV2) SetShardID(shardId uint8)                  { c.shardId = shardId }
func (c *PlainCoinV2) SetIndex(index uint8)                      { c.index = index }
func (c *PlainCoinV2) SetValue(value uint64)                     { c.value = value }
func (c *PlainCoinV2) SetPublicKey(publicKey *operation.Point)   { c.publicKey = publicKey }
func (c *PlainCoinV2) SetRandomness(r *operation.Scalar)         { c.randomness = r }
func (c *PlainCoinV2) SetCommitment(commitment *operation.Point) { c.commitment = commitment }
func (c *PlainCoinV2) SetKeyImage(keyImage *operation.Point)     { c.keyImage = keyImage }
func (c *PlainCoinV2) SetInfo(v []byte) {
	c.info = make([]byte, len(v))
	copy(c.info, v)
}

func (c PlainCoinV2) Bytes() []byte {
	return nil
}

func (c *PlainCoinV2) SetBytes(b []byte) error {
	return nil
}

// CoinV2 is the struct that will be stored to db
// When store to db, raw will be nil
type CoinV2 struct {
	// SetBytes and FromBytes of CoinV2 will use this first byte as version
	version    uint8
	shardID    uint8
	mask       *operation.Scalar
	amount     *operation.Scalar
	txRandom   *operation.Point // rG
	publicKey  *operation.Point // R^o = H_n(r * K_B^v, index)G + K_B^s
	commitment *operation.Point
	keyImage   *operation.Point

	raw   *PlainCoinV2
	index uint8
	info  []byte //256 bytes
}

// Init (OutputCoin) initializes a output coin
func (c *CoinV2) Init() *CoinV2 {
	if c == nil {
		c = new(CoinV1)
	}
	c.version = 2
	c.shardID = 0
	c.mask = new(operation.Scalar)
	c.amount = new(operation.Scalar)
	c.txRandom = new(operation.Point).Identity()
	c.publicKey = new(operation.Point).Identity()
	c.commitment = new(operation.Point).Identity()
	c.keyImage = new(operation.Point).Identity()
	c.raw = new(PlainCoinV2).Init()
	c.index = 0
	c.info = []byte{}
	return c
}

// Get SND will be nil for ver 2
func (c CoinV2) GetSNDerivator() *operation.Scalar { return nil }

func (c CoinV2) GetVersion() uint8               { return 2 }
func (c CoinV2) GetShardID() uint8               { return c.shardID }
func (c CoinV2) GetMask() *operation.Scalar      { return c.mask }
func (c CoinV2) GetAmount() *operation.Scalar    { return c.amount }
func (c CoinV2) GetTxRandom() *operation.Point   { return c.txRandom }
func (c CoinV2) GetPublicKey() *operation.Point  { return c.publicKey }
func (c CoinV2) GetCommitment() *operation.Point { return c.commitment }
func (c CoinV2) GetKeyImage() *operation.Point   { return c.keyImage }
func (c CoinV2) GetRawCoin() *PlainCoinV2        { return c.raw }
func (c CoinV2) GetIndex() *uint8                { return c.index }
func (c CoinV2) GetInfo() []byte                 { return c.info }

func (c *CoinV2) SetVersion(uint8)                          { c.version = 2 }
func (c *CoinV2) SetMask(mask *operation.Scalar)            { c.mask = mask }
func (c *CoinV2) SetShardId(shardId uint8)                  { c.shardId = shardId }
func (c *CoinV2) SetAmount(amount *operation.Scalar)        { c.amount = amount }
func (c *CoinV2) SetTxRandom(txRandom *operation.Point)     { c.txRandom = txRandom }
func (c *CoinV2) SetPublicKey(publicKey *operation.Point)   { c.publicKey = publicKey }
func (c *CoinV2) SetCommitment(commitment *operation.Point) { c.commitment = commitment }
func (c *CoinV2) SetKeyImage(keyImage *operation.Point)     { c.keyImage = keyImage }
func (c *CoinV2) SetIndex(index uint8)                      { c.index = index }
func (c *CoinV2) SetInfo(b []byte) error {
	if len(b) > MaxSizeInfoCoin {
		return errors.New("Cannot set info to CoinV2, info is longer than 255")
	}
	c.info = make([]byte, len(b))
	copy(c.info, b)
	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
// TODO Privacy
func (this *CoinV2) Bytes() []byte {
	return nil
}

func (this *CoinV2) SetBytes(coinBytes []byte) error {
	if this == nil {
		this = new(CoinV2)
	}
	return nil
}

// HashH returns the SHA3-256 hashing of coin bytes array
func (this *CoinV2) HashH() *common.Hash {
	hash := common.HashH(this.Bytes())
	return &hash
}
