package coin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type TxRandom [TxRandomGroupSize]byte

func NewTxRandom() *TxRandom {
	txRandom := new(operation.Point).Identity()
	index := uint32(0)

	res := new(TxRandom)
	res.SetTxRandomPoint(txRandom)
	res.SetIndex(index)
	return res
}

func (t TxRandom) GetTxRandomPoint() *operation.Point {
	txRandom, _ := new(operation.Point).FromBytesS(t[:operation.Ed25519KeySize])
	return txRandom
}

func (t TxRandom) GetIndex() uint32 {
	index, _ := common.BytesToUint32(t[operation.Ed25519KeySize:])
	return index
}

func (t *TxRandom) SetTxRandomPoint(txRandom *operation.Point) {
	txRandomBytes := txRandom.ToBytesS()
	copy(t[:operation.Ed25519KeySize], txRandomBytes)
}

func (t *TxRandom) SetIndex(index uint32) {
	indexBytes := common.Uint32ToBytes(index)
	copy(t[operation.Ed25519KeySize:], indexBytes)
}

func (t TxRandom) Bytes() []byte {
	return t[:]
}

func (t *TxRandom) SetBytes(coinBytes []byte) error {
	if len(coinBytes) != TxRandomGroupSize {
		return errors.New("Cannot TxRandomGroupSize.SetBytes because coinBytes is too short < TxRandomGroupSize")
	}
	_, err := new(operation.Point).FromBytesS(coinBytes[:operation.Ed25519KeySize])
	if err != nil {
		errStr := fmt.Sprintf("Cannot TxRandomGroupSize.SetBytes: bytes is not operation.Point err: %v", err)
		return errors.New(errStr)
	}
	copy(t[:], coinBytes)
	return nil
}

// CoinV2 is the struct that will be stored to db
// If not privacy, mask and amount will be the original randomness and value
// If has privacy, mask and amount will be as paper monero
type CoinV2 struct {
	// Public
	version    uint8
	info       []byte
	publicKey  *operation.Point
	commitment *operation.Point
	keyImage   *operation.Point

	// sharedRandom and txRandom is shared secret between receiver and giver
	// sharedRandom is only visible when creating coins, when it is broadcast to network, it will be set to null
	sharedRandom *operation.Scalar // r
	txRandom     *TxRandom         // rG

	// mask = randomness
	// amount = value
	mask   *operation.Scalar
	amount *operation.Scalar
}

func NewCoinFromAmountAndReceiver(amount uint64, receiver key.PaymentAddress) (*CoinV2, error) {
	fmt.Println("Creating coins where amount =", amount, "and publickey =", receiver.Pk)
	paymentInfo := key.InitPaymentInfo(receiver, amount, []byte{})
	return NewCoinFromPaymentInfo(paymentInfo)
}

// This function create new coinv2 that has same shardID with targetShardID and with info of paymentInfo
func NewCoinFromPaymentInfo(info *key.PaymentInfo) (*CoinV2, error) {
	receiverPublicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
	if err != nil {
		errStr := fmt.Sprintf("Cannot parse outputCoinV2 from PaymentInfo when parseByte PublicKey, error %v ", err)
		return nil, errors.New(errStr)
	}
	receiverPublicKeyBytes := receiverPublicKey.ToBytesS()
	targetShardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])

	c := new(CoinV2).Init()
	c.SetVersion(2)

	// Amount, Randomness, SharedRandom is transparency until we call concealData
	c.SetAmount(new(operation.Scalar).FromUint64(info.Amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetSharedRandom(operation.RandomScalar())                                  // r
	c.SetTxRandomPoint(new(operation.Point).ScalarMultBase(c.GetSharedRandom())) // rG
	c.SetInfo(info.Message)
	c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), c.GetRandomness(), operation.PedersenValueIndex))

	// If this is going to burning address then dont need to create ota
	if wallet.IsPublicKeyBurningAddress(info.PaymentAddress.Pk) {
		c.SetIndex(0)
		publicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
		if err != nil {
			panic("Something is wrong with info.paymentAddress.Pk, burning address should be a valid point")
		}
		c.SetPublicKey(publicKey)
		return c, nil
	}

	// Increase index until have the right shardID
	index := uint32(0)
	publicView := info.PaymentAddress.GetPublicView()
	publicSpend := info.PaymentAddress.GetPublicSpend()
	rK := new(operation.Point).ScalarMult(publicView, c.GetSharedRandom())
	for {
		index += 1
		c.SetIndex(index)

		// Get publickey
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := new(operation.Point).ScalarMultBase(hash)
		publicKey := new(operation.Point).Add(HrKG, publicSpend)
		c.SetPublicKey(publicKey)

		currentShardID, err := c.GetShardID()
		if err != nil {
			return nil, err
		}
		if currentShardID == targetShardID {
			break
		}
	}
	return c, nil
}

