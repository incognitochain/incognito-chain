package main

//func main() {
//	//param, _ := strconv.Atoi(os.Args[1])
//	//dcbStart := 100000000
//	govStart := 10000000
//	//	burnStart := 1000000
//
//	for i := govStart; ; i++ {
//		// dcb: 100000000
//		// gov: 10000000
//		// burn: 1000000
//		//param = 10000000
//		burnPubKeyE := privacy.PedCom.G[0].Hash(uint64(i))
//		burnPubKey := burnPubKeyE.Compress()
//		burnKey := wallet.KeyWallet{
//			KeySet: cashec.KeySet{
//				PaymentAddress: privacy.PaymentAddress{
//					Pk: burnPubKey,
//				},
//			},
//		}
//		burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
//		shardID := common.GetShardIDFromLastByte(burnPubKey[len(burnPubKey)-1])
//		fmt.Println("shardID:", shardID)
//		if shardID == 0 {
//			fmt.Printf("Burn payment address : %s %d\n", burnPaymentAddress, i)
//			goto Out
//		}
//
//		/*keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
//		fmt.Println("======================================")
//		fmt.Println(keyWalletBurningAdd.KeySet.PaymentAddress.Pk)
//		fmt.Println("======================================")*/
//	}
//Out:
//	fmt.Println("Finished")
//}
