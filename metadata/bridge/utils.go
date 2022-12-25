package bridge

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	nearclient "github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	"github.com/eteu-technologies/near-api-go/pkg/types/hash"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type EVMProof struct {
	BlockHash rCommon.Hash `json:"BlockHash"`
	TxIndex   uint         `json:"TxIndex"`
	Proof     []string     `json:"Proof"`
}

type EVMInfo struct {
	ContractAddress   string
	Prefix            string
	IsTxHashIssued    func(stateDB *statedb.StateDB, uniqTx []byte) (bool, error)
	ListTxUsedInBlock [][]byte
}

func GetEVMInfoByNetworkID(networkID uint8, ac *metadataCommon.AccumulatedValues) (*EVMInfo, error) {
	res := &EVMInfo{}
	switch networkID {
	case common.ETHNetworkID:
		res.ListTxUsedInBlock = ac.UniqETHTxsUsed
		res.ContractAddress = config.Param().EthContractAddressStr
		res.Prefix = utils.EmptyString
		res.IsTxHashIssued = statedb.IsETHTxHashIssued
	case common.BSCNetworkID:
		res.ListTxUsedInBlock = ac.UniqBSCTxsUsed
		res.ContractAddress = config.Param().BscContractAddressStr
		res.Prefix = common.BSCPrefix
		res.IsTxHashIssued = statedb.IsBSCTxHashIssued
	case common.PLGNetworkID:
		res.ListTxUsedInBlock = ac.UniqPLGTxsUsed
		res.ContractAddress = config.Param().PlgContractAddressStr
		res.Prefix = common.PLGPrefix
		res.IsTxHashIssued = statedb.IsPLGTxHashIssued
	case common.FTMNetworkID:
		res.ListTxUsedInBlock = ac.UniqFTMTxsUsed
		res.ContractAddress = config.Param().FtmContractAddressStr
		res.Prefix = common.FTMPrefix
		res.IsTxHashIssued = statedb.IsFTMTxHashIssued
	case common.AURORANetworkID:
		res.ListTxUsedInBlock = ac.UniqAURORATxsUsed
		res.ContractAddress = config.Param().AuroraContractAddressStr
		res.Prefix = common.AURORAPrefix
		res.IsTxHashIssued = statedb.IsAURORATxHashIssued
	case common.AVAXNetworkID:
		res.ListTxUsedInBlock = ac.UniqAVAXTxsUsed
		res.ContractAddress = config.Param().AvaxContractAddressStr
		res.Prefix = common.AVAXPrefix
		res.IsTxHashIssued = statedb.IsAVAXTxHashIssued
	default:
		return nil, errors.New("Invalid networkID")
	}
	return res, nil
}

func IsBridgeAggMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeAggModifyParamMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenRequestMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenResponseMeta:
		return true
	case metadataCommon.BurningUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BurningUnifiedTokenResponseMeta:
		return true
	case metadataCommon.BridgeAggAddTokenMeta:
		return true
	case metadataCommon.BurnForCallConfirmMeta, metadataCommon.BurnForCallRequestMeta, metadataCommon.BurnForCallResponseMeta:
		return true
	case metadataCommon.IssuingReshieldResponseMeta:
		return true
	default:
		return false
	}
}

func IsBridgeTxHashUsedInBlock(uniqTx []byte, uniqTxsUsed [][]byte) bool {
	for _, item := range uniqTxsUsed {
		if bytes.Equal(uniqTx, item) {
			return true
		}
	}
	return false
}

