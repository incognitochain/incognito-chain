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

type PortalCustodianDepositResponse struct {
	MetadataBase
	DepositStatus    string
	ReqTxID          common.Hash
	CustodianAddrStr string
	SharedRandom       []byte
}

func NewPortalCustodianDepositResponse(
	depositStatus string,
	reqTxID common.Hash,
	custodianAddressStr string,
	metaType int,
) *PortalCustodianDepositResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalCustodianDepositResponse{
		DepositStatus:    depositStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		CustodianAddrStr: custodianAddressStr,
	}
}

func (iRes PortalCustodianDepositResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalCustodianDepositResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalCustodianDepositResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalCustodianDepositResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalCustodianDepositResponseMeta
}

func (iRes PortalCustodianDepositResponse) Hash() *common.Hash {
	record := iRes.DepositStatus
	record += iRes.ReqTxID.String()
	record += iRes.MetadataBase.Hash().String()
	if iRes.SharedRandom != nil && len(iRes.SharedRandom) > 0 {
		record += string(iRes.SharedRandom)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalCustodianDepositResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

//todo:
func (iRes PortalCustodianDepositResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1

	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalCustodianDeposit response instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalCustodianDepositMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.DepositStatus ||
			(instDepositStatus != common.PortalCustodianDepositRefundChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var custodianAddrStrFromInst string
		var depositedAmountFromInst uint64

		contentBytes := []byte(inst[3])
		var custodianDepositContent PortalCustodianDepositContent
		err := json.Unmarshal(contentBytes, &custodianDepositContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal custodian deposit content: ", err)
			continue
		}
		shardIDFromInst = custodianDepositContent.ShardID
		txReqIDFromInst = custodianDepositContent.TxReqID
		custodianAddrStrFromInst = custodianDepositContent.IncogAddressStr
		depositedAmountFromInst = custodianDepositContent.DepositedAmount

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(custodianAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing custodian address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != common.PRVCoinID.String() {
			Logger.log.Info("WARNING - VALIDATION: an error occured while validate tx mint: ", err)
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, depositedAmountFromInst); !ok {
			Logger.log.Info("WARNING - VALIDATION: an error occured while check tx mint. CheckCoinValid return false ")
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalCustodianDeposit instruction found for PortalCustodianDepositResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalCustodianDepositResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}