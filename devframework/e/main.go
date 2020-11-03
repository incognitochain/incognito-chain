package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	k, err := wallet.Base58CheckDeserialize("12RuyEgAMWaRU5cFFMbPYKAWbM7k5UPy9UmHjtoU1iBgGnEAwzUQfxtGTx1yfTxad8X8ZAxoCg7bCZhYBmoj8guAm2Q483K5z6NQzKd")
	if err != nil {
		panic(err)
	}
	pma := k.KeySet.PaymentAddress.Pk

	fmt.Println(base58.Base58Check{}.Encode(k.KeySet.PaymentAddress.Pk, common.ZeroByte))
	// c := incognitokey.CommitteePublicKey{
	// 	IncPubKey: pma,
	// 	MiningPubKey: map[string][]byte{}{"bls":},
	// }

	r1, _, _ := base58.Base58Check{}.Decode("12T4uhqGf7t5RthAJn8G1W29rCASQhg2QsFzBbcCTNxoGbGok7f")
	c, err := incognitokey.NewCommitteeKeyFromSeed(r1, pma)
	if err != nil {
		panic(err)
	}
	fmt.Println(c.ToBase58())
}
