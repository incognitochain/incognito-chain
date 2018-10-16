package jsonresult

type GetNetworkInfoResult struct {
	Version         string                   `json:"version"`
	SubVersion      string                   `json:"SubVersion"`
	ProtocolVersion string                   `json:"ProtocolVersion"`
	NetworkActive   bool                     `json:"NetworkActive"`
	Connections     int                      `json:"Connections"`
	Networks        []map[string]interface{} `json:"Networks"`
	LocalAddresses  []string                 `json:"LocalAddresses"`
	RelayFee        int                      `json:"RelayFee"`
	IncrementalFee  uint64                   `json:"IncrementalFee"`
	Warnings        string                   `json:"Warnings"`
}
