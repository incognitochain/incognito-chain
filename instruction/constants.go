package instruction

const (
	SWAP_SHARD_ACTION          = "swap_shard"
	SWAP_ACTION                = "swap"
	RANDOM_ACTION              = "random"
	STAKE_ACTION               = "stake"
	ASSIGN_ACTION              = "assign"
	ASSIGN_SYNC_ACTION         = "assign_sync"
	STOP_AUTO_STAKE_ACTION     = "stopautostake"
	SET_ACTION                 = "set"
	RETURN_ACTION              = "return"
	UNSTAKE_ACTION             = "unstake"
	SHARD_INST                 = "shard"
	BEACON_INST                = "beacon"
	SPLITTER                   = ","
	TRUE                       = "true"
	FALSE                      = "false"
	FINISH_SYNC_ACTION         = "finish_sync"
	ACCEPT_BLOCK_REWARD_ACTION = "accept_block_reward"
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
