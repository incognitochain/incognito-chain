package privacy

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/common"
)

var isTesting bool

//#if isTesting
var pubkeyTest []*EllipticPoint
var RTest []*EllipticPoint
var mutex sync.Mutex
var counter int
var wg sync.WaitGroup
var wgchild sync.WaitGroup
var Numbs int

//#endif

type PrivateKeySchnorr struct {
	V  *big.Int
	pk *EllipticPoint
}

func generateRandom() *big.Int {
	//not safe, just for testing
	res := RandInt()
	return res
}

func generateCommonParams(listPubkey, listR []*EllipticPoint, mess []byte) (*EllipticPoint, *big.Int, *EllipticPoint, *EllipticPoint) {
	aggPubkey := new(EllipticPoint)
	aggPubkey.X = big.NewInt(0)
	aggPubkey.Y = big.NewInt(0)

	for i := 0; i < len(listPubkey); i++ {
		aggPubkey = aggPubkey.Add(listPubkey[i])
	}

	R := new(EllipticPoint)
	R.X = big.NewInt(0)
	R.Y = big.NewInt(0)
	for i := 0; i < len(listR); i++ {
		R = R.Add(listR[i])
	}
	//Curve.ScalarMult(Curve.Params().Gx, Curve.Params().Gy, ri)

	X := new(EllipticPoint)
	X.X = big.NewInt(0)
	X.Y = big.NewInt(0)

	for i := 0; i < len(listPubkey); i++ {
		temp := aggPubkey.Add(listPubkey[i])
		a := common.DoubleHashB(temp.Compress())
		aInt := big.NewInt(0)
		aInt.SetBytes(a)
		X = X.Add(listPubkey[i].ScalarMul(aInt))
	}

	Cbyte := X.Compress()
	Cbyte = append(Cbyte, R.Compress()...)
	Cbyte = append(Cbyte, mess...)
	C := big.NewInt(0)
	C.SetBytes(Cbyte)
	C.Mod(C, Curve.Params().N)

	return aggPubkey, C, R, X
}

func getListPublicKey() []*EllipticPoint {
	if isTesting {
		return pubkeyTest
	}
	//todo
	return nil
}

func getListR() []*EllipticPoint {
	if isTesting {
		return RTest
	}
	//todo
	return nil
}

func broadcastR(R *EllipticPoint) {
	if isTesting {
		mutex.Lock()
		RTest[counter] = R
		counter++
		mutex.Unlock()
	}
	//todo
}

func (priKey *PrivateKeySchnorr) SignMultiSig(mess []byte) (*big.Int, *EllipticPoint) {
	r := generateRandom()
	selfR := new(EllipticPoint)
	selfR.X = big.NewInt(0)
	selfR.Y = big.NewInt(0)
	selfR.X.Set(Curve.Params().Gx)
	selfR.Y.Set(Curve.Params().Gy)
	selfR = selfR.ScalarMul(r)

	broadcastR(selfR)

	time.Sleep(800 * time.Millisecond)

	for counter < Numbs {
	}
	listPubkey := getListPublicKey()
	listR := getListR()

	aggKey, C, R, _ := generateCommonParams(listPubkey, listR, mess)
	temp := aggKey.Add(priKey.pk)
	a := common.DoubleHashB(temp.Compress())
	aInt := big.NewInt(0)
	aInt.SetBytes(a)
	aInt.Mod(aInt, Curve.Params().N)

	sig := big.NewInt(0)
	sig.Set(aInt)
	sig.Mul(sig, C)
	sig.Mod(sig, Curve.Params().N)
	sig.Mul(sig, priKey.V)
	sig.Mod(sig, Curve.Params().N)
	sig.Add(sig, r)
	sig.Mod(sig, Curve.Params().N)
	return sig, R
}

func VerifyMultiSig(R *EllipticPoint, S *big.Int, mess []byte) bool {
	listPubkey := getListPublicKey()
	listR := getListR()
	_, C, R, X := generateCommonParams(listPubkey, listR, mess)
	//GSPoint is G^S
	GSPoint := new(EllipticPoint)
	GSPoint.X, GSPoint.Y = big.NewInt(0), big.NewInt(0)
	GSPoint.X.Set(Curve.Params().Gx)
	GSPoint.Y.Set(Curve.Params().Gy)
	GSPoint = GSPoint.ScalarMul(S)
	//RXCPoint is R.X^C
	RXCPoint := X.ScalarMul(C)
	RXCPoint = RXCPoint.Add(R)
	return GSPoint.IsEqual(RXCPoint)
}

func TestMultiSig() {
	isTesting = true
	Numbs = 40
	counter = 0
	listSigners := make([]PrivateKeySchnorr, Numbs)
	pubkeyTest = make([]*EllipticPoint, Numbs)
	RTest = make([]*EllipticPoint, Numbs)
	Sig := make([]*big.Int, Numbs)
	R := new(EllipticPoint)
	for i := 0; i < Numbs; i++ {
		listSigners[i].V = RandInt()
		listSigners[i].pk = new(EllipticPoint)
		listSigners[i].pk.X = big.NewInt(0)
		listSigners[i].pk.Y = big.NewInt(0)
		listSigners[i].pk.X.Set(Curve.Params().Gx)
		listSigners[i].pk.Y.Set(Curve.Params().Gy)
		listSigners[i].pk = listSigners[i].pk.ScalarMul(listSigners[i].V)
		pubkeyTest[i] = listSigners[i].pk
	}
	for i := 0; i < Numbs; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			Sig[j], R = listSigners[j].SignMultiSig([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
		}(i)
	}
	wg.Wait()
	aggSig := big.NewInt(0)
	for i := 0; i < Numbs; i++ {
		aggSig.Add(aggSig, Sig[i])
		aggSig.Mod(aggSig, Curve.Params().N)
	}
	fmt.Printf("EC Schnorr testing: %v\n", VerifyMultiSig(R, aggSig, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
}
