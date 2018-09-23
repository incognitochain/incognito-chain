package wallet

type KeySerializedData struct {
	PrivateKey  string `json:"PrivateKey"`
	PublicKey   string `json:"PublicKey"`
	ReadonlyKey string `json:"ReadonlyKey"`
}
