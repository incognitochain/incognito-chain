package tx_ver2

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/stretchr/testify/assert"

	"encoding/json"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPrivacyV2TxToken(t *testing.T) {
	var err error
	var numOfPrivateKeys int
	var numOfInputs int
	var dummyPrivateKeys []*privacy.PrivateKey
	var keySets []*incognitokey.KeySet
	var paymentInfo []*privacy.PaymentInfo
	var pastCoins, pastTokenCoins []privacy.Coin
	var txParams *tx_generic.TxTokenParams
	var msgCipherText []byte
	var boolParams map[string]bool
	tokenID := &common.Hash{56}
	tx2 := &TxToken{}

	Convey("Tx Token Main Test", t, func() {
		numOfPrivateKeys = RandInt()%(maxPrivateKeys-minPrivateKeys+1) + minPrivateKeys
		numOfInputs = 2
		Convey("prepare keys", func() {
			dummyPrivateKeys, keySets, paymentInfo = preparePaymentKeys(numOfPrivateKeys)
			boolParams = make(map[string]bool)
		})

		Convey("create & store PRV UTXOs", func() {
			pastCoins = make([]privacy.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i := range pastCoins {
				tempCoin, err := privacy.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)
				tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				pastCoins[i] = tempCoin
			}
			// store a bunch of sample OTA coins in PRV
			So(storeCoins(dummyDB, pastCoins, 0, common.PRVCoinID), ShouldBeNil)
		})

		Convey("create & store Token UTXOs", func() {
			// now store the token
			err := statedb.StorePrivacyToken(dummyDB, *tokenID, "NameName", "SYM", statedb.InitToken, false, uint64(100000), []byte{}, common.Hash{66})
			So(err, ShouldBeNil)

			pastTokenCoins = make([]privacy.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i := range pastTokenCoins {
				tempCoin, _, err := privacy.NewCoinCA(paymentInfo[i%len(dummyPrivateKeys)], tokenID)
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)
				tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				pastTokenCoins[i] = tempCoin
			}
			// store a bunch of sample OTA coins in PRV
			So(storeCoins(dummyDB, pastTokenCoins, 0, common.ConfidentialAssetID), ShouldBeNil)
		})

		Convey("create salary transaction", func() {
			testTxTokenV2Salary(tokenID, dummyPrivateKeys, keySets, paymentInfo, dummyDB)
		})

		Convey("transfer token", func() {
			Convey("create TX with params", func() {
				txParams, _ = getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				exists := statedb.PrivacyTokenIDExisted(dummyDB, *tokenID)
				So(exists, ShouldBeTrue)
				err = tx2.Init(txParams)
				So(err, ShouldBeNil)
			})

			Convey("should verify & accept transaction", func() {
				msgCipherText = []byte("doing a transfer")
				So(bytes.Equal(msgCipherText, tx2.GetTxNormal().GetProof().GetOutputCoins()[0].GetInfo()), ShouldBeTrue)

				isValidSanity, err := tx2.ValidateSanityData(nil, nil, nil, 0)
				So(isValidSanity, ShouldBeTrue)
				So(err, ShouldBeNil)

				boolParams["hasPrivacy"] = hasPrivacyForToken
				// before the token init tx is written into db, this should not pass
				isValidTxItself, err := tx2.ValidateTxByItself(boolParams, dummyDB, nil, nil, shardID, nil, nil)
				So(isValidTxItself, ShouldBeTrue)
				So(err, ShouldBeNil)
				err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
				So(err, ShouldBeNil)
			})

			Convey("should reject tampered TXs", func() {
				testTxTokenV2JsonMarshaler(tx2, 10, dummyDB)
				testTxTokenV2DeletedProof(tx2, dummyDB)
				testTxTokenV2InvalidFee(tx2, dummyDB)
				myParams, _ := getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				testTxTokenV2OneFakeOutput(tx2, keySets, dummyDB, myParams, *tokenID)
				myParams, _ = getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				indexForAnotherCoinOfMine := len(dummyPrivateKeys)
				testTxTokenV2OneDoubleSpentInput(myParams, pastCoins[indexForAnotherCoinOfMine], pastTokenCoins[indexForAnotherCoinOfMine], keySets, dummyDB)
			})
		})
	})
}

