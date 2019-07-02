package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"log"
	"math/rand"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

/*
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

*/
func (httpServer *HttpServer) handleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.ListAccounts{
		Accounts:   make(map[string]uint64),
		WalletName: httpServer.config.Wallet.Name,
	}
	accounts := httpServer.config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		prvCoinID.SetBytes(common.PRVCoinID[:])
		outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		amount := uint64(0)
		for _, out := range outCoins {
			amount += out.CoinDetails.Value
		}
		result.Accounts[accountName] = amount
	}

	return result, nil
}

/*
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
*/
func (httpServer *HttpServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramTemp, ok := params.(string)
	if !ok {
		return nil, nil
	}
	for _, account := range httpServer.config.Wallet.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(wallet.PaymentAddressType)
		if address == paramTemp {
			return account.Name, nil
		}
	}
	return nil, nil
}

/*
getaddressesbyaccount RPC returns a list of every address assigned to a particular account.

Parameter #1—the account name
Result—a list of addresses
*/
func (httpServer *HttpServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramTemp, ok := params.(string)
	if !ok {
		return nil, nil
	}
	result := jsonresult.GetAddressesByAccount{}
	result.Addresses = httpServer.config.Wallet.GetAddressesByAccName(paramTemp)
	return result, nil
}

/*
getaccountaddress RPC returns the current coin address for receiving payments to this account.
If the account doesn’t exist, it creates both the account and a new address for receiving payment.
Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a incognito address
*/
func (httpServer *HttpServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramTemp, ok := params.(string)
	if !ok {
		return nil, nil
	}
	activeShards := httpServer.config.BlockChain.BestState.Beacon.ActiveShards
	shardID := httpServer.config.Wallet.GetConfig().ShardID
	// if shardID is nil -> create with any shard
	if shardID != nil {
		// if shardID is configured with not nil
		shardIDInt := int(*shardID)
		// check with activeshards
		if shardIDInt >= activeShards || shardIDInt <= 0 {
			randShard := rand.Int31n(int32(activeShards))
			temp := byte(randShard)
			shardID = &temp
		}
	}
	result := httpServer.config.Wallet.GetAddressByAccName(paramTemp, shardID)
	return result, nil
}

/*
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
*/
func (httpServer *HttpServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramTemp, ok := params.(string)
	if !ok {
		return nil, nil
	}
	result := httpServer.config.Wallet.DumpPrivkey(paramTemp)
	return result, nil
}

/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (httpServer *HttpServer) handleImportAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleImportAccount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("params is invalid"))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("privateKey is invalid"))
	}
	accountName, ok := arrayParams[1].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("accountName is invalid"))
	}
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("passPhrase is invalid"))
	}
	account, err := httpServer.config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := wallet.KeySerializedData{
		PaymentAddress: account.Key.Base58CheckSerialize(wallet.PaymentAddressType),
		Pubkey:         hex.EncodeToString(account.Key.KeySet.PaymentAddress.Pk),
		ReadonlyKey:    account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}
	Logger.log.Infof("handleImportAccount result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleRemoveAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleRemoveAccount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("params is invalid"))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("privateKey is invalid"))
	}
	_, ok = arrayParams[1].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("accountName is invalid"))
	}
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("passPhrase is invalid"))
	}
	err := httpServer.config.Wallet.RemoveAccount(privateKey, passPhrase)
	if err != nil {
		return false, NewRPCError(ErrUnexpected, err)
	}
	return true, nil
}

// handleGetBalanceByPrivatekey -  return balance of private key
func (httpServer *HttpServer) handleGetBalanceByPrivatekey(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	log.Println(params)
	balance := uint64(0)

	// all component
	arrayParams := common.InterfaceSlice(params)

	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	log.Println(err)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	log.Println(senderKey)

	// get balance for accountName in wallet
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, shardIDSender, prvCoinID)
	log.Println(err)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, out := range outcoints {
		balance += out.CoinDetails.Value
	}
	log.Println(balance)

	return balance, nil
}

