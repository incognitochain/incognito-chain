package rpcserver

import (
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

/*
handleGetInOutPeerMessageCount - return all inbound/outbound message count by peer which this node connected
*/
func (httpServer *HttpServer) handleGetInOutMessageCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetInOutMessageCount by Peer params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	result, err := jsonresult.NewGetInOutMessageCountResult(paramsArray)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	// Logger.log.Debugf("handleGetInOutPeerMessages result: %+v", result)
	return result, nil
}

/*
handleGetInOutPeerMessages - return all inbound/outbound messages peer which this node connected
*/
func (httpServer *HttpServer) handleGetInOutMessages(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetInOutPeerMessagess params: %+v", params)

	paramsArray := common.InterfaceSlice(params)
	result, err := jsonresult.NewGetInOutMessageResult(paramsArray)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	//Logger.log.Debugf("handleGetInOutPeerMessages result: %+v", result)
	return result, nil
}

/*
handleGetAllConnectedPeers - return all connnected peers which this node connected
*/
func (httpServer *HttpServer) handleGetAllConnectedPeers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetAllConnectedPeers params: %+v", params)
	result := jsonresult.NewGetAllConnectedPeersResult(*httpServer.config.ConnMgr)
	Logger.log.Debugf("handleGetAllPeers result: %+v", result)
	return result, nil
}

/*
handleGetAllPeers - return all peers which this node connected
*/
func (httpServer *HttpServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetAllPeers params: %+v", params)
	result := jsonresult.NewGetAllPeersResult(*httpServer.config.AddrMgr)
	Logger.log.Debugf("handleGetAllPeers result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetNodeRole(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return httpServer.config.Server.GetNodeRole(), nil
}

func (httpServer *HttpServer) handleGetNetWorkInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result, err := jsonresult.NewGetNetworkInfoResult(httpServer.config.ProtocolVersion, *httpServer.config.ConnMgr, httpServer.config.Wallet)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	return result, nil
}

//handleListUnspentOutputCoins - use private key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//component:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list priv-key which be used to view utxo
//
func (httpServer *HttpServer) handleListUnspentOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleListUnspentOutputCoins params: %+v", params)
	result := jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}

	// get component
	paramsArray := common.InterfaceSlice(params)
	var min int
	var max int
	if len(paramsArray) > 0 && paramsArray[0] != nil {
		min = int(paramsArray[0].(float64))
	}
	if len(paramsArray) > 1 && paramsArray[1] != nil {
		max = int(paramsArray[1].(float64))
	}
	_ = min
	_ = max
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain pri-key by deserializing
		priKeyStr := keys["PrivateKey"].(string)
		keyWallet, err := wallet.Base58CheckDeserialize(priKeyStr)
		if err != nil {
			log.Println("Check Deserialize err", err)
			continue
		}
		if keyWallet.KeySet.PrivateKey == nil {
			log.Println("Private key empty")
			continue
		}

		err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}
		shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
		tokenID := &common.Hash{}
		err = tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, NewRPCError(TokenIsInvalidError, err)
		}
		outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&keyWallet.KeySet, shardID, tokenID)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.OutCoin{
				SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSerialNumber().Compress(), common.ZeroByte),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.GetPublicKey().Compress(), common.ZeroByte),
				Value:          strconv.FormatUint(outCoin.CoinDetails.GetValue(), 10),
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.GetInfo()[:], common.ZeroByte),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.GetCoinCommitment().Compress(), common.ZeroByte),
				Randomness:     base58.Base58Check{}.Encode(outCoin.CoinDetails.GetRandomness().Bytes(), common.ZeroByte),
				SNDerivator:    base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSNDerivator().Bytes(), common.ZeroByte),
			})
		}
		result.Outputs[priKeyStr] = item
	}
	Logger.log.Debugf("handleListUnspentOutputCoins result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleCheckHashValue(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCheckHashValue params: %+v", params)
	var (
		isTransaction bool
		isBlock       bool
		isBeaconBlock bool
	)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Expected array component"))
	}
	hashParams, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Expected hash string value"))
	}
	// param #1: transaction Hash
	Logger.log.Debugf("Check hash value  input Param %+v", arrayParams[0].(string))
	log.Printf("Check hash value  input Param %+v", hashParams)
	hash, err2 := common.Hash{}.NewHashFromStr(hashParams)
	if err2 != nil {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Expected hash string value"))
	}
	// Check block
	_, _, err := httpServer.config.BlockChain.GetShardBlockByHash(*hash)
	if err != nil {
		isBlock = false
		_, _, err = httpServer.config.BlockChain.GetBeaconBlockByHash(*hash)
		if err != nil {
			isBeaconBlock = false
		} else {
			result := jsonresult.HashValueDetail{
				IsBlock:       isBlock,
				IsTransaction: false,
				IsBeaconBlock: true,
			}
			Logger.log.Debugf("handleCheckHashValue result: %+v", result)
			return result, nil
		}
	} else {
		isBlock = true
		result := jsonresult.HashValueDetail{
			IsBlock:       isBlock,
			IsTransaction: false,
		}
		Logger.log.Debugf("handleCheckHashValue result: %+v", result)
		return result, nil
	}
	_, _, _, _, err1 := httpServer.config.BlockChain.GetTransactionByHash(*hash)
	if err1 != nil {
		isTransaction = false
	} else {
		isTransaction = true
		result := jsonresult.HashValueDetail{
			IsBlock:       false,
			IsTransaction: isTransaction,
		}
		Logger.log.Debugf("handleCheckHashValue result: %+v", result)
		return result, nil
	}
	result := jsonresult.HashValueDetail{
		IsBlock:       isBlock,
		IsTransaction: isTransaction,
		IsBeaconBlock: isBeaconBlock,
	}
	Logger.log.Debugf("handleCheckHashValue result: %+v", result)
	return result, nil
}