func TestTxToken_FromCompactBytes(t *testing.T) {
	encodedTxStr := "12osZaZq5MLYv4hnkeSoqVyqjhb1Be4gxKbQt4pwNRMQPP3RMuQAFSDuqdfaJ4zht2MWJXFjMGxQc9hBX8BBDznu7KZvXYTLv1urTX618q9Lu7JkWLVUn8zB4R8tSpBB4DaAJ8PFCsR5NArP11Mx41zTGhSj1DGBD4gyarPnAvb6SJaJDTnRxrcQaBzcWkGrZfHoZf3jnT1NwAZLhsR7KoR5hv1GxvrHhJB1C6a2RV5ZNS5G6zWWpAcWyGBxb9kzGJo1adj4Bcqe9kxu3FAouucFhwT81WLS3KYJ4MyJK9Nn5ZUjwB1AgbEhA4NA3TG6CKJpgnnHhmpNAMMg3FAkmVa4pbBdauuBxTR8cxcyCrFgaQ7Q8fANn6x9WYyFqYGnfWXv9hudZN33zQ5zqksvmM4iiBELbZCyGdAVAeiC5ayiEctxWTPMjfz4iDSVF9yRC6fBYXHtocHTi8QMSUjnwXBuspX77VmCztLM5v9D94npF1XTS5ATFcw3qGoEK1HcjG5a61YME19co3sjG2EHE5vGsXTFBBK5Jzsfd3XFWjKY2VcjzjTrroJLc2rwjj23sdSG9bmSF3j54ZbFVwAbnVBqpAuRro17qS31JaoymssCVuhniXFkhMFMJsqvRegG41TAn2r1psHatRvmUJhRFR6oi9FJgg5cg3X97bYbPgrKVKmqGnAduPy6UFHZn3StqGZLvzoajSbNum3ZsHfqboCeh6VMpCGB3C5TDd9K9GD4xLoc36Pf72eJqVD1bCBSpsLu9Co8KgosK1hffANNVagvYJa9vW2V9J3J3jx8gt5MhwGSAwKT1YbCgwyEE5zvFM1Du4Ky8HgePWQFXU8krLV7tqfMZ9ZayLd7buc6Mn6qerbXF6ZYkqo1m6kPgPQArUyvpEXqjbyAyAFsmuChzExDCQKzcfiWtfWZMKHdaD2btQutpYYvJW1UzEaQGQPdLdnnVXTAoKoxvjjnBwnpEpN7XJpaGMx5mjJtr9umJm3EwQw6JTFARVMwxuW8Lki5pVLnUD1mQHnkDHLGmeiLB91NEZCsEHKer4aTs71NpEFvKrQzM5pRqM8P3b1tCEQUAr4XHXt1eesh24zfWH41nYCgDv9Bs8wjH49zdMkKDsjvuYAijteMt26S5sTCycJRb3NxNLQh3f6ivyZkwGzm3nRSebwMz5KFonLR7SXfocNfovMoUVAwZz6DnPJsiV3mdarWs8E7TLRcKXE2LHH5fWkjnnjygZyJ89Ksc62ddKsAwvk6Pf1TSm2jKYB5vYMGWpTFCgMXi9NkAVFvvxUiB8KXzswyjab75qrsT8dVrtmA6gbRrq8R3eP6tnjppYLh1h8vQwbsE4MA4XoYfGeTqvPm8LiYosTGCYtmeaRxZvHAtBZ4Y18Z9eLLeF7HU1VArfRoyyQtxsdkjwdkcxb8NZMqtUynMj6KTkohRRABR4BBJWor4kSWUjyFBD8tyWuWwD8aHsT5mytejujYTRGW54MnzcWukQBL96zPDHpcA8vhiKogd2tQWEXSSu828yp6ptXrd9fiq6b55LDkUHZx6QMhc2ZxiJGzLZ8dWgAHu2kmFf6u6Qo2NVV2aWjxaUwUqZp9PXGFiybHpWuDxVRowcKtkExNedGikpBNDxAcVU12taP2FM87JPX78UYvA4uneaqWRs5SxW41dLXgPVmXnzpumJqmMGwUEB3AEzsWj8WcbUWMrMsayqkcZua5JaPqmq23JVaErGsLmgeuhe6YGsbjTRJQwen9337R2mWw1QnZuboDeK6vEk3JYH2bVbevaP1LfougAZDv4o8TtRNbxU58CkwtuqX9iK6SoY2JjEf57QVMAceyvhfAvrhTutRU4E4W4iQAUYdATuuYpaKYS4PPTox76n3HYpqRVYS9TrEUAmHM8byqkd2gU5rqfypopSBo6mnWdY7UT8vtADE9oF6TNbeu4sizPjL9PBVGNZ7b9paSsABUaeRL4JaGEmSoiDfqEeva7nbSncgPRtGT28Zb2FSTXJAconRhUecLJd97EyzaE1XyPSzQCGxt9FWRSBKcCq5GvvgkBf3AJyzDWWfRud7iVvL1UvWAYhkzd4ygFWZ9RXC7F7ayfVG25pfTPkxVPP3TqwMWgwUXpETN3ZA63KBSgGGXcSLPJqbUUscKfaBGPS3ibGMQpKg854ghsr1wh8NWBKFckPS85Wo94yCqUG1PtxDCzxM3YaZh7AoVZL2c8ueRrGq7tVJYdZL4VGCeeTLsX1C7M7dkjg8t9XZyEFmhNoCmqxxsWXEV8s7xri85eB1pTGFDFMBvX6nxDvNhHVZAKsMKi4EhgbwssMCiy6EsoMGXvvRdod42rN96tck9NUd3uhvc9qBiijR6VjNJCuSc1z8kmtfv8gXcxuPZoSRgohDUDRCSeULJvHrjSYtZcFc3jeSeRHsfgCijVEPzX5NLsg33CzNCxCHqTF9D9NgZnKvmugXw5JzwpkuFPaDTsppcKfvZaRTbLszMKFLXLe99BLL342mZoKxFXHEyif6WvFrCDccRzxmXzKrJHqdRFLP9oTbMui5CQx2FSoKtNWMcQEEH2YjfoEUskRbT5rx8mikFhmKrNmz2wXXcmfFv7a5B1Yeesda3cw6CY96wr8MxSjKom77PdsAaDVkbZyetGRn7MwGKE43fHBdZWwADJfuRrWBrjH95sb386j7N6voHgW3tQFPtnevzwLowZdFSzzpJ1DY6zXFBfm4nh66nbTyePqL9BpHa9WEyjpeAnGBxJTmroWjFSoet7Y36wcYMZryYkz9aq1YGpeYQoVhCinF7kFvf3aQ28Pum3Kkk9xNTfnWc6PP7osAbpMPZsJfShaEKQer4JVdv2Ptjvkhks3cX2SMrGgCZDpsmERGan5rbrC9dF5FX9shAx7mPvYG7x2M7UhLxAzRR6GFWKxHQt51XhWg1fh5Yrk3X9ytbnFPBuZPxpv9pDAnFNMQnymBnLv5wZfzXZmrJ5HKRvE7JL4QjRXdmu3XAeQmeWEzV59zxCPH4kXTAL62a3FfLgjoZykXExYgyZjpZsPQ3faz63cXEgpXmw3EvPKCkqkz2dxnyDhwHQbKiwzhnjb3pZRe9UujKMq2h5wzt3UQREgitE6ipNGobuyzokXT3vrssfuapcVzWnGygSLykdFxDnxxesq4Gyjp77LRGC5DNNjrLRDwhc9KLszEGQWnkuUK8FTqcRCxyNFHcH6AazPnv46nB35M467mymAD9HKTfMiWM77jv2KF6S2KDcxyna9DDWF4d2YdACoS7MhGbFg28TeUyaP2YtyxWPX2NGngJE9Pxz5KYTw7gywkiMHdfcMsaz4GTLEntkXxiL83ebvb6iK2vKD8s8d7q1sNm6Q7ne3USEtjPb55LiH8AE3KHJbwuF8eAYXaioAHLnB5Wf5Z5t17Ck3eyZXfdqbBHLEv9B6EmCvHjNVwLrZFxqv9irbMGBsNeHfjWvHpReAZJ5NemjyC9EkTQT73PJtgv3M5czebwgGhncf5jHAvtEKzKKorkG45kDJqCtvXXKkemStaUm9wqNPM1cg951xoJc4kLMjAHwYMs48PNEumg2WB6s6fU6DrA7rRuLfLf7T9iaVEZYBywUh1mGdkGVej3KMVC5gWWeeLsxBp7LX5KhfDdtFx8LPEiCBWjDK6FY3yBepUTqD4TysWDhGRcSrnjuH8cGxfEZGsPK7DjvUpZGNhVdckNeXG6izviv5JsnuPdyf76BSCUzGnX9FvpmC3oTQAdHk8xPw5gung9Fm1LWMmmfrvz7LURJEbLLU1UPTDMUWLWXik5tita1cVq6kDGNTzhcmjdKrZi6d3EhtvM2gy2dfFRfARAxSwCVACpPgS42jsBUeij2BPbviwEGzVYNFcU5SGpz3PBP9QFXA37eNXqaM7eFBnBBjP8km6xmywjoHMczQQYrFSHPyQM5f8avHcZKDFBePip5zWojNt9SvoDRm2KaYpWA5eeF6WhQHMhMokEnFRT1TbPcWSiB2J67b57ZaMfNPqK81Z3xoWKDTCS4nte8vXD2gwUTAFEG9fRftVF4LrcagBhayLNkuRuYYx5PpNoojDi5Ti3aSnZNDYdDLQomKXViZoyeNJavUYuHoRA1o8JBGvm2YxiQirzHqQFagLev2XrKgz3e5gJARhSbeXdicMndTXz9Yz4N7uHnxGH4DRzWzWNAxuNZLxcUTsos1pTiMcY2Fu9sYwBogDqpFnjD5oZqo1ReU4inQ3w9Jt1JpPzuihtKAiagSqBzwU7GQRtDVk3fU3h25Kj46AkYU1YzWvtTNf2ACe6KxucvZHZ6s1e3qnSSdCVPzP5ja4oLkdrmuGLABprKcicLhXsvkxzz6RH4vJhgVcs8PaU2THSbjtByrKSNqnPW1LWiozdfnLxSegW4iqQWtgEHuenGzRaMfK4nAMps6SQYHrvrNW9PEt7gizMoZnizujxkV3XwxxfJjo3iP1jcJBzHLZHh8DiMF1iwAg4uT521ssmvBQFCVqxbXX4k5ukja7GVRkvmhnswNqbvNvmqrVTzV2vEr6KjkvhJNoatBCKFWNp5CvZpCqFGFY5EPj1sgsCDD5MAXURFgFRsi8mFU7NDjcKktZ3SmrtW45DhpBwLXatDhVgSnJzdtF1Tk34tBTPxB6ST3on9HsKySL8HYpwE9bFcEwxRvD3669WeVfJiLRTUQEt3BxGghbRrTEdJ9JmBnitkHaRsVDvFSKVRanW9akRBbVUWsGgUJbEbt5eKnVuNyiNuD3LsbACG94s8qF7WakxAtgG7Txem55vG2JR6RCRNA8qEXitrrmKmMP9XFYKBJsGjJhTvx8J6hqB6RSRZq5uisKwEmKV1qMaP83Krd2rW3demXtbKKtPihFdgcMueJcSnExRkzGKbWL4Hwm8tBKNTGJdm5zU2csPLWy67gdjNEgQmssPQ1JY8sgVrN2j7aK1Xm8YDRwk5yMQDSxGPkkP4Wu5sjYor4m94Bm1zV3SvAYx99hMNTbhkG639w111snV3iQ4LVvWY4Vd25NNULUUuiAj8b66ULeviwwyyrDwQgut23mqQAw7PzTEJAVSKtnS3NDPHG7PJgVTDQ5eEQJGKL6Htv68mt3xh8GEgvXo5EJw9fKPxqMe1PAmd6RQFEK8n8rQH6w9u5fpLSB3yEYJW7zsA8Wxq1sk8s3twzsbrKcNcNS96gUgntrST7ZTnPs3Gy8bLPRLoBB5um3EPJPSQJBErMLYgQXkmbAyt7xhESVfeAQdirJgDg5WvAexdoFkayTDM8Wv8CSQF63uw2tF7TLEAEKnmWZ3U6LLXdfLCmSw5tfLCSgHLxnT9yRCQQZqHNmrW92YgqTjZ1DKtvNos2CzPTmLf1PMxjSgfDfB4ACkgvJKbeFfqTakW5szZREzA4ebLmDG2UUNVrzTsCPKX5tQ17oVMAkHxMhj2C5N4VAKN3GJJdqyNfkUoa1uHsTWaP9Tm91caYH1wNRCCXPZby6wUU3cVZWynxXXdBe25EyPKiMgKETzgFSUFkTkaaB99k77ou5jPncxTtUN5ivXRtRxwDrNNKv5nGz9FoLPrRc2tdmgpZMCVQ4ejSY6tRD6FNYZLi1eHLeQk3oGgLBbQQPj7vkcdbPVnzoB6fqihjNszQboR3hPsacnLqok6kruUm5ifRXpkDHFxjFtCmu9qynADpzsRg3ooaXAYKD254DgHUdmZhpRwqMoyGoCXSrd13JFrv2T2BUPN2af8LvkvzztGSgcGkZm92vA7Fzkq8uB8XRbdReoHkMcUK6KC1Rcc8siYeJoYxe6KFxo5EeVLtzjeFVah2rAkLHnsJBH8LEEa3uDqCp7e5yEhchtFSr4yXY9wQXtCp5PTtyASMLeN5MtXZ8VnXz6gesXFV9fyxbeWKufsccxvbYC3knB5Gn83UpifK5PsNHV9L4HMzX6UtTpJHaPQE3Htt9Ag9NwPDTPbfww7eyG6KEvwr8sLTW3cvmijB8bJQaaFJHd9hxVwDhoQ66oSvXour5VmdsPTF5U9SxHGmZpQm7bxv8XkkJNeJqPWDpeUDXsjxvrRbUGQV1tCAgnoJs9uf559XXy8URGLdi4uofXBMTwcY8UrNo4qysv19swLBqKn6ZDxSJy5rrxzg3BXuoj8RdGEtur14czEyDdKg1Nps4sMypj2ZVuouDCdCU6CcBg7tzYTUxJgTu1Zc86tRyiZpuc8yiPVdnQzrSMsAcVGBfDgwkUdF2wuJZ3yqw8zj5sc2sZXEBvxdKJjEX2bxk8vsd4dmsb9SPY1P4xpEHQKCUmGdisVHQv7CAEo9JDPkbWALCutpUDRspbDMxmtrUBitMKP1pwdQrNDrekdiPrB6SUAJ1FqhqjMZ2uxHAcJ4jFUjzsFptLvWuX4ufA1Lg5zdmFCPNWB4Y75Wz7G6SPYwCUpV441B5FvpfcNWGdacjquUzy58EFK3imSHq2Xm61CE4scPcx3fS6VkgV4YKHSCDE5bLPUSr8QNqwhCeGpEPSGQZqmQG9qsNXPcRy979Jkb9nbSuboDGz6DQ3nbQPsrkHw48CsGYoefykJdJPmurXc9rDxCBo7eoHV4kefVx9VnFTUhsb2XiPRTGWyX7i62gNE9Qo4TJEPNKDV7czRRj21Rk4H4bSovvchh2omE11FqZbz8vRffuwBhZXQGHoHv8wAP87pGzYSf5HdtnvVf6uWgdrZGJc1q9pZ4UZkAUZikVrJTHRWDz9ZKbvZ8Epdkd2uuxWNnNnnom6V1T4EEEThN452eaMdhqZquMtiRNc8wUfutPxoMrDQhPdyPuz65LvBy1pK7HP5Ngaw455Du3Rvnv3gayADB6cCjAoGaZ7FpmkLiXWLemMCam1hgwA9Hippwfvep1Jr27pp7XDPwRPLBSCHzztDxQVVeFcUzRbHQgTUARd7BXc8bc94bzaY2TH29Lcq2fUC8rMELrMPvfZWFrnJsbCrj8GscBDt7Y6QobcPePen6HSoHKBP88z8rxvfr2jBTkvmUbFZNjyMDSCQPS92xmA6baagY2RsZL6ffg3trzaDnQzexVMmweVRQmiwWWcTAiYMyzncQzksMkv27TkJmoYvdsUkhEu7F348NxbvpPe3DRU3JAcrgg3p6YAVz3gfc9EturWmSavN9DPAvwJVKNuLrhcCPxBMeLRUWS5oVJ7jbapztrHpDdyuqHe1piLGkbYb4GBAYzDB9Egwhe53wCNip2W39MVkfoR9CU3MWvyt2Gh5L5CtxVGXTW8qgyZzw6XhcAm56A7PiFSoms9nFAbEC2ZvMGFNk2pj2ixkscD9sgk5MRtHnA99DqP32UsiBDEbPnDQR7KP4eihBbiaX76PVhgr1kp9JApXjZVUddActMYN2eSdE421UZbzJ9BbDSELcw96dasqB8PicrnipwdzQxLWGSh54VUVXPhW4zqjg6udqBSQeG5ZYBVsv1KXUFdF351bumj3zy6uATJ5eZD5jbYvpfDgM97YjL5R87GaEK9tYTbV7PqZ25XwYA242XNPYm4sgb9fF4njtxNytowobxWHRBfLrSCFzFaUCZdWAuYM8o4EGjSwvfqP3DtPTt7rBiSFEAxUDtMmDyJvczT8QZjCH5D7d9bCioefjchZ5zAnWmp5gyiUJNbQhk8gC4TCLeKzRQ6td8iAbfxSV3ufF5yoKosRd2ZkXwwDYBUcDZYDEz4wuMyLDHNgiqbSsoWkFkZvLoSzaMXMp45xb4X8M4y1S9iHGUh2MpuRmaf14K65HdZU36xn5iUUUgnujWd53xmXsjkZmLTPiJdu6w4uY5MUcDYhy7q7h6CcgSHxaS6dhX2J8KL3am1KL9g3P2qe67PEwVNiiia7ruYZCw5cNszGVg1uVeDiASHYAV65CEpq71N3v7YMDVvMa7iFChGijVcGjzwZ9iKxPchvgWHh9Tm3HEKRr9SiD5VdJgw2aojPVcKGhVPyDaDZxQ4zDmfxv2btvUWPrpXVcy1Manhb7Kr91LyX53VFVBp6NzugAqEGzxyHFbo8ThwLbddu7fXheLvaybPC2Bsv3uFS5erux3vNb2QBWpugo4Tbn2XLdM2i5rLjvEC2JdCVeuss7NLVEo4LYFDvsMPiyQHHxVnMnA2WCu47BMTa42j3cE9f6qSg3GiXUcejGSq1wxQUDPCPFmp7sEgqfhjVocU9nohPuc2KGTHQ6SiRb4siw1t8ftsJgmpTNz8qxCeoRMruUFTTfSuYBtBo6iVjJSGz7CLaQ9sCZfPj8hspx7jRXHrD1H99Q2xmYVSZHy813bsj8Fppmx9XH83xVxcAvRs88ckp9tmgpQGjcQXzDppsaBTVcubb7HwgfoAy4Qsizntq5iJUKwdWxHQiv1Sg7WbCSvJo3VAj4wwx3HsbsBW6XvgGt2GAVxYeY9gJo5vSVEgJwfE7w1deihWpGxK1V1U2NQmutkdPxPicmpc4YcuGT2wNS95whB3qMjM6SoPBiyCyvMqc6QhUjaSZxgU7eM1WgJNH9cWksFLALvdKP8cby7Fm2SJd9FtN8BBdXURzYAPgj9J6HKeXJZhja8pG9isVNoxZY2tLWMMiwmN26BWypv91hii1xUK6oM3gWVwwBWWsLYutNLjWesuiyYY7DuRMBQSYRUo2669qyENpAF6XyYX8MNkwLL78oWFRosbsdcpykWExdMKcwDdjtsbGBQDrU3VZhxjfUjHtf8JMe2jGNnLjWa7eJ314dLDTNt8i2rdJGc77VCxEn1BYcV9hE27vA4E7YFkX75A131VQBC2KeJkyUewT2uA7ExSgPF6LYTXqArW57S3LMP3wnXTKFW4YJHmZikv8QymrRJuGhEKPkm1L6jCNzpvGdXYvwzRyT1JpNYgpi5iu2sDpAfdskWocHPK3RdGrFaRKpBt7L4npKH6qbr8AqBFf1C71ExktcrY2aQQhWbYEdyvnZDoYg1r3pG94rBfu4a48tnwfwa3nprw4NLuAUUCaiTibKvU99Z1CTUJWjM6rCFyLmoh5CC7iiepYA8n5C4Nd6FbyHkBpd82kNPKkocfUfWraUAt5hwabSPTBCUeZF6PnW57D8RD2C5mqHFhPoZv9msoT1KwArBpz8Su6qyDDBL1uDD1YDmrAUUwBo5HFpewGDArtDHQe1zzED4B4MgLizMSccggXvgBF2qdaeJ5dq9nnRQYyx4XUum9AhAaugwNpSqUo1ggKsbQwMm2wxrfcGsxLLVY2pi6masEbDY6M5oF1TWAj8im2ovPLd6hkuKaP7jEkhQBRZxNUondygCUFLT76W1VomzR4omNMmCrsw5CYPGPJr4nRy5oJoudY4kvkvVBduwXCEAAFudkBWhnWkCwpRaG9bDaqzMETu9xzzLMm35iUCm856rjggECXvzaSxYpgqqPBwKcYaLTZYYJ6Y2LQ7jsiz2E1HL1qMmEDG9YNSfchJKvbMeYcAb9g6cYJo8NdZNVtsgjmY9Cdy87dQZwiZf7Z3cXhsW5VtezWQ9bSDAjrzUzErBTJVhv49NvAHh1uco17sYYAAXWR2wVbfdukiQaWvmCCpnLP4xifA4en4CDfoKq1WytCf7FS9wpBCWisCiVASoqecYhycJuYcEkfPrv2ZeFBpfuHxQPkXnzo2yQfVLJwnisXPvjXd8ntfvt6g6gfoiEtQS9o2QuRVVzfTFjwiLfb5wdkUeU7jNeLAeigVasM3ziUSjxT2o7dAfDjaAyUAjTjN9bz2Sx8BfsYXYKCBCUiy27TZvkkeZe1PGU6LSjsp9WX6W949Ne9zYPfHH4idxtMommpj97RdJT5GqjyVMUf13TKwGaiPiecy7FmcyqWnGHK4vRuVUpmogspAq6ZBwE7SFouHNJLVDcfsdhsA3YHR8Y5v8Pd2oKzqwQ2YhTjQ6c2uT6jvMJxfrnvZ9vTPf9n2BMK7QAgdqHC36xaycBL2ynM2wCNGS3eNf8NFVZwGVfey63h9WbCK2W3vVYVz6g136P69cTkQtHhRwMsxgz33QVxVECQUfwoHAeipNbbFngYB1QqSRHj3SSyUXGqonrCwsY178piy3gvciqkBXCUd2B3SFps1DiuByLPABQeSUVLwXzkF16wNnzwjwDHULsUS4yXaTDHTPX391VvU3utCNyFXaSzCA6APNtzDc5mPLwQLUyf2aNHHJqH1kDiArNXGbndSooH94xZwPFHN9Gw1jovj5enrgnd24dPMWqGc8eU3JCzhu5MQStKeyB4riDz7vq1hoqRXQCzDj6ENBB6chCaA6WrBp3WhMVvcWXbJcZTTsZjUwGQdxc2sez3FAjmyR6TrBBTKUiWSmiKnLtdX9YsA8AqSQVPQEE6Xi57euxadnD8SAQ1mhqDtvqk8QVyjDJEeq7bYGoqCMBskNuHnMujGqgP9wSbynJU9RJ7XF4WSp1yFaaUDZG6oM5FjY89s6ibM5cDXsWgvq58sEF4ng8EskkL4B2sY4P7bTshhKvrghmDTC5WXv2g3GxCpBCYSn4wb4ciwGA5a8MpAgXcPh6Hx3WrQLrJA4FLuifSL9LbW47xLKAAWdg1mrM65XaJ8cRnGm8Kaj7RQFLhefGDBg6NEFmqY3J21maDnoXJGbGST5G2jXrf67CjFEZAKVnibMvxCUjEXerTEbdJiUznAsFogxNEqudZyai8ZAq8TsAFTbqjn3Leduw9jsCmrnajM14JEh1D7Zk5neUC26u7KFQBriSbgpdHPN2zMgTpfg8dBjPucJRUNpQBnxV8KmAtD9NdU84PbJCe8QvoKvCwKzDRQpZX6wXcN89yFnWWWAeYc7kwQPD1nQUpqsdhbA3J84RWctmmeoRpVd2sZpBHTcNhXYKmXW8NnPmesDXYoam2FRkvSsNfUXi4Ln2CAEoykoHBY2SQ4V6zhnkdpzuDuSPgxDyRkwactGK8MJtWfERUDhSUrH5BJaHCvcftiPzaAUPGLepe7o5rkZraGm4EnmfrcjA21xuRdkjbXXwQxzPsnM1jXGm1wSq6uYAuxk2cT2gJiHzs17ok6NnZtS6TTNMBBwjhmHSU9W39a45xbNVgp7SYEpLKKS2FXkjpQtBxn9XMSPbA63eRVakJiG1LBjvBorqYNDvzmNgnJJHm6WoAnGxRhac9vZpBDbpZ9jN9sFD77TW842LPMCpkL5nscRiAWxdzBE3wBmd3sQULWWT8XjmhujZybZke8FArvnRiK9zsg7F5S26tC6GfhibaiTmwKyymgNNQmmyBk4esE4oaVfbGDHhqBGiZaGuUXqT4bv9aeuxGg7DFdcdnWTSUks38skiUiQ7kvoKwm3WPrpff9fyR4FuEWfjeDNbuGzvpdKRT7Rg9XdQ9MEpDzkQidJmQd9RxJHg9WDgFmNeFNNH1X92MiYkFmWg78L6hKqqywY9FYi63BbV7rYuU3KAXCfEwtcPhFauyQwi2ZSSFc8RT6tYMQK9VGfxPX8SxEm9QfQfeTgMhJHxGjaN3cH4gsKmmHfMx6BrXw13iWnhmFLpUGYeqAXeocJgTFTGwmCpgPzBxYHKvZCq2ddx3KoWUT1Cm1rEsRhtSRRvUa8tkM7Gdc3Le5mMZgXfdAsRMu2Smjg3w1b5ww5cHMs2eZhymhjxnieFaW7gCJNDNFG15F4FLy5VM63m53tyQ6ZXch3oNDvCQ5rT2zSMXiqAkW9cmJj7ZSrcHiCbKhXaPpe5GCGVCjA3moeiHsSiVhSMHsUPEN577RWupTAEo8gQLE5d3SNVLo7zmvdBnSUKETXjz23a4Rz1qeS3bfn3zjjUUGSore3AYFJXD8i2ciJGyMpUMbyv7kF5chzFaf2yTw1Wfe8Wu2xDzfrU5mEHL8o9UZ2dSvmSJHGwjTxKs7yjLMHJhsxJQNPBzXtgbkbThu9t3ejEZfymNgMBaSE2TSf4DtXuFtys63fWoxx5ikZiwnXDW6o2scBp1bxtyp2fHtbLcPL8EuLef9vPpSoQ4WzZNnYdUCrB3y8g41QiMUEXJ8zUmwsceKqU4dkRPsQdhpyPaiXS7dFf4BcUfCiy4vXwWtbPKkcSoFZXR46EMU5JV97HEgLjY6qse3TDwoDL1Jygw8gghKqgNFx5tJqWt1RRw8ET2pZXxMZkRCc7VumU8U8p2g3gTxEjfTHzGePpmMXe1xUyQXExYziSHeuhrrGY1YymoH3g1A7rEJDUoYx9Yg11tM5otstK8aZzMvassDCz3iN7XYaz8fuqrYSkjJBvT23tiLwzymGxmdzULaGf6iQxrTpswkKj47jaamEuJCuMPzQ99HhTVqDd8rjuFwYm9AY1LNEzZG4BQ3aBjG7uPn4PHe8QAkQ6VPDVsksmnfqUw12RSJZEhPZHNmcbfJbvEZM7DF5BVSzv2UvZewZJ26S5qjbdx1g1BMRZAzdHX9skbVsGeG6J86QBUbjbiwdnbQxBHqaL1AjP8JsmbZe5NRaHuvL31uSPYBM2imVwmLPGuTfZNfL3AXQNu7gpcGWM8WQebVbRAe1H5G8kr1iGj6SD5XrMzzDGh8dXRMhirMJRcLzz17ajFUA3tw9ySZ7W4wz3vvKYpm1PQbENd6XUdQmUJWGGPWxm1KnYEyn8fowNjKzpq9DcMPX6K24buXQBJ8tMEBPhm1dPNS37y8fNTPgLTzGbM6zwqHZ5i4FvS82ufL3mNgkKMvyoKRyNTgtUGB62ykoXJmVQEXiNhQVVLWgSegno1y3rthVWYecPRYfyomAs7743EwH2CxRUkaMFkHxMcqoRRFWo5GLFj7sQqY1N6rT7ATGBZV93NAHQ6CGMxNaHGML2Xspc7HrPSe9cp6QS4F1N2uH5hGgxwg2vkir1gBydStcXx5yJdWze8DC8iKfYSXpNMxEDzxNmjgSmpri24ECNvESnHkNQfTqfDvvJMknBGy3rUAArwBie5Suv4WLxG2UZdpGuUB4zyDJG26NDbx7NJ6jVQ9TXmkVFtHLFar5dTdWRVUtMygWxdiyuNZiXrn3VB3n11328gaTMgFe4Jdi4bCdcXSSCjWngEU7oEx5CE2E3KgT3ZsS731xfMtffHXfRpcaS8Lou6rCQfykJYsHgSJmVKdyJPnDh2YjETCqXYtVpkrKd7AKSPiFtXPgeEnknLVY1GoNKYFsNpLQ12qXo9GpPaPkrKyvbQVE9SDgPVo4Qj1fVy2meb9MdnpzLWaBVeALiavPJeLLLXacge6ChgoL8kZVj48wfAwjSpJAFu8p8h8RYQW1PWM4q57MgyfvstGB4jrwCTHE6KujE3BoWDy3XJDWEozmwrnRnMfGXxNBL6AvduKCa5nGkeEauJz1EE8oS6ooTNEuyjAD4dXvXctUvBPvZJQuCHKYpgrkNY5SzjWshm6v9hoNUwnhmYPwEoJQNiLAnd2HHBz9ECXV9u4v8BCfUirpBeSsousD7CiM1YDWK4uB6FXJvSoyyB14uLswWJZE5V8Snqqz4CzhUv9dKSPEYb3jWGsPA2eZ1CK9GHpkbMmL4eV7ks8caJgJzTN7UBCDPQrukxejsUosLSJdaSdiESjUJaEwSzymktXjFMPPwY66XUbn8xd7DhHL31uvTbxeTnB5KE9ah8WZFnPR1FGRn18uLt859S39EskiLSUfRkSf8e9vX6tVmMZQQhaLapakA3fXLHyyobJeqknZZexh5WqyyyFo4vaaDsN4Ez9xNpyhwkdCW9fB8k2EimDShDgi2ceTLpuv7rUGvdMXLMdYXxpCjqJh4wfRnDPDWEroCnP5WWKDHfDGXGuEUwqfMkNmp4N1Ci8QWjXXC6nwBn2hdoSzkuTqdFCTayszouryPQeaBVqRjL3JNGgZWN1v8K5Di8pcUV3bgjeeZQ1HuKyZVdMTdrEXCxsU4rUWLHDrqqZ1sWfeNWwKPavt6V8knUoZ1gXqEHHCgJcgokfhxvFhypNCgTAAF6G5Zxq96eiWoNvRiYNPjdFq6Ksq"
	encodedTx, _, err := base58.Base58Check{}.Decode(encodedTxStr)
	if err != nil {
		panic(err)
	}

	tx := new(TxToken)
	err = json.Unmarshal(encodedTx, &tx)
	if err != nil {
		panic(err)
	}

	compactBytes, err := tx.ToCompactBytes()
	if err != nil {
		panic(err)
	}

	reductionRate := 1 - float64(len(compactBytes))/float64(len(encodedTx))
	fmt.Printf("jsonSize: %v, compactSize: %v, reductionRate: %v\n", len(encodedTx), len(compactBytes), reductionRate)

	newTx := new(TxToken)
	err = newTx.FromCompactBytes(compactBytes)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, tx.Hash().String(), newTx.Hash().String(), "tx hashes mismatch")
}

