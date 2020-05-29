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

type PortalLiquidationCustodianDepositV2Response struct {
	MetadataBase
	DepositStatus    string
	ReqTxID          common.Hash
	CustodianAddrStr string
	DepositedAmount  uint64
}

func NewPortalLiquidationCustodianDepositV2Response(
	depositStatus string,
	reqTxID common.Hash,
	custodianAddressStr string,
	depositedAmount uint64,
	metaType int,
) *PortalLiquidationCustodianDepositV2Response {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	return &PortalLiquidationCustodianDepositV2Response{
		DepositStatus:    depositStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		CustodianAddrStr: custodianAddressStr,
		DepositedAmount:  depositedAmount,
	}
}

func (iRes PortalLiquidationCustodianDepositV2Response) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalLiquidationCustodianDepositV2Response) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalLiquidationCustodianDepositV2Response) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalLiquidationCustodianDepositV2Response) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalLiquidationCustodianDepositV2ResponseMeta
}

func (iRes PortalLiquidationCustodianDepositV2Response) Hash() *common.Hash {
	record := iRes.DepositStatus
	record += strconv.FormatUint(iRes.DepositedAmount, 10)
	record += iRes.ReqTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalLiquidationCustodianDepositV2Response) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalLiquidationCustodianDepositV2Response) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PortalCustodianDeposit response instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalLiquidationCustodianDepositV2Meta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.DepositStatus ||
			(instDepositStatus != common.PortalLiquidationCustodianDepositRejectedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var custodianAddrStrFromInst string
		var depositedAmountFromInst uint64

		contentBytes := []byte(inst[3])
		var custodianDepositContent PortalLiquidationCustodianDepositContentV2
		err := json.Unmarshal(contentBytes, &custodianDepositContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal liquidation custodian deposit content: ", err)
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
			Logger.log.Info("WARNING - VALIDATION: an error occurred while deserializing custodian address string: ", err)
			continue
		}

		// collateral must be PRV
		PRVIDStr := common.PRVCoinID.String()
		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			depositedAmountFromInst != paidAmount ||
			PRVIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalLiquidationCustodianDepositV2 instruction found for PortalLiquidationCustodianDepositV2Response tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
