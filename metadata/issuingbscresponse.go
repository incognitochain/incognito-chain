package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type IssuingBSCResponse struct {
	MetadataBase
	RequestedTxID   common.Hash
	UniqBSCTx       []byte
	ExternalTokenID []byte
	SharedRandom    []byte `json:"SharedRandom,omitempty"`
}

type IssuingBSCResAction struct {
	Meta       *IssuingBSCResponse `json:"meta"`
	IncTokenID *common.Hash        `json:"incTokenID"`
}

func NewIssuingBSCResponse(
	requestedTxID common.Hash,
	uniqBSCTx []byte,
	externalTokenID []byte,
	metaType int,
) *IssuingBSCResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingBSCResponse{
		RequestedTxID:   requestedTxID,
		UniqBSCTx:       uniqBSCTx,
		ExternalTokenID: externalTokenID,
		MetadataBase:    metadataBase,
	}
}

func (iRes IssuingBSCResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingBSCResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingBSCResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingBSCResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes IssuingBSCResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += string(iRes.UniqBSCTx)
	record += string(iRes.ExternalTokenID)
	record += iRes.MetadataBase.Hash().String()
	if iRes.SharedRandom != nil && len(iRes.SharedRandom) > 0 {
		record += string(iRes.SharedRandom)
	}

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingBSCResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes IssuingBSCResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not IssuingBSCRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(IssuingBSCRequestMeta) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING BSC - VALIDATION: an error occurred while parsing instruction content: ", err)
			continue
		}
		var issuingBSCAcceptedInst IssuingBSCAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingBSCAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING BSC - VALIDATION: an error occurred while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], issuingBSCAcceptedInst.TxReqID[:]) ||
			!bytes.Equal(iRes.UniqBSCTx, issuingBSCAcceptedInst.UniqBSCTx) ||
			!bytes.Equal(iRes.ExternalTokenID, issuingBSCAcceptedInst.ExternalTokenID) ||
			shardID != issuingBSCAcceptedInst.ShardID {
			continue
		}

		addressStr := issuingBSCAcceptedInst.ReceiverAddrStr
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			Logger.log.Info("WARNING BSC- VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != issuingBSCAcceptedInst.IncTokenID.String() {
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, issuingBSCAcceptedInst.IssuingAmount); !ok {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingBSCRequest tx found for IssuingBSCResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *IssuingBSCResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}