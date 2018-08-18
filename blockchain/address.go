package blockchain

type Address interface {
	// String returns the string encoding of the transaction output
	// destination.
	String() string

	EncodeAddress() string

	ScriptAddress() []byte

	IsForNet(*Params) bool
}
