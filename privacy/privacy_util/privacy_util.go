package privacy_util

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/pkg/errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
)

func ScalarToBigInt(sc *operation.Scalar) *big.Int {
	keyR := operation.Reverse(sc.GetKey())
	keyRByte := keyR.ToBytes()
	bi := new(big.Int).SetBytes(keyRByte[:])
	return bi
}

func BigIntToScalar(bi *big.Int) *operation.Scalar {
	biByte := common.AddPaddingBigInt(bi, operation.Ed25519KeySize)
	var key curve25519.Key
	key.FromBytes(SliceToArray(biByte))
	keyR := operation.Reverse(key)
	sc, err := new(operation.Scalar).SetKey(&keyR)
	if err != nil {
		return nil
	}
	return sc
}

// ConvertIntToBinary represents a integer number in binary array with little endian with size n
func ConvertIntToBinary(inum int, n int) []byte {
	binary := make([]byte, n)

	for i := 0; i < n; i++ {
		binary[i] = byte(inum % 2)
		inum = inum / 2
	}

	return binary
}

// ConvertIntToBinary represents a integer number in binary
func ConvertUint64ToBinary(number uint64, n int) []*operation.Scalar {
	if number == 0 {
		res := make([]*operation.Scalar, n)
		for i := 0; i < n; i++ {
			res[i] = new(operation.Scalar).FromUint64(0)
		}
		return res
	}

	binary := make([]*operation.Scalar, n)

	for i := 0; i < n; i++ {
		binary[i] = new(operation.Scalar).FromUint64(number % 2)
		number = number / 2
	}
	return binary
}

// isOdd check a big int is odd or not
func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

