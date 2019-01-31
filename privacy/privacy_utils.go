package privacy

import (
	"errors"
	"math/big"
	rand2 "math/rand"
	"time"
)

// RandBytes generates random bytes
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

// ConvertIntToBinary represents a integer number in binary
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

// checkZeroArray check whether all ellement of values array are zero value or not
func checkZeroArray(values []*big.Int) bool {
	for i := 0; i < len(values); i++ {
		if values[i].Cmp(big.NewInt(0)) != 0 {
			return false
		}
	}
	return true
}

func MaxBitLen(values []*big.Int) int {
	res := 0
	for i := 0; i < len(values); i++ {
		if values[i].BitLen() > res {
			res = values[i].BitLen()
		}
	}

	return res
}

// MultiScalar2 uses Shamir's simultanenous Squaring Multi-Exponentiation Algorithm
func MultiScalar2(g []*EllipticPoint, values []*big.Int) (*EllipticPoint, error) {
	// Check inputs
	if len(g) != len(values) {
		return nil, errors.New("wrong inputs")
	}

	//convert value array to binary array
	maxBitLen := MaxBitLen(values)
	valueBinary := make([][]*big.Int, len(values))
	for i := range values {
		valueBinary[i] = ConvertBigIntToBinary(values[i], maxBitLen)
	}

	// generator result point
	res := new(EllipticPoint).Zero()

	oneNumber := big.NewInt(1)

	for i := maxBitLen - 1; i >= 0; i-- {
		// res = 2*res
		res = res.ScalarMult(big.NewInt(2))

		for j := 0; j < len(values); j++ {
			if valueBinary[j][i].Cmp(oneNumber) == 0 {
				res = res.Add(g[j])
			}
		}
	}
	return res, nil
}

func MultiScalarmult(bases []*EllipticPoint, exponents []*big.Int) (*EllipticPoint, error) {
	n := len(bases)
	if n != len(exponents) {
		return nil, errors.New("wrong inputs")
	}

	//count := 0

	baseTmp := make([]*EllipticPoint, n)
	for i := 0; i < n; i++ {
		baseTmp[i] = new(EllipticPoint)
		baseTmp[i].Set(bases[i].X, bases[i].Y)
	}

	expTmp := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		expTmp[i] = new(big.Int)
		expTmp[i].Set(exponents[i])
	}
	//start1 := time.Now()

	result := new(EllipticPoint).Zero()

	for !checkZeroArray(expTmp) {
		for i := 0; i < n; i++ {
			if new(big.Int).And(expTmp[i], big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
				result = result.Add(baseTmp[i])
			}

			expTmp[i].Rsh(expTmp[i], uint(1))
			baseTmp[i] = baseTmp[i].Add(baseTmp[i])
		}
	}

	//end1 := time.Since(start1)
	//fmt.Printf(" time faster: %v\n", end1)

	return result, nil
}
