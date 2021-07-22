package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleGetPDexV3State(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(httpServer.config.BlockChain.GetBeaconBestState(), uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDexV3StateError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDexV3StateError, err)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.PDexV3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDexV3StateError, fmt.Errorf("pDEX v3 is not available"))
	}
	pDexv3State, err := pdex.InitStateFromDB(beaconFeatureStateDB, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDexV3StateError, err)
	}

	beaconBlocks, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDexV3StateError, err)
	}
	beaconBlock := beaconBlocks[0]
	result := jsonresult.PDexV3State{
		BeaconTimeStamp: beaconBlock.Header.Timestamp,
		Params:          pDexv3State.Reader().Params(),
	}
	return result, nil
}

func (httpServer *HttpServer) handleAddLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, isPRV, err := httpServer.createRawTxAddLiquidityV3(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	if isPRV {
		res, err = httpServer.handleSendRawTransaction(newParam, closeChan)
	} else {
		res, err = httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	}
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) createRawTxAddLiquidityV3(
	params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	isPRV := false
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}

	if len(arrayParams) != 5 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}
	addLiquidityParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata type is invalid"))
	}
	addLiquidityRequest := PDEAddLiquidityV3Request{}
	// Convert map to json string
	addLiquidityParamData, err := json.Marshal(addLiquidityParam)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(addLiquidityParamData, &addLiquidityRequest)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private key %v: %v", privateKey, err))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("private key length not valid: %v", keyWallet.KeySet.PrivateKey))
	}
	senderAddress := keyWallet.KeySet.PaymentAddress

	tokenAmount, err := common.AssertAndConvertNumber(addLiquidityRequest.TokenAmount)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	amplifier, err := common.AssertAndConvertNumber(addLiquidityRequest.Amplifier)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	tokenFee, err := common.AssertAndConvertNumber(addLiquidityRequest.Fee)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tokenHash, err := common.Hash{}.NewHashFromStr(addLiquidityRequest.TokenID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	receiverAddress := privacy.OTAReceiver{}
	refundAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromAddress(senderAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = refundAddress.FromAddress(senderAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	receiverAddressStr, err := receiverAddress.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	refundAddressStr, err := refundAddress.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	metaData := metadataPdexV3.NewAddLiquidityWithValue(
		addLiquidityRequest.PoolPairID,
		addLiquidityRequest.PairHash,
		receiverAddressStr, refundAddressStr,
		tokenHash.String(),
		tokenAmount,
		uint(amplifier),
	)

	if addLiquidityRequest.TokenID == common.PRVIDStr {
		isPRV = true
	}

	var byteArrays []byte
	var txHashStr string
	if isPRV {
		// create new param to build raw tx from param interface
		rawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
		if errNewParam != nil {
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
		}
		tx, rpcErr := httpServer.txService.BuildRawTransaction(rawTxParam, metaData)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
		}
		byteArrays, err = json.Marshal(tx)
		if err != nil {
			Logger.log.Error(err)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		txHashStr = tx.Hash().String()
	} else {
		receiverAddresses, ok := arrayParams[1].(map[string]interface{})
		if !ok {
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
		}

		customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyTokenTransaction(
			params,
			metaData,
			receiverAddresses,
			addLiquidityRequest.TokenID,
			tokenAmount,
			tokenFee,
		)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
		}
		byteArrays, err = json.Marshal(customTokenTx)
		if err != nil {
			Logger.log.Error(err)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		txHashStr = customTokenTx.Hash().String()
	}

	res := &jsonresult.CreateTransactionResult{
		TxID:            txHashStr,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return res, isPRV, nil
}
