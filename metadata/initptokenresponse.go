package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"
)

type InitPTokenResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewInitPTokenResponse(requestedTxID common.Hash, metaType int) *InitPTokenResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &InitPTokenResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes InitPTokenResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes InitPTokenResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes InitPTokenResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes InitPTokenResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes InitPTokenResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *InitPTokenResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes InitPTokenResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not InitPTokenETHRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(InitPTokenRequestMeta) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var initPTokenAcceptedInst InitPTokenAcceptedInst
		err = json.Unmarshal(contentBytes, &initPTokenAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], initPTokenAcceptedInst.TxReqID[:]) {
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(initPTokenAcceptedInst.ReceiverAddr.Pk[:], pk[:]) ||
			initPTokenAcceptedInst.Amount != amount ||
			!bytes.Equal(initPTokenAcceptedInst.IncTokenID[:], assetID[:]) ||
			initPTokenAcceptedInst.ShardID != shardID {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no InitPTokenRequest tx found for the InitPTokenResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
