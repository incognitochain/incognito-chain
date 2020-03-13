package coin

type Coin interface {
	Init() Coin
	GetVersion() uint8
}
