package rpcserver

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2/ring_selection"
	"github.com/incognitochain/incognito-chain/wallet"
)

// handleListUnspentOutputCoins - use private key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
// component:
// Parameter #1—the minimum number of confirmations an output must have
// Parameter #2—the maximum number of confirmations an output may have
// Parameter #3—the list priv-key which be used to view utxo which also includes the fromHeight of each key
// From height is used to efficiently fetch onetimeaddress outputCoins
func (httpServer *HttpServer) handleListUnspentOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	// get component
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 3 elements"))
	}

	var min, max uint64

	if paramsArray[0] != nil {
		minParam, ok := paramsArray[0].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min param is invalid"))
		}
		min = uint64(minParam)
	}

	if paramsArray[1] != nil {
		maxParam, ok := paramsArray[1].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("max param is invalid"))
		}
		max = uint64(maxParam)
	}
	_ = min
	// _ = max

	listKeyParams := common.InterfaceSlice(paramsArray[2])
	if listKeyParams == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("list key is invalid"))
	}

	tokenID := &common.Hash{}
	err1 := tokenID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err1)
	}
	if len(paramsArray) == 4 {
		tokenIDStr, ok := paramsArray[3].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
		}
		if tokenIDStr != "" {
			tokenIDHash, err2 := common.Hash{}.NewHashFromStr(tokenIDStr)
			if err2 != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
			}
			tokenID = tokenIDHash
		}
	}

	result, err := httpServer.outputCoinService.ListUnspentOutputCoinsByKey(listKeyParams, tokenID, max)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// handleListUnspentOutputCoinsFromCache - use private key to get all tx which contains cached output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
// component:
// Parameter #1—the minimum number of confirmations an output must have
// Parameter #2—the maximum number of confirmations an output may have
// Parameter #3—the list priv-key which be used to view utxo which also includes the fromHeight of each key
// From height is used to efficiently fetch onetimeaddress outputCoins
func (httpServer *HttpServer) handleListUnspentOutputCoinsFromCache(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	// get component
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 3 elements"))
	}

	var min, max uint64

	if paramsArray[0] != nil {
		minParam, ok := paramsArray[0].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min param is invalid"))
		}
		min = uint64(minParam)
	}

	if paramsArray[1] != nil {
		maxParam, ok := paramsArray[1].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("max param is invalid"))
		}
		max = uint64(maxParam)
	}
	_ = min
	// _ = max

	listKeyParams := common.InterfaceSlice(paramsArray[2])
	if listKeyParams == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("list key is invalid"))
	}

	tokenID := &common.Hash{}
	err1 := tokenID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err1)
	}
	if len(paramsArray) == 4 {
		tokenIDStr, ok := paramsArray[3].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
		}
		if tokenIDStr != "" {
			tokenIDHash, err2 := common.Hash{}.NewHashFromStr(tokenIDStr)
			if err2 != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
			}
			tokenID = tokenIDHash
		}
	}

	result, err := httpServer.outputCoinService.ListCachedUnspentOutputCoinsByKey(listKeyParams, tokenID, max)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// handleListOutputCoins - use readonly key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
// component:
// Parameter #1—the minimum number of confirmations an output must have
// Parameter #2—the maximum number of confirmations an output may have
// Parameter #3—the list paymentaddress-readonlykey which be used to view list outputcoin
// Parameter #4 - optional - token id - default prv coin
func (httpServer *HttpServer) handleListOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	// get component
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 3 elements"))
	}

	minTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min param is invalid"))
	}
	min := uint64(minTemp)

	maxTemp, ok := paramsArray[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("max param is invalid"))
	}
	max := uint64(maxTemp)

	_ = min
	// _ = max

	//#3: list key component
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	if listKeyParams == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("list key is invalid"))
	}

	//#4: optional token type - default prv coin
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(paramsArray) > 3 {
		var err1 error
		tokenIdParam, ok := paramsArray[3].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
		}

		tokenID, err1 = common.Hash{}.NewHashFromStr(tokenIdParam)
		if err1 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.ListTokenNotFoundError, err1)
		}
	}
	result, err1 := httpServer.outputCoinService.ListOutputCoinsByKey(listKeyParams, *tokenID, max)
	if err1 != nil {
		return nil, err1
	}
	return result, nil
}

