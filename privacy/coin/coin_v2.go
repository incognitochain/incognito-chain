package coin

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

// CoinV2 is the struct that will be stored to db
// If not privacy, mask and amount will be the original randomness and value
// If has privacy, mask and amount will be as paper monero
type CoinV2 struct {
	// Public
	version    uint8
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

func ParseOnetimeAddress(pubSpend, pubView *operation.Point, randomness *operation.Scalar, index uint8) *operation.Point {
	rK := new(operation.Point).ScalarMult(pubView, randomness)
	hash := operation.HashToScalar(append(rK.ToBytesS(), index))
	HrKG := new(operation.Point).ScalarMultBase(hash)
	return new(operation.Point).Add(HrKG, pubSpend)
}

func ArrayCoinToCoinV2(inputCoins []Coin) []*CoinV2 {
	res := make([]*CoinV2, len(inputCoins))
	for i := 0; i < len(inputCoins); i += 1 {
		res[i] = inputCoins[i].(*CoinV2)
	}
	return res
}

func (c CoinV2) ParsePrivateKeyOfCoin(privKey key.PrivateKey) (*operation.Scalar, error) {
	tempCoin := c
	keySet := new(incognitokey.KeySet)
	err := keySet.InitFromPrivateKey(&privKey)
	if err != nil {
		err := errors.New("Cannot init keyset from privateKey")
		return nil, errhandler.NewPrivacyErr(errhandler.InvalidPrivateKeyErr, err)
	}
	tempPlainCoin, _ := tempCoin.Decrypt(keySet)

	paymentAddress := key.GeneratePaymentAddress(privKey)
	publicView := paymentAddress.GetPublicView()

	rK := new(operation.Point).ScalarMult(publicView, tempPlainCoin.GetRandomness())
	H := operation.HashToScalar(append(rK.ToBytesS(), c.GetIndex()))

	k := new(operation.Scalar).FromBytesS(privKey)
	return new(operation.Scalar).Add(H, k), nil
}

func (c CoinV2) ParseKeyImageWithPrivateKey(privKey key.PrivateKey) (*operation.Point, error) {
	k, err := c.ParsePrivateKeyOfCoin(privKey)
	if err != nil {
		if err != nil {
			err := errors.New("Cannot init keyset from privateKey")
			return nil, errhandler.NewPrivacyErr(errhandler.InvalidPrivateKeyErr, err)
		}
	}
	Hp := operation.HashToPoint(c.GetPublicKey().ToBytesS())
	return new(operation.Point).ScalarMult(Hp, k), nil
}

// AdditionalData of concealData should be publicView of the receiver
func (c *CoinV2) ConcealData(additionalData interface{}) {
	if c.IsEncrypted() {
		return
	}
	publicView := additionalData.(*operation.Point)

	rK := new(operation.Point).ScalarMult(publicView, c.GetMask())
	hash := operation.HashToScalar(append(rK.ToBytesS(), c.GetIndex()))
	hash = operation.HashToScalar(hash.ToBytesS())
	mask := new(operation.Scalar).Add(c.GetMask(), hash)

	hash = operation.HashToScalar(hash.ToBytesS())
	amount := new(operation.Scalar).Add(c.GetAmount(), hash)
	c.SetRandomness(mask)
	c.SetAmount(amount)
}

func (c *CoinV2) Decrypt(keySet *incognitokey.KeySet) (PlainCoin, error) {
	if c.IsEncrypted() == false {
		return c, nil
	}
	if keySet == nil {
		err := errors.New("Cannot Decrypt CoinV2: Keyset is empty")
		return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
	}
	viewKey := keySet.ReadonlyKey
	if len(viewKey.Rk) == 0 && len(keySet.PrivateKey) == 0 {
		err := errors.New("Cannot Decrypt CoinV2: Keyset does not contain viewkey or privatekey")
		return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
	}

	if len(viewKey.Rk) > 0 {
		rK := new(operation.Point).ScalarMult(c.GetTxRandom(), viewKey.GetPrivateView())

		// Hash multiple times
		hash := operation.HashToScalar(append(rK.ToBytesS(), c.GetIndex()))
		hash = operation.HashToScalar(hash.ToBytesS())
		randomness := c.GetMask().Sub(c.GetMask(), hash)

		// Hash 1 more time to get value
		hash = operation.HashToScalar(hash.ToBytesS())
		value := c.GetAmount().Sub(c.GetAmount(), hash)

		commitment := operation.PedCom.CommitAtIndex(value, randomness, operation.PedersenValueIndex)
		if !operation.IsPointEqual(commitment, c.GetCommitment()) {
			err := errors.New("Cannot Decrypt CoinV2: Commitment is not the same after decrypt")
			return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
		}
		c.SetRandomness(randomness)
		c.SetAmount(value)
	}
	if len(keySet.PrivateKey) > 0 {
		keyImage, err := c.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
		if err != nil {
			errReturn := errors.New("Cannot parse key image with privateKey CoinV2" + err.Error())
			return nil, errhandler.NewPrivacyErr(errhandler.ParseKeyImageWithPrivateKeyErr, errReturn)
		}
		c.SetKeyImage(keyImage)
	}
	return c, nil
}

// Init (OutputCoin) initializes a output coin
func (c *CoinV2) Init() *CoinV2 {
	if c == nil {
		c = new(CoinV2)
	}
	c.version = 2
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
	commitment := operation.PedCom.CommitAtIndex(c.amount, c.mask, operation.PedersenValueIndex)
	return !operation.IsPointEqual(commitment, c.commitment)
}

func (c CoinV2) GetVersion() uint8 { return 2 }
func (c CoinV2) GetShardID() (uint8, error) {
	if c.publicKey == nil {
		return 255, errors.New("Cannot get ShardID because PublicKey of PlainCoin is concealed")
	}
	pubKeyBytes := c.publicKey.ToBytes()
	lastByte := pubKeyBytes[operation.Ed25519KeySize-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func (c CoinV2) GetMask() *operation.Scalar       { return c.mask }
func (c CoinV2) GetRandomness() *operation.Scalar { return c.mask }
func (c CoinV2) GetAmount() *operation.Scalar     { return c.amount }
func (c CoinV2) GetTxRandom() *operation.Point    { return c.txRandom }
func (c CoinV2) GetPublicKey() *operation.Point   { return c.publicKey }
func (c CoinV2) GetCommitment() *operation.Point  { return c.commitment }
func (c CoinV2) GetKeyImage() *operation.Point    { return c.keyImage }
func (c CoinV2) GetIndex() uint8                  { return c.index }
func (c CoinV2) GetInfo() []byte                  { return c.info }
func (c CoinV2) GetValue() uint64 {
	if c.IsEncrypted() {
		return 0
	}
	return c.amount.ToUint64Little()
}

func (c *CoinV2) SetVersion(uint8)                          { c.version = 2 }
func (c *CoinV2) SetRandomness(mask *operation.Scalar)      { c.mask = mask }
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

func (c *CoinV2) Bytes() []byte {
	coinBytes := []byte{c.GetVersion(), c.GetIndex()}

	info := c.GetInfo()
	byteLengthInfo := byte(getMin(len(info), MaxSizeInfoCoin))
	coinBytes = append(coinBytes, byteLengthInfo)
	coinBytes = append(coinBytes, info[:byteLengthInfo]...)

	if c.publicKey != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.publicKey.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.commitment != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.commitment.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.keyImage != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.keyImage.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.txRandom != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.txRandom.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.mask != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.mask.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.amount != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.amount.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}

func (c *CoinV2) SetBytes(coinBytes []byte) error {
	var err error
	if c == nil {
		c = new(CoinV2)
	}
	if len(coinBytes) < 3 {
		return errors.New("coinBytes length is too small < 3")
	}
	if coinBytes[0] != 2 {
		return errors.New("coinBytes version is not 2")
	}
	c.SetVersion(coinBytes[0])
	c.SetIndex(coinBytes[1])

	offset := 2
	c.info, err = parseInfoForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 info error: " + err.Error())
	}

	c.publicKey, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 publicKey error: " + err.Error())
	}
	c.commitment, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 commitment error: " + err.Error())
	}
	c.keyImage, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 keyImage error: " + err.Error())
	}
	c.txRandom, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 txRandom error: " + err.Error())
	}
	c.mask, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 mask error: " + err.Error())
	}
	c.amount, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 amount error: " + err.Error())
	}
	return nil
}

// HashH returns the SHA3-256 hashing of coin bytes array
func (this *CoinV2) HashH() *common.Hash {
	hash := common.HashH(this.Bytes())
	return &hash
}
