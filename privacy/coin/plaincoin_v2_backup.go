package coin

// PlainCoin is the struct that we use when we process, it contains explicitly what a coin should have
// type PlainCoinV2 struct {
// 	// Public
// 	version    uint8
// 	shardID    uint8
// 	index      uint8
// 	info       []byte
// 	publicKey  *operation.Point
// 	commitment *operation.Point
// 	keyImage   *operation.Point

// 	// not visible for hasPrivacy, visible for no privacy
// 	value      uint64
// 	randomness *operation.Scalar // r (not rG)
// }

// // Init (Coin) initializes a coin
// func (c *PlainCoinV2) Init() *PlainCoinV2 {
// 	if c == nil {
// 		c = new(PlainCoinV2)
// 	}
// 	c.version = 2
// 	c.shardID = 0
// 	c.index = 0
// 	c.value = 0
// 	c.publicKey = new(operation.Point).Identity()
// 	c.randomness = new(operation.Scalar)
// 	c.commitment = new(operation.Point).Identity()
// 	c.keyImage = new(operation.Point).Identity()
// 	c.info = []byte{}
// 	return c
// }

// func (c PlainCoinV2) GetVersion() uint8                { return 2 }
// func (c PlainCoinV2) GetShardID() uint8                { return c.shardID }
// func (c PlainCoinV2) GetIndex() uint8                  { return c.index }
// func (c PlainCoinV2) GetPublicKey() *operation.Point   { return c.publicKey }
// func (c PlainCoinV2) GetCommitment() *operation.Point  { return c.commitment }
// func (c PlainCoinV2) GetInfo() []byte                  { return c.info }
// func (c PlainCoinV2) GetKeyImage() *operation.Point    { return c.keyImage }
// func (c PlainCoinV2) GetValue() uint64                 { return c.value }
// func (c PlainCoinV2) GetRandomness() *operation.Scalar { return c.randomness }

// func (c *PlainCoinV2) SetVersion(version uint8)                  { c.version = 2 }
// func (c *PlainCoinV2) SetShardID(shardID uint8)                  { c.shardID = shardID }
// func (c *PlainCoinV2) SetIndex(index uint8)                      { c.index = index }
// func (c *PlainCoinV2) SetValue(value uint64)                     { c.value = value }
// func (c *PlainCoinV2) SetPublicKey(publicKey *operation.Point)   { c.publicKey = publicKey }
// func (c *PlainCoinV2) SetRandomness(r *operation.Scalar)         { c.randomness = r }
// func (c *PlainCoinV2) SetCommitment(commitment *operation.Point) { c.commitment = commitment }
// func (c *PlainCoinV2) SetKeyImage(keyImage *operation.Point)     { c.keyImage = keyImage }
// func (c *PlainCoinV2) SetInfo(v []byte) {
// 	c.info = make([]byte, len(v))
// 	copy(c.info, v)
// }

// func (c *PlainCoinV2) ConcealData() {
// 	c.SetValue(0)
// 	c.SetRandomness(nil)
// }

// func (c PlainCoinV2) Bytes() []byte {
// 	return nil
// }

// func (c *PlainCoinV2) SetBytes(b []byte) error {
// 	return nil
// }
