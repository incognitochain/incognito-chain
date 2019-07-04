package transaction

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
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
	listUsableCommitments := make(map[common.Hash][]byte)
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for _, in := range usableInputCoins {
		usableCommitment := in.CoinDetails.CoinCommitment.Compress()
		listUsableCommitments[common.HashH(usableCommitment)] = usableCommitment
		index, err := db.GetCommitmentIndex(*tokenID, usableCommitment, shardID)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)] = index
	}

	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * randNum) - len(listUsableCommitments)
	fmt.Printf("cpRandNum: %d\n", cpRandNum)
	lenCommitment, err1 := db.GetCommitmentLength(*tokenID, shardID)
	if err1 != nil {
		Logger.log.Error(err1)
		return
	}
	if lenCommitment == nil {
		Logger.log.Error(errors.New("Commitments is empty"))
		return
	}
	if lenCommitment.Uint64() == 1 {
		commitmentIndexs = []uint64{0, 0, 0, 0, 0, 0, 0}
		temp := usableInputCoins[0].CoinDetails.CoinCommitment.Compress()
		commitments = [][]byte{temp, temp, temp, temp, temp, temp, temp}
	} else {
		for i := 0; i < cpRandNum; i++ {
			for {
				lenCommitment, _ = db.GetCommitmentLength(*tokenID, shardID)
				index, _ := common.RandBigIntMaxRange(lenCommitment)
				ok, err := db.HasCommitmentIndex(*tokenID, index.Uint64(), shardID)
				if ok && err == nil {
					temp, _ := db.GetCommitmentByIndex(*tokenID, index.Uint64(), shardID)
					if _, found := listUsableCommitments[common.HashH(temp)]; !found {
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
	j := 0
	for _, value := range listUsableCommitments {
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(value, common.ZeroByte)]
		rand := rand.Intn(randNum)
		i := (j * randNum) + rand
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		commitments = append(commitments[:i], append([][]byte{value}, commitments[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
		j++
	}
	return commitmentIndexs, myCommitmentIndexs, commitments
}

// CheckSNDerivatorExistence return true if snd exists in snDerivators list
func CheckSNDerivatorExistence(tokenID *common.Hash, snd *big.Int, shardID byte, db database.DatabaseInterface) (bool, error) {
	ok, err := db.HasSNDerivator(*tokenID, privacy.AddPaddingBigInt(snd, privacy.BigIntSize), shardID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(inputCoins []*privacy.OutputCoin, payments []*privacy.PaymentInfo,
	hasPrivacy bool, metadata metadata.Metadata,
	customTokenParams *CustomTokenParamTx,
	privacyCustomTokenParams *CustomTokenPrivacyParamTx,
	limitFee uint64) uint64 {

	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64

	sizeInfo := uint64(512)

	sizeSigPubKey := uint64(privacy.SigPubKeySize)
	sizeSig := uint64(privacy.SigNoPrivacySize)
	if hasPrivacy {
		sizeSig = uint64(privacy.SigPrivacySize)
	}

	sizeProof := uint64(0)
	if len(inputCoins) != 0 || len(payments) != 0 {
		sizeProof = zkp.EstimateProofSize(len(inputCoins), len(payments), hasPrivacy)
	} else {
		if limitFee > 0 {
			sizeProof = zkp.EstimateProofSize(1, 1, hasPrivacy)
		}
	}

	sizePubKeyLastByte := uint64(1)

	sizeMetadata := uint64(0)
	if metadata != nil {
		sizeMetadata += metadata.CalculateSize()
	}

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata

	// size of custom token data
	if customTokenParams != nil {
		customTokenDataSize := uint64(0)

		customTokenDataSize += uint64(len(customTokenParams.PropertyID))
		customTokenDataSize += uint64(len(customTokenParams.PropertySymbol))
		customTokenDataSize += uint64(len(customTokenParams.PropertyName))

		customTokenDataSize += 8 // for amount
		customTokenDataSize += 4 // for TokenTxType

		for _, out := range customTokenParams.Receiver {
			customTokenDataSize += uint64(len(out.PaymentAddress.Bytes()))
			customTokenDataSize += 8 //out.Value
		}

		for _, in := range customTokenParams.vins {
			customTokenDataSize += uint64(len(in.PaymentAddress.Bytes()))
			customTokenDataSize += uint64(len(in.TxCustomTokenID[:]))
			customTokenDataSize += uint64(len(in.Signature))
			customTokenDataSize += uint64(4) //in.VoutIndex
		}
		sizeTx += customTokenDataSize
	}

	// size of privacy custom token  data
	if privacyCustomTokenParams != nil {
		customTokenDataSize := uint64(0)

		customTokenDataSize += uint64(len(privacyCustomTokenParams.PropertyID))
		customTokenDataSize += uint64(len(privacyCustomTokenParams.PropertySymbol))
		customTokenDataSize += uint64(len(privacyCustomTokenParams.PropertyName))

		customTokenDataSize += 8 // for amount
		customTokenDataSize += 4 // for TokenTxType

		customTokenDataSize += uint64(1) // int8 version
		customTokenDataSize += uint64(5) // string, max : 5 type
		customTokenDataSize += uint64(8) // int64 locktime
		customTokenDataSize += uint64(8) // uint64 fee

		customTokenDataSize += uint64(64) // info

		customTokenDataSize += uint64(privacy.SigPubKeySize)  // sig pubkey
		customTokenDataSize += uint64(privacy.SigPrivacySize) // sig

		// Proof
		customTokenDataSize += zkp.EstimateProofSize(len(privacyCustomTokenParams.TokenInput), len(privacyCustomTokenParams.Receiver), true)

		customTokenDataSize += uint64(1) //PubKeyLastByte

		sizeTx += customTokenDataSize
	}

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
