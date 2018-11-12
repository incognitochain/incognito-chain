package privacy

import (
	"fmt"
	"math/big"
	"strconv"
)

//inversePoint return inverse point of ECC Point input
func inversePoint(eccPoint EllipticPoint) (*EllipticPoint, error) {
	//Check that input is ECC point
	if !Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
		return nil, fmt.Errorf("Input is not ECC Point")
	}
	//Create result point
	resPoint := new(EllipticPoint)
	resPoint.X = big.NewInt(0)
	resPoint.Y = big.NewInt(0)

	//inverse point of A(x,y) in ECC is A'(x, P - y) with P is order of Curve
	resPoint.X.SetBytes(eccPoint.X.Bytes())
	resPoint.Y.SetBytes(eccPoint.Y.Bytes())
	resPoint.Y.Sub(Curve.Params().P, resPoint.Y)

	return resPoint, nil
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

func ConvertAsciiToInt(c uint8) byte {
	return byte(c - 48)
}

// ConvertIntToBinany represents a integer number in binary
func ConvertIntToBinany(i int) []byte{

	binary := make([]byte, 32)
	str := strconv.FormatInt(int64(i), 2)
	for j := 0; j < len(str); j++ {
		binary[j] = ConvertAsciiToInt(str[j])
	}
	return binary
	//fmt.Printf("inddex in binary: %v\n", binary)
}
