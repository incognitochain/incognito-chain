package wallet

type KeySerializedData struct {
	PrivateKey     string `json:"PrivateKey"`
	PaymentAddress string `json:"PaymentAddress"`
	ReadonlyKey    string `json:"ReadonlyKey"`
}
