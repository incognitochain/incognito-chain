package coin

type Coin interface {
	Init() *Coin
	GetVersion() uint8
	Bytes() []byte
	SetBytes([]byte) error
}
