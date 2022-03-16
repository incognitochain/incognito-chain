package rpcserver

import (
	"encoding/json"
	"fmt"
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

// EVMDepositParams consists of parameters for creating an EVM shielding transaction.
// An EVMDepositParams is valid if at least one of the following conditions hold:
//	- Signature is not empty
//		- Receiver must not be empty
//	- Signature is empty
//		- Either DepositPrivateKey or DepositKeyIndex must not be empty; if DepositPrivateKey is empty, it will be
// 	derived from the DepositKeyIndex
//		- Receiver will be generated from the sender's private key
//		- Signature will be signed using the DepositPrivateKey
type EVMDepositParams struct {
	// MetadataType is the type of the shielding request.
	MetadataType int

	// BlockHash is the hash of the block where the public depositing transaction resides in.
	BlockHash rCommon.Hash

	// TxIndex is the index of the public depositing transaction in the BlockHash.
	TxIndex uint

	// TokenID is the shielding asset ID.
	TokenID string

	// ShieldingProofs is a proof for the shielding request.
	ShieldingProofs []string

	// DepositPrivateKey is a base58-encoded deposit privateKey used to sign the request.
	// If set empty, it will be derived from the DepositKeyIndex.
	DepositPrivateKey string

	// DepositKeyIndex is the index of the OTDepositKey. It is used to generate DepositPrivateKey and DepositPubKey when the DepositPrivateKey is not supply.
	DepositKeyIndex uint64

	// Receiver is a base58-encoded OTAReceiver. If set empty, it will be generated from the sender's privateKey.
	Receiver string

	// Signature is a valid signature signed by the owner of the shielding asset.
	// If Signature is not empty, Receiver must not be empty.
	Signature string
}

// IsValid checks if a EVMDepositParams is valid.
func (dp EVMDepositParams) IsValid() (bool, error) {
	var err error
	_, err = common.Hash{}.NewHashFromStr(dp.TokenID)
	if err != nil || dp.TokenID == "" {
		return false, fmt.Errorf("invalid tokenID %v", dp.TokenID)
	}

	if len(dp.ShieldingProofs) == 0 {
		return false, fmt.Errorf("invalid proofs")
	}

	if dp.Signature != "" {
		_, _, err = base58.Base58Check{}.Decode(dp.Signature)
		if err != nil {
			return false, fmt.Errorf("invalid signature")
		}
		if dp.Receiver == "" {
			return false, fmt.Errorf("must have `Receiver`")
		}
	} else {
		if dp.DepositPrivateKey != "" {
			_, _, err = base58.Base58Check{}.Decode(dp.DepositPrivateKey)
			if err != nil {
				return false, fmt.Errorf("invalid DepositPrivateKey")
			}
		}
	}

	if dp.Receiver != "" {
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(dp.Receiver)
		if err != nil {
			return false, fmt.Errorf("invalid receiver: %v", err)
		}
	}

	return true, nil
}

func (httpServer *HttpServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	constructor := metaConstructors[createAndSendIssuingRequest]
	return httpServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (httpServer *HttpServer) handleCreateIssuingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateIssuingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleSendRawTransaction(params, closeChan)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (httpServer *HttpServer) handleCreateAndSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateIssuingRequest,
		(*HttpServer).handleSendIssuingRequest,
	)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (httpServer *HttpServer) handleCreateAndSendIssuingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendIssuingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithContractingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// check privacy mode param
	if len(arrayParams) > 6 {
		privacyTemp, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be valid"))
		}
		hasPrivacyToken := int(privacyTemp) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}

	senderPrivateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token receivers is invalid"))
	}

	tokenID, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	var txVersion int8
	tmpVersionParam, ok := tokenParamsRaw["TxVersion"]
	if !ok {
		txVersion = 2
	} else {
		tmpVersion, ok := tmpVersionParam.(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("txVersion must be a float64"))
		}
		txVersion = int8(tmpVersion)
	}

	meta, err := rpcservice.NewContractingRequestMetadata(senderPrivateKeyParam, tokenReceivers, tokenID)
	if err != nil {
		return nil, err
	}
	if txVersion == 1 {
		meta.BurnerAddress.OTAPublic = nil
	}

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithContractingReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// sendResult, err1 := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendContractingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendContractingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningRequestMetaV2,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// sendResult, err1 := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendBurningRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendBurningRequestV2(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithIssuingEVMReq(params interface{}, closeChan <-chan struct{}, metatype int) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	meta, err := metadata.NewIssuingEVMRequestFromMap(data, metatype)
	if err != nil {
		rpcErr := rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingETHReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingETHRequestMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingETHReqV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendTxWithIssuingETHReq(params, closeChan)
}