func (c CoinV2) ParsePrivateKeyOfCoin(privKey key.PrivateKey) (*operation.Scalar, error) {
	keySet := new(incognitokey.KeySet)
	err := keySet.InitFromPrivateKey(&privKey)
	if err != nil {
		err := errors.New("Cannot init keyset from privateKey")
		return nil, errhandler.NewPrivacyErr(errhandler.InvalidPrivateKeyErr, err)
	}
	rK := new(operation.Point).ScalarMult(c.GetTxRandomPoint(), keySet.ReadonlyKey.GetPrivateView()) //rG * k = rK
	H := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(c.GetIndex())...))        // Hash(rK, index)

	k := new(operation.Scalar).FromBytesS(privKey)
	return new(operation.Scalar).Add(H, k), nil // Hash(rK, index) + privSpend
}

func (c CoinV2) ParseKeyImageWithPrivateKey(privKey key.PrivateKey) (*operation.Point, error) {
	k, err := c.ParsePrivateKeyOfCoin(privKey)
	if err != nil {
		err := errors.New("Cannot init keyset from privateKey")
		return nil, errhandler.NewPrivacyErr(errhandler.InvalidPrivateKeyErr, err)
	}
	Hp := operation.HashToPoint(c.GetPublicKey().ToBytesS())
	return new(operation.Point).ScalarMult(Hp, k), nil
}

// AdditionalData of concealData should be publicView of the receiver
func (c *CoinV2) ConcealData(additionalData interface{}) {
	// If this coin is already encrypted or it is created by other person then cannot conceal
	if c.IsEncrypted() || c.GetSharedRandom() == nil {
		return
	}
	publicView := additionalData.(*operation.Point)

	rK := new(operation.Point).ScalarMult(publicView, c.GetSharedRandom())
	hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(c.GetIndex())...))
	hash = operation.HashToScalar(hash.ToBytesS())
	mask := new(operation.Scalar).Add(c.GetRandomness(), hash)

	hash = operation.HashToScalar(hash.ToBytesS())
	amount := new(operation.Scalar).Add(c.GetAmount(), hash)
	c.SetRandomness(mask)
	c.SetAmount(amount)
	c.SetSharedRandom(nil)
}

func (c *CoinV2) Decrypt(keySet *incognitokey.KeySet) (PlainCoin, error) {
	// Must parse keyImage first in any situation
	if len(keySet.PrivateKey) > 0 {
		keyImage, err := c.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
		if err != nil {
			errReturn := errors.New("Cannot parse key image with privateKey CoinV2" + err.Error())
			return nil, errhandler.NewPrivacyErr(errhandler.ParseKeyImageWithPrivateKeyErr, errReturn)
		}
		c.SetKeyImage(keyImage)
	}
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
		rK := new(operation.Point).ScalarMult(c.GetTxRandomPoint(), viewKey.GetPrivateView())

		// Hash multiple times
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(c.GetIndex())...))
		hash = operation.HashToScalar(hash.ToBytesS())
		randomness := c.GetRandomness().Sub(c.GetRandomness(), hash)

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
	return c, nil
}

