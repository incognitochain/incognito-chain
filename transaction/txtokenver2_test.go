package transaction

import (
	"bytes"
	// "math/big"
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	// "encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/stretchr/testify/assert"
)

var _ = func() (_ struct{}) {
// initialize a `test` db in the OS's tempdir
// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestInitTxPrivacyToken(t *testing.T) {
	for loop := 0; loop < numOfLoops; loop++ {
		numOfPrivateKeys := 4
		numOfInputs := 2
		dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
		for loop := 0; loop < numOfLoops; loop++ {
	// first, we randomize some dummy private keys in the form of bytes32
			dummyPrivateKeys := []*key.PrivateKey{}
			for i := 0; i < numOfPrivateKeys; i += 1 {
				privateKey := key.GeneratePrivateKey(common.RandBytes(32))
				dummyPrivateKeys = append(dummyPrivateKeys, &privateKey)
			}

	// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
			keySets := make([]*incognitokey.KeySet,len(dummyPrivateKeys))
	// PaymentInfo is like `intent` for making Coin.
	// the paymentInfo slice here will be used to create pastCoins & inputCoins 
	// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
	// `message` field can be anything
			paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
			for i,_ := range keySets{
				keySets[i] = new(incognitokey.KeySet)
				err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
				assert.Equal(t, nil, err)

				paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(4000), []byte("test in"))
			}
			pastCoins := make([]coin.PlainCoin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i, _ := range pastCoins {
				tempCoin,err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
				assert.Equal(t, nil, err)
				assert.Equal(t, false, tempCoin.IsEncrypted())

	// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
				tempCoin.ConcealData(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				assert.Equal(t, true, tempCoin.IsEncrypted())
				assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
				pastCoins[i], err = tempCoin.Decrypt(keySets[i%len(dummyPrivateKeys)])
				assert.Equal(t, nil, err)
			}

	// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
	// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
			paymentInfoOut := make([]*key.PaymentInfo, len(dummyPrivateKeys))
			for i, _ := range dummyPrivateKeys {
				paymentInfoOut[i] = key.InitPaymentInfo(keySets[i].PaymentAddress,uint64(3900),[]byte("test out"))
				// fmt.Println(paymentInfo[i])
			}

	// use the db's interface to write our simulated pastCoins to the database
	// we do need to re-format the data into bytes first
			coinsToBeSaved := make([][]byte, 0)
			otas := make([][]byte, 0)
			for _, inputCoin := range pastCoins {
				if inputCoin.GetVersion() != 2 {
					continue
				}
				coinsToBeSaved = append(coinsToBeSaved, inputCoin.Bytes())
				otas = append(otas, inputCoin.GetPublicKey().ToBytesS())
			}
			err := statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, common.PRVCoinID, 0, coinsToBeSaved, otas, 0)

			// message to receiver
			msgCipherText := []byte("haha dummy ciphertext")
			// receiverTK, _ := new(privacy.Point).FromBytesS(keySets[0].PaymentAddress.Tk)

			initAmount := uint64(10000)
			tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

			inputCoinsPRV := []coin.PlainCoin{pastCoins[0]}
			paymentInfoPRV := []*privacy.PaymentInfo{paymentInfoOut[0]}

			// token param for init new token
			tokenParam := &CustomTokenPrivacyParamTx{
				PropertyID:     "",
				PropertyName:   "Token 1",
				PropertySymbol: "T1",
				Amount:         initAmount,
				TokenTxType:    CustomTokenInit,
				Receiver:       tokenPayments,
				TokenInput:     []coin.PlainCoin{},
				Mintable:       false,
				Fee:            0,
			}

			hasPrivacyForPRV := true
			hasPrivacyForToken := true
			shardID := byte(0)
			paramToCreateTx := NewTxPrivacyTokenInitParams(&keySets[0].PrivateKey,
				paymentInfoPRV, inputCoinsPRV, 100, tokenParam, dummyDB, nil,
				hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},nil)

			// init tx
			tx := &TxTokenVersion2{}
			err = tx.Init(paramToCreateTx)
			// fmt.Println(err)
			assert.Equal(t, nil, err)

			// assert.Equal(t, len(msgCipherText), 
			inf := tx.TxTokenData.TxNormal.GetProof().GetOutputCoins()[0].GetInfo()
	
			assert.Equal(t,true,bytes.Equal([]byte(inf),msgCipherText))
			// convert to JSON string and revert
			txJsonString := tx.JSONString()
			txHash := tx.Hash()
			tx1 := new(TxTokenBase)
			tx1.UnmarshalJSON([]byte(txJsonString))
			txHash1 := tx1.Hash()
			assert.Equal(t, txHash, txHash1)

			// // get actual tx size
			txActualSize := tx.GetTxActualSize()
			assert.Greater(t, txActualSize, uint64(0))

			txPrivacyTokenActualSize := tx.GetTxPrivacyTokenActualSize()
			assert.Greater(t, txPrivacyTokenActualSize, uint64(0))

			retrievedFee := tx.GetTxFee()
			assert.Equal(t, uint64(100),retrievedFee)


// TODO: check double spend by faking spent coins in db
			err = tx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
			assert.Equal(t, nil, err)

			isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
			assert.Equal(t, true, isValidSanity)
			assert.Equal(t, nil, err)

			isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
			assert.Equal(t, true, isValidTxItself)
			assert.Equal(t, nil, err)

			theAmount := tx.GetTxPrivacyTokenData().GetAmount()

			assert.Equal(t, initAmount, theAmount)

			isUniqueReceiver, uniquePubKey, uniqueAmount, tokenID := tx.GetTransferData()
			assert.Equal(t, true, isUniqueReceiver)
			assert.Equal(t, initAmount, uniqueAmount)
			assert.Equal(t, tx.GetTokenID(), tokenID)

			_ = uniquePubKey

			sigPubKey := tx.GetSigPubKey()
			assert.Equal(t, common.SigPubKeySize, len(sigPubKey))

			// // store init tx

			// get output coin token from tx
			outputCoins := tx.GetTxPrivacyTokenData().TxNormal.GetProof().GetOutputCoins()
			coinsToBeSaved = make([][]byte, 0)
			otas = make([][]byte, 0)
			for _, coin := range outputCoins {
				if coin.GetVersion() != 2 {
					continue
				}
				coinsToBeSaved = append(coinsToBeSaved, coin.Bytes())
				otas = append(otas, coin.GetPublicKey().ToBytesS())
			}
			err = statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, *tx.GetTokenID(), 0, coinsToBeSaved, otas, 0)

			err = statedb.StorePrivacyTokenTx(dummyDB,*tx.GetTokenID(), *tx.Hash())
			assert.Equal(t,nil,err)
			statedb.StoreCommitments(dummyDB,*tx.GetTokenID(), [][]byte{outputCoins[0].GetCommitment().ToBytesS()}, shardID)

			exists := statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
			// val,exists := tokenMap[*tx.GetTokenID()]
			fmt.Printf("Token created : %v\n",exists)
			assert.Equal(t, true, exists)

			transferAmount := uint64(69)

			paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[1].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

			prvCoinsToPayTransfer := make([]coin.PlainCoin,0)
			outerOuts := tx.GetTxBase().GetProof().GetOutputCoins()
			coinsToBeSaved = make([][]byte, 0)
			otas = make([][]byte, 0)
			for _, coin := range outerOuts {
				if coin.GetVersion() != 2 {
					continue
				}
				coinsToBeSaved = append(coinsToBeSaved, coin.Bytes())
				otas = append(otas, coin.GetPublicKey().ToBytesS())
			}
			err = statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, common.PRVCoinID, 0, coinsToBeSaved, otas, 0)
			tokenCoinsToTransfer := make([]coin.PlainCoin,0)
			for _,c := range outerOuts{
				pc,_ := c.Decrypt(keySets[0])
				// s,_ := json.Marshal(pc.(*coin.CoinV2))
				fmt.Printf("Tx Fee : %x has received %d in PRV\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
				prvCoinsToPayTransfer = append(prvCoinsToPayTransfer,pc)
			}
			for _,c := range outputCoins{
				pc,err := c.Decrypt(keySets[0])
				// s,_ := json.Marshal(pc.(*coin.CoinV2))
				fmt.Printf("Tx Token : %x has received %d in token T1\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
				assert.Equal(t,nil,err)
				tokenCoinsToTransfer = append(tokenCoinsToTransfer,pc)
			}
			// // token param for transfer token
			tokenParam2 := &CustomTokenPrivacyParamTx{
				PropertyID:     tx.GetTokenID().String(),
				PropertyName:   "Token 1",
				PropertySymbol: "T1",
				Amount:         transferAmount,
				TokenTxType:    CustomTokenTransfer,
				Receiver:       paymentInfo2,
				TokenInput:     tokenCoinsToTransfer,
				Mintable:       false,
				Fee:            0,
			}

			paramToCreateTx2 := NewTxPrivacyTokenInitParams(&keySets[0].PrivateKey,
				[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, dummyDB, nil,
				hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},nil)
				// paramToCreateTx := NewTxPrivacyTokenInitParams(&keySets[0].PrivateKey,
				// paymentInfoPRV, inputCoinsPRV, 10, tokenParam, dummyDB, nil,
				// hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},nil)

			// // init tx
			tx2 := &TxTokenVersion2{}
			err = tx2.Init(paramToCreateTx2)
			// fmt.Println(err)
			assert.Equal(t, nil, err)

			assert.Equal(t, true, bytes.Equal(msgCipherText, tx2.GetTxPrivacyTokenData().TxNormal.GetProof().GetOutputCoins()[0].GetInfo()))
			// outputCoins = tx2.GetTxPrivacyTokenData().TxNormal.GetProof().GetOutputCoins()
			// outerOuts = tx2.GetTxBase().GetProof().GetOutputCoins()
			// for _,c := range outerOuts{
			// 	pc,_ := c.Decrypt(keySets[0])
			// 	// s,_ := json.Marshal(pc.(*coin.CoinV2))
			// 	fmt.Printf("Transfer, outer out : %v - %d - supposedly in PRV\n",pc.GetPublicKey(),pc.GetValue())
			// }
			// for i,c := range outputCoins{
			// 	pc,err := c.Decrypt(keySets[1-i])
			// 	// s,_ := json.Marshal(pc.(*coin.CoinV2))
			// 	fmt.Printf("Transfer, out : %v - %d - supposedly in custom token\n",pc.GetPublicKey(),pc.GetValue())
			// 	assert.Equal(t,nil,err)
			// }

			err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
			assert.Equal(t, nil, err)

			isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
			assert.Equal(t, true, isValidSanity)
			assert.Equal(t, nil, err)

			isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
			// assert.Equal(t, true, isValidTxItself)
			// fmt.Println(err)
			// fmt.Println(isValidTxItself)
		}
	}
}