func testTxTokenV2DeletedProof(txv2 *TxToken, db *statedb.StateDB) {
	// try setting the proof to nil, then verify
	// it should not go through
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isBatch"] = false
	txn, ok := txv2.GetTxNormal().(*Tx)
	So(ok, ShouldBeTrue)
	savedProof := txn.GetProof()
	txn.SetProof(nil)
	txv2.SetTxNormal(txn)
	isValid, _ := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValid, ShouldBeTrue)
	isValidTxItself, err := txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Missing token proof -> %v", err)
	txn.SetProof(savedProof)
	txv2.SetTxNormal(txn)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)

	savedProof = txv2.GetTxBase().GetProof()
	txv2.GetTxBase().SetProof(nil)
	isValid, _ = txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValid, ShouldBeTrue)

	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Missing PRV proof -> %v", err)
	// undo the tampering
	txv2.GetTxBase().SetProof(savedProof)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)

}

func testTxTokenV2InvalidFee(txv2 *TxToken, db *statedb.StateDB) {
	// a set of init params where fee is changed so mlsag should verify to false
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here

	// set fee to increase by 1000PRV
	savedFee := txv2.GetTxBase().GetTxFee()
	txv2.GetTxBase().SetTxFee(savedFee + 1000)

	// sanity should pass
	isValidSanity, err := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isBatch"] = false

	// should reject at signature since fee & output doesn't sum to input
	isValidTxItself, err := txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Invalid fee -> %v", err)

	// undo the tampering
	txv2.GetTxBase().SetTxFee(savedFee)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
}

