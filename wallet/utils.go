package wallet

import (
	"encoding/binary"
	"math/big"
	"bytes"
	"github.com/ninjadotorg/cash-prototype/common/base58"
)

//
// Encoding
//

func addChecksumToBytes(data []byte) ([]byte, error) {
	checksum := base58.ChecksumFirst4Bytes(data)
	return append(data, checksum...), nil
}

// Keys
func compressPublicKey(x *big.Int, y *big.Int) []byte {
	var key bytes.Buffer

	// Write header; 0x2 for even y value; 0x3 for odd
	key.WriteByte(byte(0x2) + byte(y.Bit(0)))

	// Write X coord; Pad the Key so x is aligned with the LSB. Pad size is Key length - header size (1) - xBytes size
	xBytes := x.Bytes()
	for i := 0; i < (PublicKeyCompressedLength - 1 - len(xBytes)); i++ {
		key.WriteByte(0x0)
	}
	key.Write(xBytes)

	return key.Bytes()
}

//
//// As described at https://crypto.stackexchange.com/a/8916
//func expandPublicKey(Key []byte) (*big.Int, *big.Int) {
//	Y := big.NewInt(0)
//	X := big.NewInt(0)
//	X.SetBytes(Key[1:])
//
//	// y^2 = x^3 + ax^2 + b
//	// a = 0
//	// => y^2 = x^3 + b
//	ySquared := big.NewInt(0)
//	ySquared.Exp(X, big.NewInt(3), nil)
//	ySquared.Add(ySquared, curveParams.B)
//
//	Y.ModSqrt(ySquared, curveParams.P)
//
//	Ymod2 := big.NewInt(0)
//	Ymod2.Mod(Y, big.NewInt(2))
//
//	signY := uint64(Key[0]) - 2
//	if signY != Ymod2.Uint64() {
//		Y.Sub(curveParams.P, Y)
//	}
//
//	return X, Y
//}
//
//func validatePrivateKey(Key []byte) error {
//	if fmt.Sprintf("%x", Key) == "0000000000000000000000000000000000000000000000000000000000000000" || //if the Key is zero
//		bytes.Compare(Key, curveParams.N.Bytes()) >= 0 || //or is outside of the curve
//		len(Key) != 32 { //or is too short
//		return ErrInvalidPrivateKey
//	}
//
//	return nil
//}
//
//func validateChildPublicKey(Key []byte) error {
//	x, y := expandPublicKey(Key)
//
//	if x.Sign() == 0 || y.Sign() == 0 {
//		return ErrInvalidPublicKey
//	}
//
//	return nil
//}
//

/**
Numerical
 */
func uint32Bytes(i uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, i)
	return bytes
}
