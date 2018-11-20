package privacy

import "fmt"

type SerialNumber []byte   //33 bytes
type CoinCommitment []byte //67 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes


// Coin represents a coin
type Coin struct {
	PublicKey      PublicKey      // 33 bytes
	SerialNumber   SerialNumber   // 32 bytes
	CoinCommitment CoinCommitment // 34 bytes
	Randomness     Random         // Random for coin commitment
	Value          Value          // 32 bytes
	Info           []byte
}

// CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
func (coin *Coin) CommitAll() {
	var values [CM_CAPACITY-1][]byte
	values = [CM_CAPACITY-1][]byte{coin.PublicKey, coin.Value, coin.SerialNumber, coin.Randomness}
	fmt.Printf("coin info: %v\n", values)
	coin.CoinCommitment = append(coin.CoinCommitment, FULL_CM)
	coin.CoinCommitment = append(coin.CoinCommitment, Elcm.Commit(values)...)
}

// CommitPublicKey commits a public key's coin
func (coin *Coin) CommitPublicKey() []byte {
	var values [CM_CAPACITY-1][]byte
	values = [CM_CAPACITY-1][]byte{coin.PublicKey, nil, nil, coin.Randomness}


	var commitment []byte
	commitment = append(commitment, PK_CM)
	commitment = append(commitment, Elcm.Commit(values)...)
	return commitment
}

// CommitValue commits a value's coin
func (coin *Coin) CommitValue() []byte {
	var values [CM_CAPACITY-1][]byte
	values = [CM_CAPACITY-1][]byte{nil, coin.Value, nil, coin.Randomness}

	var commitment []byte
	commitment = append(commitment, VALUE_CM)
	commitment = append(commitment, Elcm.Commit(values)...)
	return commitment
}

// CommitSerialNumber commits a serial number's coin
func (coin *Coin) CommitSerialNumber() []byte {
	var values [CM_CAPACITY-1][]byte
	values = [CM_CAPACITY-1][]byte{nil, nil, coin.SerialNumber, coin.Randomness}

	var commitment []byte
	commitment = append(commitment, SN_CM)
	commitment = append(commitment, Elcm.Commit(values)...)
	return commitment
}
