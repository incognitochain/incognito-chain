package privacy

import "fmt"

type SerialNumber []byte   //32 bytes
type CoinCommitment []byte //3 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes4

// Check type of commitments
const (
	PK_CM    = byte(0x00)
	VALUE_CM = byte(0x01)
	SN_CM    = byte(0x02)
	FULL_CM  = byte(0x03)
)

// Coin represents a coin
type Coin struct {
	PublicKey      PublicKey
	SerialNumber   SerialNumber
	CoinCommitment CoinCommitment
	R              Random // Random for coin commitment
	Value, Info    []byte
}

// CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
func (coin *Coin) CommitAll() {
	var values [CM_CAPACITY][]byte
	values = [CM_CAPACITY][]byte{coin.PublicKey, coin.Value, coin.SerialNumber, coin.R}
	fmt.Printf("cin info: %v\n", values)
	coin.CoinCommitment = append(coin.CoinCommitment, FULL_CM)
	coin.CoinCommitment = append(coin.CoinCommitment, Pcm.Commit(values)...)
}

// CommitPublicKey commits a public key's coin
func (coin *Coin) CommitPublicKey() []byte {
	var values [CM_CAPACITY][]byte
	values = [CM_CAPACITY][]byte{coin.PublicKey, nil, nil, nil}

	var commitment []byte
	commitment = append(commitment, PK_CM)
	commitment = append(commitment, Pcm.Commit(values)...)
	return commitment
}

// CommitValue commits a value's coin
func (coin *Coin) CommitValue() []byte {
	var values [CM_CAPACITY][]byte
	values = [CM_CAPACITY][]byte{nil, coin.Value, nil, nil}

	var commitment []byte
	commitment = append(commitment, VALUE_CM)
	commitment = append(commitment, Pcm.Commit(values)...)
	return commitment
}

// CommitSerialNumber commits a serial number's coin
func (coin *Coin) CommitSerialNumber() []byte {
	var values [CM_CAPACITY][]byte
	values = [CM_CAPACITY][]byte{nil, nil, coin.SerialNumber, nil}

	var commitment []byte
	commitment = append(commitment, SN_CM)
	commitment = append(commitment, Pcm.Commit(values)...)
	return commitment
}