func (httpServer *HttpServer) handleCreateEVMDepositTxWithDepositKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	privateKeyStr := arrayParams[0].(string)
	w, err := wallet.Base58CheckDeserialize(privateKeyStr)
	if err != nil || len(w.KeySet.PrivateKey[:]) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid privateKey: %v", privateKeyStr))
	}
	privateKey := w.KeySet.PrivateKey[:]

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("metadata param is invalid"))
	}
	jsb, _ := json.Marshal(data)

	var dp *EVMDepositParams
	err = json.Unmarshal(jsb, &dp)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot unmarshal depositParams"))
	}
	if _, err = dp.IsValid(); err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid EVMDepositParams: %v", err))
	}

	receiver := dp.Receiver
	var sig []byte
	if dp.Signature != "" {
		sig, _, _ = base58.Base58Check{}.Decode(dp.Signature)
	} else {
		if receiver == "" {
			otaReceiver := new(privacy.OTAReceiver)
			err = otaReceiver.FromAddress(w.KeySet.PaymentAddress)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("generate OTAReceiver error: %v", err))
			}
			receiver, _ = otaReceiver.String()
		}
		otaReceiver := new(coin.OTAReceiver)
		_ = otaReceiver.FromString(receiver)

		var depositPrivateKey *privacy.Scalar
		if dp.DepositPrivateKey != "" {
			tmp, _, _ := base58.Base58Check{}.Decode(dp.DepositPrivateKey)
			depositPrivateKey = new(privacy.Scalar).FromBytesS(tmp)
		} else {
			depositKey, err := incognitokey.GenerateOTDepositKeyFromPrivateKey(privateKey, dp.TokenID, dp.DepositKeyIndex)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("GenerateOTDepositKeyFromPrivateKey error: %v", err))
			}
			depositPrivateKey = new(privacy.Scalar).FromBytesS(depositKey.PrivateKey)
		}

		schnorrPrivateKey := new(privacy.SchnorrPrivateKey)
		r := new(privacy.Scalar).FromUint64(0) // must use r = 0
		schnorrPrivateKey.Set(depositPrivateKey, r)
		metaDataBytes, _ := otaReceiver.Bytes()
		tmpSig, err := schnorrPrivateKey.Sign(common.HashB(metaDataBytes))
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		sig = tmpSig.Bytes()
	}

	tokenID, _ := new(common.Hash).NewHashFromStr(dp.TokenID)

	meta, err := metadata.NewIssuingEVMRequest(
		dp.BlockHash,
		dp.TxIndex,
		dp.ShieldingProofs,
		*tokenID,
		receiver,
		sig,
		dp.MetadataType,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendEVMDepositTxWithDepositKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateEVMDepositTxWithDepositKey(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCheckETHHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckETHHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleCheckBSCHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckBSCHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleCheckPrvPeggingHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckPRVPeggingHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleGetAllBridgeTokens(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	allBridgeTokens, err := httpServer.blockService.GetAllBridgeTokens()
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return allBridgeTokens, nil
}

func (httpServer *HttpServer) handleGetAllBridgeTokensByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	if len(arrayParams) == 0 {
		return httpServer.handleGetAllBridgeTokens(params, closeChan)
	}
	height, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse height param, must be a float"))
	}

	allBridgeTokens, err := httpServer.blockService.GetAllBridgeTokensByHeight(uint64(height))
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return allBridgeTokens, nil
}

func (httpServer *HttpServer) handleGetETHHeaderByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	ethBlockHash := arrayParams[0].(string)

	ethHeader, err := rpcservice.GetETHHeaderByHash(ethBlockHash)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return ethHeader, nil
}

