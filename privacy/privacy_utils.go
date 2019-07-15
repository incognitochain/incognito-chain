package privacy

import (
	"math/big"
	rand2 "math/rand"
	"time"
)

// RandBytes generates random bytes with length
func RandBytes(length int) []byte {
	seed := time.Now().UnixNano()
	b := make([]byte, length)
	reader := rand2.New(rand2.NewSource(int64(seed)))

	for n := 0; n < length; {
		read, err := reader.Read(b[n:])
		if err != nil {
			Logger.Log.Errorf("[PRIVACY LOG] Rand byte error : %v\n", err)
			return nil
		}
		n += read
	}
	return b
}

// RandScalar generates a big int with value less than order of group of elliptic points
func RandScalar() *big.Int {
	randNum := new(big.Int)
	for {
		randNum.SetBytes(RandBytes(BigIntSize))
		if randNum.Cmp(Curve.Params().N) == -1 {
			return randNum
		}
	}
}

// IsPowerOfTwo checks whether n is power of two or not
func IsPowerOfTwo(n int) bool {
	if n < 2 {
		return false
	}
	for n > 2 {
		if n%2 == 0 {
			n = n / 2
		} else {
			return false
		}
	}
	return true
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
func ConvertBigIntToBinary(number *big.Int, n int) []*big.Int {
	if number.Cmp(big.NewInt(0)) == 0 {
		res := make([]*big.Int, n)
		for i := 0; i < n; i++ {
			res[i] = big.NewInt(0)
		}
		return res
	}

	binary := make([]*big.Int, n)
	numberClone := new(big.Int)
	numberClone.Set(number)

	zeroNumber := big.NewInt(0)
	twoNumber := big.NewInt(2)

	for i := 0; i < n; i++ {
		binary[i] = new(big.Int)
		binary[i] = new(big.Int).Mod(numberClone, twoNumber)
		numberClone.Div(numberClone, twoNumber)

		if numberClone.Cmp(zeroNumber) == 0 && i != n-1 {
			for j := i + 1; j < n; j++ {
				binary[j] = zeroNumber
			}
			break
		}
	}
	return binary
}

// AddPaddingBigInt adds padding to big int to it is fixed size
// and returns bytes array
func AddPaddingBigInt(numInt *big.Int, fixedSize int) []byte {
	numBytes := numInt.Bytes()
	lenNumBytes := len(numBytes)
	zeroBytes := make([]byte, fixedSize-lenNumBytes)
	numBytes = append(zeroBytes, numBytes...)
	return numBytes
}

// IntToByteArr converts an integer number to 2-byte array in big endian
func IntToByteArr(n int) []byte {
	if n == 0 {
		return []byte{0, 0}
	}

	a := big.NewInt(int64(n))

	if len(a.Bytes()) > 2 {
		return []byte{}
	}

	if len(a.Bytes()) == 1 {
		return []byte{0, a.Bytes()[0]}
	}

	return a.Bytes()
}

// ByteArrToInt reverts an integer number from 2-byte array
func ByteArrToInt(bytesArr []byte) int {
	if len(bytesArr) != 2 {
		return 0
	}

	numInt := new(big.Int).SetBytes(bytesArr)
	return int(numInt.Int64())
}

// isOdd check a big int is odd or not
func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

// PAdd1Div4 computes (p + 1) / 4
func PAdd1Div4(p *big.Int) (res *big.Int) {
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