package mempool

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
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
)

var (
	db     database.DatabaseInterface
	bc     = &blockchain.BlockChain{}
	pbMempool     = pubsub.NewPubSubManager()
	tp = &TxPool{}
	privateKeyShard0 = []string{
		"112t8rqdy2bgV3kf9qb8eso8jJkgEw1RKSTqtxRNoGobZtK7YeJfzE4rPX1uYZynzP6Ym5EMjEUMGGdgeGH1pxryCU22QmtgxoMPLyaaP1J8",
		"112t8rrGixbjxd7Fh8NoECAqX6mfgjkMRDygcejkXt8NCqZVU7BjFNRaDMjdGao5KRiRg7Dn7gQdsYrXLzz5yxsryTUNLWkq9GaSyMGYKxtT",
		"112t8rxr9EZUQtW2q5om7CfDZyJNF9bNWKyjYAxbsk6SrKRi4QzXLX1SabamCZ1TBJCJNvB98CNQuPLxo7fQvVVctmsBF282FBwZWtfsuRU5",
		"112t8rqHziexNp48PRHtnqASEAchfRaWM2QTtk9eBbaqCZdUMZ4LHgAesBW7AfPKAc97mn7smoGr8SKiiXKmuvaHDKNJYK2zT7oAHDVvpXmc",
		"112t8rsCuDdsPecRrinj5n23onjKaCanM4JTUUyiU2rgjAL3nhEJH7VX1TYazxdWnvBudQvCvEjfhJ4hVjrdAqVK1s3a8fecmYXd8HWNHitC",
	}
	receiverPaymentAddress1 = "1Uv34F64ktQkX1eyd6YEG8KTENV8W5w48LRsi6oqqxVm65uvcKxEAzL2dp5DDJTqAQA7HANfQ1enKXCh2EvVdvBftko6GtGnjSZ1KqJhi"
	receiverPaymentAddress2 = "1Uv2wgU5FR5jjeN3uY3UJ4SYYyjqj97spYBEDa6cTLGiP3w6BCY7mqmASKwXz8hXfLr6mpDjhWDJ8TiM5v5U5f2cxxqCn5kwy5JM9wBgi"
	commonFee = 10
	noFee = 0
	feeEstimator = make(map[byte]*FeeEstimator)
)
var _ = func() (_ struct{}) {
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
	})
	var transactions []metadata.Transaction
	for _, privateKey := range privateKeyShard0 {
		txs := initTx("1000", privateKey, db)
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
	for shardID, _ := range feeEstimator {
		feeEstimator[shardID] = NewFeeEstimator(
			DefaultEstimateFeeMaxRollback,
			DefaultEstimateFeeMinRegisteredBlocks,
			1)
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	privacy.Logger.Init(common.NewBackend(nil).Logger("test", true))
	transaction.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()



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
func TestTxPoolGetTxsInMem(t *testing.T) {
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], 10,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], 10,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], 10,false)
	txDesc1 := createTxDescMempool(tx1, 1, 10, 0)
	txDesc2 := createTxDescMempool(tx2, 1, 10, 0)
	txDesc3 := createTxDescMempool(tx3, 1, 10, 0)
	tp.pool[*tx1.Hash()] = txDesc1
	tp.pool[*tx2.Hash()] = txDesc2
	tp.pool[*tx3.Hash()] = txDesc3
	txs := tp.GetTxsInMem()
	if len(txs) != 3 {
		t.Fatalf("Expect 3 transaction from mempool but get %+v", len(txs))
	}
}
func TestTxPoolGetSerialNumbersHashH(t *testing.T) {
	tx1 := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], 10,false)
	tx2 := CreateAndSaveTestNormalTransaction(privateKeyShard0[1], 10,false)
	tx3 := CreateAndSaveTestNormalTransaction(privateKeyShard0[2], 10,false)
	tp.poolSerialNumbersHashH[*tx1.Hash()] = tx1.ListSerialNumbersHashH()
	tp.poolSerialNumbersHashH[*tx2.Hash()] = tx2.ListSerialNumbersHashH()
	tp.poolSerialNumbersHashH[*tx3.Hash()] = tx3.ListSerialNumbersHashH()
	serialNumberList := tp.GetSerialNumbersHashH()
	if !reflect.DeepEqual(serialNumberList, tp.poolSerialNumbersHashH) {
		t.Fatalf("Something wrong with return serial list")
	}
}
func TestTxPoolIsTxInPool(t *testing.T) {
	tx := CreateAndSaveTestNormalTransaction(privateKeyShard0[0], 10,false)
	if tp.isTxInPool(tx.Hash()) {
		t.Fatalf("Expect %+v to be NOT in pool", *tx.Hash())
	}
	txDesc := createTxDescMempool(tx, 1, 10, 0)
	tp.pool[*tx.Hash()] = txDesc
	tp.poolSerialNumbersHashH[*tx.Hash()] = tx.ListSerialNumbersHashH()
	if !tp.isTxInPool(tx.Hash()) {
		t.Fatalf("Expect %+v to be in pool", *tx.Hash())
	}
}