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

type IssuingSOLResponse struct {
	MetadataBase
	RequestedTxID   common.Hash `json:"RequestedTxID"`
	UniqTx          []byte      `json:"UniqETHTx"`
	ExternalTokenID []byte      `json:"ExternalTokenID"`
	SharedRandom    []byte      `json:"SharedRandom,omitempty"`
}

type IssuingSOLResAction struct {
	Meta       *IssuingSOLResponse `json:"Meta"`
	IncTokenID *common.Hash        `json:"IncTokenID"`
}

func NewIssuingSOLResponse(
	requestedTxID common.Hash,
	uniqTx []byte,
	externalTokenID []byte,
	metaType int,
) *IssuingSOLResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingSOLResponse{
		RequestedTxID:   requestedTxID,
		UniqTx:          uniqTx,
		ExternalTokenID: externalTokenID,
		MetadataBase:    metadataBase,
	}
}

func (iRes IssuingSOLResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingSOLResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingSOLResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingSOLResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes *IssuingSOLResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&iRes)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (iRes *IssuingSOLResponse) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		MetadataBase
		RequestedTxID   common.Hash `json:"RequestedTxID"`
		UniqTx          []byte      `json:"UniqTx"`
		ExternalTokenID []byte      `json:"ExternalTokenID"`
		SharedRandom    []byte      `json:"SharedRandom,omitempty"`
	}{
		RequestedTxID:   iRes.RequestedTxID,
		UniqTx:          iRes.UniqTx,
		ExternalTokenID: iRes.ExternalTokenID,
		SharedRandom:    iRes.SharedRandom,
		MetadataBase:    iRes.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (iRes *IssuingSOLResponse) UnmarshalJSON(data []byte) error {
	temp := struct {
		MetadataBase
		RequestedTxID   common.Hash `json:"RequestedTxID"`
		UniqTx          []byte      `json:"UniqTx"`
		ExternalTokenID []byte      `json:"ExternalTokenID"`
		SharedRandom    []byte      `json:"SharedRandom,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	iRes.MetadataBase = temp.MetadataBase
	iRes.RequestedTxID = temp.RequestedTxID
	iRes.UniqTx = temp.UniqTx
	iRes.ExternalTokenID = temp.ExternalTokenID
	iRes.SharedRandom = temp.SharedRandom
	return nil
}

func (iRes *IssuingSOLResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes IssuingSOLResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not IssuingSOLRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			(instMetaType != strconv.Itoa(IssuingSOLRequestMeta)) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var issuingSOLAcceptedInst IssuingSOLAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingSOLAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], issuingSOLAcceptedInst.TxReqID[:]) ||
			!bytes.Equal(iRes.UniqTx, issuingSOLAcceptedInst.UniqExternalTx) ||
			!bytes.Equal(iRes.ExternalTokenID, issuingSOLAcceptedInst.ExternalTokenID) ||
			shardID != issuingSOLAcceptedInst.ShardID {
			continue
		}

		addressStr := issuingSOLAcceptedInst.ReceivingIncAddrStr
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != issuingSOLAcceptedInst.IncTokenID.String() {
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, issuingSOLAcceptedInst.IssuingAmount); !ok {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingSOLRequest tx found for IssuingSOLResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *IssuingSOLResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