func (httpServer *HttpServer) handleListOutputCoinsFromCache(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	// get component
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 3 elements"))
	}

	minTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min param is invalid"))
	}
	min := int(minTemp)

	maxTemp, ok := paramsArray[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("max param is invalid"))
	}
	max := int(maxTemp)

	_ = min
	_ = max

	//#3: list key component
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	if listKeyParams == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("list key is invalid"))
	}

	//#4: optional token type - default prv coin
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(paramsArray) > 3 {
		var err1 error
		tokenIdParam, ok := paramsArray[3].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token id param is invalid"))
		}

		tokenID, err1 = common.Hash{}.NewHashFromStr(tokenIdParam)
		if err1 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.ListTokenNotFoundError, err1)
		}
	}
	result, err1 := httpServer.outputCoinService.ListCachedOutputCoinsByKey(listKeyParams, *tokenID)
	if err1 != nil {
		return nil, err1
	}
	return result, nil
}

// handleRandomCommitments - from input of outputcoin, random to create data for create new tx
func (httpServer *HttpServer) handleRandomCommitments(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 element"))
	}

	// #1: ShardID
	shardID, ok := arrayParams[0].(float64)
	if !ok {
		//If no direct shardID provided, try a payment address
		paymentAddressStr, ok := arrayParams[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("shardID is invalid: expect a shardID or a payment address, have %v", arrayParams[0])))
		}

		tmpWallet, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("error when deserialized payment address %v: %v", paymentAddressStr, err)))
		}

		pk := tmpWallet.KeySet.PaymentAddress.Pk
		if len(pk) == 0 {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("payment address %v invalid: no public key found", paymentAddressStr)))
		}

		shardID = float64(common.GetShardIDFromLastByte(pk[len(pk)-1]))
	}

	// #2: available inputCoin from old outputcoin
	outputs, ok := arrayParams[1].([]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("outputs is invalid"))
	}
	if len(outputs) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("len of outputs must be greater than zero"))
	}

	//#3 - tokenID - default PRV
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("tokenID is invalid"))
		}
		tokenID, err = common.Hash{}.NewHashFromStr(tokenIDTemp)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.ListTokenNotFoundError, err)
		}
	}

	commitmentIndexs, myCommitmentIndexs, commitments, err2 := httpServer.txService.RandomCommitments(byte(shardID), outputs, tokenID)
	if err2 != nil {
		return nil, err2
	}

	result := jsonresult.NewRandomCommitmentResult(commitmentIndexs, myCommitmentIndexs, commitments)
	return result, nil
}

// handleRandomCommitmentsAndPublicKey - returns a list of random commitments, public keys and indices for creating txver2
func (httpServer *HttpServer) handleRandomCommitmentsAndPublicKeys(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Infof("new request: %v\n", params)
	if config.Config().IsMainNet {
		return httpServer.handleRandomDecoysSelection(params, closeChan)
	}

	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 element"))
	}

	// #1: ShardID
	shardID, ok := arrayParams[0].(float64)
	if !ok {
		//If no direct shardID provided, try a payment address
		paymentAddressStr, ok := arrayParams[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("shardID is invalid: expect a shardID or a payment address, have %v", arrayParams[0])))
		}

		tmpWallet, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("error when deserialized payment address %v: %v", paymentAddressStr, err)))
		}

		pk := tmpWallet.KeySet.PaymentAddress.Pk
		if len(pk) == 0 {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("payment address %v invalid: no public key found", paymentAddressStr)))
		}

		shardID = float64(common.GetShardIDFromLastByte(pk[len(pk)-1]))
	}

	// #2: Number of commitments
	numOutputs, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Number of commitments is invalid"))
	}

	//#3 - tokenID - default PRV
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("tokenID is invalid"))
		}
		tokenID, err = common.Hash{}.NewHashFromStr(tokenIDTemp)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.ListTokenNotFoundError, err)
		}
	}

	commitmentIndices, publicKeys, commitments, assetTags, err2 := httpServer.txService.RandomCommitmentsAndPublicKeys(byte(shardID), int(numOutputs), tokenID)
	if err2 != nil {
		return nil, err2
	}

	result := jsonresult.NewRandomCommitmentAndPublicKeyResult(commitmentIndices, publicKeys, commitments, assetTags)
	return result, nil
}

