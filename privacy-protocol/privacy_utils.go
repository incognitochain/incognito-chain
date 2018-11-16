package privacy

import (
	"crypto/rand"
	"fmt"
)

//inversePoint return inverse point of ECC Point input
//func inversePoint(eccPoint EllipticPoint) (*EllipticPoint, error) {
//	//Check that input is ECC point
//	if !Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
//		return nil, fmt.Errorf("Input is not ECC Point")
//	}
//	//Create result point
//	resPoint := new(EllipticPoint)
//	resPoint.X = big.NewInt(0)
//	resPoint.Y = big.NewInt(0)
//
//	//inverse point of A(x,y) in ECC is A'(x, P - y) with P is order of Curve
//	resPoint.X.SetBytes(eccPoint.X.Bytes())
//	resPoint.Y.SetBytes(eccPoint.Y.Bytes())
//	resPoint.Y.Sub(Curve.Params().P, resPoint.Y)
//
//	return resPoint, nil
//}

// RandBytes generates random bytes
func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return b
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
func ConvertIntToBinary(i int, n int) []byte{
	binary := make([]byte, n)
	tmp := i

	for i := n-1; i >= 0; i--{
		binary[i] = byte(tmp % 2)
		tmp = tmp / 2
	}

	return binary
}

// ConvertIntToBinary represents a integer number in binary with specific length
//func ConvertBigIntToBinany(i big.Int, len int) ([]byte, error){
//
//	if len%8 != 0 {
//		return nil, fmt.Errorf("length must be divided by 8")
//	}
//
//	binary := make([]byte, len/8)
//
//	str := strconv.FormatInt(int64(i), 2)
//	for j := 0; j < len(str); j++ {
//		binary[j] = ConvertAsciiToInt(str[j])
//	}
//	return binary
//	//fmt.Printf("inddex in binary: %v\n", binary)
//}
