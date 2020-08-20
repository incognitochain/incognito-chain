package debugtool

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"io/ioutil"
	"os"
	"time"
)

// RandBytes generates random bytes with length
func RandBytes(length int) []byte {
	rbytes := make([]byte, length)
	rand.Read(rbytes)
	return rbytes
}

// RandIntInterval returns a random int in range [L; R]
func RandIntInterval(L, R int) int {
	length := R - L + 1
	r := common.RandInt() % length
	return L + r
}

func ParseOutputCoins(paymentInfo []*privacy.PaymentInfo) ([]*coin.CoinV1, error) {
	outputCoins := []*coin.CoinV1{}
	for _, pInfo := range paymentInfo {
		sndOut := privacy.RandomScalar()
		tmpCoin := new(coin.CoinV1)
		tmpCoin.CoinDetails = new(coin.PlainCoinV1)
		tmpCoin.CoinDetails.SetValue(pInfo.Amount)
		if len(pInfo.Message) > 0 {
			if len(pInfo.Message) > privacy.MaxSizeInfoCoin {
				return nil, errors.New("size of info is too big")
			}
		}
		tmpCoin.CoinDetails.SetInfo(pInfo.Message)

		PK, err := new(privacy.Point).FromBytesS(pInfo.PaymentAddress.Pk)
		if err != nil {
			return nil, errors.New("cannot parse public key")
		}
		tmpCoin.CoinDetails.SetPublicKey(PK)
		tmpCoin.CoinDetails.SetSNDerivator(sndOut)
		outputCoins = append(outputCoins, tmpCoin)
	}
	return outputCoins, nil
}

func ParseOutputCoinV2s(paymentInfo []*privacy.PaymentInfo) ([]*coin.CoinV2, error) {
	outputCoins := []*coin.CoinV2{}
	for _, pInfo := range paymentInfo {
		c, err := coin.NewCoinFromPaymentInfo(pInfo)
		if err != nil {
			return nil, err
		}
		outputCoins = append(outputCoins, c)
	}
	return outputCoins, nil
}

func ParseResponse(respondInBytes []byte) (*rpcserver.JsonResponse, error){
	var respond rpcserver.JsonResponse
	err := json.Unmarshal(respondInBytes, &respond)
	if err != nil {
		return nil, err
	}

	return &respond, nil
}

func CreateTxPrivacyInitParams(db *statedb.StateDB, keySet *incognitokey.KeySet, paymentAddress privacy.PaymentAddress, inputCoins []coin.PlainCoin, hasPrivacy bool, fee uint64, tokenID common.Hash) ([]*privacy.PaymentInfo, *transaction.TxPrivacyInitParams, error) {
	//initialize payment info of input coins
	paymentInfos := make([]*privacy.PaymentInfo, 1)
	sumAmount := uint64(0)
	for _, inputCoin := range inputCoins {
		sumAmount += inputCoin.GetValue()
	}

	paymentInfos[0] = key.InitPaymentInfo(paymentAddress, sumAmount - fee, []byte("I want to change this world"))

	//create privacyinitparam
	txPrivacyInitParam := transaction.NewTxPrivacyInitParams(&keySet.PrivateKey,
		paymentInfos,
		inputCoins,
		fee,
		hasPrivacy,
		db,
		&tokenID,
		nil,
		nil)

	return paymentInfos, txPrivacyInitParam, nil
}

