package coin

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type Coin_v2 struct {
	version    uint8
	index      uint8
	mask       *operation.Scalar
	amount     *operation.Scalar
	txRandom   *operation.Point
	publicKey  *operation.Point // K^o = H_n(r * K_B^v )G + K_B^s
	commitment *operation.Point
	info       []byte //256 bytes
}

func (this Coin_v2) Getinfo() []byte                 { return this.info }
func (this Coin_v2) GetIndex() uint8                 { return this.index }
func (this Coin_v2) GetVersion() uint8               { return this.version }
func (this Coin_v2) GetMask() *operation.Scalar      { return this.mask }
func (this Coin_v2) GetAmount() *operation.Scalar    { return this.amount }
func (this Coin_v2) GetTxRandom() *operation.Point   { return this.txRandom }
func (this Coin_v2) GetPublicKey() *operation.Point  { return this.publicKey }
func (this Coin_v2) GetCommitment() *operation.Point { return this.commitment }

func (this *Coin_v2) SetInfo(b []byte)                          { this.info = b }
func (this *Coin_v2) SetIndex(index uint8)                      { this.index = index }
func (this *Coin_v2) SetVersion(version uint8)                  { this.version = version }
func (this *Coin_v2) SetMask(mask *operation.Scalar)            { this.mask.Set(mask) }
func (this *Coin_v2) SetAmount(amount *operation.Scalar)        { this.amount.Set(amount) }
func (this *Coin_v2) SetTxRandom(txRandom *operation.Point)     { this.txRandom.Set(txRandom) }
func (this *Coin_v2) SetPublicKey(publicKey *operation.Point)   { this.publicKey.Set(publicKey) }
func (this *Coin_v2) SetCommitment(commitment *operation.Point) { this.commitment.Set(commitment) }

func NewCoinv2(index uint8, mask *operation.Scalar, amount *operation.Scalar, txRandom *operation.Point, addressee *operation.Point, commitment *operation.Point) *Coin_v2 {
	return &Coin_v2{
		index,
		mask,
		amount,
		txRandom,
		addressee,
		commitment,
	}
}

// Init (Coin) initializes a coin
func (this *Coin_v2) Init() *Coin_v2 {
	this.index = uint8(0)
	this.mask = new(operation.Scalar).FromUint64(0)
	this.amount = new(operation.Scalar).FromUint64(0)
	this.txRandom = new(operation.Point).Identity()
	this.addressee = new(operation.Point).Identity()
	this.commitment = new(operation.Point).Identity()

	return this
}

// MarshalJSON (Coin_v1) converts coin to bytes array,
// base58 check encode that bytes array into string
// json.Marshal the string
func (coin Coin_v2) MarshalJSON() ([]byte, error) {
	data := coin.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON (Coin) receives bytes array of coin (it was be MarshalJSON before),
// json.Unmarshal the bytes array to string
// base58 check decode that string to bytes array
// and set bytes array to coin
func (coin *Coin_v2) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	coin.SetBytes(temp)
	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (coin *Coin_v2) Bytes() []byte {
	var coinBytes []byte

	if coin.publicKey != nil {
		publicKey := coin.publicKey.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.coinCommitment != nil {
		coinCommitment := coin.coinCommitment.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.snDerivator != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coin.snDerivator.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.serialNumber != nil {
		serialNumber := coin.serialNumber.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.randomness != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coin.randomness.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.value > 0 {
		value := new(big.Int).SetUint64(coin.value).Bytes()
		coinBytes = append(coinBytes, byte(len(value)))
		coinBytes = append(coinBytes, value...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if len(coin.info) > 0 {
		byteLengthInfo := byte(0)
		if len(coin.info) > MaxSizeInfoCoin {
			// only get 255 byte of info
			byteLengthInfo = byte(MaxSizeInfoCoin)
		} else {
			lengthInfo := len(coin.info)
			byteLengthInfo = byte(lengthInfo)
		}
		coinBytes = append(coinBytes, byteLengthInfo)
		infoBytes := coin.info[0:byteLengthInfo]
		coinBytes = append(coinBytes, infoBytes...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}
