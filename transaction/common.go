package transaction

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/privacy/zeroknowledge"
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
func RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	commitmentIndexs = []uint64{} // : list commitment indexes which: random from full db commitments + commitments of usableInputCoins
	commitments = [][]byte{}
	myCommitmentIndexs = []uint64{} // : list indexes of commitments(usableInputCoins) in {commitmentIndexs}
	if randNum == 0 {
		randNum = privacy.CMRingSize // default
	}

	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := [][]byte{}
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for _, in := range usableInputCoins {
		usableCommitment := in.CoinDetails.CoinCommitment.Compress()
		listUsableCommitments = append(listUsableCommitments, usableCommitment)
		index, _ := db.GetCommitmentIndex(tokenID, usableCommitment, shardID)
		mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)] = index
	}

	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * randNum) - len(listUsableCommitments)
	fmt.Printf("cpRandNum: %d\n", cpRandNum)
	lenCommitment, _ := db.GetCommitmentLength(tokenID, shardID)
	if lenCommitment.Uint64() == 1 {
		commitmentIndexs = []uint64{0, 0, 0, 0, 0, 0, 0}
	} else {
		for i := 0; i < cpRandNum; i++ {
			for {
				lenCommitment, _ = db.GetCommitmentLength(tokenID, shardID)
				index, _ := common.RandBigIntN(lenCommitment)
				ok, err := db.HasCommitmentIndex(tokenID, index.Uint64(), shardID)
				if ok && err == nil {
					temp, _ := db.GetCommitmentByIndex(tokenID, index.Uint64(), shardID)
					if index2, err := common.SliceBytesExists(listUsableCommitments, temp); index2 == -1 && err == nil {
						// random commitment not in commitments of usableinputcoin
						commitmentIndexs = append(commitmentIndexs, index.Uint64())
						commitments = append(commitments, temp)
						break
					}
				} else {
					continue
				}
			}
		}
	}

	// loop to insert usable commitments into commitmentIndexs for every group
	for j, temp := range listUsableCommitments {
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(temp, common.ZeroByte)]
		rand := rand.Intn(randNum)
		i := (j * randNum) + rand
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
	}
	return commitmentIndexs, myCommitmentIndexs, commitments
}

// CheckSNDerivatorExistence return true if snd exists in snDerivators list
func CheckSNDerivatorExistence(tokenID *common.Hash, snd *big.Int, shardID byte, db database.DatabaseInterface) (bool, error) {
	ok, err := db.HasSNDerivator(tokenID, *snd, shardID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(inputCoins []*privacy.OutputCoin, payments []*privacy.PaymentInfo, hasPrivacy bool) uint64 {
	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64

	sizeInfo := uint64(0)
	if hasPrivacy {
		sizeInfo = uint64(64)
	}

	sizeSigPubKey := uint64(privacy.SigPubKeySize)
	sizeSig := uint64(privacy.SigNoPrivacySize)
	if hasPrivacy {
		sizeSig = uint64(privacy.SigPrivacySize)
	}

	sizeProof := zkp.EstimateProofSize(len(inputCoins), len(payments), hasPrivacy)

	sizePubKeyLastByte := uint64(1)

	// TODO 0xjackpolope
	sizeMetadata := uint64(0)

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}

// SortTxsByLockTime sorts txs by lock time
func SortTxsByLockTime(txs []metadata.Transaction, isDesc bool) []metadata.Transaction {
	sort.Slice(txs, func(i, j int) bool {
		if isDesc {
			return txs[i].GetLockTime() > txs[j].GetLockTime()
		}
		return txs[i].GetLockTime() <= txs[j].GetLockTime()
	})
	return txs
}

func TxToIns(tx metadata.Transaction) []string {
	//todo @0xjackalope
	a := []string{""}
	return a
}
