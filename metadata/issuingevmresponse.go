package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type IssuingEVMResponse struct {
	MetadataBase
	RequestedTxID   common.Hash `json:"RequestedTxID"`
	UniqTx          []byte      `json:"UniqETHTx"`
	ExternalTokenID []byte      `json:"ExternalTokenID"`
	SharedRandom    []byte      `json:"SharedRandom,omitempty"`
}

type IssuingEVMResAction struct {
	Meta       *IssuingEVMResponse `json:"meta"`
	IncTokenID *common.Hash        `json:"incTokenID"`
}

func NewIssuingEVMResponse(
	requestedTxID common.Hash,
	uniqTx []byte,
	externalTokenID []byte,
	metaType int,
) *IssuingEVMResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingEVMResponse{
		RequestedTxID:   requestedTxID,
		UniqTx:          uniqTx,
		ExternalTokenID: externalTokenID,
		MetadataBase:    metadataBase,
	}
}

func (iRes IssuingEVMResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingEVMResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingEVMResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingEVMResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes IssuingEVMResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += string(iRes.UniqTx)
	record += string(iRes.ExternalTokenID)
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingEVMResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes IssuingEVMResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not IssuingEVMRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			(instMetaType != strconv.Itoa(IssuingETHRequestMeta) && instMetaType != strconv.Itoa(IssuingBSCRequestMeta) &&
				instMetaType != strconv.Itoa(IssuingPRVERC20RequestMeta) && instMetaType != strconv.Itoa(IssuingPRVBEP20RequestMeta) &&
				instMetaType != strconv.Itoa(IssuingPLGRequestMeta) && instMetaType != strconv.Itoa(IssuingFantomRequestMeta)) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var issuingEVMAcceptedInst IssuingEVMAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingEVMAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], issuingEVMAcceptedInst.TxReqID[:]) ||
			!bytes.Equal(iRes.UniqTx, issuingEVMAcceptedInst.UniqTx) ||
			!bytes.Equal(iRes.ExternalTokenID, issuingEVMAcceptedInst.ExternalTokenID) ||
			shardID != issuingEVMAcceptedInst.ShardID {
			continue
		}

		addressStr := issuingEVMAcceptedInst.ReceiverAddrStr
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != issuingEVMAcceptedInst.IncTokenID.String() {
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, issuingEVMAcceptedInst.IssuingAmount); !ok {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingETHRequest tx found for IssuingEVMResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *IssuingEVMResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