/*
handleGetConnectionCount - RPC returns the number of connections to other nodes.
*/
func (httpServer *HttpServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetConnectionCount params: %+v", params)
	if httpServer.config.ConnMgr == nil || httpServer.config.ConnMgr.GetListeningPeer() == nil {
		return 0, nil
	}
	result := 0
	listeningPeer := httpServer.config.ConnMgr.GetListeningPeer()
	result += len(listeningPeer.GetPeerConns())
	Logger.log.Debugf("handleGetConnectionCount result: %+v", result)
	return result, nil
}

/*
handleEstimateFee - RPC estimates the transaction fee per kilobyte that needs to be paid for a transaction to be included within a certain number of blocks.
*/
func (httpServer *HttpServer) handleEstimateFee(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleEstimateFee params: %+v", params)
	/******* START Fetch all component to ******/
	// all component
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Not enough params"))
	}
	// param #1: private key of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Sender private key is invalid"))
	}
	// param #3: estimation fee coin per kb
	defaultFeeCoinPerKbtemp, ok := arrayParams[2].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Default FeeCoinPerKbtemp is invalid"))
	}
	defaultFeeCoinPerKb := int64(defaultFeeCoinPerKbtemp)
	// param #4: hasPrivacy flag for PRV
	hashPrivacyTemp, ok := arrayParams[3].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("hasPrivacy is invalid"))
	}
	hasPrivacy := int(hashPrivacyTemp) > 0

	senderKeySet, err := httpServer.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, NewRPCError(InvalidSenderPrivateKeyError, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

	prvCoinID := &common.Hash{}
	err = prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, NewRPCError(TokenIsInvalidError, err)
	}
	outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, prvCoinID)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	// remove out coin in mem pool
	outCoins, err = httpServer.filterMemPoolOutCoinsToSpent(outCoins)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}

	estimateFeeCoinPerKb := uint64(0)
	estimateTxSizeInKb := uint64(0)
	if len(outCoins) > 0 {
		// param #2: list receiver
		receiversPaymentAddressStrParam := make(map[string]interface{})
		if arrayParams[1] != nil {
			receiversPaymentAddressStrParam = arrayParams[1].(map[string]interface{})
		}
		paymentInfos := make([]*privacy.PaymentInfo, 0)
		for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
			keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
			if err != nil {
				return nil, NewRPCError(InvalidReceiverPaymentAddressError, err)
			}
			paymentInfo := &privacy.PaymentInfo{
				Amount:         uint64(amount.(float64)),
				PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
			}
			paymentInfos = append(paymentInfos, paymentInfo)
		}

		// Check custom token param
		var customTokenParams *transaction.CustomTokenParamTx
		var customPrivacyTokenParam *transaction.CustomTokenPrivacyParamTx
		if len(arrayParams) > 4 {
			// param #5: token params
			tokenParamsRaw := arrayParams[4].(map[string]interface{})
			isPrivacy := tokenParamsRaw["Privacy"].(bool)
			if !isPrivacy {
				// Check normal custom token param
				customTokenParams, _, err = httpServer.buildCustomTokenParam(tokenParamsRaw, senderKeySet)
				if err.(*RPCError) != nil {
					return nil, err.(*RPCError)
				}
			} else {
				// Check privacy custom token param
				customPrivacyTokenParam, _, _, err = httpServer.buildPrivacyCustomTokenParam(tokenParamsRaw, senderKeySet, shardIDSender)
				if err.(*RPCError) != nil {
					return nil, err.(*RPCError)
				}
			}
		}

		// check real fee(nano PRV) per tx
		_, estimateFeeCoinPerKb, estimateTxSizeInKb = httpServer.estimateFee(defaultFeeCoinPerKb, outCoins, paymentInfos, shardIDSender, 8, hasPrivacy, nil, customTokenParams, customPrivacyTokenParam)
	}
	result := jsonresult.NewEstimateFeeResult(estimateFeeCoinPerKb, estimateTxSizeInKb)
	Logger.log.Debugf("handleEstimateFee result: %+v", result)
	return result, nil
}

