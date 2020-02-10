package rpcservice

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math/big"
)

type DatabaseService struct {
	DB incdb.Database
}

//func (dbService DatabaseService) CheckETHHashIssued(data map[string]interface{}) (bool, error) {
//	blockHashParam, ok := data["BlockHash"].(string)
//	if !ok {
//		return false, errors.New("Block hash param is invalid")
//	}
//	blockHash := rCommon.HexToHash(blockHashParam)
//
//	txIdxParam, ok := data["TxIndex"].(float64)
//	if !ok {
//		return false, errors.New("Tx index param is invalid")
//	}
//	txIdx := uint(txIdxParam)
//	uniqETHTx := append(blockHash[:], []byte(strconv.Itoa(int(txIdx)))...)
//
//	issued, err := rawdb.IsETHTxHashIssued(dbService.DB, uniqETHTx)
//	return issued, err
//}

//func (dbService DatabaseService) GetBurningConfirm(txID common.Hash) (uint64, error) {
//	return rawdb.GetBurningConfirm(dbService.DB, txID)
//}

//func (dbService DatabaseService) ListSerialNumbers(tokenID common.Hash, shardID byte) (map[string]uint64, error) {
//	return rawdb.ListSerialNumber(dbService.DB, tokenID, shardID)
//}

func (dbService DatabaseService) ListSNDerivator(tokenID common.Hash) ([]big.Int, error) {
	resultInBytes, err := rawdb.ListSNDerivator(dbService.DB, tokenID)
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
	return rawdb.ListCommitment(dbService.DB, tokenID, shardID)
}

func (dbService DatabaseService) ListCommitmentIndices(tokenID common.Hash, shardID byte) (map[uint64]string, error) {
	return rawdb.ListCommitmentIndices(dbService.DB, tokenID, shardID)
}

func (dbService DatabaseService) ListCommitmentsV2(tokenID common.Hash, shardID byte, stateDB *statedb.StateDB) (map[string]uint64, error) {
	return statedb.ListCommitment(stateDB, tokenID, shardID)
}

func (dbService DatabaseService) ListCommitmentIndicesV2(tokenID common.Hash, shardID byte, stateDB *statedb.StateDB) (map[uint64]string, error) {
	return statedb.ListCommitmentIndices(stateDB, tokenID, shardID)
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
		ok, _ := rawdb.HasSerialNumber(dbService.DB, tokenID, serialNumber, shardIDSender)
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

		ok, err := rawdb.HasSNDerivator(dbService.DB, tokenID, common.AddPaddingBigInt(new(big.Int).SetBytes(snderivator), common.BigIntSize))
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

func (dbService DatabaseService) GetProducersBlackList(beaconHeight uint64) (map[string]uint8, error) {
	return rawdb.GetProducersBlackList(dbService.DB, beaconHeight)
}

func (dbService DatabaseService) GetPDEStatus(pdePrefix []byte, pdeSuffix []byte) (byte, error) {
	return rawdb.GetPDEStatus(dbService.DB, pdePrefix, pdeSuffix)
}

func (dbService DatabaseService) GetPDEContributionStatus(pdePrefix []byte, pdeSuffix []byte) (*metadata.PDEContributionStatus, error) {
	pdeStatusContentBytes, err := rawdb.GetPDEContributionStatus(dbService.DB, pdePrefix, pdeSuffix)
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
