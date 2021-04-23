package instruction

const (
	SWAP_SHARD_ACTION              = "swap_shard"
	SWAP_ACTION                    = "swap"
	RANDOM_ACTION                  = "random"
	STAKE_ACTION                   = "stake"
	ASSIGN_ACTION                  = "assign"
	ASSIGN_SYNC_ACTION             = "assign_sync"
	STOP_AUTO_STAKE_ACTION         = "stopautostake"
	SET_ACTION                     = "set"
	RETURN_ACTION                  = "return"
	UNSTAKE_ACTION                 = "unstake"
	SHARD_INST                     = "shard"
	BEACON_INST                    = "beacon"
	SPLITTER                       = ","
	TRUE                           = "true"
	FALSE                          = "false"
	FINISH_SYNC_ACTION             = "finish_sync"
	ACCEPT_BLOCK_REWARD_V3_ACTION  = "accept_block_reward"
	SHARD_RECEIVE_REWARD_V3_ACTION = "shard_subset_reward"

	SHARD_RECEIVE_REWARD_V1_ACTION = 43
	ACCEPT_BLOCK_REWARD_V1_ACTION  = 37
	SHARD_REWARD_INST              = "shardRewardInst"
)

const (
	BEACON_CHAIN_ID = -1
)

//Swap Instruction Sub Type
const (
	SWAP_BY_END_EPOCH = iota
	SWAP_BY_SLASHING
	SWAP_BY_INCREASE_COMMITTEES_SIZE
	SWAP_BY_DECREASE_COMMITTEES_SIZE
)
