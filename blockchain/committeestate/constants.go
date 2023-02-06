package committeestate

var (
	MAX_SWAP_OR_ASSIGN_PERCENT_V2 = 6
	MAX_SLASH_PERCENT_V3          = 3
	MAX_SWAP_OUT_PERCENT_V3       = 8
	MAX_SWAP_IN_PERCENT_V3        = 8
	MAX_ASSIGN_PERCENT_V3         = 8
)

const (
	STATE_TEST_VERSION      = 0
	SELF_SWAP_SHARD_VERSION = 1
	STAKING_FLOW_V2         = 2
	STAKING_FLOW_V3         = 3
	STAKING_FLOW_V4         = 4
)

const (
	swapRuleTestVersion     = 0
	swapRuleSlashingVersion = 2
	swapRuleDCSVersion      = 3
)

const (
	unstakeRuleTestVersion     = 0
	unstakeRuleSlashingVersion = 1
	unstakeRuleDCSVersion      = 2
)

const (
	syncTerm      = 17280  //2 days with block time = 10s
	committeeTerm = 259200 //30 days with block time = 10s

)

const (
	ASSIGN_RULE_V1 = 1
	ASSIGN_RULE_V2 = 2
	ASSIGN_RULE_V3 = 3
)
