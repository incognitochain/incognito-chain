package privacy

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSpendingKey(t *testing.T) {
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	fmt.Printf("\nSpending key: %v\n", spendingKey)
	fmt.Println(len(spendingKey))
	expectedResult := []byte{2, 31, 181, 150, 219, 129, 230, 208, 43, 243, 210, 88, 110, 227, 152, 31, 229, 25, 242, 117, 192, 172, 156, 167, 107, 188, 242, 235, 180, 9, 125, 150}

	assert.Equal(t, expectedResult, spendingKey)
}

func TestPAdd1Div4(t *testing.T) {
	res := new(big.Int)
	res = PAdd1Div4(new(big.Int).SetInt64(123))
	expectedResult := new(big.Int).SetInt64(31)
	assert.Equal(t, expectedResult, res)

}

// func TestGetCurve(t *testing.T) {
// 	curve1 := GetCurve()
// 	curve2 := GetCurve()
// 	fmt.Printf("Pk curve 1: %v\n", &curve1)
// 	fmt.Printf("Pk curve 2: %v\n", &curve2)
// 	assert.Equal(t, &curve1, &curve2)
// }