func testTxTokenV2OneFakeOutput(txv2 *TxToken, keySets []*incognitokey.KeySet, db *statedb.StateDB, params *tx_generic.TxTokenParams, fakingTokenID common.Hash) {
	// similar to the above. All these verifications should fail
	var err error
	var isValid bool
	txn, ok := txv2.GetTxNormal().(*Tx)
	So(ok, ShouldBeTrue)
	outs := txn.Proof.GetOutputCoins()
	tokenOutput, ok := outs[0].(*coin.CoinV2)
	savedCoinBytes := tokenOutput.Bytes()
	So(ok, ShouldBeTrue)
	tokenOutput.Decrypt(keySets[0])
	// set amount from 69 to 690
	tokenOutput.SetValue(690)
	tokenOutput.SetSharedRandom(operation.RandomScalar())
	tokenOutput.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	txv2.SetTxNormal(txn)
	// here ring is broken so signing will err
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldNotBeNil)
	// isValid, err = txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, 0, false, nil, nil)
	// verify must fail
	// So(isValid, ShouldBeFalse)
	logger.Infof("TEST RESULT : Fake output (wrong amount) -> %v", err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	txn.Proof.SetOutputCoins(outs)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = true
	boolParams["isBatch"] = false

	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	So(isValid, ShouldBeTrue)

	// now instead of changing amount, we change the OTA public key
	theProof := txn.GetProof()
	outs = theProof.GetOutputCoins()
	tokenOutput, ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = tokenOutput.Bytes()
	So(ok, ShouldBeTrue)
	payInf := &privacy.PaymentInfo{PaymentAddress: keySets[0].PaymentAddress, Amount: uint64(69), Message: []byte("doing a transfer")}
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin, _, err := createUniqueOTACoinCA(payInf, &fakingTokenID, db)
	So(err, ShouldBeNil)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	theProofSpecific, ok := theProof.(*privacy.ProofV2)
	theBulletProof, ok := theProofSpecific.GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2)
	cmsv := theBulletProof.GetCommitments()
	cmsv[0] = newCoin.GetCommitment()
	outs[0] = newCoin
	theProof.SetOutputCoins(outs)
	txv2.SetTxNormal(txn)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	// verify must fail
	So(isValid, ShouldBeFalse)
	logger.Infof("Fake output (wrong receiving OTA) -> %v", err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	cmsv[0] = tokenOutput.GetCommitment()
	theProof.SetOutputCoins(outs)
	txv2.SetTxNormal(txn)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	So(isValid, ShouldBeTrue)
}