func CreateCoinFromJSONOutcoin(jsonOutCoin jsonresult.OutCoin) (coin.PlainCoin, error) {
	var output coin.PlainCoin
	var keyImage, pubkey, cm *operation.Point
	var snd, randomness *operation.Scalar
	var info []byte
	var err error

	value, ok := math.ParseUint64(jsonOutCoin.Value)
	if !ok {
		return nil, errors.New("Cannot parse value")
	}

	if len(jsonOutCoin.KeyImage) == 0 {
		keyImage = nil
	} else {
		keyImageInBytes, err := DecodeBase58Check(jsonOutCoin.KeyImage)
		if err != nil {
			return nil, err
		}
		keyImage, err = new(operation.Point).FromBytesS(keyImageInBytes)
		if err != nil {
			return nil, err
		}
	}

	if len(jsonOutCoin.Commitment) == 0 {
		cm = nil
	} else {
		cmInbytes, err := DecodeBase58Check(jsonOutCoin.Commitment)
		if err != nil {
			return nil, err
		}
		cm, err = new(operation.Point).FromBytesS(cmInbytes)
		if err != nil {
			return nil, err
		}
	}

	if len(jsonOutCoin.PublicKey) == 0 {
		pubkey = nil
	} else {
		pubkeyInBytes, err := DecodeBase58Check(jsonOutCoin.PublicKey)
		if err != nil {
			return nil, err
		}
		pubkey, err = new(operation.Point).FromBytesS(pubkeyInBytes)
		if err != nil {
			return nil, err
		}
	}

	if len(jsonOutCoin.Randomness) == 0 {
		randomness = nil
	} else {
		randomnessInBytes, err := DecodeBase58Check(jsonOutCoin.Randomness)
		if err != nil {
			return nil, err
		}
		randomness = new(operation.Scalar).FromBytesS(randomnessInBytes)
	}

	if len(jsonOutCoin.SNDerivator) == 0 {
		snd = nil
	} else {
		sndInBytes, err := DecodeBase58Check(jsonOutCoin.SNDerivator)
		if err != nil {
			return nil, err
		}
		snd = new(operation.Scalar).FromBytesS(sndInBytes)
	}

	if len(jsonOutCoin.Info) == 0 {
		info = []byte{}
	} else {
		info, err = DecodeBase58Check(jsonOutCoin.Info)
		if err != nil {
			return nil, err
		}
	}

	switch jsonOutCoin.Version {
	case "1":
		output = new(coin.PlainCoinV1)
		tmpOutput, _ := output.(*coin.PlainCoinV1)
		tmpOutput.SetSNDerivator(snd)

	case "2":
		output = new(coin.CoinV2)
	}
	//output = new(coin.PlainCoinV1)
	//tmpOutput, _ := output.(*coin.PlainCoinV1)
	//tmpOutput.SetSNDerivator(snd)

	output.SetCommitment(cm)
	output.SetValue(value)
	output.SetKeyImage(keyImage)
	output.SetRandomness(randomness)
	output.SetPublicKey(pubkey)
	output.SetInfo(info)

	return output, nil
}

func DivideCoins(coins []coin.PlainCoin) ([]coin.PlainCoin, []coin.PlainCoin){
	coinV1s := []coin.PlainCoin{}
	coinV2s := []coin.PlainCoin{}

	for _, coin := range coins{
		if coin.GetVersion() == 1{
			coinV1s = append(coinV1s, coin)
		}else{
			coinV2s = append(coinV2s, coin)
		}
	}

	return coinV1s, coinV2s
}

func ChooseCoinsToSpend(coins []coin.PlainCoin, amount uint64, overbalanced bool) ([]coin.PlainCoin, error){
	sumValue := uint64(0)
	res := []coin.PlainCoin{}
	for _, coin := range coins {
		sumValue += coin.GetValue()

		if sumValue > amount{
			res = append(res, coin)
		}
	}
	if sumValue < amount && !overbalanced{
		return nil, errors.New("not enough coins to spend")
	}

	return res, nil
}

