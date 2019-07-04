package mempool

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

var (
	db     database.DatabaseInterface
	bc     = &blockchain.BlockChain{}
	pbMempool     = pubsub.NewPubSubManager()
	tp = &TxPool{}
	feeEstimator = make(map[byte]*FeeEstimator)
	privateKeyShard0 = []string{
		"112t8rqdy2bgV3kf9qb8eso8jJkgEw1RKSTqtxRNoGobZtK7YeJfzE4rPX1uYZynzP6Ym5EMjEUMGGdgeGH1pxryCU22QmtgxoMPLyaaP1J8",
		"112t8rrGixbjxd7Fh8NoECAqX6mfgjkMRDygcejkXt8NCqZVU7BjFNRaDMjdGao5KRiRg7Dn7gQdsYrXLzz5yxsryTUNLWkq9GaSyMGYKxtT",
		"112t8rxr9EZUQtW2q5om7CfDZyJNF9bNWKyjYAxbsk6SrKRi4QzXLX1SabamCZ1TBJCJNvB98CNQuPLxo7fQvVVctmsBF282FBwZWtfsuRU5",
		"112t8rqHziexNp48PRHtnqASEAchfRaWM2QTtk9eBbaqCZdUMZ4LHgAesBW7AfPKAc97mn7smoGr8SKiiXKmuvaHDKNJYK2zT7oAHDVvpXmc",
		"112t8rsCuDdsPecRrinj5n23onjKaCanM4JTUUyiU2rgjAL3nhEJH7VX1TYazxdWnvBudQvCvEjfhJ4hVjrdAqVK1s3a8fecmYXd8HWNHitC",
	}
	stakingPublicKey = "151vzKx6AaQs8Jw5Q8PefGSPu3E16w2E2tSRXd1tEyM1qUA4H1r" // public key of 112t8rsCuDdsPecRrinj5n23onjKaCanM4JTUUyiU2rgjAL3nhEJH7VX1TYazxdWnvBudQvCvEjfhJ4hVjrdAqVK1s3a8fecmYXd8HWNHitC
	receiverPaymentAddress1 = "1Uv34F64ktQkX1eyd6YEG8KTENV8W5w48LRsi6oqqxVm65uvcKxEAzL2dp5DDJTqAQA7HANfQ1enKXCh2EvVdvBftko6GtGnjSZ1KqJhi"
	receiverPaymentAddress2 = "1Uv2wgU5FR5jjeN3uY3UJ4SYYyjqj97spYBEDa6cTLGiP3w6BCY7mqmASKwXz8hXfLr6mpDjhWDJ8TiM5v5U5f2cxxqCn5kwy5JM9wBgi"
	commonFee = int64(10)
	noFee = int64(0)
	defaultTokenParams = make(map[string]interface{})
	defaultTokenReceiver = make(map[string]interface{})
)
var _ = func() (_ struct{}) {
	for i:=0; i< 255; i++ {
		shardID := byte(i)
		feeEstimator[shardID] = NewFeeEstimator(
			DefaultEstimateFeeMaxRollback,
			DefaultEstimateFeeMinRegisteredBlocks,
			1)
	}
	db, err = database.Open("leveldb", filepath.Join("./", "./testdatabase/"))
	if err != nil {
		fmt.Println("Could not open database connection")
		return
	}
	err = bc.InitForTest(&blockchain.Config{
		DataBase: db,
		PubSubManager: pbMempool,
		ChainParams: &blockchain.ChainTestParam,
	})
	if err != nil {
		panic("Could not init blockchain")
	}
	tp.Init(&Config{
		DataBase: db,
		BlockChain: bc,
		PubSubManager: pbMempool,
		IsLoadFromMempool: false,
		PersistMempool: false,
		FeeEstimator: feeEstimator,
		ChainParams: &blockchain.ChainTestParam,
	})
	var transactions []metadata.Transaction
	for _, privateKey := range privateKeyShard0 {
		txs := initTx("3000000", privateKey, db)
		transactions = append(transactions, txs[0])
	}
	err = tp.config.BlockChain.CreateAndSaveTxViewPointFromBlock(&blockchain.ShardBlock{
		Header: blockchain.ShardHeader{ ShardID: 0},
		Body: blockchain.ShardBody{
			Transactions: transactions,
		},
	})
	if err != nil {
		fmt.Println("Can not fetch transaction")
		return
	}
	defaultTokenParams["TokenID"] = ""
	defaultTokenParams["TokenName"] = "ABCD123"
	defaultTokenParams["TokenSymbol"] = "ABCDF123"
	defaultTokenParams["TokenAmount"] = float64(1000)
	defaultTokenParams["TokenTxType"] = float64(0)
	defaultTokenReceiver[receiverPaymentAddress1] = float64(1000)
	defaultTokenParams["TokenReceivers"] = defaultTokenReceiver
	// token id: 6efff7b815f2890758f55763c53c4563feada766726ea4c08fe04dba8fd11b89
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	privacy.Logger.Init(common.NewBackend(nil).Logger("test", true))
	transaction.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func ResetMempoolTest()  {
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.poolSerialNumbersHashH = make(map[common.Hash][]common.Hash)
	tp.TokenIDPool = make(map[common.Hash]string)
	tp.CandidatePool = make(map[common.Hash]string)
	tp.DuplicateTxs = make(map[common.Hash]uint64)
	tp.config.RoleInCommittees = -1
	tp.IsBlockGenStarted = false
	tp.IsUnlockMempool = true
	_, subChanRole, _ := tp.config.PubSubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
	tp.config.RoleInCommitteesEvent = subChanRole
	tp.IsTest = false
}
func initTx(amount string, privateKey string, db database.DatabaseInterface) []metadata.Transaction {
	var initTxs []metadata.Transaction
	var initAmount, _ = strconv.Atoi(amount) // amount init
	testUserkeyList := []string{
		privateKey,
	}
	for _, val := range testUserkeyList {
		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testUserKey.KeySet.ImportFromPrivateKey(&testUserKey.KeySet.PrivateKey)
		testSalaryTX := transaction.Tx{}
		testSalaryTX.InitTxSalary(uint64(initAmount), &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			db,
			nil,
		)
		initTxs = append(initTxs, &testSalaryTX)
	}
	return initTxs
}
// chooseBestOutCoinsToSpent returns list of unspent coins for spending with amount
func chooseBestOutCoinsToSpent(outCoins []*privacy.OutputCoin, amount uint64) (resultOutputCoins []*privacy.OutputCoin, remainOutputCoins []*privacy.OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*privacy.OutputCoin, 0)
	remainOutputCoins = make([]*privacy.OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)
	
	// either take the smallest coins, or a single largest one
	var outCoinOverLimit *privacy.OutputCoin
	outCoinsUnderLimit := make([]*privacy.OutputCoin, 0)
	
	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value < amount {
			outCoinsUnderLimit = append(outCoinsUnderLimit, outCoin)
		} else if outCoinOverLimit == nil {
			outCoinOverLimit = outCoin
		} else if outCoinOverLimit.CoinDetails.Value > outCoin.CoinDetails.Value {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
			outCoinOverLimit = outCoin
		}
	}
	
	sort.Slice(outCoinsUnderLimit, func(i, j int) bool {
		return outCoinsUnderLimit[i].CoinDetails.Value < outCoinsUnderLimit[j].CoinDetails.Value
	})
	
	for _, outCoin := range outCoinsUnderLimit {
		if totalResultOutputCoinAmount < amount {
			totalResultOutputCoinAmount += outCoin.CoinDetails.Value
			resultOutputCoins = append(resultOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	
	if outCoinOverLimit != nil && (outCoinOverLimit.CoinDetails.Value > 2*amount || totalResultOutputCoinAmount < amount) {
		remainOutputCoins = append(remainOutputCoins, resultOutputCoins...)
		resultOutputCoins = []*privacy.OutputCoin{outCoinOverLimit}
		totalResultOutputCoinAmount = outCoinOverLimit.CoinDetails.Value
	} else if outCoinOverLimit != nil {
		remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
	}
	
	if totalResultOutputCoinAmount < amount {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, errors.New("Not enough coin")
	} else {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
	}
}

func CreateAndSaveTestNormalTransaction(privateKey string, fee int64, hasPrivacyCoin bool) metadata.Transaction{
	// get sender key set from private key
	senderKeySet, _ := wallet.Base58CheckDeserialize(privateKey)
	senderKeySet.KeySet.ImportFromPrivateKey(&senderKeySet.KeySet.PrivateKey)
	lastByte := senderKeySet.KeySet.PaymentAddress.Pk[len(senderKeySet.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	
	receiversPaymentAddressStrParam := make(map[string]interface{})
	receiversPaymentAddressStrParam[receiverPaymentAddress2] = 50
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, _ := wallet.Base58CheckDeserialize(paymentAddressStr)
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(int)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}
	estimateFeeCoinPerKb := fee
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outCoins, err := tp.config.BlockChain.GetListOutputCoinsByKeyset(&senderKeySet.KeySet, shardIDSender, prvCoinID)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	remainOutputCoins := make([]*privacy.OutputCoin, 0)
	for _, outCoin := range outCoins {
		if tp.ValidateSerialNumberHashH(outCoin.CoinDetails.SerialNumber.Compress()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		fmt.Println("Can't create transaction")
		return nil
	}
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, paymentInfos, hasPrivacyCoin, nil, nil, nil, 1)
	realFee := uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err := chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				fmt.Println("Can't create transaction", err)
				return nil
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	tx := transaction.Tx{}
	err1 := tx.Init(
		&senderKeySet.KeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacyCoin,
		db,
		nil, // use for prv coin -> nil is valid
		nil,
	)
	if err1 != nil {
		panic("no tx found")
	}
	return &tx
}

func CreateAndSaveTestStakingTransaction(privateKey string, fee int64, isBeacon bool) metadata.Transaction{
	// get sender key set from private key
	hasPrivacyCoin := false
	senderKeySet, _ := wallet.Base58CheckDeserialize(privateKey)
	senderKeySet.KeySet.ImportFromPrivateKey(&senderKeySet.KeySet.PrivateKey)
	lastByte := senderKeySet.KeySet.PaymentAddress.Pk[len(senderKeySet.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	
	receiversPaymentAddressStrParam := make(map[string]interface{})
	if isBeacon {
		receiversPaymentAddressStrParam[common.BurningAddress] = tp.config.ChainParams.StakingAmountShard * 3
	} else {
		receiversPaymentAddressStrParam[common.BurningAddress] = tp.config.ChainParams.StakingAmountShard
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, _ := wallet.Base58CheckDeserialize(paymentAddressStr)
		paymentInfo := &privacy.PaymentInfo{
			Amount:         amount.(uint64),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}
	estimateFeeCoinPerKb := fee
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outCoins, err := tp.config.BlockChain.GetListOutputCoinsByKeyset(&senderKeySet.KeySet, shardIDSender, prvCoinID)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	remainOutputCoins := make([]*privacy.OutputCoin, 0)
	for _, outCoin := range outCoins {
		if tp.ValidateSerialNumberHashH(outCoin.CoinDetails.SerialNumber.Compress()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		fmt.Println("Can't create transaction")
		return nil
	}
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	paymentAddress, _ := senderKeySet.Serialize(wallet.PaymentAddressType)
	var stakingMetadata *metadata.StakingMetadata
	if isBeacon {
		stakingMetadata, _ = metadata.NewStakingMetadata(64, base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte), tp.config.ChainParams.StakingAmountShard)
	} else {
		stakingMetadata, _ = metadata.NewStakingMetadata(63, base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte), tp.config.ChainParams.StakingAmountShard)
	}
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, paymentInfos, hasPrivacyCoin, stakingMetadata, nil, nil, 1)
	realFee := uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err := chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				fmt.Println("Can't create transaction", err)
				return nil
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	tx := transaction.Tx{}
	err1 := tx.Init(
		&senderKeySet.KeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacyCoin,
		db,
		nil, // use for prv coin -> nil is valid
		stakingMetadata,
	)
	if err1 != nil {
		panic("no tx found")
	}
	return &tx
}
func CreateAndSaveTestInitCustomTokenTransaction(privateKey string, fee int64, tokenParamsRaw map[string]interface{}) metadata.Transaction{
	var hasPrivacyCoin = false
	// get sender key set from private key
	senderKeySet, _ := wallet.Base58CheckDeserialize(privateKey)
	senderKeySet.KeySet.ImportFromPrivateKey(&senderKeySet.KeySet.PrivateKey)
	lastByte := senderKeySet.KeySet.PaymentAddress.Pk[len(senderKeySet.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	
	receiversPaymentAddressStrParam := make(map[string]interface{})
	receiversPaymentAddressStrParam[receiverPaymentAddress2] = 50
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, _ := wallet.Base58CheckDeserialize(paymentAddressStr)
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(int)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}
	estimateFeeCoinPerKb := fee
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outCoins, err := tp.config.BlockChain.GetListOutputCoinsByKeyset(&senderKeySet.KeySet, shardIDSender, prvCoinID)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	remainOutputCoins := make([]*privacy.OutputCoin, 0)
	for _, outCoin := range outCoins {
		if tp.ValidateSerialNumberHashH(outCoin.CoinDetails.SerialNumber.Compress()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		fmt.Println("Can't create transaction")
		return nil
	}
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		fmt.Println("Can't create transaction", err)
		return nil
	}
	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:     tokenParamsRaw["TokenID"].(string),
		PropertyName:   tokenParamsRaw["TokenName"].(string),
		PropertySymbol: tokenParamsRaw["TokenSymbol"].(string),
		TokenTxType:    int(tokenParamsRaw["TokenTxType"].(float64)),
		Amount:         uint64(tokenParamsRaw["TokenAmount"].(float64)),
	}
	tokenParams.Receiver, _ = transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"])
	
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, paymentInfos, hasPrivacyCoin, nil, tokenParams, nil, 1)
	realFee := uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err := chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				fmt.Println("Can't create transaction", err)
				return nil
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	tx := &transaction.TxCustomToken{}
	err1 := tx.Init(
		&senderKeySet.KeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		db,
		nil,
		hasPrivacyCoin,
		shardIDSender,
	)
	if err1 != nil {
		panic("no tx found")
	}
	return tx
}
func TestTxPoolGetTxsInMem(t *testing.T) {
	ResetMempoolTest()
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], commonFee,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], commonFee,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], commonFee,false)
	txDesc1 := createTxDescMempool(tx1, 1, uint64(commonFee), 0)
	txDesc2 := createTxDescMempool(tx2, 1, uint64(commonFee), 0)
	txDesc3 := createTxDescMempool(tx3, 1, uint64(commonFee), 0)
	tp.pool[*tx1.Hash()] = txDesc1
	tp.pool[*tx2.Hash()] = txDesc2
	tp.pool[*tx3.Hash()] = txDesc3
	txs := tp.GetTxsInMem()
	if len(txs) != 3 {
		t.Fatalf("Expect 3 transaction from mempool but get %+v", len(txs))
	}
}
func TestTxPoolGetSerialNumbersHashH(t *testing.T) {
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], commonFee,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], commonFee,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], commonFee,false)
	tp.poolSerialNumbersHashH[*tx1.Hash()] = tx1.ListSerialNumbersHashH()
	tp.poolSerialNumbersHashH[*tx2.Hash()] = tx2.ListSerialNumbersHashH()
	tp.poolSerialNumbersHashH[*tx3.Hash()] = tx3.ListSerialNumbersHashH()
	serialNumberList := tp.GetSerialNumbersHashH()
	if !reflect.DeepEqual(serialNumberList, tp.poolSerialNumbersHashH) {
		t.Fatalf("Something wrong with return serial list")
	}
}
func TestTxPoolIsTxInPool(t *testing.T) {
	ResetMempoolTest()
	tx := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], commonFee,false)
	if tp.isTxInPool(tx.Hash()) {
		t.Fatalf("Expect %+v to be NOT in pool", *tx.Hash())
	}
	txDesc := createTxDescMempool(tx, 1, uint64(commonFee), 0)
	tp.pool[*tx.Hash()] = txDesc
	tp.poolSerialNumbersHashH[*tx.Hash()] = tx.ListSerialNumbersHashH()
	if !tp.isTxInPool(tx.Hash()) {
		t.Fatalf("Expect %+v to be in pool", *tx.Hash())
	}
}
func TestTxPoolAddTx(t *testing.T) {
	ResetMempoolTest()
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], 10,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], 10,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], 10,false)
	txDesc1 := createTxDescMempool(tx1, 1, 10, 0)
	txDesc2 := createTxDescMempool(tx2, 1, 10, 0)
	txDesc3 := createTxDescMempool(tx3, 1, 10, 0)
	tp.addTx(txDesc1, false)
	tp.addTx(txDesc2, false)
	tp.addTx(txDesc3, false)
	if len(tp.pool) != 3 {
		t.Fatalf("Expect 3 transaction from mempool but get %+v", len(tp.pool))
	}
	if len(tp.poolSerialNumbersHashH) != 3 {
		t.Fatalf("Expect 3 transaction from mempool but get %+v", len(tp.poolSerialNumbersHashH))
	}
}
func TestTxPoolValidateTransaction(t *testing.T) {
	ResetMempoolTest()
	salaryTx := initTx("100", privateKeyShard0[0], db)
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], commonFee,false)
	tx1DS := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], commonFee + 5,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], commonFee,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], commonFee,false)
	tx4 := CreateAndSaveTestNormalTransaction(privateKeyShard0[3], noFee,false)
	tx5 := CreateAndSaveTestNormalTransaction(privateKeyShard0[4], commonFee,false)
	txInitCustomToken := CreateAndSaveTestInitCustomTokenTransaction(privateKeyShard0[3], commonFee, defaultTokenParams)
	txInitCustomTokenFailed := CreateAndSaveTestInitCustomTokenTransaction(privateKeyShard0[4], commonFee, defaultTokenParams)
	txStakingShard := CreateAndSaveTestStakingTransaction(privateKeyShard0[4], commonFee, false)
	txStakingBeacon := CreateAndSaveTestStakingTransaction(privateKeyShard0[4], commonFee, true)
	txDesc1 := createTxDescMempool(tx1, 1, uint64(commonFee), 0)
	// Check condition 1: Sanity - Max version error
	ResetMempoolTest()
	tx1.(*transaction.Tx).Version = 2
	err1 := tp.validateTransaction(tx1)
	if err1 == nil {
		t.Fatal("Expect max version error error but no error")
	} else {
		if err1.(MempoolTxError).Code != ErrCodeMessage[RejectSansityTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSansityTx], err)
		}
	}
	tx1.(*transaction.Tx).Version = 1
	// Check condition 1: Size - Invalid size error
	ResetMempoolTest()
	common.MaxTxSize = 0
	common.MaxBlockSize = 2000
	err2 := tp.validateTransaction(tx2)
	if err2 == nil {
		t.Fatal("Expect size error error but no error")
	} else {
		if err2.(MempoolTxError).Code != ErrCodeMessage[RejectSansityTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSansityTx], err)
		}
	}
	common.MaxTxSize = 100
	common.MaxBlockSize = 2000
	// Check Condition 1: Sanity Validate type
	ResetMempoolTest()
	tx3.(*transaction.Tx).Type = "abc"
	err3 := tp.validateTransaction(tx3)
	if err3 == nil {
		t.Fatal("Expect type error error but no error")
	} else {
		if err3.(MempoolTxError).Code != ErrCodeMessage[RejectSansityTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSansityTx], err)
		}
	}
	tx3.(*transaction.Tx).Type = common.TxNormalType
	// Check Condition 1: Sanity Validate type
	ResetMempoolTest()
	tempLockTime := tx4.(*transaction.Tx).LockTime
	tx4.(*transaction.Tx).LockTime = time.Now().Unix() + 1000000
	err4 := tp.validateTransaction(tx4)
	if err4 == nil {
		t.Fatal("Expect type error error but no error")
	} else {
		if err4.(MempoolTxError).Code != ErrCodeMessage[RejectSansityTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSansityTx], err)
		}
	}
	tx4.(*transaction.Tx).LockTime = tempLockTime
	// Check Condition 1: Sanity Validate Info Length
	ResetMempoolTest()
	tempByte := []byte{}
	for i:=0;i<514;i++{
		tempByte = append(tempByte, byte(i))
	}
	tx4.(*transaction.Tx).Info = tempByte
	err5 := tp.validateTransaction(tx4)
	if err5 == nil {
		t.Fatal("Expect type error error but no error")
	} else {
		if err5.(MempoolTxError).Code != ErrCodeMessage[RejectSansityTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSansityTx], err)
		}
	}
	tx4.(*transaction.Tx).Info = []byte{}
	// Check condition 2: tx exist in pool
	tp.pool[*tx1.Hash()] = txDesc1
	tp.poolSerialNumbersHashH[*tx1.Hash()] = tx1.ListSerialNumbersHashH()
	err6 := tp.validateTransaction(tx1)
	if err6 == nil {
		t.Fatal("Expect reject duplicate error but no error")
	} else {
		if err6.(MempoolTxError).Code != ErrCodeMessage[RejectDuplicateTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDuplicateTx], err)
		}
	}
	// Check Condition 3: Salary Transaction
	ResetMempoolTest()
	err7 := tp.validateTransaction(salaryTx[0])
	if err7 == nil {
		t.Fatal("Expect salary error error but no error")
	} else {
		if err7.(MempoolTxError).Code != ErrCodeMessage[RejectSalaryTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectSalaryTx], err)
		}
	}
	// Check Condition 4: Validate fee
	ResetMempoolTest()
	err8 := tp.validateTransaction(tx4)
	if err8 == nil {
		t.Fatal("Expect fee error error but no error")
	} else {
		if err8.(MempoolTxError).Code != ErrCodeMessage[RejectInvalidFee].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectInvalidFee], err)
		}
	}
	tx5.(*transaction.Tx).Type = common.TxNormalType
	// Check Condition 5: Double spend
	ResetMempoolTest()
	tp.addTx(txDesc1, false)
	err9 := tp.validateTransaction(tx1DS)
	if err9 == nil {
		t.Fatal("Expect double spend error error but no error")
	} else {
		if err9.(MempoolTxError).Code != ErrCodeMessage[RejectDoubleSpendWithMempoolTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDoubleSpendWithMempoolTx], err)
		}
	}
	// Check Condition 5: Check double spend with mempool
	ResetMempoolTest()
	tp.addTx(txDesc1, false)
	err10 := tp.validateTransaction(tx1DS)
	if err10 == nil {
		t.Fatal("Expect double spend in mempool error error but no error")
	} else {
		if err10.(MempoolTxError).Code != ErrCodeMessage[RejectDoubleSpendWithMempoolTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDoubleSpendWithMempoolTx], err)
		}
	}
	// check Condition 7: Check double spend with blockchain
	ResetMempoolTest()
	err = tp.config.BlockChain.CreateAndSaveTxViewPointFromBlock(&blockchain.ShardBlock{
		Header: blockchain.ShardHeader{ ShardID: 0},
		Body: blockchain.ShardBody{
			Transactions: []metadata.Transaction{tx1},
		},
	})
	if err != nil {
		t.Fatalf("Expect no error but get %+v", err)
	}
	err11 := tp.validateTransaction(tx1)
	if err11 == nil {
		t.Fatal("Expect double spend with blockchain error error but no error")
	} else {
		if err11.(MempoolTxError).Code != ErrCodeMessage[RejectDoubleSpendWithBlockchainTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDoubleSpendWithBlockchainTx], err)
		}
	}
	// check Condition 8: Check Init Custom Token
	ResetMempoolTest()
	tp.TokenIDPool[*txInitCustomToken.Hash()] = "6efff7b815f2890758f55763c53c4563feada766726ea4c08fe04dba8fd11b89"
	err12 := tp.validateTransaction(txInitCustomTokenFailed)
	if err12 == nil {
		t.Fatal("Expect duplicate init token error error but no error")
	} else {
		if err12.(MempoolTxError).Code != ErrCodeMessage[RejectDuplicateInitTokenTx].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDuplicateInitTokenTx], err)
		}
	}
	// check Condition 9: Check Init Custom Token
	ResetMempoolTest()
	tp.CandidatePool[*txStakingShard.Hash()] = stakingPublicKey
	err13 := tp.validateTransaction(txStakingShard)
	if err13 == nil {
		t.Fatal("Expect duplicate staking pubkey error error but no error")
	} else {
		if err13.(MempoolTxError).Code != ErrCodeMessage[RejectDuplicateStakePubkey].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDuplicateStakePubkey], err)
		}
	}
	err13 = tp.validateTransaction(txStakingBeacon)
	if err13 == nil {
		t.Fatal("Expect duplicate staking pubkey error error but no error")
	} else {
		if err13.(MempoolTxError).Code != ErrCodeMessage[RejectDuplicateStakePubkey].Code {
			t.Fatalf("Expect Error %+v but get %+v", ErrCodeMessage[RejectDuplicateStakePubkey], err)
		}
	}
	ResetMempoolTest()
	// Pass all case
	err14 := tp.validateTransaction(txStakingShard)
	if err14 != nil {
		t.Fatal("Expect no err but get ", err14)
	}
	err14 = tp.validateTransaction(tx3)
	if err14 != nil {
		t.Fatal("Expect no err but get ", err14)
	}
	err14 = tp.validateTransaction(txInitCustomToken)
	if err14 != nil {
		t.Fatal("Expect no err but get ", err14)
	}
}