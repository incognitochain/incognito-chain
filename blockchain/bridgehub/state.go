package bridgehub

type BridgeHubState struct {
	// TODO: staking asset is PRV or others?
	StakingInfos map[string]uint64 // bridgePubKey : amount PRV stake

	// bridgePubKey only belongs one Bridge
	BridgeInfos map[string]BridgeInfo // BridgeID : BridgeInfo

	TokenPrices map[string]uint64 // pTokenID: price * 1e6
}

type BridgeInfo struct {
	ExtChainID    string
	BriValidators []string          // array of bridgePubKey
	BriPubKey     string            // Public key of TSS that used to validate sig from validators by TSS
	PTokenAmounts map[string]uint64 // pTokenID : amount

	// info of previous bridge validators that are used to slashing if they haven't completed their remain tasks
	PrevBriValidators []string // array of bridgePubKey
	PrevBriPubKey     string   // Public key of TSS that used to validate sig from validators by TSS
}

func NewBrigdeHubState() *BridgeHubState {
	return &BridgeHubState{}
}
