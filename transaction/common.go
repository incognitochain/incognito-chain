package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
	"math"
	"math/big"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/common"
	"math/rand"
)

// ConvertOutputCoinToInputCoin - convert output coin from old tx to input coin for new tx
func ConvertOutputCoinToInputCoin(usableOutputsOfOld []*privacy.OutputCoin) []*privacy.InputCoin {
	var inputCoins []*privacy.InputCoin

	for _, coin := range usableOutputsOfOld {
		inCoin := new(privacy.InputCoin)
		inCoin.CoinDetails = coin.CoinDetails
		inputCoins = append(inputCoins, inCoin)
	}
	return inputCoins
}

// RandomCommitmentsProcess - process list commitments and useable tx to create
// a list commitment random which be used to create a proof for new tx
// result contains
// commitmentIndexs = [{1,2,3,4,myindex1,6,7,8}{9,10,11,12,13,myindex2,15,16}...]
// myCommitmentIndexs = [4, 13, ...]
func RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, db database.DatabaseInterface, chainID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64) {
	commitmentIndexs = []uint64{}   // : list commitment indexes which: random from full db commitments + commitments of usableInputCoins
	myCommitmentIndexs = []uint64{} // : list indexes of commitments(usableInputCoins) in {commitmentIndexs}
	if randNum == 0 {
		randNum = 8 // default
	}

	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := [][]byte{}
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for _, in := range usableInputCoins {
		usableCommitment := in.CoinDetails.CoinCommitment.Compress()
		listUsableCommitments = append(listUsableCommitments, usableCommitment)
		index, _ := db.GetCommitmentIndex(tokenID, usableCommitment, chainID)
		mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(usableCommitment, byte(0x00))] = index
	}

	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * randNum) - len(listUsableCommitments)
	for i := 0; i < cpRandNum; i++ {
		for true {
			lenCommitment, _ := db.GetCommitmentLength(tokenID, chainID)
			index, _ := common.RandBigIntN(lenCommitment)
			ok, err := db.HasCommitmentIndex(tokenID, index.Uint64(), chainID)
			if ok && err == nil {
				temp, _ := db.GetCommitmentByIndex(tokenID, index.Uint64(), chainID)
				if index2, err := common.SliceBytesExists(listUsableCommitments, temp); index2 == -1 && err == nil {
					// random commitment not in commitments of usableinputcoin
					commitmentIndexs = append(commitmentIndexs, index.Uint64())
					break
				}
			} else {
				continue
			}
		}
	}

	// loop to insert usable commitments into commitmentIndexs for every group
	for j, temp := range listUsableCommitments {
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(temp, byte(0x00))]
		rand := rand.Intn(randNum)
		i := (j * randNum) + rand
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
	}
	return commitmentIndexs, myCommitmentIndexs
}

// CheckSNDerivatorExistence return true if snd exists in snDerivators list
func CheckSNDerivatorExistence(tokenID *common.Hash, snd *big.Int, chainID byte, db database.DatabaseInterface) (bool, error) {
	ok, err := db.HasSNDerivator(tokenID, *snd, chainID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(inputCoins []*privacy.OutputCoin, payments []*privacy.PaymentInfo) uint64 {
	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64

	sizeSigPubKey := uint64(privacy.SigPubKeySize)
	sizeSig := uint64(privacy.SigSize)
	sizeProof := zkp.EstimateProofSize(inputCoins, payments)

	sizePubKeyLastByte := uint64(1)
	// TODO 0xjackpolope
	//sizeMetadata :=

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}
