package zkp

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/wallet"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init() starting payment v1 logger for test !")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()


type CoinObject struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    string
	SerialNumber   string
	Randomness     string
	Value          uint64
	Info           string
}

func createSamplePlainCoinV1(privKey privacy.PrivateKey, pubKey *operation.Point, amount uint64, msg []byte) (*coin.PlainCoinV1, error) {
	c := new(coin.PlainCoinV1).Init()

	c.SetValue(amount)
	c.SetInfo(msg)
	c.SetPublicKey(pubKey)
	c.SetSNDerivator(operation.RandomScalar())
	c.SetRandomness(operation.RandomScalar())

	//Derive serial number from snDerivator
	c.SetKeyImage(new(operation.Point).Derive(privacy.PedCom.G[0], new(operation.Scalar).FromBytesS(privKey), c.GetSNDerivator()))

	//Create commitment
	err := c.CommitAll()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func ParseCoinObjectToStruct(coinObjects []CoinObject) ([]*coin.PlainCoinV1, uint64) {
	coins := make([]*coin.PlainCoinV1, len(coinObjects))
	sumValue := uint64(0)

	for i := 0; i < len(coins); i++ {

		publicKey, _, _ := base58.Base58Check{}.Decode(coinObjects[i].PublicKey)
		publicKeyPoint := new(operation.Point)
		publicKeyPoint.FromBytesS(publicKey)

		coinCommitment, _, _ := base58.Base58Check{}.Decode(coinObjects[i].CoinCommitment)
		coinCommitmentPoint := new(operation.Point)
		coinCommitmentPoint.FromBytesS(coinCommitment)

		snd, _, _ := base58.Base58Check{}.Decode(coinObjects[i].SNDerivator)
		sndBN := new(operation.Scalar).FromBytesS(snd)

		serialNumber, _, _ := base58.Base58Check{}.Decode(coinObjects[i].CoinCommitment)
		serialNumberPoint := new(operation.Point)
		serialNumberPoint.FromBytesS(serialNumber)

		randomness, _, _ := base58.Base58Check{}.Decode(coinObjects[i].Randomness)
		randomnessBN := new(operation.Scalar).FromBytesS(randomness)

		coins[i] = new(coin.PlainCoinV1).Init()
		coins[i].SetPublicKey(publicKeyPoint)
		coins[i].SetCommitment(coinCommitmentPoint)
		coins[i].SetSNDerivator(sndBN)
		coins[i].SetKeyImage(serialNumberPoint)
		coins[i].SetRandomness(randomnessBN)
		coins[i].SetValue(coinObjects[i].Value)

		sumValue += coinObjects[i].Value

	}

	return coins, sumValue
}

func TestPaymentProofToBytes(t *testing.T) {
	//witness := new(PaymentWitness)
	witnessParam := new(PaymentWitnessParam)

	keyWallet, _ := wallet.Base58CheckDeserialize("112t8rnXHD9s2MXSXigMyMtKdGFtSJmhA9cCBN34Fj55ox3cJVL6Fykv8uNWkDagL56RnA4XybQKNRrNXinrDDfKZmq9Y4LR18NscSrc9inc")
	_ = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	senderKeyBN := new(operation.Scalar).FromBytesS(keyWallet.KeySet.PrivateKey)
	senderPKPoint := new(operation.Point)
	senderPKPoint.FromBytesS(keyWallet.KeySet.PaymentAddress.Pk)

	coinStrs := []CoinObject{
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "16rDxiXDg9AhyC3o3XiBQZAtg4P2x1ER9umyspRFC4AUWGj9LnK",
			SNDerivator:    "12bf2zoKdYw8c8BT3YMKNaVkLppoQqEkLtSCymEa6EK65FSowV7",
			SerialNumber:   "17ioQJTBFV8HGK6TYQn9mWfdT8Z7wRCMyn9GjFYhMx6dP8UrnJp",
			Randomness:     "13CyLqj6BErihknHV7AWqHdAodLAwRwGuqkEdDqFb5chS5uhLN",
			Value:          13063917525,
			Info:           "13PMpZ4",
		},
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17pb83j2YcrB8WLr1jPNGsT6Qgo3dEan7U6NsJwR2QAY1PcmXWa",
			SNDerivator:    "12M48gjxpPUkb69ieMLc9EhBDcCerTbhtHnAgdoaEToXUYhFiCb",
			SerialNumber:   "18fJDPSbjLnTCxk2QUrEig4Ai5kWbPYediD1KhKinKm142smQVs",
			Randomness:     "12cyHe5MyGLDGeKDZSknP2DEny48mNMC49Rd8CHhdiBCh35bnTs",
			Value:          4230769230,
			Info:           "13PMpZ4",
		},
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17tcbagBHAjG8fr2RGLc3FjAJ5Mkbqitdv6KCtQWLiBydHoHpRP",
			SNDerivator:    "1SXpgdZKqwENjSYgLhaam6PS3u5CciYMHuwyt1ipr5SUQQMYGn",
			SerialNumber:   "15Mnm5Do1Np3eoPdGRvECJb8mjhHLgvDYoWxNQgAxXTLCUh2MYa",
			Randomness:     "12Br463SeHFafpPEntE1L81S6vk5HShUtgE7tiCfPzr1aWiSZMU",
			Value:          13395348837,
			Info:           "13PMpZ4",
		},
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "16ep85MLtTigBiwPf1b6bRcKJ9NJfVazxjC1GCzEsqKK1J5927t",
			SNDerivator:    "1EXomopZG5uUDbC9fRWyg4UnvboqY3PQnmjF1srRyRUDFXaUzd",
			SerialNumber:   "16PpZUXgsQntxB8Js6yoPRzyZEiQyQTGtSXUEYnvV1uEbv7wPdw",
			Randomness:     "12uWji6kLpUo8Xg5AJx1odDkeP2ZQ7g9p2tnSMwLJfvgDDmkWVS",
			Value:          466999090249,
			Info:           "13PMpZ4",
		},
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17Pw2SmoW4zXojM8HHHMEpX5k3SjKL8UAeGXBDKjqpJBJKtHSkf",
			SNDerivator:    "122qVAS24X5AjWdWsiX54npCN7WDrAyDk4VmGSbFNexWcofzNXa",
			SerialNumber:   "18LrSQofiFy9HuiCbdPJZp7nFKg9z6xNiN1EoeRVWdCiMf6Yyrm",
			Randomness:     "12XdvDLJ2UKASYX2wCSEKvda3xYrJKeUaP4XXmQ3f6f5hA399pg",
			Value:          13423728813,
			Info:           "13PMpZ4",
		},
		{
			PublicKey:      "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "188C79Y2jJmKNxxuGN56S5rSXDYAZqP7erMEebmui74DaS7qf4V",
			SNDerivator:    "1xD4hPppKFkwTkUK2GkR6VVEszhF94KZEFqpqvSynqUePGKnrh",
			SerialNumber:   "16TKsDv351rbn64bw4CTnwfSd626oJ6bYRYjUQqYP2dmyRqYpXn",
			Randomness:     "1WPLdUVWt6566hpjENoNkSmukPSyYBjbWwrv2nyQx49DByPR36",
			Value:          6285714285,
			Info:           "13PMpZ4",
		},
	}

	inputCoins, sumValue := ParseCoinObjectToStruct(coinStrs)

	keyWalletReceiver, _ := wallet.Base58CheckDeserialize("112t8rnXHD9s2MXSXigMyMtKdGFtSJmhA9cCBN34Fj55ox3cJVL6Fykv8uNWkDagL56RnA4XybQKNRrNXinrDDfKZmq9Y4LR18NscSrc9inc")
	_ = keyWalletReceiver.KeySet.InitFromPrivateKey(&keyWalletReceiver.KeySet.PrivateKey)
	//receiverKeyBN := new(big.Int).SetBytes(keyWalletReceiver.KeySet.PrivateKey)
	receiverPublicKey := keyWalletReceiver.KeySet.PaymentAddress.Pk
	receiverPublicKeyPoint := new(operation.Point)
	receiverPublicKeyPoint.FromBytesS(receiverPublicKey)

	amountTransfer := uint64(1000000000)

	outputCoins := make([]*coin.CoinV1, 2)
	outputCoins[0] = new(coin.CoinV1)
	outputCoins[0].Init()
	outputCoins[0].CoinDetails.SetValue(uint64(amountTransfer))
	outputCoins[0].CoinDetails.SetPublicKey(receiverPublicKeyPoint)
	outputCoins[0].CoinDetails.SetSNDerivator(operation.RandomScalar())

	changeAmount := sumValue - amountTransfer

	outputCoins[1] = new(coin.CoinV1)
	outputCoins[1].Init()
	outputCoins[1].CoinDetails.SetValue(changeAmount)
	outputCoins[1].CoinDetails.SetPublicKey(senderPKPoint)
	outputCoins[1].CoinDetails.SetSNDerivator(operation.RandomScalar())

	//HasPrivacy              bool
	//PrivateKey              *big.Int
	//InputCoins              []*coin.PlainCoinV1
	//OutputCoins             []*privacy.OutputCoin
	//PublicKeyLastByteSender byte
	//Commitments             []*privacy.Point
	//CommitmentIndices       []uint64
	//MyCommitmentIndices     []uint64
	//Fee                     uint64
	var inputCoinsGeneric []coin.PlainCoin
	for _,c := range inputCoins{
		inputCoinsGeneric = append(inputCoinsGeneric, c)
	}
	witnessParam.HasPrivacy = true
	witnessParam.PrivateKey = senderKeyBN
	witnessParam.InputCoins = inputCoinsGeneric
	witnessParam.OutputCoins = outputCoins
	witnessParam.PublicKeyLastByteSender = keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]

	//witness.Init()

}