func (httpServer *HttpServer) handleGetBridgeReqWithStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	status, err := httpServer.blockService.GetBridgeReqWithStatus(data["TxReqID"].(string))
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return status, nil
}

func processBurningReq(
	burningMetaType int,
	params interface{},
	closeChan <-chan struct{},
	httpServer *HttpServer,
	isPRV bool,
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	if len(arrayParams) > 6 {
		privacyTemp, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be valid "))
		}
		hasPrivacyToken := int(privacyTemp) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}

	senderPrivateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token receivers is invalid"))
	}

	tokenID, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	tokenName, ok := tokenParamsRaw["TokenName"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token name is invalid"))
	}

	remoteAddress, ok := tokenParamsRaw["RemoteAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("remote address is invalid"))
	}

	var txVersion int8
	tmpVersionParam, ok := tokenParamsRaw["TxVersion"]
	if !ok {
		txVersion = 2
	} else {
		tmpVersion, ok := tmpVersionParam.(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("txVersion must be a float64"))
		}
		txVersion = int8(tmpVersion)
	}

	meta, err := rpcservice.NewBurningRequestMetadata(
		senderPrivateKeyParam,
		tokenReceivers,
		tokenID,
		tokenName,
		remoteAddress,
		burningMetaType,
		httpServer.GetBlockchain(),
		httpServer.GetBlockchain().BeaconChain.CurrentHeight(),
		txVersion)
	if err != nil {
		return nil, err
	}
	var byteArrays []byte
	var err2 error
	var txHash string
	if isPRV {
		// create new param to build raw tx from param interface
		createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
		if errNewParam != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
		}
		rawTx, rpcErr := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, rpcErr
		}
		byteArrays, err2 = json.Marshal(rawTx)
		txHash = rawTx.Hash().String()
	} else {
		customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, rpcErr
		}
		byteArrays, err2 = json.Marshal(customTokenTx)
		txHash = customTokenTx.Hash().String()
	}

	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            txHash,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningForDepositToSCRequestMetaV2,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendBurningForDepositToSCRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendBurningForDepositToSCRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingBSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingBSCRequestMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningBSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPBSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningBSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningBSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPRVERC20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPRVERC20RequestMeta,
		params,
		closeChan,
		httpServer,
		true,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPRVERC20Request(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPRVERC20Req(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPRVBEP20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPRVBEP20RequestMeta,
		params,
		closeChan,
		httpServer,
		true,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPRVBEP20Request(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPRVBEP20Req(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPRVERC20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPRVERC20RequestMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPRVBEP20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPRVBEP20RequestMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPLGReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPLGRequestMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPLGReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPLGRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPLGRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPLGReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPBSCForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPBSCForDepositToSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPBSCForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPBSCForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPLGForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPLGForDepositToSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPLGForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPLGForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleGenerateOTDepositKey(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}

	paramList, ok := paramsArray[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be a map[string]interface{}"))
	}

	type TmpParam struct {
		PrivateKey string
		Index      uint64
		TokenID    string
	}
	var tmpParam *TmpParam
	jsb, _ := json.Marshal(paramList)
	err := json.Unmarshal(jsb, &tmpParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("unmarshal parameters error: %v", err))
	}

	w, err := wallet.Base58CheckDeserialize(tmpParam.PrivateKey)
	if err != nil || len(w.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("invalid privateKey %v", tmpParam.PrivateKey))
	}
	depositKey, err := incognitokey.GenerateOTDepositKeyFromPrivateKey(w.KeySet.PrivateKey[:], tmpParam.TokenID, tmpParam.Index)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("GenerateNextOTDepositKey encoured an error: %v", err))
	}

	type Res struct {
		DepositKey     *incognitokey.OTDepositKey
		DepositAddress string `json:"DepositAddress,omitempty"`
	}

	// check is Portal token
	latestBeaconHeight := httpServer.config.BlockChain.GetBeaconBestState().BeaconHeight
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(latestBeaconHeight)
	if portalParamV4.IsPortalToken(tmpParam.TokenID) {
		// generate shielding multiSig address
		depositPubKeyStr := base58.Base58Check{}.NewEncode(depositKey.PublicKey, 0)
		_, depositAddress, err := portalParamV4.PortalTokens[tmpParam.TokenID].GenerateOTMultisigAddress(
			portalParamV4.MasterPubKeys[tmpParam.TokenID], int(portalParamV4.NumRequiredSigs), depositPubKeyStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
				fmt.Errorf("error when generating multisig address %v", err))
		}

		return Res{DepositKey: depositKey, DepositAddress: depositAddress}, nil
	}

	return Res{DepositKey: depositKey}, nil
}

