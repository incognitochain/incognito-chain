package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalRequestUnlockCollateral - portal custodian requests unlock collateral (after returning pubToken to user)
// metadata - custodian requests unlock collateral - create normal tx with this metadata
type PortalWithdrawRewardResponse struct {
	MetadataBase
	CustodianAddressStr string
	TokenID             common.Hash
	RewardAmount        uint64
	TxReqID             common.Hash
	SharedRandom       []byte
}

func NewPortalWithdrawRewardResponse(
	reqTxID common.Hash,
	custodianAddressStr string,
	tokenID common.Hash,
	rewardAmount uint64,
	metaType int,
) *PortalWithdrawRewardResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalWithdrawRewardResponse{
		MetadataBase:        metadataBase,
		CustodianAddressStr: custodianAddressStr,
		TokenID:             tokenID,
		RewardAmount:        rewardAmount,
		TxReqID:             reqTxID,
	}
}

func (iRes PortalWithdrawRewardResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalWithdrawRewardResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalWithdrawRewardResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalWithdrawRewardResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalRequestWithdrawRewardResponseMeta
}

func (iRes PortalWithdrawRewardResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.TxReqID.String()
	record += iRes.CustodianAddressStr
	record += iRes.TokenID.String()
	record += strconv.FormatUint(iRes.RewardAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalWithdrawRewardResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalWithdrawRewardResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1

	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalWithdrawReward instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalRequestWithdrawRewardMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != common.PortalReqWithdrawRewardAcceptedChainStatus {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var custodianAddrStrFromInst string
		var rewardAmountFromInst uint64
		var tokenIDFromInst common.Hash

		contentBytes := []byte(inst[3])
		var reqWithdrawRewardContent PortalRequestWithdrawRewardContent
		err := json.Unmarshal(contentBytes, &reqWithdrawRewardContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal request withdraw reward content: ", err)
			continue
		}
		shardIDFromInst = reqWithdrawRewardContent.ShardID
		txReqIDFromInst = reqWithdrawRewardContent.TxReqID
		custodianAddrStrFromInst = reqWithdrawRewardContent.CustodianAddressStr
		rewardAmountFromInst = reqWithdrawRewardContent.RewardAmount
		tokenIDFromInst = reqWithdrawRewardContent.TokenID

		if !bytes.Equal(iRes.TxReqID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(custodianAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing custodian address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted {
			Logger.log.Info("WARNING - VALIDATION: Error occured while validate tx mint.  ", err)
			continue
		}
		if coinID.String() != tokenIDFromInst.String() {
			Logger.log.Info("WARNING - VALIDATION: Receive Token ID in tx mint maybe not correct. Must be PRV")
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, rewardAmountFromInst); !ok {
			Logger.log.Info("WARNING - VALIDATION: Error occured while check receiver and amount. CheckCoinValid return false ")
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalWithdrawReward instruction found for PortalWithdrawRewardResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalWithdrawRewardResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}