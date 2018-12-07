package privacy

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	rand2 "math/rand"
	"time"
)

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
func RandByte()  byte{
	var res byte
	res = 0
	var bit byte
	rand2.Seed(time.Now().UnixNano())
	for i:=0;i<8;i++{
		bit=byte(rand2.Intn(2))
		res += bit<<byte(i)
	}
	return res
}
// RandInt generates a big int with value less than order of group of elliptic points
func RandInt() *big.Int {
	//Todo: thunderbird
	//Thunderbird had done: random a 32-bytes big interger
	for {
	Int_bytes:=make([]byte,BigIntSize)
	for i:=0;i<BigIntSize;i++{
		Int_bytes[i] = RandByte()
		}
		randNum :=new(big.Int).SetBytes(Int_bytes)
		if(TestRandInt(randNum) && randNum.Cmp(Curve.Params().N)==-1){
			return randNum
		}
	}
}
func TestRandInt(a *big.Int) bool{
	threshold_test:= 0.01
	length:=a.BitLen()
	zero_count:=0
	one_count:=0
	for i:=0;i<length;i++{
		if(a.Bit(i)==1) {
			one_count++
		}
		if(a.Bit(i)==0){
			zero_count++
		}
	}
	if math.Abs(1-float64(zero_count)/float64(one_count))<=threshold_test{
		return true
	}
	return false
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
func getindex(bigint *big.Int, stableSz int) int {
	return  stableSz - len(bigint.Bytes())
}
func PadBigInt(fckBigInt *big.Int, stableSz int) []byte{
	idx:=getindex(fckBigInt,stableSz)
	paddedBig:=make([]byte, stableSz)
	for i:=0;i<idx;i++{
		paddedBig[i] = byte(0x00)
	}
	for i:=idx;i<stableSz;i++{
		paddedBig[i] = fckBigInt.Bytes()[i-idx]
	}
	return paddedBig
}
//
