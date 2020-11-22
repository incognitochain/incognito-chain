package txpool

import (
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
) (
	*PoolManager,
	error,
) {
	res := &PoolManager{}
	for i := 0; i < activeShards; i++ {
		txV := &TxsVerifier{}
		res.ShardTxsPool = append(res.ShardTxsPool, NewTxsPool(txV, make(chan metadata.Transaction)))
	}

	return res, nil
}

func (pm *PoolManager) Start() error {
	_, newRoleECh, err := pm.ps.RegisterNewSubscriber(pubsub.NodeRoleDetailTopic)
	if err != nil {
		return err
	}
	for msg := range newRoleECh {
		newRole, ok := msg.Value.(*pubsub.NodeRole)
		if ok {
			if (newRole.CID == -1) && (newRole.Role == common.CommitteeRole) {
				for _, txPool := range pm.ShardTxsPool {
					txPool.Start()
				}
			}
			if (newRole.CID > -1) && (newRole.CID < len(pm.ShardTxsPool)) {
				if (newRole.Role == common.SyncingRole) || (newRole.Role == common.CommitteeRole) /*|| (newRole.Role == common.NodeModeRelay) */ {
					pm.ShardTxsPool[newRole.CID].Start()
				} else {
					pm.ShardTxsPool[newRole.CID].Stop()
				}
			}
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
