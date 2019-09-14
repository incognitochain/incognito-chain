package rpcservice
import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/privacy"
)

type TxMemPoolService struct {
	TxMemPool * mempool.TxPool
}

func (txMemPoolService TxMemPoolService) GetPoolCandidate() map[common.Hash]string {
	return txMemPoolService.TxMemPool.GetClonedPoolCandidate()
}


func (txMemPoolService TxMemPoolService) FilterMemPoolOutcoinsToSpent(outCoins []*privacy.OutputCoin) ([]*privacy.OutputCoin, error) {
	remainOutputCoins := make([]*privacy.OutputCoin, 0)

	for _, outCoin := range outCoins {
		if txMemPoolService.TxMemPool.ValidateSerialNumberHashH(outCoin.CoinDetails.GetSerialNumber().Compress()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	return remainOutputCoins, nil
}



