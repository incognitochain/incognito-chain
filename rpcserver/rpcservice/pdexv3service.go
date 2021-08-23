package rpcservice

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/transaction"
)

func (blockService BlockService) GetPdexv3ParamsModifyingRequestStatus(reqTxID string) (*metadataPdexv3.ParamsModifyingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3ParamsModifyingStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.ParamsModifyingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// paramSelector helps to wrap transaction creation steps (v2 only)
type paramSelector struct {
	TokenID        common.Hash
	PRV            *bean.CreateRawTxParam
	Token          *bean.CreateRawPrivacyTokenTxParam
	TokenReceivers []*privacy.PaymentInfo
	Metadata       metadataCommon.Metadata
}

// SetTokenID, SetTokenReceivers, SetMetadata add necessary data for token tx creation that are missing from the params struct
func (sel *paramSelector) SetTokenID(id common.Hash)                  { sel.TokenID = id }
func (sel *paramSelector) SetTokenReceivers(r []*privacy.PaymentInfo) { sel.TokenReceivers = r }
func (sel *paramSelector) SetMetadata(md metadataCommon.Metadata)     { sel.Metadata = md }

// PdexTxService extends TxService with wrappers to build TX with cleaner syntax
type PdexTxService struct {
	*TxService
}

func (svc PdexTxService) ReadParamsFrom(raw interface{}, metadataReader interface{}) (*paramSelector, error) {
	var err error
	// token id defaults to PRV
	sel := paramSelector{TokenID: common.PRVCoinID}
	sel.PRV, err = bean.NewCreateRawTxParam(raw)
	if err != nil {
		return nil, err
	}
	sel.Token, err = bean.NewCreateRawPrivacyTokenTxParam(raw)
	if err != nil {
		return nil, err
	}

	arrayParams := common.InterfaceSlice(raw)
	if len(arrayParams) >= 5 {
		rawMd, err := json.Marshal(arrayParams[4])
		if err != nil {
			return nil, fmt.Errorf("Cannot parse metadata - %v", err)
		}
		err = json.Unmarshal(rawMd, metadataReader)
	}

	return &sel, err
}

func (svc PdexTxService) BuildTransaction(
	sel *paramSelector, md metadataCommon.Metadata,
) (metadataCommon.Transaction, *RPCError) {
	if sel.TokenID == common.PRVCoinID {
		return svc.BuildRawTransaction(sel.PRV, md)
	} else {
		return buildTokenTransaction(svc, sel)
	}
}

func (svc PdexTxService) GenerateOTAReceivers(
	tokens []common.Hash, addr privacy.PaymentAddress,
) (map[common.Hash]privacy.OTAReceiver, error) {
	result := make(map[common.Hash]privacy.OTAReceiver)
	var err error
	for _, tokenID := range tokens {
		temp := privacy.OTAReceiver{}
		err = temp.FromAddress(addr)
		if err != nil {
			return nil, err
		}
		result[tokenID] = temp
	}
	return result, nil
}

func buildTokenTransaction(svc PdexTxService, sel *paramSelector) (metadataCommon.Transaction, *RPCError) {
	params := sel.Token

	// choose token inputs
	outputTokens, err := svc.BlockChain.TryGetAllOutputCoinsByKeyset(
		params.SenderKeySet, params.ShardIDSender, &sel.TokenID, true,
	)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	outputTokens, err = svc.filterMemPoolOutcoinsToSpent(outputTokens)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}

	var totalTokenTransferred uint64
	for _, payment := range sel.TokenReceivers {
		totalTokenTransferred += payment.Amount
	}
	candidateOutputTokens, _, _, err := svc.chooseBestOutCoinsToSpent(
		outputTokens, totalTokenTransferred,
	)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}

	tokenParams := &transaction.TokenParam{
		TokenTxType: int(transaction.CustomTokenTransfer),
		// amount will default to sum of payments
		Amount:     0,
		Receiver:   sel.TokenReceivers,
		Fee:        0,
		TokenInput: candidateOutputTokens,
		PropertyID: sel.TokenID.String(),
	}

	// choose PRV inputs
	inputCoins, realFeePRV, err1 := svc.chooseOutsCoinByKeyset(params.PaymentInfos,
		params.EstimateFeeCoinPerKb, 0, params.SenderKeySet,
		params.ShardIDSender, params.HasPrivacyCoin, nil, tokenParams)
	if err1 != nil {
		return nil, err1
	}
	if len(params.PaymentInfos) == 0 && realFeePRV == 0 {
		params.HasPrivacyCoin = false
	}

	// tx creation params
	txTokenParams := transaction.NewTxTokenParams(&params.SenderKeySet.PrivateKey,
		params.PaymentInfos,
		inputCoins,
		realFeePRV,
		tokenParams,
		svc.BlockChain.GetBestStateShard(params.ShardIDSender).GetCopiedTransactionStateDB(),
		sel.Metadata,
		params.HasPrivacyCoin,
		params.HasPrivacyToken,
		params.ShardIDSender, params.Info,
		svc.BlockChain.BeaconChain.GetFinalViewState().GetBeaconFeatureStateDB(),
	)

	tx := &transaction.TxTokenVersion2{}
	errTx := tx.Init(txTokenParams)
	if errTx != nil {
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	return tx, nil
}

func (blockService BlockService) GetPdexv3WithdrawalLPFeeStatus(reqTxID string) (*metadataPdexv3.WithdrawalLPFeeStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalLPFeeStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.WithdrawalLPFeeStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPdexv3WithdrawalProtocolFeeStatus(reqTxID string) (*metadataPdexv3.WithdrawalProtocolFeeStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalProtocolFeeStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.WithdrawalProtocolFeeStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
