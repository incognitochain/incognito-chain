package tx_ver2

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func reconstructRingCAV2(txEnv metadata.ValidationEnviroment, sumOutputsWithFee, sumOutputAssetTags *privacy.Point, numOfOutputs *privacy.Scalar, transactionStateDB *statedb.StateDB) (*mlsag.Ring, [][]*privacy.CoinV2, error) {
	txSigPubKey := new(SigPubKey)
	if err := txSigPubKey.SetBytes(txEnv.SigPubKey()); err != nil {
		errStr := fmt.Sprintf("Error when parsing bytes of txSigPubKey %v", err)
		return nil, nil, utils.NewTransactionErr(utils.UnexpectedError, errors.New(errStr))
	}
	indexes := txSigPubKey.Indexes
	n := len(indexes)
	if n == 0 {
		return nil, nil, errors.New("Cannot get ring from Indexes: Indexes is empty")
	}

	m := len(indexes[0])
	OTAData := txEnv.DBData()
	if m*n != len(OTAData) {
		return nil, nil, fmt.Errorf("Cached OTA data not match with indexes")
	}
	ring := make([][]*privacy.Point, n)
	coinsInRing := make([][]*privacy.CoinV2, n)
	
	for i := 0; i < n; i++ {
		sumCommitment := new(privacy.Point).Identity()
		sumCommitment.Sub(sumCommitment, sumOutputsWithFee)
		sumAssetTags := new(privacy.Point).Identity()
		sumAssetTags.Sub(sumAssetTags, sumOutputAssetTags)
		row := make([]*privacy.Point, m+2)
		coinsInRing[i] = make([]*privacy.CoinV2, m)

		for j := 0; j < m; j++ {
			randomCoinBytes := OTAData[i*m+j]
			randomCoin := new(privacy.CoinV2)
			if err := randomCoin.SetBytes(randomCoinBytes); err != nil {
				utils.Logger.Log.Errorf("Set coin Byte error %v ", err)
				return nil, nil, err
			}
			row[j] = randomCoin.GetPublicKey()
			sumCommitment.Add(sumCommitment, randomCoin.GetCommitment())
			temp := new(privacy.Point).ScalarMult(randomCoin.GetAssetTag(), numOfOutputs)
			sumAssetTags.Add(sumAssetTags, temp)

			coinsInRing[i][j] = randomCoin
		}

		row[m] = new(privacy.Point).Set(sumAssetTags)
		row[m+1] = new(privacy.Point).Set(sumCommitment)
		ring[i] = row
	}
	return mlsag.NewRing(ring), coinsInRing, nil
}
