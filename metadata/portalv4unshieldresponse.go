package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type PortalUnshieldResponse struct {
	MetadataBase
	RequestStatus  string
	ReqTxID        common.Hash
	OTAPubKeyStr   string
	TxRandomStr    string
	UnshieldAmount uint64
	IncTokenID     string
}

func NewPortalV4UnshieldResponse(
	requestStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	txRandomStr string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalUnshieldResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalUnshieldResponse{
		RequestStatus:  requestStatus,
		ReqTxID:        reqTxID,
		MetadataBase:   metadataBase,
		OTAPubKeyStr:   requesterAddressStr,
		TxRandomStr:    txRandomStr,
		UnshieldAmount: amount,
		IncTokenID:     tokenID,
	}
}

func (iRes PortalUnshieldResponse) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalUnshieldResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalUnshieldResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalUnshieldResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == metadataCommon.PortalV4UnshieldingResponseMeta
}

func (iRes PortalUnshieldResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.OTAPubKeyStr
	record += iRes.TxRandomStr
	record += strconv.FormatUint(iRes.UnshieldAmount, 10)
	record += iRes.IncTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalUnshieldResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalUnshieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *MintData,
	shardID byte, tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalUnshieldResponse instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.PortalV4UnshieldingRequestMeta)) {
			continue
		}
		instReqStatus := inst[2]
		if iRes.RequestStatus != "refunded" ||
			(instReqStatus != portalcommonv4.PortalV4RequestRefundedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var receiverOTAPubKeyFromInst string
		var receiverTxRandomFromInst string
		var unshieldAmountFromInst uint64
		var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var unshieldReqContent PortalUnshieldRequestContent
		err := json.Unmarshal(contentBytes, &unshieldReqContent)
		if err != nil {
			Logger.log.Error("[VerifyUnshieldResponse] WARNING - VALIDATION: an error occured while parsing portal v4 unshield request content: ", err)
			continue
		}
		shardIDFromInst = unshieldReqContent.ShardID
		txReqIDFromInst = unshieldReqContent.TxReqID
		receiverOTAPubKeyFromInst = unshieldReqContent.OTAPubKeyStr
		receiverTxRandomFromInst = unshieldReqContent.TxRandomStr
		unshieldAmountFromInst = unshieldReqContent.UnshieldAmount
		tokenIDStrFromInst = unshieldReqContent.TokenID

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, assetID, err := tx.GetTxMintData()
		if err != nil {
			Logger.log.Error("[VerifyUnshieldResponse] ERROR - VALIDATION: an error occured while get tx mint data: ", err)
			continue
		}
		if !isMinted {
			Logger.log.Info("[VerifyUnshieldResponse] WARNING - VALIDATION: this is not Tx Mint: ")
			continue
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()

		publicKey, txRandom, err := coin.ParseOTAInfoFromString(receiverOTAPubKeyFromInst, receiverTxRandomFromInst)
		if err != nil {
			Logger.log.Errorf("[VerifyUnshieldResponse] Wrong request info's txRandom - Cannot set txRandom from bytes: %+v", err)
			continue
		}

		txR := mintCoin.(*coin.CoinV2).GetTxRandom()
		if !bytes.Equal(publicKey.ToBytesS(), pk[:]) ||
			unshieldAmountFromInst != paidAmount ||
			!bytes.Equal(txR[:], txRandom[:]) ||
			tokenIDStrFromInst != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("[VerifyUnshieldResponse] no PortalV4UnshieldRequest instruction found for PortalUnshieldResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
