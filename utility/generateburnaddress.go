package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	temp := 0
	for i := 0; ; i++ {
		burnPubKeyE := privacy.PedCom.G[0].Hash(int64(i))
		burnPubKey := burnPubKeyE.Compress()
		if burnPubKey[len(burnPubKey)-1] == 0 {
			burnKey := wallet.KeyWallet{
				KeySet: incognitokey.KeySet{
					PaymentAddress: privacy.PaymentAddress{
						Pk: burnPubKey,
					},
				},
			}
			burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
			fmt.Printf("Special payment address : %s %d\n", burnPaymentAddress, i)
			keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(burnPaymentAddress)
			fmt.Println("======================================")
			fmt.Println(keyWalletBurningAdd.KeySet.PaymentAddress.Pk)
			fmt.Println("======================================")
			temp += 1
			if temp == 3 {
				goto Out
			}
		}
	}
Out:
	fmt.Println("Finished")
}

// result
/*Special payment address : 1NHpWKZYCLQeGKSSsJewsA8p3nsPoAZbmEmtsuBqd6yU7KJnzJZVt39b7AgP 897
======================================
[3 216 102 168 155 36 135 253 243 218 209 109 131 105 134 20 37 96 171 207 227 0 236 35 158 98 24 188 68 32 191 99 0]
======================================
Special payment address : 1NHoFQ3Nr8fQm3ZLk2ACSgZXjVH6JobpuV65RD3QAEEGe76KknMQhGbc4g8P 934
======================================
[2 139 228 54 166 17 174 144 173 226 219 133 190 139 124 163 248 126 160 134 203 144 245 151 192 33 123 164 122 174 191 255 0]
======================================
Special payment address : 1NHp2EKw7ALdXUzBfoRJvKrBBM9nkejyDcHVPvUjDcWRyG22dHHyiBKQGL1c 1103
======================================
[3 88 81 5 73 221 37 60 3 235 186 153 134 46 49 204 20 7 56 45 188 104 251 173 161 65 36 61 137 49 132 248 0]
======================================*/
//
