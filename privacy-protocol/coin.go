package privacy

import (
	"math/big"
)

type SerialNumber []byte   //33 bytes
type CoinCommitment []byte //67 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes
type SNDerivator []byte


// Coin represents a coin
type Coin struct {
	PublicKey      EllipticPoint      	// 33 bytes
	SNDerivator    big.Int   		// 32 bytes
	CoinCommitment EllipticPoint 	// 33 bytes
	Randomness     big.Int         	// Random for coin commitment
	Value          big.Int          	// 32 bytes
	Info           []byte
	MerkleRoot			MerkleRoot
}

//CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
//func (coin *Coin) CommitAll() {
//	//var values [PCM_CAPACITY-1][]byte
//	values := [PCM_CAPACITY][]byte{coin.PublicKey, coin.Value, coin.SNDerivator, coin.Randomness}
//	fmt.Printf("coin info: %v\n", values)
//	coin.CoinCommitment = append(coin.CoinCommitment, FULL)
//	coin.CoinCommitment = append(coin.CoinCommitment, Pcm.Commit(values)...)
//}
//
//// CommitPublicKey commits a public key's coin
//func (coin *Coin) CommitPublicKey() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{coin.PublicKey, nil, nil, coin.Randomness}
//
//
//	var commitment []byte
//	commitment = append(commitment, PK)
//	commitment = append(commitment, Pcm.Commit(values)...)
//	return commitment
//}
//
//// CommitValue commits a value's coin
//func (coin *Coin) CommitValue() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{nil, coin.Value, nil, coin.Randomness}
//
//	var commitment []byte
//	commitment = append(commitment, VALUE)
//	commitment = append(commitment, Pcm.Commit(values)...)
//	return commitment
//}
//
//// CommitSNDerivator commits a serial number's coin
//func (coin *Coin) CommitSNDerivator() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{nil, nil, coin.SNDerivator, coin.Randomness}
//
//	var commitment []byte
//	commitment = append(commitment, SND)
//	commitment = append(commitment, Pcm.Commit(values)...)
//	return commitment
//}

// SpendingCoin represents a list of coins to be spent corresponding to spending key
type SpendingCoin struct{
	Coins []Coin
	SpendingKey SpendingKey
}

