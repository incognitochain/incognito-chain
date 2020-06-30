package privacy_v2

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)
// TEST DURATION NOTE : 100 iterations of 1-to-12 coins = 15sec
var (
	numOfLoops = 100
	minOutCoinCount = 1
	maxOutCoinCount = 12

)

func TestPaymentV2InitAndMarshalling(t *testing.T) {
	for loop:=0;loop<=numOfLoops;loop++{
		outCoinCount := common.RandInt() % (maxOutCoinCount-minOutCoinCount+1) + minOutCoinCount
		// make some dummy private keys for our dummy users
		dummyPrivateKeys := make([]*operation.Scalar,outCoinCount)
		for i,_ := range dummyPrivateKeys{
			dummyPrivateKeys[i] = operation.RandomScalar()
		}
		// each of these dummy users are provided a (not confirmed by blockchain) coinv2 of value 3000
		// paymentAdress is persistent and held by this user, while the OTA is inside the coin
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, pk := range dummyPrivateKeys {
			pkb := pk.ToBytes()
			paymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(pkb[:]),3000,[]byte{})
		}
		inputCoins := make([]coin.PlainCoin, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			inputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
		// in this test, each user will send themselves 2000 and the rest is txfee
		for _,pInf := range paymentInfo{
			pInf.Amount = 2000
		}
		outputCoins := make([]*coin.CoinV2, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			outputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
	// prove and verify without privacy (no bulletproof)
	// also marshal to byte and back
		proof, err := Prove(inputCoins, outputCoins, false, paymentInfo)
		assert.Equal(t, nil, err)
		b := proof.Bytes()

		temp := new(PaymentProofV2)
		err = temp.SetBytes(b)
		b2 := temp.Bytes()
		assert.Equal(t, true, bytes.Equal(b2, b))

		correct,err := proof.Verify(false, nil, uint64(1000*outCoinCount), byte(0), nil, false, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t,true,correct)
	}
}

func TestPaymentV2ProveWithPrivacy(t *testing.T) {
	outCoinCount := common.RandInt() % (maxOutCoinCount-minOutCoinCount+1) + minOutCoinCount
	for loop:=0;loop<numOfLoops;loop++{
		// make some dummy private keys for our dummy users
		dummyPrivateKeys := make([]*operation.Scalar,outCoinCount)
		for i,_ := range dummyPrivateKeys{
			dummyPrivateKeys[i] = operation.RandomScalar()
		}
		// each of these dummy users are provided a (not confirmed by blockchain) coinv2 of value 3000
		// paymentAdress is persistent and held by this user, while the OTA is inside the coin
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, pk := range dummyPrivateKeys {
			pkb := pk.ToBytes()
			paymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(pkb[:]),3000,[]byte{})
		}
		inputCoins := make([]coin.PlainCoin, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			inputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
		// in this test, each user will send some other rando 2500 and the rest is txfee
		outPaymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, _ := range dummyPrivateKeys {
			otherPriv := operation.RandomScalar()
			pkb := otherPriv.ToBytes()
			outPaymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(pkb[:]),2500,[]byte{})
		}
		outputCoins := make([]*coin.CoinV2, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			outputCoins[i],err = coin.NewCoinFromPaymentInfo(outPaymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
	// prove and verify with privacy using bulletproof
	// note that bulletproofs only assure each outcoin amount is in uint64 range
	// while the equality suminput = suminput + sumfee must be checked using mlsag later
	// here our mock scenario has out+fee<in but passes anyway
		proof, err := Prove(inputCoins, outputCoins, true, paymentInfo)
		assert.Equal(t, nil, err)
		isSane, err := proof.ValidateSanity()
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isSane)

		isValid,err := proof.Verify(true, nil, uint64(200*outCoinCount), byte(0), nil, false, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t,true,isValid)
	}
}
