package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleGetBridgeAggState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Beacon height is invalid"))
	}
	result, err := httpServer.blockService.GetBridgeAggState(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBridgeAggStateError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetBridgeAggConvertStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	sDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()

	data, err := statedb.GetBridgeAggStatus(
		sDB,
		statedb.BridgeAggConvertStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleBridgeAggConvert(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.createBridgeAggConvertTransaction(params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, false, closeChan)
}

func (httpServer *HttpServer) createBridgeAggConvertTransaction(params interface{}) (
	*jsonresult.CreateTransactionResult, *rpcservice.RPCError,
) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param to be %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}

	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		metadataBridge.ConvertTokenToUnifiedTokenRequest
	}{}
	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	recv := privacy.OTAReceiver{}
	err = recv.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	md := metadataBridge.NewConvertTokenToUnifiedTokenRequestWithValue(
		mdReader.TokenID, mdReader.UnifiedTokenID, mdReader.Amount, recv,
	)
	paramSelect.SetTokenID(mdReader.TokenID)
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		{
			PaymentAddress: burnAddr,
			Amount:         md.Amount,
		},
	}
	paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
	paramSelect.SetTokenReceivers(burnPayments)

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handleBridgeAggShield(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createBridgeAggShieldTransaction(params)
	if err != nil {
		return nil, err
	}

	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return res, nil
}

func (httpServer *HttpServer) createBridgeAggShieldTransaction(params interface{}) (
	*jsonresult.CreateTransactionResult, *rpcservice.RPCError,
) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param to be %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}

	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		Data []struct {
			BlockHash  string      `json:"BlockHash"`
			TxIndex    *uint       `json:"TxIndex"`
			Proof      []string    `json:"Proof"`
			NetworkID  uint        `json:"NetworkID"`
			IncTokenID common.Hash `json:"IncTokenID"`
		} `json:"Data"`
		UnifiedTokenID common.Hash `json:"UnifiedTokenID"`
	}{}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// parse params & metadata
	_, err = httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	data := []metadataBridge.ShieldRequestData{}
	for _, v := range mdReader.Data {
		temp := metadataBridge.ShieldRequestData{
			NetworkID:  v.NetworkID,
			IncTokenID: v.IncTokenID,
		}
		type EVMProof struct {
			BlockHash string   `json:"BlockHash"`
			TxIndex   uint     `json:"TxIndex"`
			Proof     []string `json:"Proof"`
		}
		proof := EVMProof{
			BlockHash: v.BlockHash,
			TxIndex:   *v.TxIndex,
			Proof:     v.Proof,
		}
		proofData, err := json.Marshal(proof)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		temp.Proof = proofData
		data = append(data, temp)
	}

	md := metadataBridge.NewShieldRequestWithValue(data, mdReader.UnifiedTokenID)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, md)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleBridgeAggUnshield(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.createBridgeAggUnshieldTransaction(params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, false, closeChan)
}

