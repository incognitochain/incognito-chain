package debugtool

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"
)

//OutCoinKey is used to retrieve output coins via RPC.
//
//Payment address must always be present. Other fields are optional
type OutCoinKey struct {
	paymentAddress string
	otaKey         string
	readonlyKey	   string
}

func (outCoinKey *OutCoinKey) SetOTAKey(otaKey string){
	outCoinKey.otaKey = otaKey
}

func (outCoinKey *OutCoinKey) SetPaymentAddress(paymentAddress string){
	outCoinKey.paymentAddress = paymentAddress
}

func (outCoinKey *OutCoinKey) SetReadonlyKey(readonlyKey string){
	outCoinKey.readonlyKey = readonlyKey
}

func NewOutCoinKey(paymentAddress, otaKey, readonlyKey string) *OutCoinKey {
	return &OutCoinKey{paymentAddress: paymentAddress, otaKey: otaKey, readonlyKey: readonlyKey}
}

func NewOutCoinKeyFromPrivateKey(privateKey string) (*OutCoinKey, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil{
		return nil, err
	}
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	otaSecretKey := keyWallet.Base58CheckSerialize(wallet.OTAKeyType)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)

	return &OutCoinKey{paymentAddress: paymentAddStr, otaKey: otaSecretKey, readonlyKey: viewingKeyStr}, nil
}

func ParseCoinFromJsonResponse(b []byte) ([]jsonresult.ICoinInfo, []*big.Int, error){
	response, err := ParseResponse(b)
	fmt.Println(err)
	if err != nil{
		return nil, nil, err
	}

	var tmp jsonresult.ListOutputCoins
	err = json.Unmarshal(response.Result, &tmp)
	if err != nil {
		return nil, nil, err
	}

	resultOutCoins := make([]jsonresult.ICoinInfo, 0)
	listOutputCoins := tmp.Outputs
	for _, value := range listOutputCoins {
		for _, outCoin := range value {
			out, _, err := jsonresult.NewCoinFromJsonOutCoin(outCoin)
			if err != nil {
				return nil, nil, err
			}

			resultOutCoins = append(resultOutCoins, out)
		}
	}

	return resultOutCoins, nil, nil
}

func GetListDecryptedCoins(privateKey string, listOutputCoins []jsonresult.ICoinInfo) ([]coin.PlainCoin, []string, error){
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, nil, err
	}

	listDecyptedOutCoins := make([]coin.PlainCoin, 0)
	listKeyImages := make([]string, 0)
	for _, outCoin := range listOutputCoins {
		if outCoin.GetVersion() == 1{
			if outCoin.IsEncrypted() {
				tmpCoin, ok := outCoin.(*coin.CoinV1)
				if !ok {
					return nil, nil, errors.New("invalid CoinV1")
				}

				decryptedCoin, err := tmpCoin.Decrypt(&keyWallet.KeySet)
				if err != nil{
					return nil, nil, err
				}

				keyImage, err := decryptedCoin.ParseKeyImageWithPrivateKey(keyWallet.KeySet.PrivateKey)
				if err != nil{
					return nil, nil, err
				}
				decryptedCoin.SetKeyImage(keyImage)

				keyImageString := base64.RawStdEncoding.EncodeToString(keyImage.ToBytesS())


				listKeyImages = append(listKeyImages, keyImageString)
				listDecyptedOutCoins = append(listDecyptedOutCoins, decryptedCoin)
			}else{
				tmpPlainCoinV1, ok := outCoin.(*coin.PlainCoinV1)
				if !ok {
					return nil, nil, errors.New("invalid PlaincoinV1")
				}

				keyImage, err := tmpPlainCoinV1.ParseKeyImageWithPrivateKey(keyWallet.KeySet.PrivateKey)
				if err != nil{
					return nil, nil, err
				}
				tmpPlainCoinV1.SetKeyImage(keyImage)

				keyImageString := base58.Base58Check{}.Encode(keyImage.ToBytesS(), common.ZeroByte)

				listKeyImages = append(listKeyImages, keyImageString)
				listDecyptedOutCoins = append(listDecyptedOutCoins, tmpPlainCoinV1)
			}
		}else if outCoin.GetVersion() == 2 {
			tmpCoinV2, ok := outCoin.(*coin.CoinV2)
			if !ok {
				return nil, nil, errors.New("invalid CoinV2")
			}
			decryptedCoin, err := tmpCoinV2.Decrypt(&keyWallet.KeySet)
			if err != nil {
				return nil, nil, err
			}

			keyImage := decryptedCoin.GetKeyImage()
			keyImageString := base58.Base58Check{}.Encode(keyImage.ToBytesS(), common.ZeroByte)

			listKeyImages = append(listKeyImages, keyImageString)
			listDecyptedOutCoins = append(listDecyptedOutCoins, decryptedCoin)
		}
	}

	return listDecyptedOutCoins, listKeyImages, nil
}

