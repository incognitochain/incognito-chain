package v2utils

type MintNftStatus struct {
	NftID       string `json:"NftID"`
	Status      byte   `json:"Status"`
	BurntAmount uint64 `json:"BurntAmount"`
}

type ContributionStatus struct {
	Status                  byte   `json:"Status"`
	Token0ID                string `json:"Token0ID"`
	Token0ContributedAmount uint64 `json:"Token0ContributedAmount"`
	Token0ReturnedAmount    uint64 `json:"Token0ReturnedAmount"`
	Token1ID                string `json:"Token1ID"`
	Token1ContributedAmount uint64 `json:"Token1ContributedAmount"`
	Token1ReturnedAmount    uint64 `json:"Token1ReturnedAmount"`
	PoolPairID              string `json:"PoolPairID"`
}

type WithdrawStatus struct {
	Status       byte   `json:"Status"`
	Token0ID     string `json:"Token0ID"`
	Token0Amount uint64 `json:"Token0Amount"`
	Token1ID     string `json:"Token1ID"`
	Token1Amount uint64 `json:"Token1Amount"`
}

type StakingStatus struct {
	Status        byte   `json:"Status"`
	NftID         string `json:"NftID"`
	StakingPoolID string `json:"StakingPoolID"`
	Liquidity     uint64 `json:"Liquidity"`
}

type UnstakingStatus struct {
	Status        byte   `json:"Status"`
	NftID         string `json:"NftID"`
	StakingPoolID string `json:"StakingPoolID"`
	Liquidity     uint64 `json:"Liquidity"`
}
