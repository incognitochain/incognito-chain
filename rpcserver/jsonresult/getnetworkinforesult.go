package jsonresult

type GetNetworkInfoResult struct {
	Version         int                      `json:"version"`
	SubVersion      string                   `json:"SubVersion"`
	ProtocolVersion string                   `json:"ProtocolVersion"`
	LocalServices   string                   `json:"LocalServices"`
	LocalRelay      bool                     `json:"LocalRelay"`
	TimeOffset      int                      `json:"TimeOffset"`
	NetworkActive   bool                     `json:"NetworkActive"`
	Connections     int                      `json:"Connections"`
	Networks        []map[string]interface{} `json:"Networks"`
	LocalAddresses  []string                 `json:"LocalAddresses"`
	RelayFee        int                      `json:"RelayFee"`
	IncrementalFee  int                      `json:"IncrementalFee"`
	Warnings        string                   `json:"Warnings"`
}
