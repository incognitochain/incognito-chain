package metadata
//
//import (
//	"fmt"
//	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
//	"github.com/incognitochain/incognito-chain/retriever"
//	"math"
//	"strconv"
//
//	"github.com/incognitochain/incognito-chain/common"
//)
//
//type MetadataBase struct {
//	Type int
//}
//
//func NewMetadataBase(thisType int) *MetadataBase {
//	return &MetadataBase{Type: thisType}
//}
//
//func (mb MetadataBase) IsMinerCreatedMetaType() bool {
//	metaType := mb.GetType()
//	for _, mType := range minerCreatedMetaTypes {
//		if metaType == mType {
//			return true
//		}
//	}
//	return false
//}
//
//func (mb *MetadataBase) CalculateSize() uint64 {
//	return 0
//}
//
//func (mb *MetadataBase) Validate() error {
//	return nil
//}
//
//func (mb *MetadataBase) Process() error {
//	return nil
//}
//
//func (mb MetadataBase) GetType() int {
//	return mb.Type
//}
//
//func (mb MetadataBase) Hash() *common.Hash {
//	record := strconv.Itoa(mb.Type)
//	data := []byte(record)
//	hash := common.HashH(data)
//	return &hash
//}
//
//func (mb MetadataBase) CheckTransactionFee(tx retriever.Transaction, minFeePerKbTx uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
//	if tx.GetType() == common.TxCustomTokenPrivacyType {
//		feeNativeToken := tx.GetTxFee()
//		feePToken := tx.GetTxFeeToken()
//		if feePToken > 0 {
//			tokenID := tx.GetTokenID()
//			feePTokenToNativeTokenTmp, err := ConvertPrivacyTokenToNativeToken(feePToken, tx.GetTokenID(), beaconHeight, stateDB)
//			if err != nil {
//				fmt.Printf("transaction %+v: %+v %v can not convert to native token",
//					tx.Hash().String(), feePToken, tokenID)
//				return false
//			}
//			feePTokenToNativeToken := uint64(math.Ceil(feePTokenToNativeTokenTmp))
//			feeNativeToken += feePTokenToNativeToken
//		}
//		// get limit fee in native token
//		actualTxSize := tx.GetTxActualSize()
//		// check fee in native token
//		minFee := actualTxSize * minFeePerKbTx
//		if feeNativeToken < minFee {
//			fmt.Printf("transaction %+v has %d fees PRV which is under the required amount of %d, tx size %d",
//				tx.Hash().String(), feeNativeToken, minFee, actualTxSize)
//			return false
//		}
//		return true
//	}
//	// normal privacy tx
//	txFee := tx.GetTxFee()
//	fullFee := minFeePerKbTx * tx.GetTxActualSize()
//	return !(txFee < fullFee)
//}
//
//func (mb *MetadataBase) BuildReqActions(tx retriever.Transaction, chainRetriever retriever.ChainRetriever, shardViewRetriever retriever.ShardViewRetriever, beaconViewRetriever retriever.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
//	return [][]string{}, nil
//}
//
//func (mb MetadataBase) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []retriever.Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx retriever.Transaction, chainRetriever retriever.ChainRetriever, ac *retriever.AccumulatedValues, shardViewRetriever retriever.ShardViewRetriever, beaconViewRetriever retriever.BeaconViewRetriever) (bool, error) {
//	return true, nil
//}