// handleRandomCommitmentsAndPublicKey - returns a list of random commitments, public keys and indices for creating txver2
func (httpServer *HttpServer) handleRandomDecoysSelection(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Infof("[handleRandomDecoysSelection] new request: %v\n", params)
	if !config.Config().IsMainNet {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidMethodPermissionError, fmt.Errorf("RPC not supported by the network configuration"))
	}

	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 element"))
	}

	// #1: ShardID
	shardID, ok := arrayParams[0].(float64)
	if !ok {
		//If no direct shardID provided, try a payment address
		paymentAddressStr, ok := arrayParams[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("shardID is invalid: expect a shardID or a payment address, have %v", arrayParams[0])))
		}

		tmpWallet, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("error when deserialized payment address %v: %v", paymentAddressStr, err)))
		}

		pk := tmpWallet.KeySet.PaymentAddress.Pk
		if len(pk) == 0 {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("payment address %v invalid: no public key found", paymentAddressStr)))
		}

		shardID = float64(common.GetShardIDFromLastByte(pk[len(pk)-1]))
	}

	// #2: Number of commitments
	numOutputs, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Number of commitments is invalid"))
	}

	//#3 - tokenID - default PRV
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("tokenID is invalid"))
		}
		tokenID, err = common.Hash{}.NewHashFromStr(tokenIDTemp)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.ListTokenNotFoundError, err)
		}
	}

	start := time.Now()
	commitmentIndices, publicKeys, commitments, assetTags, err := httpServer.randomDecoysFromGamma(int(numOutputs), byte(shardID), tokenID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}
	Logger.log.Infof("[RandomDecoysFromGamma] Shard: %v, TokenID: %v, #Coins: %v, timeElapsed: %v\n", shardID, tokenID.String(), numOutputs, time.Since(start).Seconds())

	result := jsonresult.NewRandomCommitmentAndPublicKeyResult(commitmentIndices, publicKeys, commitments, assetTags)
	return result, nil
}

func (httpServer *HttpServer) randomDecoysFromGamma(numOutputs int, shardID byte, tokenID *common.Hash) ([]uint64, [][]byte, [][]byte, [][]byte, error) {
	if int(shardID) >= common.MaxShardNumber {
		return nil, nil, nil, nil, fmt.Errorf("shardID %v is out of range, maxShardNumber is %v", shardID, common.MaxShardNumber)
	}
	shardState := httpServer.blockService.BlockChain.GetBestStateShard(shardID)
	db := shardState.GetCopiedTransactionStateDB()
	latestHeight := shardState.ShardHeight

	indices := make([]uint64, 0)
	publicKeys := make([][]byte, 0)
	commitments := make([][]byte, 0)
	assetTags := make([][]byte, 0)

	// these coins either all have asset tags or none does
	hasAssetTags := true
	failedCount := 0
	var idx *big.Int
	var coinDB *coin.CoinV2
	var err error
	for i := 0; i < numOutputs; i++ {
		for {
			idx, coinDB, err = ring_selection.Pick(db, shardID, *tokenID, latestHeight)
			if err != nil {
				failedCount++
				if failedCount > ring_selection.MaxGammaTries*numOutputs {
					return nil, nil, nil, nil, fmt.Errorf("max attempt exceeded")
				}
			} else {
				break
			}

		}

		commitment := coinDB.GetCommitment()
		indices = append(indices, idx.Uint64())
		publicKeys = append(publicKeys, coinDB.GetPublicKey().ToBytesS())
		commitments = append(commitments, commitment.ToBytesS())

		if hasAssetTags {
			assetTag := coinDB.GetAssetTag()
			if assetTag != nil {
				assetTags = append(assetTags, assetTag.ToBytesS())
			} else {
				hasAssetTags = false
			}
		}
	}

	return indices, publicKeys, commitments, assetTags, nil
}
