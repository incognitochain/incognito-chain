package privacy

type SerialNumber []byte   //33 bytes
type CoinCommitment []byte //67 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes
type SNDerivator []byte


// Coin represents a coin
type Coin struct {
	PublicKey      PublicKey      // 33 bytes
	SNDerivator    SNDerivator   // 32 bytes
	CoinCommitment CoinCommitment // 34 bytes
	Randomness     Random         // Random for coin commitment
	Value          Value          // 32 bytes
	Info           []byte
}

// CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
//func (coin *Coin) CommitAll() {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{coin.PublicKey, coin.Value, coin.SNDerivator, coin.Randomness}
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
//// CommitSerialNumber commits a serial number's coin
//func (coin *Coin) CommitSerialNumber() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{nil, nil, coin.SNDerivator, coin.Randomness}
//
//	var commitment []byte
//	commitment = append(commitment, SND)
//	commitment = append(commitment, Pcm.Commit(values)...)
//	return commitment
//}
