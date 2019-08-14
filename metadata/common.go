package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
)

func ParseMetadata(meta interface{}) (Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md Metadata
	switch int(mtTemp["Type"].(float64)) {
	case IssuingRequestMeta:
		md = &IssuingRequest{}

	case IssuingResponseMeta:
		md = &IssuingResponse{}

	case ContractingRequestMeta:
		md = &ContractingRequest{}

	case IssuingETHRequestMeta:
		md = &IssuingETHRequest{}

	case IssuingETHResponseMeta:
		md = &IssuingETHResponse{}

	case BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}

	case BurningRequestMeta:
		md = &BurningRequest{}

	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}
	case ReturnStakingMeta:
		md = &ReturnStakingMetadata{}

	case WithDrawRewardRequestMeta:
		md = &WithDrawRewardRequest{}
	case WithDrawRewardResponseMeta:
		md = &WithDrawRewardResponse{}
	default:
		fmt.Printf("[db] parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

func GetETHHeader(
	ethBlockHash rCommon.Hash,
) (*types.Header, error) {
	rpcClient := rpccaller.NewRPCClient()
	params := []interface{}{ethBlockHash, false}
	var getBlockByNumberRes GetBlockByNumberRes
	err := rpcClient.RPCCall(
		EthereumLightNodeProtocol,
		EthereumLightNodeHost,
		EthereumLightNodePort,
		"eth_getBlockByHash",
		params,
		&getBlockByNumberRes,
	)
	if err != nil {
		return nil, err
	}
	if getBlockByNumberRes.RPCError != nil {
		fmt.Printf("WARNING: an error occured during calling eth_getBlockByHash: %s", getBlockByNumberRes.RPCError.Message)
		return nil, nil
	}
	return getBlockByNumberRes.Result, nil
}

func PickAndParseLogMapFromReceipt(constructedReceipt *types.Receipt) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		Logger.log.Debug("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(rCommon.HexToAddress(common.EthContractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		return nil, nil
	}
	return ParseETHLogData(logData)
}
