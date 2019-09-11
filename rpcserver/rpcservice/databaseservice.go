package rpcservice

import (
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"strconv"
)

type DatabaseService struct {
	DB database.DatabaseInterface
}

func (dbService DatabaseService) CheckETHHashIssued(data map[string]interface{}) (bool, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	uniqETHTx := append(blockHash[:], []byte(strconv.Itoa(int(txIdx)))...)

	issued, err := dbService.DB.IsETHTxHashIssued(uniqETHTx)
	return issued, err
}

func (dbService DatabaseService) GetAllBridgeTokens() ([]byte, error) {
	allBridgeTokensBytes, err := dbService.DB.GetAllBridgeTokens()
	return allBridgeTokensBytes, err
}

func (dbService DatabaseService) GetBridgeReqWithStatus(txID string) (byte, error) {
	txIDHash, err := common.Hash{}.NewHashFromStr(txID)
	if err != nil {
		return byte(0), err
	}

	status, err := dbService.DB.GetBridgeReqWithStatus(*txIDHash)
	return status, err
}
