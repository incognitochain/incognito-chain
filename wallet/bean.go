package wallet

type KeySerializedData struct {
	PrivateKey     string `json:"PrivateKey"`
	PaymentAddress string `json:"PaymentAddress"`
	Pubkey         string `json:"Pubkey"`
	ReadonlyKey    string `json:"ReadonlyKey"`
}