func GetShardIDFromPaymentAddress(paymentAddress key.PaymentAddress) (byte, error) {
	// calculate shard ID
	lastByte := paymentAddress.Pk[len(paymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func GetShardIDFromPaymentAddressStr(addressStr string) (byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return byte(0), errors.New("Payment address' public key must not be empty")
	}
	// calculate shard ID
	return GetShardIDFromPaymentAddress(keyWallet.KeySet.PaymentAddress)
}

func ExtractIssueEVMDataFromReceipt(
	txReceipt *types.Receipt,
	contractAddress string, prefix string,
	expectedIncAddrStr string,
) (*big.Int, string, []byte, error) {
	if txReceipt == nil {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: bridge tx receipt is null")
	}

	logMap, err := PickAndParseLogMapFromReceiptByContractAddr(txReceipt, contractAddress, "Deposit")
	if err != nil {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: an error occurred while parsing log map from receipt: %v", err)
	}
	if logMap == nil {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: could not find log map out from receipt")
	}

	logMapBytes, _ := json.Marshal(logMap)
	metadataCommon.Logger.Log.Warn("INFO: eth logMap json - ", string(logMapBytes))

	// the token might be ETH/ERC20 BNB/BEP20
	tokenAddr, ok := logMap["token"].(rCommon.Address)
	if !ok {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: could not parse evm token id from log map.")
	}
	extTokenID := append([]byte(prefix), tokenAddr.Bytes()...)

	incAddrStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: could not parse incognito address from bridge log map.")
	}
	if expectedIncAddrStr != "" && incAddrStr != expectedIncAddrStr {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: different incognito address from bridge log map.")
	}
	if expectedIncAddrStr == "" {
		key, err := wallet.Base58CheckDeserialize(incAddrStr)
		if err != nil || len(key.KeySet.PaymentAddress.Pk) == 0 {
			return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: invalid incognito address from bridge log map.")
		}
	}

	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		return nil, utils.EmptyString, nil, fmt.Errorf("WARNING: could not parse amount from bridge log map.")
	}
	return amt, incAddrStr, extTokenID, nil
}

type DepositEventData struct {
	Amount          *big.Int
	ReceiverStr     string
	ExternalTokenID []byte
	IncTxID         []byte
	ShardID         byte
	IsOneTime       bool
}

func ExtractRedepositEVMDataFromReceipt(
	txReceipt *types.Receipt,
	contractAddress string, prefix string,
) ([]DepositEventData, error) {
	if txReceipt == nil {
		return nil, fmt.Errorf("bridge tx receipt is null")
	}
	if len(txReceipt.Logs) == 0 {
		metadataCommon.Logger.Log.Errorf("WARNING: LOG data is invalid.")
		return nil, nil
	}

	var result []DepositEventData
	for _, log := range txReceipt.Logs {
		if log, err := ParseEVMLogDataByEventName(log, contractAddress, "Redeposit"); err != nil {
			metadataCommon.Logger.Log.Errorf("WARNING: try parse Redeposit event - %v", err)
		} else {
			logMapBytes, _ := json.Marshal(log)
			metadataCommon.Logger.Log.Warn("INFO: eth redeposit log json - ", string(logMapBytes))

			// the token might be ETH/ERC20 BNB/BEP20
			tokenAddr, ok := log["token"].(rCommon.Address)
			if !ok {
				return nil, fmt.Errorf("could not parse evm token id from log map")
			}
			extTokenID := append([]byte(prefix), tokenAddr.Bytes()...)

			incAddr, ok := log["redepositIncAddress"].([]byte)
			if !ok {
				return nil, fmt.Errorf("could not parse incognito address from bridge log map")
			}
			var recv privacy.OTAReceiver
			err := recv.SetBytes(incAddr)
			if err != nil {
				return nil, fmt.Errorf("invalid incognito receiver from bridge reshield log map.")
			}
			incAddrStr, _ := recv.String()

			amt, ok := log["amount"].(*big.Int)
			if !ok {
				return nil, fmt.Errorf("could not parse amount from bridge log map.")
			}
			incTxID, ok := log["itx"].([32]byte)
			if !ok {
				return nil, fmt.Errorf("could not parse itx from bridge log map.")
			}

			result = append(result, DepositEventData{
				Amount:          amt,
				ReceiverStr:     incAddrStr,
				ExternalTokenID: extTokenID,
				IncTxID:         incTxID[:],
				ShardID:         recv.GetShardID(),
				IsOneTime:       true,
			})
		}
	}

	return result, nil
}

