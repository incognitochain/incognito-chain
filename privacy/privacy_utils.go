package privacy

import (
	"errors"
	"fmt"
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

// RandInt generates a big int with value less than order of group of elliptic points
func RandInt() *big.Int {
	for {
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

// ConvertIntToBinary represents a integer number in binary
func ConvertBigIntToBinary(number *big.Int, n int) []*big.Int {
	binary := make([]*big.Int, n)
	numberClone := new(big.Int)
	numberClone.Set(number)

	tmp := big.NewInt(0)
	twoNumber := big.NewInt(2)

	for i := n - 1; i >= 0; i-- {
		tmp.Mod(numberClone, twoNumber)
		binary[i] = new(big.Int).Set(tmp)
		numberClone.Div(numberClone, twoNumber)
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

//func multiExp(g []*EllipticPoint, values []*big.Int) (*EllipticPoint, error) {
//	// Check inputs
//	if len(g) != len(values) {
//		return nil, errors.New("wrong inputs")
//	}
//
//	maxBitLen := MaxBitLen(values)
//
//	// save values
//	valuesTmp := make([]*big.Int, len(values))
//	for i := 0; i < len(valuesTmp); i++ {
//		valuesTmp[i] = new(big.Int)
//		valuesTmp[i].Set(values[i])
//	}
//
//	// generator result point
//	res := new(EllipticPoint).Zero()
//
//	for i := maxBitLen -1; i >= 0; i-- {
//		// res = 2*res
//		res = res.ScalarMult(big.NewInt(2))
//
//		// res = res + e1,i*g1 + ... + en,i*gn
//		for j := 0; j < len(valuesTmp); j++ {
//			r := valuesTmp[i].Bit(i)
//			if r == 1{
//				res = res.Add(g[j])
//			}
//			//valuesTmp[j].Div(valuesTmp[j], big.NewInt(2))
//		}
//
//		//if checkZeroArray(valuesTmp){
//		//	//res = res.ScalarMult(big.NewInt(2))
//		//	break
//		//}
//	}
//	return res, nil
//}

//func multiExp2(g []*EllipticPoint, values []*big.Int) (*EllipticPoint, error) {
//	// Check inputs
//	if len(g) != len(values) {
//		return nil, errors.New("wrong inputs")
//	}
//
//	// save values
//	valuesTmp := make([]*big.Int, len(values))
//	for i := 0; i < len(valuesTmp); i++ {
//		valuesTmp[i] = new(big.Int)
//		valuesTmp[i].Set(values[i])
//	}
//
//	// generator result point
//	r := new(big.Int)
//	res := new(EllipticPoint).Zero()
//
//	y := big.NewInt(1)
//	for values
//
//	//for i := 0; i < Curve.Params().BitSize; i++ {
//	//	// res = 2*res
//	//	res = res.ScalarMult(big.NewInt(2))
//	//
//	//	// res = res + e1,i*g1 + ... + en,i*gn
//	//	for j := 0; j < len(valuesTmp); j++ {
//	//		r.Mod(valuesTmp[j], big.NewInt(2))
//	//		if r.Cmp(big.NewInt(1)) ==0{
//	//			res = res.Add(g[j])
//	//		}
//	//		valuesTmp[j].Div(valuesTmp[j], big.NewInt(2))
//	//	}
//	//
//	//	if checkZeroArray(valuesTmp){
//	//		//res = res.ScalarMult(big.NewInt(2))
//	//		break
//	//	}
//	//}
//	return res, nil
//}

//func exp (x * EllipticPoint, n *big.Int) *EllipticPoint{
//	if n.Cmp(big.NewInt(0)) == 0{
//		return x
//	}
//
//	nTmp := new(big.Int)
//	nTmp.Set(n)
//
//	xTmp := new(EllipticPoint)
//	xTmp.Set(x.X, x.Y)
//
//	y := new(EllipticPoint).Zero()
//
//	r := big.NewInt(0)
//
//	for nTmp.Cmp(big.NewInt(1)) == 1{
//		// nTmp is even
//		if r.Mod(nTmp, big.NewInt(2)).Cmp(big.NewInt(1)) == 0 {
//			y = xTmp.Add(y)
//		}
//		xTmp = xTmp.Add(xTmp)
//		nTmp.Div(nTmp, big.NewInt(2))
//	}
//
//	return xTmp.Add(y)
//}

func multiScalarmult(bases []*EllipticPoint, exponents []*big.Int) (*EllipticPoint, error) {
	n := len(bases)
	if n != len(exponents) {
		return nil, errors.New("wrong inputs")
	}

	//count := 0

	baseTmp := make([]*EllipticPoint, n)
	for i:=0; i<n; i++{
		baseTmp[i] = new(EllipticPoint)
		baseTmp[i].Set(bases[i].X, bases[i].Y)
	}

	expTmp := make([]*big.Int, n)
	for i:=0; i<n; i++{
		expTmp[i] = new(big.Int)
		expTmp[i].Set(exponents[i])
	}
	start1 := time.Now()

	result := new(EllipticPoint).Zero()

	for !checkZeroArray(expTmp) {
		for i := 0; i < n; i++ {
			if new(big.Int).And(expTmp[i], big.NewInt(1)).Cmp(big.NewInt(1)) ==0 {
				result = result.Add(baseTmp[i])
			}

			expTmp[i].Rsh(expTmp[i], uint(1))
			baseTmp[i] = baseTmp[i].Add(baseTmp[i])
		}
	}

	end1 := time.Since(start1)
	fmt.Printf(" time faster: %v\n", end1)

	return result, nil
}

