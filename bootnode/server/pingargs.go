package server

type PingArgs struct {
	rawAddress string
	publicKey  string
	signData   string
}

func (ping *PingArgs) Init(RawAddress string, PublicKey string, SignData string) {
	ping.publicKey = PublicKey
	ping.signData = SignData
	ping.rawAddress = RawAddress
}