// happens after txTransfer in test
// we create a second transfer, then try to reuse fee input / token input
func testTxTokenV2OneDoubleSpentInput(pr *tx_generic.TxTokenParams, dbCoin privacy.Coin, dbTokenCoin privacy.Coin, keySets []*incognitokey.KeySet, db *statedb.StateDB) {
	feeOutputSerialized := dbCoin.Bytes()
	tokenOutputSerialized := dbTokenCoin.Bytes()

	// now we try to use them as input
	doubleSpendingFeeInput := &coin.CoinV2{}
	doubleSpendingFeeInput.SetBytes(feeOutputSerialized)
	_, err := doubleSpendingFeeInput.Decrypt(keySets[0])
	So(err, ShouldBeNil)
	doubleSpendingTokenInput := &coin.CoinV2{}
	doubleSpendingTokenInput.SetBytes(tokenOutputSerialized)
	_, err = doubleSpendingTokenInput.Decrypt(keySets[0])
	So(err, ShouldBeNil)
	// save both fee&token outputs from previous tx
	otaBytes := [][]byte{doubleSpendingFeeInput.GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.PRVCoinID, otaBytes, 0)
	otaBytes = [][]byte{doubleSpendingTokenInput.GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.ConfidentialAssetID, otaBytes, 0)

	pc := doubleSpendingFeeInput
	pr.InputCoin = []coin.PlainCoin{pc}
	tx := &TxToken{}
	err = tx.Init(pr)
	So(err, ShouldBeNil)
	isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	isValidTxItself, err := tx.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
	So(err, ShouldBeNil)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	logger.Infof("Swap with spent Fee Input -> %v", err)
	So(err, ShouldNotBeNil)

	// now we try to swap in a used token input
	pc = doubleSpendingTokenInput
	pr.TokenParams.TokenInput = []coin.PlainCoin{pc}
	tx = &TxToken{}
	err = tx.Init(pr)
	So(err, ShouldBeNil)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)
	isValidTxItself, err = tx.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
	So(err, ShouldBeNil)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	logger.Infof("Swap with spent Token Input of same TokenID underneath -> %v", err)
	So(err, ShouldNotBeNil)
}