func (httpServer *HttpServer) handleGetNextOTDepositKey(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}

	paramList, ok := paramsArray[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be a map[string]interface{}"))
	}

	type TmpParam struct {
		PrivateKey string
		TokenID    string
	}
	var tmpParam *TmpParam
	jsb, _ := json.Marshal(paramList)
	err := json.Unmarshal(jsb, &tmpParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("unmarshal parameters error: %v", err))
	}

	depositKey, err := httpServer.blockService.GenerateNextOTDepositKey(tmpParam.PrivateKey, tmpParam.TokenID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("GenerateNextOTDepositKey encoured an error: %v", err))
	}

	type Res struct {
		DepositKey     *incognitokey.OTDepositKey
		DepositAddress string
	}

	// check is Portal token
	latestBeaconHeight := httpServer.config.BlockChain.GetBeaconBestState().BeaconHeight
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(latestBeaconHeight)
	if portalParamV4.IsPortalToken(tmpParam.TokenID) {
		// generate shielding multiSig address
		depositPubKeyStr := base58.Base58Check{}.NewEncode(depositKey.PublicKey, 0)
		_, depositAddress, err := portalParamV4.PortalTokens[tmpParam.TokenID].GenerateOTMultisigAddress(
			portalParamV4.MasterPubKeys[tmpParam.TokenID], int(portalParamV4.NumRequiredSigs), depositPubKeyStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
				fmt.Errorf("error when generating multisig address %v", err))
		}

		return Res{DepositKey: depositKey, DepositAddress: depositAddress}, nil
	}

	return Res{DepositKey: depositKey}, nil
}

func (httpServer *HttpServer) handleHasOTDepositPubKeys(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}

	paramList, ok := paramsArray[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be a map[string]interface{}"))
	}

	type TmpParam struct {
		DepositPubKeys []string
	}
	var tmpParam *TmpParam
	jsb, _ := json.Marshal(paramList)
	err := json.Unmarshal(jsb, &tmpParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("unmarshal parameters error: %v", err))
	}

	beaconState, err := httpServer.blockService.BlockChain.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("error while retrieving beaconBestState: %v", err))
	}
	portalV4StateDB := beaconState.GetBeaconFeatureStateDB()
	res := make(map[string]bool)
	for _, pubKeyStr := range tmpParam.DepositPubKeys {
		pubKey, _, err := base58.Base58Check{}.Decode(pubKeyStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot decode depositPubKey %v, must be a base58-encoded string", pubKeyStr))
		}
		exists := statedb.OTDepositPubKeyExists(portalV4StateDB, pubKey)
		res[pubKeyStr] = exists
	}

	return res, nil
}

func (httpServer *HttpServer) handleGetDepositTxsByPubKeys(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}

	paramList, ok := paramsArray[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be a map[string]interface{}"))
	}

	type TmpParam struct {
		DepositPubKeys []string
	}
	var tmpParam *TmpParam
	jsb, _ := json.Marshal(paramList)
	err := json.Unmarshal(jsb, &tmpParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("unmarshal parameters error: %v", err))
	}

	beaconState, err := httpServer.blockService.BlockChain.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("error while retrieving beaconBestState: %v", err))
	}
	featureStateDB := beaconState.GetBeaconFeatureStateDB()
	res := make(map[string][]string)
	for _, pubKeyStr := range tmpParam.DepositPubKeys {
		pubKey, _, err := base58.Base58Check{}.Decode(pubKeyStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot decode depositPubKey %v, must be a base58-encoded string", pubKeyStr))
		}
		txList, _ := statedb.GetShieldRequestIDsByOTDepositPubKey(featureStateDB, pubKey)
		res[pubKeyStr] = txList
	}

	return res, nil
}
