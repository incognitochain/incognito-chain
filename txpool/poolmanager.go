package txpool

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/pkg/errors"
)

type PoolManager struct {
	ShardTxsPool []TxPool
	// newRoleEChs  pubsub.EventChannel
	ps *pubsub.PubSubManager
}

func NewPoolManager(
	activeShards int,
	ps *pubsub.PubSubManager,
	ttl time.Duration,
) (
	*PoolManager,
	error,
) {
	res := &PoolManager{
		ps: ps,
	}
	for i := 0; i < activeShards; i++ {
		res.ShardTxsPool = append(res.ShardTxsPool, NewTxsPool(nil, make(chan metadata.Transaction, 128), ttl))
	}

	return res, nil
}

func (pm *PoolManager) Start(relayShards []byte) error {
	_, newRoleECh, err := pm.ps.RegisterNewSubscriber(pubsub.NodeRoleDetailTopic)
	if err != nil {
		Logger.Errorf("Register receieved error", err)
		return err
	}
	relayShardM := map[byte]interface{}{}
	for _, sID := range relayShards {
		if int(sID) < len(pm.ShardTxsPool) {
			go pm.ShardTxsPool[sID].Start()
			relayShardM[sID] = nil
		}
	}
	for msg := range newRoleECh {
		newRole, ok := msg.Value.(*pubsub.NodeRole)
		if ok {
			Logger.Infof(" Received new role %v %v\n", newRole.CID, newRole.Role)
			// Enable this for beacon full validation
			// if (newRole.CID == -1) && (newRole.Role == common.CommitteeRole) {
			// 	for _, txPool := range pm.ShardTxsPool {
			// 		if !txPool.IsRunning() {
			// 			txPool.Start()
			// 		}
			// 	}
			// }
			if (newRole.CID > -1) && (newRole.CID < len(pm.ShardTxsPool)) {
				if (newRole.Role == common.SyncingRole) || (newRole.Role == common.CommitteeRole) /*|| (newRole.Role == common.NodeModeRelay) */ {
					if !pm.ShardTxsPool[newRole.CID].IsRunning() {
						go pm.ShardTxsPool[newRole.CID].Start()
					}
				} else {
					if pm.ShardTxsPool[newRole.CID].IsRunning() {
						if _, ok := relayShardM[byte(newRole.CID)]; !ok {
							go pm.ShardTxsPool[newRole.CID].Stop()
						}
					}
				}
			}
		} else {
			Logger.Errorf("Cannot parse new role %v\n", *newRole)
		}
	}
	return nil
}

func (pm *PoolManager) GetShardTxsPool(shardID byte) (TxPool, error) {
	if int(shardID) >= len(pm.ShardTxsPool) {
		return nil, errors.Errorf("Can not get tx pool for this shard ID %v", shardID)
	}
	return pm.ShardTxsPool[shardID], nil
}

func (pm *PoolManager) GetMempoolInfo() MempoolInfo {
	res := &GetMempoolInfo{
		Size:          0,
		Bytes:         0,
		Usage:         0,
		MaxMempool:    0,
		MempoolMinFee: 1000000,
		MempoolMaxFee: 0,
		ListTxs:       []MempoolInfoTx{},
	}

	allTxsData := []TxsData{}
	for _, txPool := range pm.ShardTxsPool {
		if txPool.IsRunning() {
			allTxsData = append(allTxsData, txPool.snapshotPool())
		}
	}
	for _, txsData := range allTxsData {
		res.Size += len(txsData.TxByHash)
		for txHash, tx := range txsData.TxByHash {
			res.ListTxs = append(res.ListTxs, &GetMempoolInfoTx{
				TxID:     txHash,
				LockTime: tx.GetLockTime(),
			})
			if txInfo, ok := txsData.TxInfos[txHash]; ok {
				res.Bytes += txInfo.Size
				if res.MempoolMinFee > txInfo.Fee {
					res.MempoolMinFee = txInfo.Fee
				} else {
					if res.MempoolMaxFee < txInfo.Fee {
						res.MempoolMaxFee = txInfo.Fee
					}
				}
			}
		}
	}
	return res
}

func (pm *PoolManager) GetTransactionByHash(txHash string) (metadata.Transaction, error) {
	for _, txPool := range pm.ShardTxsPool {
		if txPool.IsRunning() {
			tx := txPool.getTxByHash(txHash)
			if tx != nil {
				return tx, nil
			}
		}
	}
	return nil, errors.Errorf("Transaction %v not found in mempool", txHash)
}

func (pm *PoolManager) RemoveTransactionInPool(txHash string) {
	for _, txPool := range pm.ShardTxsPool {
		if txPool.IsRunning() {
			txPool.RemoveTx(txHash)
		}
	}
}
