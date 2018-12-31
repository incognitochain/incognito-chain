package privacy

import (
	"math/big"
	"testing"

	"github.com/ninjadotorg/constant/common"
	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	address := GeneratePublicKey(spendingKey)
	hash := common.HashB([]byte("hello"))
	signature, _ := Sign(hash, spendingKey)
	valid := Verify(signature, hash[:], address)
	assert.Equal(t, true, valid)
}