func getParamForTxTokenTransfer(dbCoins []privacy.Coin, dbTokenCoins []privacy.Coin, keySets []*incognitokey.KeySet, db *statedb.StateDB, specifiedTokenID *common.Hash) (*tx_generic.TxTokenParams, *tx_generic.TokenParam) {
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")
	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	feeOutputs := dbCoins[:1]
	tokenOutputs := dbTokenCoins[:1]
	prvCoinsToPayTransfer := make([]coin.PlainCoin, 0)
	tokenCoinsToTransfer := make([]coin.PlainCoin, 0)
	for _, c := range feeOutputs {
		pc, err := c.Decrypt(keySets[0])
		So(err, ShouldBeNil)
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer, pc)
	}
	for _, c := range tokenOutputs {
		pc, err := c.Decrypt(keySets[0])
		So(err, ShouldBeNil)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer, pc)
	}

	tokenParam2 := &tx_generic.TokenParam{
		PropertyID:  specifiedTokenID.String(),
		Amount:      transferAmount,
		TokenTxType: utils.CustomTokenTransfer,
		Receiver:    paymentInfo2,
		TokenInput:  tokenCoinsToTransfer,
		Mintable:    false,
		Fee:         0,
	}

	txParams := tx_generic.NewTxTokenParams(&keySets[0].PrivateKey,
		[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return txParams, tokenParam2
}

func testTxTokenV2Salary(tokenID *common.Hash, privateKeys []*privacy.PrivateKey, keySets []*incognitokey.KeySet, paymentInfo []*privacy.PaymentInfo, db *statedb.StateDB) {
	Convey("Tx Salary Test", func() {
		Convey("create salary coins", func() {
			var err error
			var salaryCoin *privacy.CoinV2
			for {
				salaryCoin, _, err = privacy.NewCoinCA(paymentInfo[0], tokenID)
				So(err, ShouldBeNil)
				otaPublicKeyBytes := salaryCoin.GetPublicKey().ToBytesS()
				// want an OTA in shard 0
				if otaPublicKeyBytes[31] == 0 {
					break
				}
			}
			var c privacy.Coin = salaryCoin
			So(salaryCoin.IsEncrypted(), ShouldBeFalse)
			So(storeCoins(db, []privacy.Coin{c}, 0, common.ConfidentialAssetID), ShouldBeNil)
			Convey("create salary TX", func() {
				txsal := &TxToken{}
				// actually making the salary TX
				err := txsal.InitTxTokenSalary(salaryCoin, privateKeys[0], db, nil, tokenID, "Token 1")
				So(err, ShouldBeNil)
				testTxTokenV2JsonMarshaler(txsal, 10, db)
				// ptoken minting requires valid signed metadata, so we skip validation here
				SkipConvey("verify salary TX", func() {
					isValid, err := txsal.ValidateTxSalary(db)
					So(err, ShouldBeNil)
					So(isValid, ShouldBeTrue)
					// malTx := &TxToken{}
					// this other coin is already in db so it must be rejected
					// err = malTx.InitTxTokenSalary(salaryCoin, privateKeys[0], db, nil, tokenID, "Token 1")
					// So(err, ShouldNotBeNil)
				})
			})
		})

	})
}

func resignUnprovenTxToken(decryptingKeys []*incognitokey.KeySet, txToken *TxToken, params *tx_generic.TxTokenParams, nonPrivacyParams *tx_generic.TxPrivacyInitParams) error {
	var err error
	txOuter := &txToken.Tx
	txOuter.SetCachedHash(nil)

	txn, ok := txToken.GetTxNormal().(*Tx)
	if !ok {
		logger.Errorf("Test Error : cast")
		return utils.NewTransactionErr(-1000, nil, "Cast failed")
	}
	txn.SetCachedHash(nil)

	// NOTE : hasPrivacy has been deprecated in the real flow.
	if nonPrivacyParams == nil {
		propertyID, _ := common.TokenStringToHash(params.TokenParams.PropertyID)
		paramsInner := tx_generic.NewTxPrivacyInitParams(
			params.SenderKey,
			params.TokenParams.Receiver,
			params.TokenParams.TokenInput,
			params.TokenParams.Fee,
			true,
			params.TransactionStateDB,
			propertyID,
			nil,
			nil,
		)
		_ = paramsInner
		paramsOuter := tx_generic.NewTxPrivacyInitParams(
			params.SenderKey,
			params.PaymentInfo,
			params.InputCoin,
			params.FeeNativeCoin,
			false,
			params.TransactionStateDB,
			&common.PRVCoinID,
			params.MetaData,
			params.Info,
		)
		err = resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TokenData, false)
		err = resignUnprovenTx(decryptingKeys, txn, paramsInner, nil, true)
		txToken.SetTxNormal(txn)
		txToken.Tx = *txOuter
		return err
	} else {
		paramsOuter := nonPrivacyParams
		err := resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TokenData, false)
		txToken.Tx = *txOuter
		return err
	}

	// txTokenDataHash, err := txToken.TxTokenData.Hash()

}

