package bridgehub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type IssuingBTCResponse struct {
	metadataCommon.MetadataBase
	RequestedTxID common.Hash `json:"RequestedTxID"`
	UniqTx        []byte      `json:"UniqBTCTx"`
	SharedRandom  []byte      `json:"SharedRandom,omitempty"`
}

type IssuingBTCResAction struct {
	Meta       *IssuingBTCResponse `json:"meta"`
	IncTokenID *common.Hash        `json:"incTokenID"`
}

func NewIssuingBTCResponse(
	requestedTxID common.Hash,
	uniqTx []byte,
	metaType int,
) *IssuingBTCResponse {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	return &IssuingBTCResponse{
		RequestedTxID: requestedTxID,
		UniqTx:        uniqTx,
		MetadataBase:  metadataBase,
	}
}

func (iRes IssuingBTCResponse) CheckTransactionFee(tr metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingBTCResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingBTCResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingBTCResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes IssuingBTCResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(iRes)
	hash := common.HashH(rawBytes)
	return &hash
}

func (iRes *IssuingBTCResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iRes)
}

func (iRes IssuingBTCResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	idx := -1
	prefix := "BTC hub"
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not IssuingEVMRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.ShieldingBTCRequestMeta)) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var ShieldingBTCAcceptedInst ShieldingBTCAcceptedInst
		err = json.Unmarshal(contentBytes, &ShieldingBTCAcceptedInst)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], ShieldingBTCAcceptedInst.TxReqID[:]) ||
			!bytes.Equal(iRes.UniqTx, ShieldingBTCAcceptedInst.UniqTx) ||
			shardID != ShieldingBTCAcceptedInst.ShardID {
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != ShieldingBTCAcceptedInst.IncTokenID.String() {
			continue
		}

		otaReceiver := ShieldingBTCAcceptedInst.Receiver
		if err != nil || !otaReceiver.IsValid() {
			metadataCommon.Logger.Log.Errorf("%v parse OTAReceiver error: %v\n", prefix, err)
			continue
		}
		if mintCoin.GetValue() != ShieldingBTCAcceptedInst.IssuingAmount {
			metadataCommon.Logger.Log.Errorf("%v expected issuingAmount %v, got %v\n",
				prefix, ShieldingBTCAcceptedInst.IssuingAmount, mintCoin.GetValue())
			continue
		}
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), mintCoin.GetPublicKey().ToBytesS()) {
			metadataCommon.Logger.Log.Errorf("%v expected pubKey %v, got %v\n",
				prefix, otaReceiver.PublicKey.ToBytesS(), mintCoin.GetPublicKey().ToBytesS())
			continue
		}
		txRandom := mintCoin.GetTxRandom()
		if !bytes.Equal(txRandom.Bytes(), otaReceiver.TxRandom.Bytes()) {
			metadataCommon.Logger.Log.Errorf("%v expected txRandom %v, got %v\n",
				prefix, otaReceiver.TxRandom.Bytes(), txRandom.Bytes())
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingETHRequest tx found for IssuingBTCResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *IssuingBTCResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
