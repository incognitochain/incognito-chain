package server

type PingArgs struct {
	RawAddress    string
	PublicKeyType string
	PublicKey     string
	SignData      string
}

func (ping *PingArgs) Init(RawAddress string, PublicKeyType string, PublicKey string, SignData string) {
	ping.PublicKey = PublicKey
	ping.PublicKeyType = PublicKeyType
	ping.SignData = SignData
	ping.RawAddress = RawAddress
}
