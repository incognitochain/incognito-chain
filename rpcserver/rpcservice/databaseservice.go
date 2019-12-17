package rpcservice

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
)

type DatabaseService struct {
	DB *database.DatabaseInterface
}

func (dbService DatabaseService) CheckETHHashIssued(data map[string]interface{}) (bool, error) {
	blockHashParam, ok := data["BlockHash"].(string)
	if !ok {
		return false, errors.New("Block hash param is invalid")
	}
	blockHash := rCommon.HexToHash(blockHashParam)

	txIdxParam, ok := data["TxIndex"].(float64)
	if !ok {
		return false, errors.New("Tx index param is invalid")
	}
	txIdx := uint(txIdxParam)
	uniqETHTx := append(blockHash[:], []byte(strconv.Itoa(int(txIdx)))...)

	issued, err := (*dbService.DB).IsETHTxHashIssued(uniqETHTx)
	return issued, err
}

func (dbService DatabaseService) GetAllBridgeTokens() ([]byte, error) {
	allBridgeTokensBytes, err := (*dbService.DB).GetAllBridgeTokens()
	return allBridgeTokensBytes, err
}

func (dbService DatabaseService) GetBridgeReqWithStatus(txID string) (byte, error) {
	txIDHash, err := common.Hash{}.NewHashFromStr(txID)
	if err != nil {
		return byte(0), err
	}

	status, err := (*dbService.DB).GetBridgeReqWithStatus(*txIDHash)
	return status, err
}

func (dbService DatabaseService) GetBurningConfirm(txID common.Hash) (uint64, error) {
	return (*dbService.DB).GetBurningConfirm(txID)
}

func (dbService DatabaseService) ListSerialNumbers(tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	return (*dbService.DB).ListSerialNumber(tokenID, shardID)
}

func (dbService DatabaseService) ListSNDerivator(tokenID common.Hash) ([]big.Int, error) {
	resultInBytes, err := (*dbService.DB).ListSNDerivator(tokenID)
	if err != nil {
		return nil, err
	}

	result := []big.Int{}
	for _, v := range resultInBytes {
		result = append(result, *(new(big.Int).SetBytes(v)))
	}

	return result, nil
}

func (dbService DatabaseService) ListCommitments(tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	return (*dbService.DB).ListCommitment(tokenID, shardID)
}

func (dbService DatabaseService) ListCommitmentIndices(tokenID common.Hash, shardID byte) (map[uint64]string, error) {
	return (*dbService.DB).ListCommitmentIndices(tokenID, shardID)
}

func (dbService DatabaseService) HasSerialNumbers(paymentAddressStr string, serialNumbersStr []interface{}, tokenID common.Hash) ([]bool, error) {
	_, shardIDSender, err := GetKeySetFromPaymentAddressParam(paymentAddressStr)
	if err != nil {
		return nil, err
	}

	result := make([]bool, 0)
	for _, item := range serialNumbersStr {
		itemStr, okParam := item.(string)
		if !okParam {
			return nil, errors.New("Invalid serial number param")
		}
		serialNumber, _, err := base58.Base58Check{}.Decode(itemStr)
		if err != nil {
			return nil, errors.New("Invalid serial number param")
		}
		ok, _ := (*dbService.DB).HasSerialNumber(tokenID, serialNumber, shardIDSender)
		if ok {
			// serial number in db
			result = append(result, true)
		} else {
			// serial number not in db
			result = append(result, false)
		}
	}

	return result, nil
}

func (dbService DatabaseService) HasSnDerivators(paymentAddressStr string, snDerivatorStr []interface{}, tokenID common.Hash) ([]bool, error) {
	_, _, err := GetKeySetFromPaymentAddressParam(paymentAddressStr)
	if err != nil {
		return nil, err
	}

	result := make([]bool, 0)
	for _, item := range snDerivatorStr {
		itemStr, okParam := item.(string)
		if !okParam {
			return nil, errors.New("Invalid serial number derivator param")
		}
		snderivator, _, err := base58.Base58Check{}.Decode(itemStr)
		if err != nil {
			return nil, errors.New("Invalid serial number derivator param")
		}

		ok, err := (*dbService.DB).HasSNDerivator(tokenID, common.AddPaddingBigInt(new(big.Int).SetBytes(snderivator), common.BigIntSize))
		if ok && err == nil {
			// SnD in db
			result = append(result, true)
		} else {
			// SnD not in db
			result = append(result, false)
		}
	}
	return result, nil
}

func (dbService DatabaseService) ListRewardAmount() map[string]map[common.Hash]uint64 {
	return (*dbService.DB).ListCommitteeReward()
}

func (dbService DatabaseService) GetProducersBlackList(beaconHeight uint64) (map[string]uint8, error) {
	return (*dbService.DB).GetProducersBlackList(beaconHeight)
}

func (dbService DatabaseService) GetPDEStatus(pdePrefix []byte, pdeSuffix []byte) (byte, error) {
	return (*dbService.DB).GetPDEStatus(pdePrefix, pdeSuffix)
}

func (dbService DatabaseService) GetPDEContributionStatus(pdePrefix []byte, pdeSuffix []byte) (*metadata.PDEContributionStatus, error) {
	pdeStatusContentBytes, err := (*dbService.DB).GetPDEContributionStatus(pdePrefix, pdeSuffix)
	if err != nil {
		return nil, err
	}
	if len(pdeStatusContentBytes) == 0 {
		return nil, nil
	}
	var contributionStatus metadata.PDEContributionStatus
	err = json.Unmarshal(pdeStatusContentBytes, &contributionStatus)
	if err != nil {
		return nil, err
	}
	return &contributionStatus, nil
}
