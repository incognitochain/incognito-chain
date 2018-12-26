package privacy

import (
	"math/big"
	rand2 "math/rand"
	"time"
)

// RandBytes generates random bytes
func RandBytes(length int) []byte {
	seed:=time.Now().UnixNano()
	b := make([]byte, length)
	reader := rand2.New(rand2.NewSource(int64(seed)))
	for n := 0; n < length; {
		read, err := reader.Read(b[n:])
		if err != nil {
			panic(err)
		}
		n += read
	}
	return b
	//b := make([]byte, n)
	//_, err := rand.Read(b)
	//if err != nil {
	//	fmt.Println("error:", err)
	//	return nil
	//}
	//return b
}

//func RandByte() byte {
//	var res byte
//	res = 0
//	var bit byte
//	rand2.Seed(time.Now().UnixNano())
//	for i := 0; i < 8; i++ {
//		bit = byte(rand2.Intn(2))
//		res += bit << byte(i)
//	}
//	return res
//}

// RandInt generates a big int with value less than order of group of elliptic points
func RandInt() *big.Int {
	for {
		//bytes := make([]byte, BigIntSize)
		randNum := new(big.Int).SetBytes(RandBytes(BigIntSize))
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

// ConvertIntToBinary represents a integer number in binary
func ConvertIntToBinary(inum int, n int) []byte {
	binary := make([]byte, n)

	for i := n - 1; i >= 0; i-- {
		binary[i] = byte(inum % 2)
		inum = inum / 2
	}

	return binary
}

// AddPaddingBigInt adds padding to big int to it is fixed size
func AddPaddingBigInt(numInt *big.Int, fixedSize int) []byte {
	numBytes := numInt.Bytes()
	lenNumBytes := len(numBytes)

	for i := 0; i < fixedSize-lenNumBytes; i++ {
		numBytes = append([]byte{0}, numBytes...)
	}
	return numBytes
}

// IntToByteArr converts an integer number to 2 bytes array
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

// ByteArrToInt reverts an integer number from bytes array
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
	res = new(big.Int)
	res.Add(p, new(big.Int).SetInt64(1))
	res.Div(res, new(big.Int).SetInt64(4))
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