func (httpServer *HttpServer) createBridgeAggUnshieldTransaction(params interface{}) (
	*jsonresult.CreateTransactionResult, *rpcservice.RPCError,
) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param to be %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		metadataBridge.UnshieldRequest
		TokenReceivers map[string]uint64 `json:"TokenReceivers,omitempty"`
	}{}
	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	recv := privacy.OTAReceiver{}
	err = recv.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	md := metadataBridge.NewUnshieldRequestWithValue(mdReader.UnifiedTokenID, mdReader.Data, recv, mdReader.IsDepositToSC)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	paramSelect.SetTokenID(mdReader.UnifiedTokenID)
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress
	burningAmount := uint64(0)
	for _, data := range md.Data {
		burningAmount += data.BurningAmount
	}

	burnPayments := []*privacy.PaymentInfo{
		{
			PaymentAddress: burnAddr,
			Amount:         burningAmount,
		},
	}

	var prvReceivers []*privacy.PaymentInfo
	if arrayParams[1] != nil {
		temp := make(map[string]uint64)
		raw, err := json.Marshal(arrayParams[1])
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		err = json.Unmarshal(raw, &temp)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		if len(temp) != 0 {
			for k, v := range temp {
				key, err := wallet.Base58CheckDeserialize(k)
				if err != nil {
					return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
				}
				prvReceiver := &privacy.PaymentInfo{
					PaymentAddress: key.KeySet.PaymentAddress,
					Amount:         v,
				}
				prvReceivers = append(prvReceivers, prvReceiver)
			}
		}
	}

	if len(mdReader.TokenReceivers) != 0 {
		for k, v := range mdReader.TokenReceivers {
			key, err := wallet.Base58CheckDeserialize(k)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
			}
			temp := &privacy.PaymentInfo{
				PaymentAddress: key.KeySet.PaymentAddress,
				Amount:         v,
			}
			burnPayments = append(burnPayments, temp)
		}
	}

	if len(prvReceivers) != 0 {
		// amount by token
		paramSelect.Token.PaymentInfos = prvReceivers
		// amount by PRV
		paramSelect.SetTokenReceivers(burnPayments)
	} else {
		paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
		paramSelect.SetTokenReceivers(burnPayments)
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}
	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetBridgeAggShieldStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	sDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()

	res := bridgeagg.ShieldStatus{}
	prefixValues := [][]byte{
		{},
		{common.BoolToByte(false)},
		{common.BoolToByte(true)},
	}
	for _, prefixValue := range prefixValues {
		suffix := append(txID.Bytes(), prefixValue...)
		data, err := statedb.GetBridgeAggStatus(
			sDB,
			statedb.BridgeAggShieldStatusPrefix(),
			suffix,
		)
		if err != nil {
			continue
		}
		status := bridgeagg.ShieldStatus{}
		err = json.Unmarshal(data, &status)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		res.Status = status.Status
		if status.Status == common.RejectedStatusByte {
			res.Data = nil
			res.ErrorCode = status.ErrorCode
		} else {
			if len(res.Data) == 0 {
				res.Data = make([]bridgeagg.ShieldStatusData, len(status.Data))
			}
			for i, v := range status.Data {
				res.Data[i].Amount += v.Amount
				res.Data[i].Reward += v.Reward
			}

		}
	}
	if len(res.Data) == 0 && res.Status == 0 && res.ErrorCode == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Not found status"))
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetBridgeAggUnshieldStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	sDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()

	data, err := statedb.GetBridgeAggStatus(
		sDB,
		statedb.BridgeAggUnshieldStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleEstimateFeeByBurntAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Incorrect parameter length"))
	}
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		UnifiedTokenID common.Hash `json:"UnifiedTokenID"`
		TokenID        common.Hash `json:"TokenID"`
		BurntAmount    uint64      `json:"BurntAmount"`
	}{}
	rawMd, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(rawMd, &mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateFeeByBurntAmountError, err)
	}
	result, err := httpServer.blockService.BridgeAggEstimateFeeByBurntAmount(mdReader.UnifiedTokenID, mdReader.TokenID, mdReader.BurntAmount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateFeeByBurntAmountError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleEstimateFeeByExpectedAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Incorrect parameter length"))
	}
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		UnifiedTokenID common.Hash `json:"UnifiedTokenID"`
		TokenID        common.Hash `json:"TokenID"`
		ExpectedAmount uint64      `json:"ExpectedAmount"`
	}{}
	rawMd, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(rawMd, &mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateFeeByExpectedAmountError, err)
	}
	result, err := httpServer.blockService.BridgeAggEstimateFeeByExpectedAmount(mdReader.UnifiedTokenID, mdReader.TokenID, mdReader.ExpectedAmount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateFeeByExpectedAmountError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleBridgeAggEstimateReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Incorrect parameter length"))
	}
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		UnifiedTokenID common.Hash `json:"UnifiedTokenID"`
		TokenID        common.Hash `json:"TokenID"`
		Amount         uint64      `json:"Amount"`
	}{}
	rawMd, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(rawMd, &mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateRewardError, err)
	}
	result, err := httpServer.blockService.BridgeAggEstimateReward(mdReader.UnifiedTokenID, mdReader.TokenID, mdReader.Amount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.BridgeAggEstimateRewardError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleBridgeAggGetBurntProof(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Incorrect parameter length"))
	}
	// metadata object format to read from RPC parameters
	Reader := &struct {
		TxReqID   common.Hash `json:"TxReqID"`
		DataIndex *int        `json:"DataIndex"`
	}{}
	rawData, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(rawData, &Reader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	txReqID := Reader.TxReqID
	if Reader.DataIndex != nil {
		txReqID = common.HashH(append(txReqID.Bytes(), common.IntToBytes(*Reader.DataIndex)...))
	}
	height, onBeacon, err := httpServer.blockService.GetBurningConfirm(txReqID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return retrieveBurnProof(0, onBeacon, height, &txReqID, httpServer, false)
}