// Init (OutputCoin) initializes a output coin
func (c *CoinV2) Init() *CoinV2 {
	if c == nil {
		c = new(CoinV2)
	}
	c.version = 2
	c.info = []byte{}
	c.publicKey = new(operation.Point).Identity()
	c.commitment = new(operation.Point).Identity()
	c.keyImage = new(operation.Point).Identity()
	c.sharedRandom = new(operation.Scalar)
	c.txRandom = NewTxRandom()
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

func (c CoinV2) GetRandomness() *operation.Scalar   { return c.mask }
func (c CoinV2) GetAmount() *operation.Scalar       { return c.amount }
func (c CoinV2) GetSharedRandom() *operation.Scalar { return c.sharedRandom }
func (c CoinV2) GetTxRandom() *TxRandom             { return c.txRandom }
func (c CoinV2) GetTxRandomPoint() *operation.Point { return c.txRandom.GetTxRandomPoint() }
func (c CoinV2) GetPublicKey() *operation.Point     { return c.publicKey }
func (c CoinV2) GetCommitment() *operation.Point    { return c.commitment }
func (c CoinV2) GetKeyImage() *operation.Point      { return c.keyImage }
func (c CoinV2) GetIndex() uint32                   { return c.txRandom.GetIndex() }
func (c CoinV2) GetInfo() []byte                    { return c.info }
func (c CoinV2) GetValue() uint64 {
	if c.IsEncrypted() {
		return 0
	}
	return c.amount.ToUint64Little()
}

func (c *CoinV2) SetVersion(uint8)                               { c.version = 2 }
func (c *CoinV2) SetRandomness(mask *operation.Scalar)           { c.mask = mask }
func (c *CoinV2) SetAmount(amount *operation.Scalar)             { c.amount = amount }
func (c *CoinV2) SetSharedRandom(sharedRandom *operation.Scalar) { c.sharedRandom = sharedRandom }
func (c *CoinV2) SetTxRandom(txRandom *TxRandom) {
	c.txRandom = NewTxRandom()
	c.txRandom.SetBytes(txRandom.Bytes())
}
func (c *CoinV2) SetTxRandomPoint(txRandomPoint *operation.Point) {
	c.txRandom.SetTxRandomPoint(txRandomPoint)
}
func (c *CoinV2) SetPublicKey(publicKey *operation.Point)   { c.publicKey = publicKey }
func (c *CoinV2) SetCommitment(commitment *operation.Point) { c.commitment = commitment }
func (c *CoinV2) SetKeyImage(keyImage *operation.Point)     { c.keyImage = keyImage }
func (c *CoinV2) SetIndex(index uint32)                     { c.txRandom.SetIndex(index) }
func (c *CoinV2) SetValue(value uint64)                     { c.amount = new(operation.Scalar).FromUint64(value) }
func (c *CoinV2) SetInfo(b []byte) {
	c.info = make([]byte, len(b))
	copy(c.info, b)
}

func (c CoinV2) Bytes() []byte {
	coinBytes := []byte{c.GetVersion()}
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

	if c.sharedRandom != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.sharedRandom.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.txRandom != nil {
		coinBytes = append(coinBytes, TxRandomGroupSize)
		coinBytes = append(coinBytes, c.txRandom.Bytes()...)
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
	if len(coinBytes) == 0 {
		return errors.New("coinBytes is empty")
	}
	if coinBytes[0] != 2 {
		return errors.New("coinBytes version is not 2")
	}
	c.SetVersion(coinBytes[0])

	offset := 1
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
	c.sharedRandom, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV2 mask error: " + err.Error())
	}

	if coinBytes[offset] != TxRandomGroupSize {
		return errors.New("SetBytes CoinV2 TxRandomGroup error: length of TxRandomGroup is not correct")
	}
	offset += 1
	if offset+TxRandomGroupSize > len(coinBytes) {
		return errors.New("SetBytes CoinV2 TxRandomGroup error: length of coinBytes is too small")
	}
	c.txRandom = NewTxRandom()
	err = c.txRandom.SetBytes(coinBytes[offset : offset+TxRandomGroupSize])
	if err != nil {
		return errors.New("SetBytes CoinV2 TxRandomGroup error: " + err.Error())
	}
	offset += TxRandomGroupSize

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
func (c *CoinV2) HashH() *common.Hash {
	hash := common.HashH(c.Bytes())
	return &hash
}

func (c CoinV2) MarshalJSON() ([]byte, error) {
	data := c.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

func (c *CoinV2) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	err = c.SetBytes(temp)
	if err != nil {
		return err
	}
	return nil
}
