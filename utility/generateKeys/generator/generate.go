package generator

import (
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common/base58"
)

func GenerateAddress(seeds []string) ([]string, error) {
	var pubAddresses []string
	for _, seed := range seeds {
		keySet := (&cashec.KeySet{}).GenerateKey([]byte(seed))
		pubAddress := base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, byte(0x00))
		pubAddresses = append(pubAddresses, pubAddress)
	}
	return pubAddresses, nil
}