// padd1Div4 computes (p + 1) / 4
func padd1Div4(p *big.Int) (res *big.Int) {
	res = new(big.Int).Add(p, big.NewInt(1))
	res.Div(res, big.NewInt(4))
	return
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

func ConvertScalarArrayToBigIntArray(scalarArr []*operation.Scalar) []*big.Int {
	res := make([]*big.Int, len(scalarArr))

	for i := 0; i < len(res); i++ {
		tmp := operation.Reverse(scalarArr[i].GetKey())
		res[i] = new(big.Int).SetBytes(ArrayToSlice(tmp.ToBytes()))
	}

	return res
}

func SliceToArray(slice []byte) [operation.Ed25519KeySize]byte {
	var array [operation.Ed25519KeySize]byte
	copy(array[:], slice)
	return array
}

func ArrayToSlice(array [operation.Ed25519KeySize]byte) []byte {
	var slice []byte
	slice = array[:]
	return slice
}

const (
	seedKeyLen     = 64 // bytes
	childNumberLen = 4  // bytes
	chainCodeLen   = 32 // bytes

	privateKeySerializedLen = 108 // len string

	privKeySerializedBytesLen     = 75 // bytes
	paymentAddrSerializedBytesLen = 71 // bytes
	readOnlyKeySerializedBytesLen = 71 // bytes
	otaKeySerializedBytesLen	  = 71 // bytes

	privKeyBase58CheckSerializedBytesLen     = 107 // len string
	paymentAddrBase58CheckSerializedBytesLen = 148 // len string
	readOnlyKeyBase58CheckSerializedBytesLen = 103 // len string
	otaKeyBase58CheckSerializedBytesLen 	 = 103 // len string

	PriKeyType         = byte(0x0) // Serialize wallet account key into string with only PRIVATE KEY of account keyset
	PaymentAddressType = byte(0x1) // Serialize wallet account key into string with only PAYMENT ADDRESS of account keyset
	ReadonlyKeyType    = byte(0x2) // Serialize wallet account key into string with only READONLY KEY of account keyset
	OTAKeyType		   = byte(0x3) // Serialize wallet account key into string with only OTA KEY of account keyset
)

var burnAddress1BytesDecode = []byte{1, 32, 99, 183, 246, 161, 68, 172, 228, 222, 153, 9, 172, 39, 208, 245, 167, 79, 11, 2, 114, 65, 241, 69, 85, 40, 193, 104, 199, 79, 70, 4, 53, 0, 0, 163, 228, 236, 208}

func Base58CheckDeserializePaymetAddress(address string) (*key.PaymentAddress, error) {
	paymentAddrss := &key.PaymentAddress{}
	data, _, err := base58.Base58Check{}.Decode(address)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(burnAddress1BytesDecode, data) {
		if len(data) != paymentAddrSerializedBytesLen && len(data) != paymentAddrSerializedBytesLen+1+operation.Ed25519KeySize {
			return nil, errors.New("length ota public key not valid: "+string(len(data)))
		}
	}
	apkKeyLength := int(data[1])
	pkencKeyLength := int(data[apkKeyLength+2])
	paymentAddrss.Pk = make([]byte, apkKeyLength)
	paymentAddrss.Tk = make([]byte, pkencKeyLength)
	copy(paymentAddrss.Pk[:], data[2:2+apkKeyLength])
	copy(paymentAddrss.Tk[:], data[3+apkKeyLength:3+apkKeyLength+pkencKeyLength])
	//Deserialize OTAPublic Key
	if len(data) > paymentAddrSerializedBytesLen {
		otapkLength := int(data[apkKeyLength+pkencKeyLength+3])
		if otapkLength != operation.Ed25519KeySize {
			return nil, errors.New("length ota public key not valid: "+string(otapkLength))
		}
		paymentAddrss.OTAPublic = append([]byte{}, data[apkKeyLength+pkencKeyLength+4:apkKeyLength+pkencKeyLength+otapkLength+4]...)
	}
	return paymentAddrss, nil
}

func Base58CheckSerializePaymentAddress(paymentAddress *key.PaymentAddress, isNewCheckSum bool) string {
	buffer := new(bytes.Buffer)
	buffer.WriteByte(PaymentAddressType)

	keyBytes := make([]byte, 0)
	keyBytes = append(keyBytes, byte(len(paymentAddress.Pk))) // set length PaymentAddress
	keyBytes = append(keyBytes, paymentAddress.Pk[:]...)      // set PaymentAddress

	keyBytes = append(keyBytes, byte(len(paymentAddress.Tk))) // set length Pkenc
	keyBytes = append(keyBytes, paymentAddress.Tk[:]...)      // set Pkenc

	if len(paymentAddress.OTAPublic) > 0 {
		keyBytes = append(keyBytes, byte(len(paymentAddress.OTAPublic))) // set length OTAPublicKey
		keyBytes = append(keyBytes, paymentAddress.OTAPublic[:]...)      // set OTAPublicKey
	}

	buffer.Write(keyBytes)

	checkSum := base58.ChecksumFirst4Bytes(buffer.Bytes(), isNewCheckSum)
	serializedKey := append(buffer.Bytes(), checkSum...)

	return base58.Base58Check{}.NewEncode(serializedKey, common.ZeroByte) //Must use the new encoding algorithm from now on
}

func GetPaymentAddressV1(addr string, isNewEncoding bool) (string, error) {
	paymentAddress, err := Base58CheckDeserializePaymetAddress(addr)
	if err != nil {
		return "", err
	}

	if len(paymentAddress.Pk) == 0 || len(paymentAddress.Pk) == 0 {
		return "", errors.New(fmt.Sprintf("something must be wrong with the provided payment address: %v", addr))
	}

	//Remove the publicOTA key and try to deserialize
	paymentAddress.OTAPublic = nil

	if isNewEncoding{
		addrV1 := Base58CheckSerializePaymentAddress(paymentAddress, true)
		if len(addrV1) == 0 {
			return "", errors.New(fmt.Sprintf("cannot decode new payment address: %v", addr))
		}
		return addrV1, nil
	}else{
		addrV1 := Base58CheckSerializePaymentAddress(paymentAddress, false)
		if len(addrV1) == 0 {
			return "", errors.New(fmt.Sprintf("cannot decode new payment address: %v", addr))
		}
		return addrV1, nil
	}
}

func IsPublicKeyBurningAddress(publicKey []byte) bool {
	// get burning address
	burnAddress1, err := Base58CheckDeserializePaymetAddress(common.BurningAddress)
	if err != nil {
		return false
	}
	if bytes.Equal(publicKey, burnAddress1.Pk) {
		return true
	}
	burnAddress2, err := Base58CheckDeserializePaymetAddress(common.BurningAddress2)
	if err != nil {
		return false
	}
	if bytes.Equal(publicKey, burnAddress2.Pk) {
		return true
	}

	return false
}