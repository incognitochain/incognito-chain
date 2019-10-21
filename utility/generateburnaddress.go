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
		burnPubKeyE := privacy.RandomPoint()
		burnPubKey := burnPubKeyE.ToBytesS()
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
/*Special payment address : 15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs 1   // burning address
======================================
[99 183 246 161 68 172 228 222 153 9 172 39 208 245 167 79 11 2 114 65 241 69 85 40 193 104 199 79 70 4 53 0]
======================================
Special payment address : 15onyX2Ux1Che2mh4e4ap3D8zy68AetHkCY5jDeQQJhmKs1nT5vz6p62qEJ 21
======================================
[3 10 141 47 195 123 109 123 58 104 169 33 117 142 250 119 151 254 2 51 116 226 9 129 244 218 44 154 155 167 176 0]
======================================
Special payment address : 15pcY2K1Rs7t7HE65P3tkEziR6HjGH1h8UB8mhB9mHdmryEqww7tQBHzHWa 248
======================================
[219 231 166 157 154 232 141 223 243 199 38 25 202 237 18 255 126 111 109 146 166 92 228 73 5 48 186 134 98 55 231 0]
======================================
*/
//
