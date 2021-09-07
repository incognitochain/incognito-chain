package v2utils

type MintNftStatus struct {
	NftID       string `json:"NftID"`
	Status      string `json:"Status"`
	BurntAmount uint64 `json:"BurntAmount"`
}

type ContributionStatus struct {
	Status                  string `json:"Status"`
	Token0ID                string `json:"Token0ID"`
	Token0ContributedAmount uint64 `json:"Token0ContributedAmount"`
	Token0ReturnedAmount    uint64 `json:"Token0ReturnedAmount"`
	Token1ID                string `json:"Token1ID"`
	Token1ContributedAmount uint64 `json:"Token1ContributedAmount"`
	Token1ReturnedAmount    uint64 `json:"Token1ReturnedAmount"`
}

type WithdrawStatus struct {
	Status       string `json:"Status"`
	Token0ID     string `json:"Token0ID"`
	Token0Amount uint64 `json:"Token0Amount"`
	Token1ID     string `json:"Token1ID"`
	Token1Amount uint64 `json:"Token1Amount"`
}

type StakingStatus struct {
	Status        string `json:"Status"`
	NftID         string `json:"NftID"`
	StakingPoolID string `json:"StakingPoolID"`
	Liquidity     uint64 `json:"Liquidity"`
}