func VerifyWasmData(
	stateDB *statedb.StateDB, listTxUsed [][]byte,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error),
	externalShieldTx string, incognitoAddress string,
) (byte, error) {
	uniqTxCryptoHash, err := hash.NewCryptoHashFromBase58(externalShieldTx)
	if err != nil {
		return 0, fmt.Errorf("WARNING: invalid external shield tx request %v", externalShieldTx)
	}
	uniqTxTemp := [32]byte(uniqTxCryptoHash)
	uniqTx := uniqTxTemp[:]
	isUsedInBlock := IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		return 0, fmt.Errorf("WARNING: already issued for the hash in current block: %v", uniqTx)
	}
	isIssued, err := isTxHashIssued(stateDB, uniqTx)
	if err != nil {
		return 0, fmt.Errorf("WARNING: an issue occured while checking the bridge tx hash is issued or not: %v", err)
	}
	if isIssued {
		return 0, fmt.Errorf("WARNING: already issued for the hash in previous blocks: %v", uniqTx)
	}

	receivingShardID, err := GetShardIDFromPaymentAddressStr(incognitoAddress)
	if err != nil {
		return 0, fmt.Errorf("WARNING: an error occurred while getting shard id from payment address: %v", err)
	}
	return receivingShardID, nil
}

func VerifyTokenPair(
	stateDBs map[int]*statedb.StateDB,
	ac *metadataCommon.AccumulatedValues,
	incTokenID common.Hash,
	token []byte,
) error {
	canProcess, err := ac.CanProcessTokenPair(token, incTokenID)
	if err != nil {
		return fmt.Errorf("WARNING: an error occurred while checking it can process for token pair on the current block or not: %v", err)
	}
	if !canProcess {
		return fmt.Errorf("WARNING: pair of incognito token id & bridge's id is invalid in current block")
	}
	privacyTokenExisted, err := statedb.CheckTokenIDExisted(stateDBs, incTokenID)
	if err != nil {
		return fmt.Errorf("WARNING: Cannot find tokenID %s", incTokenID.String())
	}
	isValid, err := statedb.CanProcessTokenPair(stateDBs[common.BeaconChainID], token, incTokenID, privacyTokenExisted)
	if err != nil {
		return fmt.Errorf("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: %v", err)
	}
	if !isValid {
		return fmt.Errorf("WARNING: pair of incognito token id %s & bridge's id %x is invalid with previous blocks", incTokenID, token)
	}
	return nil
}

func FindExternalTokenID(stateDB *statedb.StateDB, incTokenID common.Hash, prefix string) ([]byte, error) {
	// Convert to external tokenID
	tokenID, err := findExternalTokenID(stateDB, &incTokenID)
	if err != nil {
		return nil, err
	}

	if prefix != common.NEARPrefix {
		if len(tokenID) < common.ExternalBridgeTokenLength {
			return nil, errors.New("invalid external token id")
		}

		prefixLen := len(prefix)
		if (prefixLen > 0 && !bytes.Equal([]byte(prefix), tokenID[:prefixLen])) || len(tokenID) != (common.ExternalBridgeTokenLength+prefixLen) {
			return nil, errors.New(fmt.Sprintf("invalid prefix in external tokenID %v", tokenID))
		}
	}
	return tokenID, nil
}

// TrimNetworkPrefix is a helper that removes the 3-byte network prefix from the full 23-byte external address (for burning confirm etc.);
// within the bridgeagg vault we only use prefixed addresses
func TrimNetworkPrefix(fullTokenID []byte, prefix string) ([]byte, error) {
	if !bytes.HasPrefix(fullTokenID, []byte(prefix)) {
		return nil, fmt.Errorf("invalid prefix in external tokenID %x, expect %s", fullTokenID, prefix)
	}
	result := fullTokenID[len(prefix):]
	if len(result) != common.ExternalBridgeTokenLength {
		return nil, fmt.Errorf("invalid length %d for un-prefixed external address", len(result))
	}
	return result, nil
}

// findExternalTokenID finds the external tokenID for a bridge token from database
func findExternalTokenID(stateDB *statedb.StateDB, tokenID *common.Hash) ([]byte, error) {
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(stateDB)
	if err != nil {
		return nil, err
	}
	var allBridgeTokens []*rawdbv2.BridgeTokenInfo
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, token := range allBridgeTokens {
		if token.TokenID.IsEqual(tokenID) && len(token.ExternalTokenID) > 0 {
			return token.ExternalTokenID, nil
		}
	}
	return nil, errors.New("invalid tokenID")
}

