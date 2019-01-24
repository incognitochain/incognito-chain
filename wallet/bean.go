package wallet

type KeySerializedData struct {
	PrivateKey     string `json:"PrivateKey"`
	PaymentAddress string `json:"PaymentAddress"`
	Pubkey         string `json:"Pubkey"` // in hex encode string
	ReadonlyKey    string `json:"ReadonlyKey"`
}
