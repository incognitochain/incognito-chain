package coin

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

// CoinV2 is the struct that will be stored to db
// If not privacy, mask and amount will be the original randomness and value
// If has privacy, mask and amount will be as paper monero
type CoinV2 struct {
	// Public
	version    uint8
	shardID    uint8
	index      uint8
	info       []byte
	publicKey  *operation.Point
	commitment *operation.Point
	keyImage   *operation.Point
	txRandom   *operation.Point // rG

	// mask = randomness
	// amount = value
	mask   *operation.Scalar
	amount *operation.Scalar
}

// Init (OutputCoin) initializes a output coin
func (c *CoinV2) Init() *CoinV2 {
	if c == nil {
		c = new(CoinV2)
	}
	c.version = 2
	c.shardID = 0
	c.index = 0
	c.info = []byte{}
	c.publicKey = new(operation.Point).Identity()
	c.commitment = new(operation.Point).Identity()
	c.keyImage = new(operation.Point).Identity()
	c.txRandom = new(operation.Point).Identity()
	c.mask = new(operation.Scalar)
	c.amount = new(operation.Scalar)
	return c
}

// Get SND will be nil for ver 2
func (c CoinV2) GetSNDerivator() *operation.Scalar { return nil }

func (c CoinV2) IsEncrypted() bool {
	commitment := operation.PedCom.CommitAtIndex(c.amount, c.mask, operation.PedersenRandomnessIndex)
	return !operation.IsPointEqual(commitment, c.commitment)
}

func (c CoinV2) GetVersion() uint8               { return 2 }
func (c CoinV2) GetShardID() uint8               { return c.shardID }
func (c CoinV2) GetMask() *operation.Scalar      { return c.mask }
func (c CoinV2) GetAmount() *operation.Scalar    { return c.amount }
func (c CoinV2) GetTxRandom() *operation.Point   { return c.txRandom }
func (c CoinV2) GetPublicKey() *operation.Point  { return c.publicKey }
func (c CoinV2) GetCommitment() *operation.Point { return c.commitment }
func (c CoinV2) GetKeyImage() *operation.Point   { return c.keyImage }
func (c CoinV2) GetIndex() uint8                 { return c.index }
func (c CoinV2) GetInfo() []byte                 { return c.info }
func (c CoinV2) GetValue() uint64                { return c.amount.ToUint64() }

func (c *CoinV2) SetVersion(uint8)                          { c.version = 2 }
func (c *CoinV2) SetMask(mask *operation.Scalar)            { c.mask = mask }
func (c *CoinV2) SetShardID(shardID uint8)                  { c.shardID = shardID }
func (c *CoinV2) SetAmount(amount *operation.Scalar)        { c.amount = amount }
func (c *CoinV2) SetTxRandom(txRandom *operation.Point)     { c.txRandom = txRandom }
func (c *CoinV2) SetPublicKey(publicKey *operation.Point)   { c.publicKey = publicKey }
func (c *CoinV2) SetCommitment(commitment *operation.Point) { c.commitment = commitment }
func (c *CoinV2) SetKeyImage(keyImage *operation.Point)     { c.keyImage = keyImage }
func (c *CoinV2) SetIndex(index uint8)                      { c.index = index }
func (c *CoinV2) SetValue(value uint64)                     { c.amount = new(operation.Scalar).FromUint64(value) }
func (c *CoinV2) SetInfo(b []byte) {
	c.info = make([]byte, len(b))
	copy(c.info, b)
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