func VerifyProofAndParseEVMReceipt(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	hosts []string,
	minEVMConfirmationBlocks int,
	networkPrefix string,
	checkEVMHarkFork bool,
) (*types.Receipt, error) {
	// get evm header result
	evmHeaderResult, err := evmcaller.GetEVMHeaderResult(blockHash, hosts, minEVMConfirmationBlocks, networkPrefix)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("Can not get EVM header result - Error: %+v", err)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	// check fork
	if evmHeaderResult.IsForked {
		metadataCommon.Logger.Log.Errorf("EVM Block hash %s is not in main chain", blockHash.String())
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt,
			fmt.Errorf("EVM Block hash %s is not in main chain", blockHash.String()))
	}

	// check min confirmation blocks
	if !evmHeaderResult.IsFinalized {
		metadataCommon.Logger.Log.Errorf("EVM block hash %s is not enough confirmation blocks %v", blockHash.String(), minEVMConfirmationBlocks)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt,
			fmt.Errorf("EVM block hash %s is not enough confirmation blocks %v", blockHash.String(), minEVMConfirmationBlocks))
	}

	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, txIndex)

	nodeList := new(light.NodeList)
	for _, proofStr := range proofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			return nil, err
		}
		nodeList.Put([]byte{}, proofBytes)
	}
	proof := nodeList.NodeSet()
	val, _, err := trie.VerifyProof(evmHeaderResult.Header.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		errMsg := fmt.Sprintf("WARNING: EVM issuance proof verification failed: %v", err)
		metadataCommon.Logger.Log.Warn(errMsg)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	// if iReq.Type == IssuingETHRequestMeta || iReq.Type == IssuingPRVERC20RequestMeta || iReq.Type == IssuingPLGRequestMeta {
	if checkEVMHarkFork {
		if len(val) == 0 {
			return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("the encoded receipt is empty"))
		}

		// hardfork london with new transaction type => 0x02 || RLP([...SenderPayload, ...SenderSignature, ...GasPayerPayload, ...GasPayerSignature])
		if val[0] == AccessListTxType || val[0] == DynamicFeeTxType {
			val = val[1:]
		}
	}

	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	if constructedReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("The constructedReceipt's status is not success"))
	}

	return constructedReceipt, nil
}

func VerifyWasmShieldTxId(
	txHash string,
	hosts []string,
	minWasmConfirmationBlocks int,
	contractId string,
) (string, string, uint64, string, error) {
	tx, err := hash.NewCryptoHashFromBase58(txHash)
	if err != nil {
		return "", "", 0, "", errors.New("Invalid transaction hash")
	}
	ctx := context.Background()
	accountId := "incognito"
	var token, incognitoAddress, extractedContractId string
	var amount uint64
	isValid := false
	for _, h := range hosts {
		rpcClient, err := nearclient.NewClient(h)
		if err != nil {
			continue
		}
		txStatus, err := rpcClient.TransactionStatus(ctx, tx, accountId)
		if err != nil {
			continue
		}
		if len(txStatus.Status.Failure) != 0 {
			return "", "", 0, "", errors.New("Transaction shield is failed")
		}
		minedBlock, err := rpcClient.BlockDetails(ctx, block.BlockHash(txStatus.TransactionOutcome.BlockHash))
		if err != nil {
			continue
		}
		// verify block not forked
		blockByHeight, err := rpcClient.BlockDetails(ctx, block.BlockID(uint(minedBlock.Header.Height)))
		if err != nil {
			continue
		}
		if blockByHeight.Header.Hash.String() != txStatus.TransactionOutcome.BlockHash.String() {
			return "", "", 0, "", errors.New("Transaction is in forked chain")
		}

		latestBlock, err := rpcClient.BlockDetails(ctx, block.FinalityFinal())
		if err != nil {
			continue
		}
		if latestBlock.Header.Height < minedBlock.Header.Height+uint64(minWasmConfirmationBlocks) {
			return "", "", 0, "", errors.New("The shield transaction is not finalized")
		}

		// detect shield event
		for _, receiptOutCome := range txStatus.ReceiptsOutcome {
			if len(receiptOutCome.Outcome.Status.Failure) != 0 {
				isValid = false
				break
			}
			if isValid {
				continue
			}
			if receiptOutCome.Outcome.ExecutorID != contractId {
				continue
			}
			if len(receiptOutCome.Outcome.Logs) == 0 {
				continue
			}
			events := strings.Split(receiptOutCome.Outcome.Logs[0], " ")
			if len(events) != 3 {
				continue
			}
			amount, err = strconv.ParseUint(events[2], 10, 64)
			if err != nil {
				continue
			}
			token = events[1]
			incognitoAddress = events[0]
			extractedContractId = receiptOutCome.Outcome.ExecutorID
			isValid = true
		}
		if isValid {
			return incognitoAddress, token, amount, extractedContractId, nil
		}
		break
	}
	return "", "", 0, "", errors.New("The endpoints are not response or set or invalid transaction")
}

