package transaction

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"math"
	"math/big"
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

type RandomCommitmentsProcessParam struct {
	usableInputCoins []*privacy.InputCoin
	randNum          int
	db               database.DatabaseInterface
	shardID          byte
	tokenID          *common.Hash
}

func NewRandomCommitmentsProcessParam(usableInputCoins []*privacy.InputCoin, randNum int,
	db database.DatabaseInterface, shardID byte, tokenID *common.Hash) *RandomCommitmentsProcessParam {
	result := &RandomCommitmentsProcessParam{
		tokenID:          tokenID,
		shardID:          shardID,
		db:               db,
		randNum:          randNum,
		usableInputCoins: usableInputCoins,
	}
	return result
}

// RandomCommitmentsProcess - process list commitments and useable tx to create
// a list commitment random which be used to create a proof for new tx
// result contains
// commitmentIndexs = [{1,2,3,4,myindex1,6,7,8}{9,10,11,12,13,myindex2,15,16}...]
// myCommitmentIndexs = [4, 13, ...]
func RandomCommitmentsProcess(param *RandomCommitmentsProcessParam) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	commitmentIndexs = []uint64{} // : list commitment indexes which: random from full db commitments + commitments of usableInputCoins
	commitments = [][]byte{}
	myCommitmentIndexs = []uint64{} // : list indexes of commitments(usableInputCoins) in {commitmentIndexs}
	if param.randNum == 0 {
		param.randNum = privacy.CommitmentRingSize // default
	}

	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := make(map[common.Hash][]byte)
	listUsableCommitmentsIndices := make([]common.Hash, len(param.usableInputCoins))
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)

	for i, in := range param.usableInputCoins {
		usableCommitment := in.CoinDetails.GetCoinCommitment().ToBytesS()
		commitmentInHash := common.HashH(usableCommitment)
		listUsableCommitments[commitmentInHash] = usableCommitment
		listUsableCommitmentsIndices[i] = commitmentInHash

		index, err := param.db.GetCommitmentIndex(*param.tokenID, usableCommitment, param.shardID)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		commitmentInBase58Check := base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)
		mapIndexCommitmentsInUsableTx[commitmentInBase58Check] = index
	}

	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * param.randNum) - len(listUsableCommitments)
	//fmt.Printf("cpRandNum: %d\n", cpRandNum)
	lenCommitment, err1 := param.db.GetCommitmentLength(*param.tokenID, param.shardID)
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
		temp := param.usableInputCoins[0].CoinDetails.GetCoinCommitment().ToBytesS()
		commitments = [][]byte{temp, temp, temp, temp, temp, temp, temp}
	} else {
		for i := 0; i < cpRandNum; i++ {
			for {
				lenCommitment, _ = param.db.GetCommitmentLength(*param.tokenID, param.shardID)
				index, _ := common.RandBigIntMaxRange(lenCommitment)
				ok, err := param.db.HasCommitmentIndex(*param.tokenID, index.Uint64(), param.shardID)
				if ok && err == nil {
					temp, _ := param.db.GetCommitmentByIndex(*param.tokenID, index.Uint64(), param.shardID)
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
	for _, commitmentInHash := range listUsableCommitmentsIndices {
		commitmentValue := listUsableCommitments[commitmentInHash]
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(commitmentValue, common.ZeroByte)]
		randInt := rand.Intn(param.randNum)
		i := (j * param.randNum) + randInt
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		commitments = append(commitments[:i], append([][]byte{commitmentValue}, commitments[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
		j++
	}
	return commitmentIndexs, myCommitmentIndexs, commitments
}

// CheckSNDerivatorExistence return true if snd exists in snDerivators list
func CheckSNDerivatorExistence(tokenID *common.Hash, snd *privacy.Scalar, db database.DatabaseInterface) (bool, error) {
	ok, err := db.HasSNDerivator(*tokenID, snd.ToBytesS())
	if err != nil {
		return false, err
	}
	return ok, nil
}

type EstimateTxSizeParam struct {
	numInputCoins            int
	numPayments              int
	hasPrivacy               bool
	metadata                 metadata.Metadata
	privacyCustomTokenParams *CustomTokenPrivacyParamTx
	limitFee                 uint64
}

func NewEstimateTxSizeParam(numInputCoins, numPayments int,
	hasPrivacy bool, metadata metadata.Metadata,
	privacyCustomTokenParams *CustomTokenPrivacyParamTx,
	limitFee uint64) *EstimateTxSizeParam {
	estimateTxSizeParam := &EstimateTxSizeParam{
		numInputCoins:            numInputCoins,
		numPayments:              numPayments,
		hasPrivacy:               hasPrivacy,
		limitFee:                 limitFee,
		metadata:                 metadata,
		privacyCustomTokenParams: privacyCustomTokenParams,
	}
	return estimateTxSizeParam
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(estimateTxSizeParam *EstimateTxSizeParam) uint64 {

	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64

	sizeInfo := uint64(512)

	sizeSigPubKey := uint64(common.SigPubKeySize)
	sizeSig := uint64(common.SigNoPrivacySize)
	if estimateTxSizeParam.hasPrivacy {
		sizeSig = uint64(common.SigPrivacySize)
	}

	sizeProof := uint64(0)
	if estimateTxSizeParam.numInputCoins != 0 || estimateTxSizeParam.numPayments != 0 {
		sizeProof = utils.EstimateProofSize(estimateTxSizeParam.numInputCoins, estimateTxSizeParam.numPayments, estimateTxSizeParam.hasPrivacy)
	} else {
		if estimateTxSizeParam.limitFee > 0 {
			sizeProof = utils.EstimateProofSize(1, 1, estimateTxSizeParam.hasPrivacy)
		}
	}

	sizePubKeyLastByte := uint64(1)

	sizeMetadata := uint64(0)
	if estimateTxSizeParam.metadata != nil {
		sizeMetadata += estimateTxSizeParam.metadata.CalculateSize()
	}

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata

	// size of privacy custom token  data
	if estimateTxSizeParam.privacyCustomTokenParams != nil {
		customTokenDataSize := uint64(0)

		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyID))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertySymbol))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyName))

		customTokenDataSize += 8 // for amount
		customTokenDataSize += 4 // for TokenTxType

		customTokenDataSize += uint64(1) // int8 version
		customTokenDataSize += uint64(5) // string, max : 5 type
		customTokenDataSize += uint64(8) // int64 locktime
		customTokenDataSize += uint64(8) // uint64 fee

		customTokenDataSize += uint64(64) // info

		customTokenDataSize += uint64(common.SigPubKeySize)  // sig pubkey
		customTokenDataSize += uint64(common.SigPrivacySize) // sig

		// Proof
		customTokenDataSize += utils.EstimateProofSize(len(estimateTxSizeParam.privacyCustomTokenParams.TokenInput), len(estimateTxSizeParam.privacyCustomTokenParams.Receiver), true)

		customTokenDataSize += uint64(1) //PubKeyLastByte

		sizeTx += customTokenDataSize
	}

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}