// handleGetBalanceByPaymentAddress -  return balance of paymentaddress
func (httpServer *HttpServer) handleGetBalanceByPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	// all component
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	// param #1: private key of sender
	paymentAddressParam := arrayParams[0]
	accountWithPaymentAddress, err := wallet.Base58CheckDeserialize(paymentAddressParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// get balance for accountName in wallet
	lastByte := accountWithPaymentAddress.KeySet.PaymentAddress.Pk[len(accountWithPaymentAddress.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&accountWithPaymentAddress.KeySet, shardIDSender, prvCoinID)
	Logger.log.Infof("OutCoins: %+v", outcoints)
	Logger.log.Infof("shardIDSender: %+v", shardIDSender)
	Logger.log.Infof("accountWithPaymentAddress.KeySet: %+v", accountWithPaymentAddress.KeySet)
	Logger.log.Infof("paymentAddressParam: %+v", paymentAddressParam)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, out := range outcoints {
		balance += out.CoinDetails.Value
	}

	return balance, nil
}

/*
handleGetBalance - RPC gets the balances in decimal
*/
func (httpServer *HttpServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	if httpServer.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("wallet is not existed"))
	}
	if len(httpServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("params is invalid"))
	}
	// Param #1: account "*" for all or a particular account
	accountName, ok := arrayParams[0].(string)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("accountName is invalid"))
	}

	// Param #2: the minimum number of confirmations an output must have
	minTemp, ok := arrayParams[1].(float64)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("min is invalid"))
	}
	min := int(minTemp)
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("passPhrase is invalid"))
	}

	if passPhrase != httpServer.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("password phrase is wrong for local wallet"))
	}

	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range httpServer.config.Wallet.MasterAccount.Child {
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, out := range outCoins {
				balance += out.CoinDetails.Value
			}
		}
	} else {
		for _, account := range httpServer.config.Wallet.MasterAccount.Child {
			if account.Name == accountName {
				// get balance for accountName in wallet
				lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
				shardIDSender := common.GetShardIDFromLastByte(lastByte)
				outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				for _, out := range outCoins {
					balance += out.CoinDetails.Value
				}
				break
			}
		}
	}

	return balance, nil
}

/*
handleGetReceivedByAccount -  RPC returns the total amount received by addresses in a
particular account from transactions with the specified number of confirmations. It does not count salary transactions.
*/
func (httpServer *HttpServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	if httpServer.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("wallet is not existed"))
	}
	if len(httpServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("params is invalid"))
	}
	// Param #1: account "*" for all or a particular account
	accountName, ok := arrayParams[0].(string)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("accountName is invalid"))
	}

	// Param #2: the minimum number of confirmations an output must have
	minTemp, ok := arrayParams[1].(float64)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("min is invalid"))
	}
	min := int(minTemp)
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return balance, NewRPCError(ErrRPCInvalidParams, errors.New("passPhrase is invalid"))
	}

	if passPhrase != httpServer.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("password phrase is wrong for local wallet"))
	}

	for _, account := range httpServer.config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			prvCoinID := &common.Hash{}
			prvCoinID.SetBytes(common.PRVCoinID[:])
			outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, out := range outCoins {
				balance += out.CoinDetails.Value
			}
			break
		}
	}
	return balance, nil
}

/*
handleSetTxFee - RPC sets the transaction fee per kilobyte paid more by transactions created by this wallet. default is 1 coin per 1 kb
*/
func (httpServer *HttpServer) handleSetTxFee(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	httpServer.config.Wallet.GetConfig().IncrementalFee = uint64(params.(float64))
	err := httpServer.config.Wallet.Save(httpServer.config.Wallet.PassPhrase)
	return err == nil, NewRPCError(ErrUnexpected, err)
}

// handleListCustomToken - return list all custom token in network
func (httpServer *HttpServer) handleListCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	temps, err := httpServer.config.BlockChain.ListCustomToken()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.ListCustomToken{ListCustomToken: []jsonresult.CustomToken{}}
	for _, token := range temps {
		item := jsonresult.CustomToken{}
		item.Init(token)
		result.ListCustomToken = append(result.ListCustomToken, item)
	}
	return result, nil
}

func (httpServer *HttpServer) handleListPrivacyCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	temps, listCustomTokenCrossShard, err := httpServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.ListCustomToken{ListCustomToken: []jsonresult.CustomToken{}}
	for _, token := range temps {
		item := jsonresult.CustomToken{}
		item.InitPrivacy(token)
		result.ListCustomToken = append(result.ListCustomToken, item)
	}
	for _, token := range listCustomTokenCrossShard {
		item := jsonresult.CustomToken{}
		item.InitPrivacyForCrossShard(token)
		result.ListCustomToken = append(result.ListCustomToken, item)
	}
	return result, nil
}