func CreateSampleCoin(privKey privacy.PrivateKey, pubKey *privacy.Point, amount uint64, msg []byte, version int) (coin.Coin, error) {
	if version == 1{
		c := new(coin.PlainCoinV1).Init()

		c.SetValue(amount)
		c.SetInfo(msg)
		c.SetPublicKey(pubKey)
		c.SetSNDerivator(privacy.RandomScalar())
		c.SetRandomness(privacy.RandomScalar())

		//Derive serial number from snDerivator
		c.SetKeyImage(new(privacy.Point).Derive(privacy.PedCom.G[0], new(privacy.Scalar).FromBytesS(privKey), c.GetSNDerivator()))

		//Create commitment
		err := c.CommitAll()

		if err != nil {
			return nil, err
		}

		cOut := new(coin.CoinV1).Init()
		cOut.CoinDetails = c

		return cOut, nil
	}else{
		c := new(coin.CoinV2).Init()

		c.SetValue(amount)
		c.SetInfo(msg)
		c.SetPublicKey(pubKey)
		c.SetRandomness(privacy.RandomScalar())

		keyImage, err := c.ParseKeyImageWithPrivateKey(privKey)
		if err != nil {
			return nil, err
		}
		c.SetKeyImage(keyImage)

		return c, nil
	}
}

func CreateSampleCoinsFromTotalAmount(senderSK privacy.PrivateKey, pubkey *privacy.Point, totalAmount uint64, numFeeInputs, version int) ([]coin.Coin, error) {
	coinList := []coin.Coin{}
	tmpAmount := totalAmount / uint64(numFeeInputs)
	for i := 0; i < numFeeInputs-1; i++ {
		amount := tmpAmount - uint64(RandIntInterval(0, int(tmpAmount)/2))
		coin, err := CreateSampleCoin(senderSK, pubkey, amount, nil, version)
		if err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
		totalAmount -= amount
	}
	coin, err := CreateSampleCoin(senderSK, pubkey, totalAmount, nil, version)
	if err != nil {
		return nil, err
	}
	coinList = append(coinList, coin)

	return coinList, nil
}

func CreateAndSaveCoins(numCoins, numEquals int, privKey privacy.PrivateKey, pubKey *privacy.Point, db *statedb.StateDB, version int, tokenID common.Hash) ([]coin.Coin, error) {
	shardID := byte(0)

	amount := uint64(numCoins * 1000)
	outCoins := []coin.Coin{}
	for i := 0; i < numEquals; i++ {
		coin, err := CreateSampleCoin(privKey, pubKey, 1000, nil, version)
		if err != nil {
			return nil, err
		}
		outCoins = append(outCoins, coin)
	}
	tmpOutCoins, err := CreateSampleCoinsFromTotalAmount(privKey, pubKey, amount, numCoins-numEquals, version)
	for _, coin := range tmpOutCoins {
		outCoins = append(outCoins, coin)
	}
	if err != nil {
		return nil, err
	}

	//save coins and commitment indices onto the database
	commitmentsToBeSaved := [][]byte{}
	coinsToBeSaved := [][]byte{}
	OTAToBeSaved := [][]byte{}
	for _, outCoin := range outCoins {
		coinsToBeSaved = append(coinsToBeSaved, outCoin.Bytes())
		commitmentsToBeSaved = append(commitmentsToBeSaved, outCoin.GetCommitment().ToBytesS())
		OTAToBeSaved = append(OTAToBeSaved, outCoin.GetPublicKey().ToBytesS())
	}
	if version == 1{
		SNDToBeSaved := [][]byte{}
		for _, outCoin := range outCoins {
			tmpCoin, ok := outCoin.(*coin.CoinV1)
			if !ok{
				return nil, errors.New("cannot parse CoinV1")
			}
			SNDToBeSaved = append(SNDToBeSaved, tmpCoin.GetSNDerivator().ToBytesS())

		}

		err = statedb.StoreSNDerivators(db, tokenID, SNDToBeSaved)
		if err != nil {
			return nil, err
		}

		err = statedb.StoreOutputCoins(db, tokenID, pubKey.ToBytesS(), coinsToBeSaved, shardID)
		if err != nil {
			return nil, err
		}
		err = statedb.StoreCommitments(db, tokenID, commitmentsToBeSaved, shardID)
		if err != nil {
			return nil, err
		}
	}else{
		err = statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, 0, coinsToBeSaved, OTAToBeSaved, shardID)
	}

	return outCoins, nil
}

