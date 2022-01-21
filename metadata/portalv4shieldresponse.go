package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalShieldingResponse struct {
	MetadataBase

	// RequestStatus is the status of the shielding request.
	RequestStatus string

	// ReqTxID is the hash of the shielding request transaction.
	ReqTxID common.Hash

	// Receiver is the same as in the request.
	// If Receiver is an Incognito payment address, SharedRandom must not be empty.
	// If Receiver is an OTAReceiver, SharedRandom is not required.
	Receiver string `json:"RequesterAddrStr"` // the json-tag is required for backward-compatibility.

	// MintingAmount is the shielding amount.
	MintingAmount uint64

	// IncTokenID is the Incognito ID of the shielding token.
	IncTokenID string

	// SharedRandom is combined with Receiver to make sure the minting amount is for the eligible party.
	SharedRandom []byte `json:"SharedRandom,omitempty"`
}

func NewPortalShieldingResponse(
	depositStatus string,
	reqTxID common.Hash,
	receiver string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalShieldingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalShieldingResponse{
		RequestStatus: depositStatus,
		ReqTxID:       reqTxID,
		MetadataBase:  metadataBase,
		Receiver:      receiver,
		MintingAmount: amount,
		IncTokenID:    tokenID,
	}
}

func (iRes PortalShieldingResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
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
	record += iRes.Receiver
	record += strconv.FormatUint(iRes.MintingAmount, 10)
	record += iRes.IncTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalShieldingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

// VerifyMinerCreatedTxBeforeGettingInBlock validates if a shieldResponse is a reply for an instruction from the beacon.
// The response is valid for a specific instruction if
//	1. the instruction has a valid metadata type with accepted status
//	2. the requested txIDs match
//	3. the tokenID is valid
//	4. receiver and amount are valid
//		4.1 If the receiver is a payment address, checkCoinValid must pass
//		4.2 If the receiver is an OTAReceiver, txRandom must match
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
	prefix := fmt.Sprintf("[PortalV4ShieldResp]")

	for i, inst := range mintData.Insts {
		// Step 1.
		if len(inst) < 4 { // this is not PortalV4ShieldingRequest response instruction
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
			Logger.log.Errorf("%v an error occurred while parsing portal request shielding content: %v\n",
				prefix, err)
			continue
		}
		instShieldingStatus := inst[2]
		if instShieldingStatus != iRes.RequestStatus ||
			(instShieldingStatus != portalcommonv4.PortalV4RequestAcceptedChainStatus) {
			Logger.log.Errorf("%v invalid shieldingStatus: %v (%v)\n",
				prefix, instShieldingStatus, iRes.RequestStatus)
			continue
		}

		// Step 2.
		if !bytes.Equal(iRes.ReqTxID[:], shieldingReqContent.TxReqID[:]) ||
			shardID != shieldingReqContent.ShardID {
			Logger.log.Errorf("%v invalid requestTx: expected %v, got %v\n",
				prefix, shieldingReqContent.TxReqID.String(), iRes.ReqTxID.String())
			continue
		}

		// Step 3.
		tokenIDPointer, _ := new(common.Hash).NewHashFromStr(shieldingReqContent.TokenID)
		tokenID := *tokenIDPointer
		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted {
			Logger.log.Errorf("%v GetTxMintData FAILED: isMinted %v, error: %v\n",
				prefix, isMinted, err)
			continue
		}
		if coinID.String() != tokenID.String() {
			Logger.log.Errorf("%v expected tokenID %v, got %v\n", prefix, tokenID.String(), coinID.String())
			continue
		}

		// Step 4.
		keyWallet, err := wallet.Base58CheckDeserialize(shieldingReqContent.Receiver)
		if err == nil { // 4.1
			paymentAddress := keyWallet.KeySet.PaymentAddress
			if ok := mintCoin.CheckCoinValid(paymentAddress, iRes.SharedRandom, shieldingReqContent.MintingAmount); !ok {
				Logger.log.Errorf("%v CheckCoinValid FAILED\n", prefix, err)
				continue
			}
		} else { // 4.2
			otaReceiver := new(privacy.OTAReceiver)
			err = otaReceiver.FromString(shieldingReqContent.Receiver)
			if err != nil || !otaReceiver.IsValid() {
				Logger.log.Errorf("%v parse OTAReceiver error: %v\n", prefix, err)
				continue
			}
			if mintCoin.GetValue() != shieldingReqContent.MintingAmount {
				Logger.log.Errorf("%v expected mintingAmount %v, got %v\n",
					prefix, shieldingReqContent.MintingAmount, mintCoin.GetValue())
				continue
			}
			if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), mintCoin.GetPublicKey().ToBytesS()) {
				Logger.log.Errorf("%v expected pubKey %v, got %v\n",
					prefix, otaReceiver.PublicKey.ToBytesS(), mintCoin.GetPublicKey().ToBytesS())
				continue
			}
			txRandom := mintCoin.GetTxRandom()
			if !bytes.Equal(txRandom.Bytes(), otaReceiver.TxRandom.Bytes()) {
				Logger.log.Errorf("%v expected txRandom %v, got %v\n",
					prefix, otaReceiver.TxRandom.Bytes(), txRandom.Bytes())
				continue
			}

		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("%v no instruction found for PortalShieldingResponse tx %s",
			prefix, tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalShieldingResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