// SortTxsByLockTime sorts txs by lock time
/*func SortTxsByLockTime(txs []metadata.Transaction, isDesc bool) []metadata.Transaction {
	sort.Slice(txs, func(i, j int) bool {
		if isDesc {
			return txs[i].GetLockTime() > txs[j].GetLockTime()
		}
		return txs[i].GetLockTime() <= txs[j].GetLockTime()
	})
	return txs
}*/

type BuildCoinBaseTxByCoinIDParams struct {
	payToAddress    *privacy.PaymentAddress
	amount          uint64
	payByPrivateKey *privacy.PrivateKey
	db              database.DatabaseInterface
	meta            metadata.Metadata
	coinID          common.Hash
	txType          int
	coinName        string
	shardID         byte
}

func NewBuildCoinBaseTxByCoinIDParams(payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
	coinID common.Hash,
	txType int,
	coinName string,
	shardID byte) *BuildCoinBaseTxByCoinIDParams {
	params := &BuildCoinBaseTxByCoinIDParams{
		db:              db,
		shardID:         shardID,
		meta:            meta,
		amount:          amount,
		coinID:          coinID,
		coinName:        coinName,
		payByPrivateKey: payByPrivateKey,
		payToAddress:    payToAddress,
		txType:          txType,
	}
	return params
}

