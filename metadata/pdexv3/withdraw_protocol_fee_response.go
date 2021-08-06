package pdexv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type WithdrawalProtocolFeeResponse struct {
	metadataCommon.MetadataBase
	TokenType string      `json:"TokenType"`
	ReqTxID   common.Hash `json:"ReqTxID"`
}

func NewPdexv3WithdrawalProtocolFeeResponse(
	metaType int,
	tokenType string,
	reqTxID common.Hash,
) *WithdrawalProtocolFeeResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalProtocolFeeResponse{
		MetadataBase: *metadataBase,
		TokenType:    tokenType,
		ReqTxID:      reqTxID,
	}
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return withdrawalResponse.Type == metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) Hash() *common.Hash {
	record := withdrawalResponse.MetadataBase.Hash().String()
	record += withdrawalResponse.TokenType
	record += withdrawalResponse.ReqTxID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawalResponse *WithdrawalProtocolFeeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawalResponse)
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte, tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	// verify mining tx with the request tx
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not WithdrawalProtocolFeeResponse instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta)) {
			continue
		}

		contentBytes := []byte(inst[3])
		var instContent WithdrawalProtocolFeeContent
		err := json.Unmarshal(contentBytes, &instContent)
		if err != nil {
			continue
		}

		tokenType := instContent.TokenType
		if withdrawalResponse.TokenType != tokenType {
			continue
		}

		shardIDFromInst := instContent.ShardID
		txReqIDFromInst := instContent.TxReqID
		receiver := instContent.Receiver

		receiverAddress, err := isValidReceiverAddressStr(receiver.AddressStr, shardIDFromInst)
		if err != nil {
			continue
		}

		if !bytes.Equal(withdrawalResponse.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, assetID, err := tx.GetTxMintData()
		if err != nil {
			continue
		}
		if !isMinted {
			continue
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()
		txR := mintCoin.(*coin.CoinV2).GetTxRandom()

		publicKey := receiverAddress.PublicKey
		txRandom := receiverAddress.TxRandom

		if !bytes.Equal(publicKey.ToBytesS(), pk[:]) ||
			receiver.Amount != paidAmount ||
			!bytes.Equal(txR[:], txRandom[:]) ||
			receiver.TokenID != *assetID {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("No WithdrawalProtocolFee instruction found for WithdrawalProtocolFeeResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