type NormalResult struct {
	Result interface{} `json:"result"`
}

func VerifyProofAndParseAuroraReceipt(
	txHash common.Hash,
	auroraHosts []string,
	nearHosts []string,
	minEVMConfirmationBlocks int,
	networkPrefix string,
) (*types.Receipt, error) {
	// query tx receipt from auora chain
	var unstructuredResult map[string]interface{}
	var res *types.Receipt
	var err error
	for _, h := range auroraHosts {
		unstructuredResult, err = getAURORATransactionReceipt(h, txHash)
		if err != nil {
			return nil, err
		}
		if unstructuredResult != nil {
			break
		}
	}
	if unstructuredResult == nil {
		return nil, fmt.Errorf("query receipt %v got nil value", txHash.String())
	}
	nearTransactionHash, exist := unstructuredResult["nearTransactionHash"]
	if !exist {
		return nil, fmt.Errorf("nearTransactionHash non-exist in aurora receipt: %v", unstructuredResult)
	}
	nearTransactionHashStr, ok := nearTransactionHash.(string)
	if !ok || nearTransactionHashStr == "" {
		return nil, fmt.Errorf("invalid nearTransactionHash: %v", unstructuredResult)
	}

	// convert map to json
	jsonString, err := json.Marshal(unstructuredResult)
	if err != nil {
		return nil, fmt.Errorf("can not convert json string with err: %v", err.Error())
	}
	err = json.Unmarshal(jsonString, &res)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json string with err: %v", err.Error())
	}

	if res == nil || res.Status == 0 {
		return nil, fmt.Errorf("transaction failed: %d", res.Status)
	}
	evmHeaderResult, err := evmcaller.GetEVMHeaderResult(res.BlockHash, auroraHosts, minEVMConfirmationBlocks, networkPrefix)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("Can not get EVM header result - Error: %+v", err)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	// check fork
	if evmHeaderResult.IsForked {
		metadataCommon.Logger.Log.Errorf("EVM Block hash %s is not in main chain", res.BlockHash.String())
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt,
			fmt.Errorf("EVM Block hash %s is not in main chain", res.BlockHash.String()))
	}

	// query near tx from above result
	ctx := context.Background()
	decodeTx, err := hex.DecodeString(nearTransactionHashStr[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid hex type with err: %v", err.Error())
	}
	nearTx := base58.Encode(decodeTx)
	tx, err := hash.NewCryptoHashFromBase58(nearTx)
	if err != nil {
		return nil, errors.New("Invalid transaction hash")
	}
	accountId := "incognito"
	hostActive := true
	for i, h := range nearHosts {
		rpcClient, err := nearclient.NewClient(h)
		if err != nil {
			if (i + 1) == len(nearHosts) {
				hostActive = false
			}
			continue
		}
		txStatus, err := rpcClient.TransactionStatus(ctx, tx, accountId)
		if err != nil {
			if (i + 1) == len(nearHosts) {
				hostActive = false
			}
			continue
		}
		// compare receipt root
		minedBlock, err := rpcClient.BlockDetails(ctx, block.BlockHash(txStatus.TransactionOutcome.BlockHash))
		if err != nil {
			if (i + 1) == len(nearHosts) {
				hostActive = false
			}
			continue
		}
		receiptRootEVMStr := base58.Encode(evmHeaderResult.Header.ReceiptHash.Bytes())
		if minedBlock.Header.ChunkReceiptsRoot.String() != receiptRootEVMStr {
			return nil, errors.New("Root on evm chain and native chain not match")
		}
		break
	}

	if !hostActive {
		return nil, fmt.Errorf("no endpoint in list %+v is active to verify or invalid near tx %s", nearHosts, tx.String())
	}

	return res, nil
}

