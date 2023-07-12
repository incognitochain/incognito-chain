package pdexv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type WithdrawalProtocolFeeResponse struct {
	metadataCommon.MetadataBase
	Address      string      `json:"Address"`
	TokenID      common.Hash `json:"TokenID"`
	Amount       uint64      `json:"Amount"`
	ReqTxID      common.Hash `json:"ReqTxID"`
	SharedRandom []byte      `json:"SharedRandom"`
}

func NewPdexv3WithdrawalProtocolFeeResponse(
	metaType int,
	address string,
	tokenID common.Hash,
	amount uint64,
	reqTxID common.Hash,
) *WithdrawalProtocolFeeResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalProtocolFeeResponse{
		MetadataBase: *metadataBase,
		Address:      address,
		TokenID:      tokenID,
		Amount:       amount,
		ReqTxID:      reqTxID,
	}
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return withdrawalResponse.Type == metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(struct {
		Type    int         `json:"Type"`
		Address string      `json:"Address"`
		TokenID common.Hash `json:"TokenID"`
		Amount  uint64      `json:"Amount"`
		ReqTxID common.Hash `json:"ReqTxID"`
	}{
		Type:    metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta,
		Address: withdrawalResponse.Address,
		TokenID: withdrawalResponse.TokenID,
		Amount:  withdrawalResponse.Amount,
		ReqTxID: withdrawalResponse.ReqTxID,
	})

	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawalResponse *WithdrawalProtocolFeeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawalResponse)
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte, tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	// verify mining tx with the request tx
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not WithdrawalProtocolFeeRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta)) {
			continue
		}

		contentBytes := []byte(inst[3])
		var instContent WithdrawalProtocolFeeContent
		err := json.Unmarshal(contentBytes, &instContent)
		if err != nil {
			continue
		}

		shardIDFromInst := instContent.ShardID
		txReqIDFromInst := instContent.TxReqID

		if !bytes.Equal(withdrawalResponse.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, assetID, err := tx.GetTxMintData()
		if err != nil || !isMinted || assetID.String() != instContent.TokenID.String() {
			continue
		}

		keyWallet, err := wallet.Base58CheckDeserialize(instContent.Address)
		if err != nil {
			continue
		}
		paymentAddress := keyWallet.KeySet.PaymentAddress

		if ok := mintCoin.CheckCoinValid(paymentAddress, withdrawalResponse.SharedRandom, instContent.Amount); !ok {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("No WithdrawalProtocolFee instruction found for WithdrawalProtocolFeeResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (withdrawalResponse *WithdrawalProtocolFeeResponse) SetSharedRandom(r []byte) {
	withdrawalResponse.SharedRandom = r
}