func InitDatabase() (*statedb.StateDB, error){
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	wrapperDB := statedb.NewDatabaseAccessWarper(diskBD)
	db, err := statedb.NewWithPrefixTrie(emptyRoot, wrapperDB)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ParseIndicesAndCommitmentsFromJson(jsonRespondInBytes []byte)([]uint64, []uint64, []*operation.Point, error){
	var jsonRespond rpcserver.JsonResponse
	err := json.Unmarshal(jsonRespondInBytes, &jsonRespond)
	if err != nil{
		return nil, nil, nil, err
	}

	var msg json.RawMessage
	err = json.Unmarshal(jsonRespond.Result, &msg)
	if err != nil{
		return nil, nil, nil, err
	}

	var result jsonresult.RandomCommitmentResult
	err = json.Unmarshal(msg, &result)
	if err != nil{
		return nil, nil, nil, err
	}

	commitmentIndices := result.CommitmentIndices
	myCommitmentIndices := result.MyCommitmentIndexs

	commitments := make([]*operation.Point, len(result.Commitments))

	for i, cm := range result.Commitments{
		cmInbytes, err := DecodeBase58Check(cm)
		if err != nil {
			return nil, nil, nil, err
		}

		commitments[i], err = new(operation.Point).FromBytesS(cmInbytes)
		if err != nil {
			return nil, nil, nil, err
		}
	}


	return commitmentIndices, myCommitmentIndices, commitments, nil
}

func InitParam(tx metadata.Transaction, fee uint64, keySet *incognitokey.KeySet, version int8) {
	if version == 1{
		tmpTx, _ := tx.(*transaction.TxVersion1)
		tmpTx.Fee = fee
		tmpTx.Type = common.TxNormalType
		tmpTx.Metadata = nil
		tmpTx.PubKeyLastByteSender = keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		if tmpTx.LockTime == 0 {
			tmpTx.LockTime = time.Now().Unix()
		}
		tmpTx.Version = version
		tmpTx.Info = nil
	}else{
		tmpTx, _ := tx.(*transaction.TxVersion2)
		tmpTx.Fee = fee
		tmpTx.Type = common.TxNormalType
		tmpTx.Metadata = nil
		tmpTx.PubKeyLastByteSender = keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		if tmpTx.LockTime == 0 {
			tmpTx.LockTime = time.Now().Unix()
		}
		tmpTx.Version = version
		tmpTx.Info = nil
	}

}

func ProveAndSign(witness *zkp.PaymentWitness, paymentInfos []*privacy.PaymentInfo, tx *transaction.TxVersion1, keySet *incognitokey.KeySet) error {
	paymentProof, err := witness.Prove(true, paymentInfos)
	if err != nil {
		return err
	}
	tx.Proof = paymentProof
	sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
	err1 := tx.Sign(sigPrivate)
	if err1 != nil {
		return errors.New(err1.Error())
	}
	return nil
}

func (this *DebugTool) GetRandomCommitment(tokenID, paymentAddress string, input []coin.PlainCoin) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	coins := []jsonresult.OutCoin{}
	for _, outCoin := range input{
		coin := jsonresult.NewOutCoin(outCoin)
		coin.Info = "18WDn5zYiHy8kdFEtiVXsEGqhjeKJToeKoeNkh8ZZnRBagFG1b"
		coins = append(coins, coin)
	}

	r := new(rpcserver.JsonRequest)
	r.Id = ""
	r.Jsonrpc = "1.0"
	r.Method = "randomcommitments"

	r.Params = []interface{}{paymentAddress, coins, tokenID}

	rInBytes, err := json.MarshalIndent(r, "\t", "\t")
	if err != nil {
		return nil, err
	}

	query := string(rInBytes)

	//fmt.Println(query)
	//return nil, nil
	return this.SendPostRequestWithQuery(query)
}

