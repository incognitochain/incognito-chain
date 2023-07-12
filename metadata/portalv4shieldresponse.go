package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalShieldingResponse struct {
	MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	RequesterAddrStr string
	MintingAmount    uint64
	IncTokenID       string
	SharedRandom     []byte
}

func NewPortalShieldingResponse(
	depositStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalShieldingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalShieldingResponse{
		RequestStatus:    depositStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		RequesterAddrStr: requesterAddressStr,
		MintingAmount:    amount,
		IncTokenID:       tokenID,
	}
}

func (iRes PortalShieldingResponse) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalShieldingResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalShieldingResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalShieldingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == metadataCommon.PortalV4ShieldingResponseMeta
}

func (iRes PortalShieldingResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.MintingAmount, 10)
	record += iRes.IncTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalShieldingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalShieldingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *MintData,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalShieldingRequest response instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(metadataCommon.PortalV4ShieldingRequestMeta) {
			continue
		}

		contentBytes := []byte(inst[3])
		var shieldingReqContent PortalShieldingRequestContent
		err := json.Unmarshal(contentBytes, &shieldingReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occurred while parsing portal request shielding content: ", err)
			continue
		}

		instShieldingStatus := inst[2]
		if instShieldingStatus != iRes.RequestStatus ||
			(instShieldingStatus != portalcommonv4.PortalV4RequestAcceptedChainStatus) {
			continue
		}

		if !bytes.Equal(iRes.ReqTxID[:], shieldingReqContent.TxReqID[:]) ||
			shardID != shieldingReqContent.ShardID {
			continue
		}

		tokenIDPointer, _ := new(common.Hash).NewHashFromStr(shieldingReqContent.TokenID)
		tokenID := *tokenIDPointer
		keyWallet, err := wallet.Base58CheckDeserialize(shieldingReqContent.IncogAddressStr)
		if err != nil {
			continue
		}
		paymentAddress := keyWallet.KeySet.PaymentAddress

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != tokenID.String() {
			continue
		}

		if ok := mintCoin.CheckCoinValid(paymentAddress, iRes.SharedRandom, shieldingReqContent.MintingAmount); !ok {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalShieldingRequest instruction found for PortalShieldingResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalShieldingResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
