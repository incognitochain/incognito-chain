package privacy

import (
	"encoding/json"
	"math/big"
)

type SerialNumber []byte   //33 bytes
type CoinCommitment []byte //67 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes
type SNDerivator []byte

// Coin represents a coin
type Coin struct {
	PublicKey      *EllipticPoint
	CoinCommitment *EllipticPoint
	SNDerivator    *big.Int
	Randomness     *big.Int
	Value          uint64
	Info           []byte
}

func (coin Coin) CoinToJson() []byte {
	var tmpnote struct {
		Value        uint64
		Rho, R, Memo []byte
	}
	tmpnote.Value = note.Value
	tmpnote.Rho = note.Rho
	tmpnote.R = note.R
	tmpnote.Memo = note.Memo

	noteJson, err := json.Marshal(&tmpnote)
	if err != nil {
		return []byte{}
	}
	// fmt.Printf("%s", noteJson)
	return noteJson
}

func ParseJsonToNote(jsonnote []byte) (*Note, error) {
	var note Note
	err := json.Unmarshal(jsonnote, &note)
	if err != nil {
		return nil, err
	}
	// fmt.Println(note)
	return &note, nil
}

func (coin *Coin) Encrypt() []byte{
	/**** Generate symmetric key of AES cryptosystem,
				it is used for encryption coin details ****/
	var point EllipticPoint
	point.Randomize()
	symKey := point.X.Bytes()

	/**** Encrypt coin details using symKey ****/
	// Convert coin details from struct to bytes array

	// Encrypt symKey using Transmission key's receiver with ElGamal cryptosystem
	return nil

}

//CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
//func (coin *Coin) CommitAll() {
//	//var values [PCM_CAPACITY-1][]byte
//	values := [PCM_CAPACITY][]byte{coin.PublicKey, coin.Value, coin.SNDerivator, coin.Randomness}
//	fmt.Printf("coin info: %v\n", values)
//	coin.CoinCommitment = append(coin.CoinCommitment, FULL)
//	coin.CoinCommitment = append(coin.CoinCommitment, PedCom.Commit(values)...)
//}

//// CommitPublicKey commits a public key's coin
//func (coin *Coin) CommitPublicKey() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{coin.PublicKey, nil, nil, coin.Randomness}
//
//
//	var commitment []byte
//	commitment = append(commitment, PK)
//	commitment = append(commitment, PedCom.Commit(values)...)
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
//	commitment = append(commitment, PedCom.Commit(values)...)
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
//	commitment = append(commitment, PedCom.Commit(values)...)
//	return commitment
//}

// UnspentCoin represents a list of coins to be spent corresponding to spending key
//type UnspentCoin struct {
//	SpendingKey SpendingKey
//	UnspentCoinList map[Coin]big.Int
//}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	BlockHeight *big.Int
	CoinDetails *Coin
}

type OutputCoin struct{
	CoinDetails   *Coin
	CoinDetailsEncrypted []byte
}