func (tool *DebugTool) InitPaymentWitness(tokenIDString string, senderPaymentAddress string, inputCoins []coin.PlainCoin, paymentInfos []*privacy.PaymentInfo, keySet *incognitokey.KeySet, fee uint64) (*zkp.PaymentWitness, error) {
	//Get random commitments to create one-of-many proofs
	jsonRespondInBytes, err := tool.GetRandomCommitment(tokenIDString, senderPaymentAddress, inputCoins)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(jsonRespondInBytes))

	commitmentIndices, myCommitmentIndices, commitments, err := ParseIndicesAndCommitmentsFromJson(jsonRespondInBytes)
	if err != nil {
		return nil, err
	}



	outputCoins, err := ParseOutputCoins(paymentInfos)
	if err != nil {
		return nil, err
	}

	// PrepareTransaction witness for proving
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              true,
		PrivateKey:              new(operation.Scalar).FromBytesS(keySet.PrivateKey),
		InputCoins:              inputCoins,
		OutputCoins:             outputCoins,
		PublicKeyLastByteSender: keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1],
		Commitments:             commitments,
		CommitmentIndices:       commitmentIndices,
		MyCommitmentIndices:     myCommitmentIndices,
		Fee:                     fee,
	}

	witness := new(zkp.PaymentWitness)
	err1 := witness.Init(paymentWitnessParam)
	if err1 != nil {
		return nil, err1
	}
	return witness, nil
}

func (tool *DebugTool) PrepareTransaction(privateKey string, tokenIDString string) (*wallet.KeyWallet, string, *operation.Point, []coin.PlainCoin, []coin.PlainCoin, error) {
	listOutputCoins, err := tool.GetPlainOutputCoin(privateKey, tokenIDString)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}

	keySet := &keyWallet.KeySet
	err = keySet.InitFromPrivateKey(&keySet.PrivateKey)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}

	senderPaymentAddress := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)

	pubkey, err := new(privacy.Point).FromBytesS(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}

	if err != nil {
		return nil, "", nil, nil, nil, err
	}

	coinV1s, coinV2s := DivideCoins(listOutputCoins)
	return keyWallet, senderPaymentAddress, pubkey, coinV1s, coinV2s, nil
}

func (tool *DebugTool) GetPlainOutputCoin(privateKey, tokenID string) ([]coin.PlainCoin, error) {
	outputCoins := []coin.PlainCoin{}
	respondInBytes, err := tool.GetListUnspentOutputTokens(privateKey, tokenID)
	if err != nil {
		return nil, err
	}

	respond, err := ParseResponse(respondInBytes)
	if err !=nil{
		return nil, err
	}

	var msg json.RawMessage
	err = json.Unmarshal(respond.Result, &msg)
	if err != nil {
		return nil, err
	}

	var tmp jsonresult.ListOutputCoins
	err = json.Unmarshal(msg, &tmp)
	if err != nil {
		return nil, err
	}

	listOutputCoins := tmp.Outputs
	for _, value := range listOutputCoins {
		for _, outcoin := range value {
			out, err := CreateCoinFromJSONOutcoin(outcoin)
			if err != nil {
				return nil, err
			}
			outputCoins = append(outputCoins, out)
		}
	}

	return outputCoins, nil
}

func (tool *DebugTool) InitTxVer1(tx *transaction.TxVersion1, keyWallet *wallet.KeyWallet, paymentString string, inputCoins []coin.PlainCoin, amount uint64, tokenIDString string) error {
	senderPaymentAddress := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	keySet := &keyWallet.KeySet

	fee := uint64(100)

	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
	if err != nil {
		return err
	}

	paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, common.PRVCoinID)
	if err != nil {
		return err
	}
	paymentInfos[0].Amount = amount - fee

	InitParam(tx, fee, keySet, 1)

	witness, err1 := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
	if err1 != nil {
		return err1
	}

	err2 := ProveAndSign(witness, paymentInfos, tx, keySet)
	if err2 != nil {
		return err2
	}
	return nil
}
