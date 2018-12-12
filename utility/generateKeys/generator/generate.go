package generator

import (
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/wallet"
)

func GenerateAddress(seeds []string) ([]string, error) {
	var pubAddresses []string
	for _, seed := range seeds {
		keySet := (&cashec.KeySet{}).GenerateKey([]byte(seed))
		key := wallet.Key{KeySet: *keySet}
		pubAddress := key.Base58CheckSerialize(wallet.PaymentAddressType)
		pubAddresses = append(pubAddresses, pubAddress)
	}
	return pubAddresses, nil
}
