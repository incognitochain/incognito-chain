package wallet

type KeySerializedData struct {
	PrivateKey  string `json:"PrivateKey"`
	PublicKey   string `json:"PaymentAddress"`
	ReadonlyKey string `json:"ReadonlyKey"`
}