func BuildCoinBaseTxByCoinID(params *BuildCoinBaseTxByCoinIDParams) (metadata.Transaction, error) {
	switch params.txType {
	case NormalCoinType:
		tx := &Tx{}
		err := tx.InitTxSalary(params.amount, params.payToAddress, params.payByPrivateKey, params.db, params.meta)
		return tx, err
	case CustomTokenPrivacyType:
		var propertyID [common.HashSize]byte
		copy(propertyID[:], params.coinID[:])
		receiver := &privacy.PaymentInfo{
			Amount:         params.amount,
			PaymentAddress: *params.payToAddress,
		}
		propID := common.Hash(propertyID)
		tokenParams := &CustomTokenPrivacyParamTx{
			PropertyID:     propID.String(),
			PropertyName:   params.coinName,
			PropertySymbol: params.coinName,
			Amount:         params.amount,
			TokenTxType:    CustomTokenInit,
			Receiver:       []*privacy.PaymentInfo{receiver},
			TokenInput:     []*privacy.InputCoin{},
			Mintable:       true,
		}
		tx := &TxCustomTokenPrivacy{}
		err := tx.Init(
			NewTxPrivacyTokenInitParams(params.payByPrivateKey,
				[]*privacy.PaymentInfo{},
				nil,
				0,
				tokenParams,
				params.db,
				params.meta,
				false,
				false,
				params.shardID,
				nil))
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return tx, nil
	}
	return nil, nil
}

// IsBridgeTokenID finds the external tokenID for a bridge token from database
func IsBridgeTokenID(tokenID common.Hash, db database.DatabaseInterface) (bool, error) {
	allBridgeTokensBytes, err := db.GetAllBridgeTokens()
	if err != nil {
		return false, err
	}
	var allBridgeTokens []*lvdb.BridgeTokenInfo
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return false, err
	}
	for _, token := range allBridgeTokens {
		if token.TokenID.IsEqual(&tokenID) {
			return true, nil
		}
	}
	return false, errors.New("invalid bridge tokenID")
}

func ValidateBatchTxsByItself(
	txList []*Tx,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever) (bool, error, int) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err, -1
	}

	bulletProofList := make([]*aggregaterange.AggregatedRangeProof, 0)
	for i, tx := range txList {
		shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		hasPrivacy := tx.IsPrivacy()
		ok, err := tx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID, true)
		if !ok {
			return false, err, i
		}
		if tx.Metadata != nil {
			if hasPrivacy {
				return false, errors.New("Metadata can not exist in not privacy tx"), i
			}
			validateMetadata := tx.Metadata.ValidateMetadataByItself()
			if validateMetadata {
				return validateMetadata, nil, i
			} else {
				return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid")), i
			}
		}
		if hasPrivacy {
			if bulletProof := tx.Proof.GetAggregatedRangeProof(); bulletProof != nil {
				bulletProofList = append(bulletProofList, bulletProof)
			}
		}
	}
	ok, error, i := aggregaterange.VerifyBatchingAggregatedRangeProofs(bulletProofList)
	if !ok {
		if error != nil {
			Logger.log.Error(error)
		}
		Logger.log.Errorf("FAILED VERIFICATION BATCH PAYMENT PROOF %d", i)
		return false, NewTransactionErr(TxProofVerifyFailError, error), -1
	}

	return true, nil, -1
}