func getAURORATransactionReceipt(url string, txHash common.Hash) (map[string]interface{}, error) {
	rpcClient := rpccaller.NewRPCClient()
	params := []interface{}{"0x" + txHash.String()}
	var res NormalResult
	err := rpcClient.RPCCall(
		"",
		url,
		"",
		"eth_getTransactionReceipt",
		params,
		&res,
	)
	if err != nil {
		return nil, err
	}
	if res.Result == nil {
		return nil, fmt.Errorf("tx aurora id %s non exist", params[0])
	}

	return res.Result.(map[string]interface{}), nil
}

func GetEVMInfoByMetadataType(metadataType int, networkID uint) ([]string, string, int, bool, error) {
	var hosts []string
	var networkPrefix string
	minConfirmationBlocks := metadataCommon.EVMConfirmationBlocks
	checkEVMHardFork := false
	isETHNetwork := false
	isBSCNetwork := false
	isPLGNetwork := false
	isFTMNetwork := false
	isAURORANetwork := false
	isAVAXNetwork := false

	if metadataType == metadataCommon.IssuingETHRequestMeta || metadataType == metadataCommon.IssuingPRVERC20RequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.ETHNetworkID) {
		isETHNetwork = true
	}
	if metadataType == metadataCommon.IssuingBSCRequestMeta || metadataType == metadataCommon.IssuingPRVBEP20RequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.BSCNetworkID) {
		isBSCNetwork = true
	}
	if metadataType == metadataCommon.IssuingPLGRequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.PLGNetworkID) {
		isPLGNetwork = true
	}
	if metadataType == metadataCommon.IssuingFantomRequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.FTMNetworkID) {
		isFTMNetwork = true
	}

	if metadataType == metadataCommon.IssuingAuroraRequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.AURORANetworkID) {
		isAURORANetwork = true
	}

	if metadataType == metadataCommon.IssuingAvaxRequestMeta || (metadataType == metadataCommon.IssuingUnifiedTokenRequestMeta && networkID == common.AVAXNetworkID) {
		isAVAXNetwork = true
	}

	if isBSCNetwork {
		evmParam := config.Param().BSCParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		networkPrefix = common.BSCPrefix

	} else if isETHNetwork {
		evmParam := config.Param().GethParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		// Ethereum network with default prefix (empty string)
		networkPrefix = ""
		checkEVMHardFork = true

	} else if isPLGNetwork {
		evmParam := config.Param().PLGParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		minConfirmationBlocks = metadataCommon.PLGConfirmationBlocks
		networkPrefix = common.PLGPrefix
		checkEVMHardFork = true

	} else if isFTMNetwork {
		evmParam := config.Param().FTMParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		minConfirmationBlocks = metadataCommon.FantomConfirmationBlocks
		networkPrefix = common.FTMPrefix
		checkEVMHardFork = true

	} else if isAURORANetwork {
		evmParam := config.Param().AURORAParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		minConfirmationBlocks = metadataCommon.AuroraConfirmationBlocks
		networkPrefix = common.AURORAPrefix
		checkEVMHardFork = true

	} else if isAVAXNetwork {
		evmParam := config.Param().AVAXParam
		evmParam.GetFromEnv()
		hosts = evmParam.Host

		minConfirmationBlocks = metadataCommon.AvaxConfirmationBlocks
		networkPrefix = common.AVAXPrefix
		checkEVMHardFork = true

	} else {
		return nil, "", 0, false, fmt.Errorf("Invalid metadata type for EVM shielding request metaType %v networkID %v", metadataType, networkID)
	}

	return hosts, networkPrefix, minConfirmationBlocks, checkEVMHardFork, nil
}

