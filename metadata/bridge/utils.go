package bridge

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	"github.com/eteu-technologies/near-api-go/pkg/types/hash"
	"math/big"
	"strconv"
	"strings"

	nearclient "github.com/eteu-technologies/near-api-go/pkg/client"
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
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func IsBridgeAggMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeAggModifyRewardReserveMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenRequestMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenResponseMeta:
		return true
	case metadataCommon.IssuingUnifiedRewardResponseMeta:
		return true
	case metadataCommon.BurningUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BurningUnifiedTokenResponseMeta:
		return true
	case metadataCommon.BridgeAggAddTokenMeta:
		return true
	default:
		return false
	}
}

type Vault struct {
	statedb.BridgeAggConvertedTokenState
	RewardReserve uint64 `json:"RewardReserve"`
	Decimal       uint   `json:"Decimal"`
	IsPaused      bool   `json:"IsPaused"`
}

func (v *Vault) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
		Decimal       uint                                  `json:"Decimal"`
		IsPaused      bool                                  `json:"IsPaused"`
	}{
		State:         &v.BridgeAggConvertedTokenState,
		RewardReserve: v.RewardReserve,
		Decimal:       v.Decimal,
		IsPaused:      v.IsPaused,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
		Decimal       uint                                  `json:"Decimal"`
		IsPaused      bool                                  `json:"IsPaused"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	v.RewardReserve = temp.RewardReserve
	if temp.State != nil {
		v.BridgeAggConvertedTokenState = *temp.State
	}
	v.Decimal = temp.Decimal
	v.IsPaused = temp.IsPaused
	return nil
}

func IsBridgeTxHashUsedInBlock(uniqTx []byte, uniqTxsUsed [][]byte) bool {
	for _, item := range uniqTxsUsed {
		if bytes.Equal(uniqTx, item) {
			return true
		}
	}
	return false
}

func getShardIDFromPaymentAddress(addressStr string) (byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return byte(0), fmt.Errorf("Payment address' public key must not be empty")
	}
	// calculate shard ID
	lastByte := keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func ExtractIssueEVMData(
	stateDB *statedb.StateDB, shardID byte, listTxUsed [][]byte, contractAddress string, prefix string,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error),
	txReceipt *types.Receipt, blockHash rCommon.Hash, txIndex uint,
) (*big.Int, byte, string, []byte, []byte, error) {
	if txReceipt == nil {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: bridge tx receipt is null")
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqTx := append(blockHash[:], []byte(strconv.Itoa(int(txIndex)))...)
	isUsedInBlock := IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: already issued for the hash in current block: ", uniqTx)
	}
	isIssued, err := isTxHashIssued(stateDB, uniqTx)
	if err != nil {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an issue occured while checking the bridge tx hash is issued or not: ", err)
	}
	if isIssued {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: already issued for the hash in previous blocks: ", uniqTx)
	}

	logMap, err := PickAndParseLogMapFromReceiptByContractAddr(txReceipt, contractAddress, "Deposit")
	if err != nil {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an error occurred while parsing log map from receipt: ", err)
	}
	if logMap == nil {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not find log map out from receipt")
	}

	logMapBytes, _ := json.Marshal(logMap)
	metadataCommon.Logger.Log.Warn("INFO: eth logMap json - ", string(logMapBytes))

	// the token might be ETH/ERC20 BNB/BEP20
	tokenAddr, ok := logMap["token"].(rCommon.Address)
	if !ok {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse evm token id from log map.")
	}
	token := append([]byte(prefix), tokenAddr.Bytes()...)

	addressStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse incognito address from bridge log map.")
	}
	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse amount from bridge log map.")
	}
	receivingShardID, err := getShardIDFromPaymentAddress(addressStr)
	if err != nil {
		return nil, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an error occurred while getting shard id from payment address: ", err)
	}
	return amt, receivingShardID, addressStr, token, uniqTx, nil
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

	receivingShardID, err := getShardIDFromPaymentAddress(incognitoAddress)
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
		return fmt.Errorf("WARNING: an error occurred while checking it can process for token pair on the current block or not: ", err)
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
		return fmt.Errorf("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: ", err)
	}
	if !isValid {
		return fmt.Errorf("WARNING: pair of incognito token id & bridge's id is invalid with previous blocks")
	}
	return nil
}

func FindExternalTokenID(stateDB *statedb.StateDB, incTokenID common.Hash, prefix string, metaType int) ([]byte, error) {
	// Convert to external tokenID
	tokenID, err := findExternalTokenID(stateDB, &incTokenID)
	if err != nil {
		return nil, err
	}

	if len(tokenID) < common.ExternalBridgeTokenLength {
		return nil, errors.New("invalid external token id")
	}

	prefixLen := len(prefix)
	if (prefixLen > 0 && !bytes.Equal([]byte(prefix), tokenID[:prefixLen])) || len(tokenID) != (common.ExternalBridgeTokenLength+prefixLen) {
		return nil, errors.New(fmt.Sprintf("metadata type %v with invalid external tokenID %v", metaType, tokenID))
	}
	return tokenID, nil
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
		latestBlock, err := rpcClient.BlockDetails(ctx, block.FinalityFinal())
		if err != nil {
			continue
		}
		if latestBlock.Header.Height < minedBlock.Header.Height+uint64(minWasmConfirmationBlocks) {
			return "", "", 0, "", errors.New("The shield transaction is not finalized")
		}

		// detect shield event
		for _, receiptOutCome := range txStatus.ReceiptsOutcome {
			if receiptOutCome.Outcome.ExecutorID != contractId {
				continue
			}
			if len(receiptOutCome.Outcome.Logs) == 0 {
				break
			}
			events := strings.Split(receiptOutCome.Outcome.Logs[0], " ")
			if len(events) != 3 {
				break
			}
			amount, err := strconv.ParseUint(events[2], 10, 64)
			if err != nil {
				break
			}
			return events[0], events[1], amount, receiptOutCome.Outcome.ExecutorID, nil
		}
	}

	return "", "", 0, "", errors.New("The endpoints are not response or set or invalid transaction")
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
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		metadataCommon.Logger.Log.Errorf("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(rCommon.HexToAddress(ethContractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		metadataCommon.Logger.Log.Errorf("WARNING: logData is empty.")
		return nil, nil
	}
	return ParseEVMLogDataByEventName(logData, eventName)
}

func ParseEVMLogDataByEventName(data []byte, name string) (map[string]interface{}, error) {
	abiIns, err := abi.JSON(strings.NewReader(common.AbiJson))
	if err != nil {
		return nil, err
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, name, data); err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.UnexpectedError, err)
	}
	return dataMap, nil
}

func GetNetworkTypeByNetworkID(networkID uint) (uint, error) {
	switch networkID {
	case common.ETHNetworkID, common.BSCNetworkID, common.PLGNetworkID, common.FTMNetworkID:
		return common.EVMNetworkType, nil
	default:
		return 0, errors.New("Not found networkID")
	}
}
