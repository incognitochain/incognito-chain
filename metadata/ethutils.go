package metadata

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/pkg/errors"
)

func VerifyProofAndParseEVMReceipt(
	blockHash eCommon.Hash,
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
		Logger.log.Errorf("Can not get EVM header result - Error: %+v", err)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	// check fork
	if evmHeaderResult.IsForked {
		Logger.log.Errorf("EVM Block hash %s is not in main chain", blockHash.String())
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt,
			fmt.Errorf("EVM Block hash %s is not in main chain", blockHash.String()))
	}

	// check min confirmation blocks
	if !evmHeaderResult.IsFinalized {
		Logger.log.Errorf("EVM block hash %s is not enough confirmation blocks %v", blockHash.String(), minEVMConfirmationBlocks)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt,
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
		Logger.log.Warn(errMsg)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	// if iReq.Type == IssuingETHRequestMeta || iReq.Type == IssuingPRVERC20RequestMeta || iReq.Type == IssuingPLGRequestMeta {
	if checkEVMHarkFork {
		if len(val) == 0 {
			return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("the encoded receipt is empty"))
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
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	if constructedReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("The constructedReceipt's status is not success"))
	}

	return constructedReceipt, nil
}

func PickAndParseLogMapFromReceiptByContractAddr(
	constructedReceipt *types.Receipt,
	ethContractAddressStr string,
	eventName string) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		Logger.log.Errorf("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(eCommon.HexToAddress(ethContractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		Logger.log.Errorf("WARNING: logData is empty.")
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
		return nil, NewMetadataTxError(UnexpectedError, err)
	}
	return dataMap, nil
}

func GetEVMInfoByMetadataType(metadataType int) ([]string, string, int, error) {
	var hosts []string
	var networkPrefix string
	minConfirmationBlocks := EVMConfirmationBlocks

	switch metadataType {
	case IssuingETHRequestMeta, IssuingPRVERC20RequestMeta:
		{
			evmParam := config.Param().GethParam
			evmParam.GetFromEnv()
			hosts = evmParam.Host

			// Ethereum network with default prefix (empty string)
			networkPrefix = ""
		}
	case IssuingBSCRequestMeta, IssuingPRVBEP20RequestMeta:
		{
			evmParam := config.Param().BSCParam
			evmParam.GetFromEnv()
			hosts = evmParam.Host

			networkPrefix = common.BSCPrefix
		}
	case IssuingPLGRequestMeta:
		{
			evmParam := config.Param().PLGParam
			evmParam.GetFromEnv()
			hosts = evmParam.Host

			minConfirmationBlocks = PLGConfirmationBlocks
			networkPrefix = common.PLGPrefix
		}
	default:
		{
			return nil, "", 0, fmt.Errorf("Invalid metadata tyep for EVM shielding request %v", metadataType)
		}
	}

	return hosts, networkPrefix, minConfirmationBlocks, nil
}