func GetWasmInfoByMetadataType(metadataType int) ([]string, string, int, string, error) {
	if metadataType == metadataCommon.IssuingNearRequestMeta {
		wasmParam := config.Param().NEARParam
		contractAddress := config.Param().NearContractAddressStr
		wasmParam.GetFromEnv()
		hosts := wasmParam.Host

		minConfirmationBlocks := metadataCommon.NearConfirmationBlocks
		networkPrefix := common.NEARPrefix

		return hosts, networkPrefix, minConfirmationBlocks, contractAddress, nil
	}
	return nil, "", 0, "", errors.New("Invalid meta data type")
}

func PickAndParseLogMapFromReceiptByContractAddr(
	constructedReceipt *types.Receipt,
	ethContractAddressStr string,
	eventName string) (map[string]interface{}, error) {
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		metadataCommon.Logger.Log.Errorf("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for i, log := range constructedReceipt.Logs {
		res, err := ParseEVMLogDataByEventName(log, ethContractAddressStr, eventName)
		if err != nil {
			metadataCommon.Logger.Log.Infof("WARNING: skip log #%d - %v", i, err)
		} else {
			return res, nil
		}
	}

	return nil, nil
}

func ParseEVMLogDataByEventName(log *types.Log, ethContractAddressStr string, name string) (map[string]interface{}, error) {
	abiIns, err := abi.JSON(strings.NewReader(common.AbiJson))
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(rCommon.HexToAddress(ethContractAddressStr).Bytes(), log.Address.Bytes()) {
		return nil, fmt.Errorf("contract address mismatch, expect %s", ethContractAddressStr)
	}
	event, exists := abiIns.Events[name]
	if !exists {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.UnexpectedError, fmt.Errorf("event %s not found in vault ABI", name))
	}
	evSigHash := event.Id()
	if !bytes.Equal(log.Topics[0][:], evSigHash[:]) {
		return nil, fmt.Errorf("event %s with topic %s not found in log %x", name, evSigHash.String(), log.Topics[0])
	}
	if len(log.Data) == 0 {
		return nil, fmt.Errorf("logData is empty")
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, name, log.Data); err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.UnexpectedError, err)
	}
	return dataMap, nil
}

func GetNetworkTypeByNetworkID(networkID uint8) (uint, error) {
	switch networkID {
	case
		common.ETHNetworkID,
		common.BSCNetworkID,
		common.PLGNetworkID,
		common.FTMNetworkID,
		common.AVAXNetworkID:
		return common.EVMNetworkType, nil
	case common.AURORANetworkID:
		return common.AURORANetworkID, nil
	default:
		return 0, errors.New("Not found networkID")
	}
}

func IsBurningConfirmMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BurningConfirmMeta, metadataCommon.BurningConfirmMetaV2:
		return true
	case metadataCommon.BurningConfirmForDepositToSCMeta, metadataCommon.BurningConfirmForDepositToSCMetaV2:
		return true
	case metadataCommon.BurningBSCConfirmMeta, metadataCommon.BurningPBSCConfirmForDepositToSCMeta:
		return true
	case metadataCommon.BurningFantomConfirmForDepositToSCMeta, metadataCommon.BurningFantomConfirmMeta:
		return true
	case metadataCommon.BurningPLGConfirmMeta, metadataCommon.BurningPLGConfirmForDepositToSCMeta:
		return true
	case metadataCommon.BurningPRVBEP20ConfirmMeta, metadataCommon.BurningPRVERC20ConfirmMeta:
		return true
	case metadataCommon.BurnForCallConfirmMeta:
		return true
	case metadataCommon.BurningAuroraConfirmMeta, metadataCommon.BurningAuroraConfirmForDepositToSCMeta:
		return true
	case metadataCommon.BurningAvaxConfirmMeta, metadataCommon.BurningAvaxConfirmForDepositToSCMeta:
		return true
	case metadataCommon.BurningNearConfirmMeta:
		return true
	case metadataCommon.BurningPRVRequestConfirmMeta:
		return true
	default:
		return false
	}
}
