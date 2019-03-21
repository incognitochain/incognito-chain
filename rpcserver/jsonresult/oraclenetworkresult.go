package jsonresult

type OracleNetworkResult struct {
	OraclePubKeys          []string `json:"OraclePubKeys"`
	WrongTimesAllowed      uint8    `json:"WrongTimesAllowed"`
	Quorum                 uint8    `json:"Quorum"`
	AcceptableErrorMargin  uint32   `json:"AcceptableErrorMargin"`
	UpdateFrequency        uint32   `json:"UpdateFrequency"`
	OracleRewardMultiplier uint8    `json:"OracleRewardMultiplier"`
}
