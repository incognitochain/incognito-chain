package rpcservice
import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
)

type TxMemPoolService struct {
	TxMemPool * mempool.TxPool
}

func (txMemPoolService TxMemPoolService) GetPoolCandidate() map[common.Hash]string {
	return txMemPoolService.TxMemPool.GetClonedPoolCandidate()
}


