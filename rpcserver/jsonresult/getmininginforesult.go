package jsonresult

type GetMiningInfoResult struct {
	ShardHeight         uint64 `json:"ShardHeight"`
	BeaconHeight        uint64 `json:"BeaconHeight"`
	CurrentShardBlockTx int    `json:"CurrentShardBlockTx"`
	PoolSize            int    `json:"PoolSize"`
	Chain               string `json:"Chain"`
	IsCommittee         bool   `json:"IsCommittee"`
	ShardID             int    `json:"ShardID"`
	Role                string `json:"Role"`
}