// handleGetPublicKeyFromPaymentAddress - return base58check encode of public key which is got from payment address
func (httpServer *HttpServer) handleGetPublicKeyFromPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("params is invalid"))
	}
	paymentAddress, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("paymentAddress is invalid"))
	}

	key, err := wallet.Base58CheckDeserialize(paymentAddress)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	return base58.Base58Check{}.Encode(key.KeySet.PaymentAddress.Pk[:], common.ZeroByte), nil
}

// ------------------------------------ Defragment output coin of account by combine many input coin in to 1 output coin --------------------
/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (httpServer *HttpServer) handleDefragmentAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	data, err := httpServer.createRawDefragmentAccountTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*RPCError) != nil {
		return nil, NewRPCError(ErrSendTxData, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:    sendResult.(jsonresult.CreateTransactionResult).TxID,
		ShardID: tx.ShardID,
	}
	return result, nil
}

/*
// createRawDefragmentAccountTransaction.
*/
func (httpServer *HttpServer) createRawDefragmentAccountTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := httpServer.buildRawDefragmentAccountTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Critical(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
		ShardID:         txShardID,
	}
	return result, nil
}

// buildRawDefragmentAccountTransaction
func (httpServer *HttpServer) buildRawDefragmentAccountTransaction(params interface{}, meta metadata.Metadata) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 4 {
		return nil, NewRPCError(ErrRPCInvalidParams, nil)
	}
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("senderKeyParam is invalid"))
	}
	maxValTemp, ok := arrayParams[1].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("maxVal is invalid"))
	}
	maxVal := uint64(maxValTemp)
	estimateFeeCoinPerKbtemp, ok := arrayParams[2].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("estimateFeeCoinPerKb is invalid"))
	}
	estimateFeeCoinPerKb := int64(estimateFeeCoinPerKbtemp)
	// param #4: hasPrivacyCoin flag: 1 or -1
	hasPrivacyCoin := int(arrayParams[3].(float64)) > 0
	/********* END Fetch all component to *******/

	// param #1: private key of sender
	senderKeySet, err := httpServer.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, prvCoinID)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	// remove out coin in mem pool
	outCoins, err = httpServer.filterMemPoolOutCoinsToSpent(outCoins)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	outCoins, amount := httpServer.calculateOutputCoinsByMinValue(outCoins, maxVal)
	if len(outCoins) == 0 {
		return nil, NewRPCError(ErrGetOutputCoin, nil)
	}
	paymentInfo := &privacy.PaymentInfo{
		Amount:         uint64(amount),
		PaymentAddress: senderKeySet.PaymentAddress,
	}
	paymentInfos := []*privacy.PaymentInfo{paymentInfo}
	// check real fee(nano PRV) per tx
	realFee, _, _ := httpServer.estimateFee(estimateFeeCoinPerKb, outCoins, paymentInfos, shardIDSender, 8, hasPrivacyCoin, nil, nil, nil)
	if len(outCoins) == 0 {
		realFee = 0
	}

	if uint64(amount) < realFee {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	paymentInfo.Amount = uint64(amount) - realFee

	inputCoins := transaction.ConvertOutputCoinToInputCoin(outCoins)

	/******* END GET output native coins(PRV), which is used to create tx *****/
	// START create tx
	// missing flag for privacy
	// false by default
	tx := transaction.Tx{}
	err = tx.Init(
		&senderKeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacyCoin,
		*httpServer.config.Database,
		nil, // use for prv coin -> nil is valid
		meta,
	)
	// END create tx

	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	return &tx, nil
}

//calculateOutputCoinsByMinValue
func (httpServer *HttpServer) calculateOutputCoinsByMinValue(outCoins []*privacy.OutputCoin, maxVal uint64) ([]*privacy.OutputCoin, uint64) {
	outCoinsTmp := make([]*privacy.OutputCoin, 0)
	amount := uint64(0)
	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value <= maxVal {
			outCoinsTmp = append(outCoinsTmp, outCoin)
			amount += outCoin.CoinDetails.Value
		}
	}
	return outCoinsTmp, amount
}

// ----------------------------- End ------------------------------------