func createTokenTransferParams(inputCoins []privacy.Coin, db *statedb.StateDB, tokenID, tokenName, symbol string, keySet *incognitokey.KeySet) (*tx_generic.TxTokenParams, *tx_generic.TokenParam, error) {
	var err error

	msgCipherText := []byte("Testing Transfer Token")
	transferAmount := uint64(0)
	plainInputCoins := make([]coin.PlainCoin, len(inputCoins))
	for i, inputCoin := range inputCoins {
		plainInputCoins[i], err = inputCoin.Decrypt(keySet)
		if err != nil {
			return nil, nil, err
		}
		if i != 0 {
			transferAmount += plainInputCoins[i].GetValue()
		}
	}

	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	inputCoinsPRV := []coin.PlainCoin{plainInputCoins[0]}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(10), []byte("test out"))}

	// token param for init new token
	tokenParam := &tx_generic.TokenParam{
		PropertyID:     tokenID,
		PropertyName:   tokenName,
		PropertySymbol: symbol,
		Amount:         transferAmount,
		TokenTxType:    utils.CustomTokenTransfer,
		Receiver:       tokenPayments,
		TokenInput:     plainInputCoins[1:len(inputCoins)],
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := tx_generic.NewTxTokenParams(&keySet.PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx, tokenParam, nil
}