// handleEstimateFeeWithEstimator -- get fee from estomator
func (httpServer *HttpServer) handleEstimateFeeWithEstimator(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleEstimateFeeWithEstimator params: %+v", params)
	// all params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 2 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Not enough params"))
	}
	// param #1: estimation fee coin per kb from client
	defaultFeeCoinPerKbTemp, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("defaultFeeCoinPerKbTemp is invalid"))
	}
	defaultFeeCoinPerKb := int64(defaultFeeCoinPerKbTemp)

	// param #2: payment address
	senderKeyParam := arrayParams[1]
	senderKeySet, err := httpServer.GetKeySetFromKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(InvalidSenderPrivateKeyError, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	// param #2: numbloc
	numblock := uint64(8)
	if len(arrayParams) >= 3 {
		numblock = uint64(arrayParams[2].(float64))
	}

	// param #3: tokenId
	var tokenId *common.Hash
	if len(arrayParams) >= 4 && arrayParams[3] != nil {
		tokenId, err = common.Hash{}.NewHashFromStr(arrayParams[3].(string))
	}
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	estimateFeeCoinPerKb := httpServer.estimateFeeWithEstimator(defaultFeeCoinPerKb, shardIDSender, numblock, tokenId)

	result := jsonresult.NewEstimateFeeResult(estimateFeeCoinPerKb, 0)
	Logger.log.Debugf("handleEstimateFeeWithEstimator result: %+v", result)
	return result, nil
}

/*func (httpServer *HttpServer) handleGetFeeEstimator(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return httpServer.config.FeeEstimator, nil
}*/

// handleGetActiveShards - return active shard num
func (httpServer *HttpServer) handleGetActiveShards(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetActiveShards params: %+v", params)
	activeShards := httpServer.config.BlockChain.BestState.Beacon.ActiveShards
	Logger.log.Debugf("handleGetActiveShards result: %+v", activeShards)
	return activeShards, nil
}

func (httpServer *HttpServer) handleGetMaxShardsNumber(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetMaxShardsNumber params: %+v", params)
	result := common.MAX_SHARD_NUMBER
	Logger.log.Debugf("handleGetMaxShardsNumber result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetStakingAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetStakingAmount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) <= 0 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("ErrRPCInvalidParams"))
	}
	stackingType := int(arrayParams[0].(float64))
	amount := uint64(0)
	stakingData, _ := metadata.NewStakingMetadata(metadata.ShardStakingMeta, "", httpServer.config.ChainParams.StakingAmountShard, "", true)
	if stackingType == 1 {
		amount = stakingData.GetBeaconStakeAmount()
	}
	if stackingType == 0 {
		amount = stakingData.GetShardStateAmount()
	}
	Logger.log.Debugf("handleGetStakingAmount result: %+v", amount)
	return amount, nil
}

func (httpServer *HttpServer) handleHashToIdenticon(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := make([]string, 0)
	for _, hash := range arrayParams {
		temp, err := common.Hash{}.NewHashFromStr(hash.(string))
		if err != nil {
			return nil, NewRPCError(UnexpectedError, errors.New("Hash string is invalid"))
		}
		result = append(result, common.Render(temp.GetBytes()))
	}
	return result, nil
}

// handleGetPublicKeyMining - return publickey mining which be used to verify block
func (httpServer *HttpServer) handleGetPublicKeyMining(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	publicKey, keyType := httpServer.config.ConsensusEngine.GetCurrentMiningPublicKey()
	if publicKey != "" {
		key := struct {
			PublicKey string
			KeyType   string
		}{
			PublicKey: publicKey,
			KeyType:   keyType,
		}
		return key, nil
	} else {
		return nil, NewRPCError(UnexpectedError, errors.New("Can not find key"))
	}
}
