package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
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
func (httpServer *HttpServer) handleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.ListAccounts{
		Accounts:   make(map[string]uint64),
		WalletName: httpServer.config.Wallet.Name,
	}
	accounts := httpServer.config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		err := prvCoinID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
		}
		outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		amount := uint64(0)
		for _, out := range outCoins {
			amount += out.CoinDetails.GetValue()
		}
		result.Accounts[accountName] = amount
	}

	return result, nil
}

/*
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
*/
func (httpServer *HttpServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
func (httpServer *HttpServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
func (httpServer *HttpServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
func (httpServer *HttpServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramTemp, ok := params.(string)
	if !ok {
		return nil, nil
	}
	result := httpServer.config.Wallet.DumpPrivateKey(paramTemp)
	return result, nil
}

/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (httpServer *HttpServer) handleImportAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleImportAccount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("params is invalid"))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privateKey is invalid"))
	}
	accountName, ok := arrayParams[1].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("accountName is invalid"))
	}
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("passPhrase is invalid"))
	}
	account, err := httpServer.config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := wallet.KeySerializedData{
		PaymentAddress: account.Key.Base58CheckSerialize(wallet.PaymentAddressType),
		Pubkey:         hex.EncodeToString(account.Key.KeySet.PaymentAddress.Pk),
		ReadonlyKey:    account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}
	Logger.log.Debugf("handleImportAccount result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleRemoveAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRemoveAccount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("params is invalid"))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privateKey is invalid"))
	}
	_, ok = arrayParams[1].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("accountName is invalid"))
	}
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("passPhrase is invalid"))
	}
	err := httpServer.config.Wallet.RemoveAccount(privateKey, passPhrase)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return true, nil
}

// handleGetBalanceByPrivatekey -  return balance of private key
func (httpServer *HttpServer) handleGetBalanceByPrivatekey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	log.Println(params)
	balance := uint64(0)

	// all component
	arrayParams := common.InterfaceSlice(params)

	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("key component invalid"))
	}
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		log.Println(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	log.Println(senderKey)

	// get balance for accountName in wallet
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	prvCoinID := &common.Hash{}
	err = prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, shardIDSender, prvCoinID)
	log.Println(err)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	for _, out := range outcoints {
		balance += out.CoinDetails.GetValue()
	}
	log.Println(balance)
	//return jsonresult.AccountBalanceResult{
	//	Account: senderKey.Base58CheckSerialize(wallet.PaymentAddressType),
	//	Balance: balance,
	//}, nil
	return balance, nil
}

// handleGetBalanceByPaymentAddress -  return balance of paymentaddress
func (httpServer *HttpServer) handleGetBalanceByPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	balance := uint64(0)

	// all component
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("key component invalid"))
	}
	// param #1: private key of sender
	paymentAddressParam := arrayParams[0]
	accountWithPaymentAddress, err := wallet.Base58CheckDeserialize(paymentAddressParam.(string))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// get balance for accountName in wallet
	lastByte := accountWithPaymentAddress.KeySet.PaymentAddress.Pk[len(accountWithPaymentAddress.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	prvCoinID := &common.Hash{}
	err1 := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err1)
	}
	outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&accountWithPaymentAddress.KeySet, shardIDSender, prvCoinID)
	Logger.log.Debugf("OutCoins: %+v", outcoints)
	Logger.log.Debugf("shardIDSender: %+v", shardIDSender)
	Logger.log.Debugf("accountWithPaymentAddress.KeySet: %+v", accountWithPaymentAddress.KeySet)
	Logger.log.Debugf("paymentAddressParam: %+v", paymentAddressParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	for _, out := range outcoints {
		balance += out.CoinDetails.GetValue()
	}

	return balance, nil
}

/*
handleGetBalance - RPC gets the balances in decimal
*/
func (httpServer *HttpServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	balance := uint64(0)

	if httpServer.config.Wallet == nil {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("wallet is not existed"))
	}
	if len(httpServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("params is invalid"))
	}
	// Param #1: account "*" for all or a particular account
	accountName, ok := arrayParams[0].(string)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("accountName is invalid"))
	}

	// Param #2: the minimum number of confirmations an output must have
	minTemp, ok := arrayParams[1].(float64)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min is invalid"))
	}
	min := int(minTemp)
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("passPhrase is invalid"))
	}

	if passPhrase != httpServer.config.Wallet.PassPhrase {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("password phrase is wrong for local wallet"))
	}

	prvCoinID := &common.Hash{}
	err1 := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err1)
	}
	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range httpServer.config.Wallet.MasterAccount.Child {
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
			}
			for _, out := range outCoins {
				balance += out.CoinDetails.GetValue()
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
					return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
				}
				for _, out := range outCoins {
					balance += out.CoinDetails.GetValue()
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
func (httpServer *HttpServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	balance := uint64(0)

	if httpServer.config.Wallet == nil {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("wallet is not existed"))
	}
	if len(httpServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("params is invalid"))
	}
	// Param #1: account "*" for all or a particular account
	accountName, ok := arrayParams[0].(string)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("accountName is invalid"))
	}

	// Param #2: the minimum number of confirmations an output must have
	minTemp, ok := arrayParams[1].(float64)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("min is invalid"))
	}
	min := int(minTemp)
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase, ok := arrayParams[2].(string)
	if !ok {
		return balance, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("passPhrase is invalid"))
	}

	if passPhrase != httpServer.config.Wallet.PassPhrase {
		return balance, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("password phrase is wrong for local wallet"))
	}

	for _, account := range httpServer.config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			prvCoinID := &common.Hash{}
			err1 := prvCoinID.SetBytes(common.PRVCoinID[:])
			if err1 != nil {
				return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err1)
			}
			outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, prvCoinID)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
			}
			for _, out := range outCoins {
				balance += out.CoinDetails.GetValue()
			}
			break
		}
	}
	return balance, nil
}

/*
handleSetTxFee - RPC sets the transaction fee per kilobyte paid more by transactions created by this wallet. default is 1 coin per 1 kb
*/
func (httpServer *HttpServer) handleSetTxFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	httpServer.config.Wallet.GetConfig().IncrementalFee = uint64(params.(float64))
	err := httpServer.config.Wallet.Save(httpServer.config.Wallet.PassPhrase)
	return err == nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
}

// handleListCustomToken - return list all custom token in network
func (httpServer *HttpServer) handleListCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	temps, err := httpServer.config.BlockChain.ListCustomToken()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.ListCustomToken{ListCustomToken: []jsonresult.CustomToken{}}
	for _, token := range temps {
		item := jsonresult.NewNormalToken(token)
		result.ListCustomToken = append(result.ListCustomToken, *item)
	}
	return result, nil
}

func (httpServer *HttpServer) handleListPrivacyCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	temps, listCustomTokenCrossShard, err := httpServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.ListCustomToken{ListCustomToken: []jsonresult.CustomToken{}}
	tokenIDs := make(map[common.Hash]interface{})
	for tokenID, token := range temps {
		item := jsonresult.NewPrivacyToken(token)
		tokenIDs[tokenID] = 0
		result.ListCustomToken = append(result.ListCustomToken, *item)
	}
	for tokenID, token := range listCustomTokenCrossShard {
		if _, ok := tokenIDs[tokenID]; ok {
			continue
		}
		item := jsonresult.CustomToken{}
		item.InitPrivacyForCrossShard(token)
		result.ListCustomToken = append(result.ListCustomToken, item)
	}
	return result, nil
}

// handleGetPublicKeyFromPaymentAddress - return base58check encode of public key which is got from payment address
func (httpServer *HttpServer) handleGetPublicKeyFromPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("params is invalid"))
	}
	paymentAddress, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("paymentAddress is invalid"))
	}

	key, err := wallet.Base58CheckDeserialize(paymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	result := jsonresult.NewGetPublicKeyFromPaymentAddressResult(key.KeySet.PaymentAddress.Pk[:])

	return result, nil
}

// ------------------------------------ Defragment output coin of account by combine many input coin in to 1 output coin --------------------
/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (httpServer *HttpServer) handleDefragmentAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	data, err := httpServer.createRawDefragmentAccountTransaction(params, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.SendTxDataError, err)
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
func (httpServer *HttpServer) createRawDefragmentAccountTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	tx, err := httpServer.buildRawDefragmentAccountTransaction(params, nil)
	if err.(*rpcservice.RPCError) != nil {
		Logger.log.Critical(err)
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
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
func (httpServer *HttpServer) buildRawDefragmentAccountTransaction(params interface{}, meta metadata.Metadata) (*transaction.Tx, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 4 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, nil)
	}
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("senderKeyParam is invalid"))
	}
	maxValTemp, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("maxVal is invalid"))
	}
	maxVal := uint64(maxValTemp)
	estimateFeeCoinPerKbtemp, ok := arrayParams[2].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("estimateFeeCoinPerKb is invalid"))
	}
	estimateFeeCoinPerKb := int64(estimateFeeCoinPerKbtemp)
	// param #4: hasPrivacyCoin flag: 1 or -1
	hasPrivacyCoin := int(arrayParams[3].(float64)) > 0
	/********* END Fetch all component to *******/

	// param #1: private key of sender
	senderKeySet, shardIDSender, err := rpcservice.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.InvalidSenderPrivateKeyError, err)
	}

	prvCoinID := &common.Hash{}
	err1 := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err1)
	}
	outCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, prvCoinID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetOutputCoinError, err)
	}
	// remove out coin in mem pool
	outCoins, err = httpServer.txMemPoolService.FilterMemPoolOutcoinsToSpent(outCoins)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetOutputCoinError, err)
	}
	outCoins, amount := httpServer.calculateOutputCoinsByMinValue(outCoins, maxVal)
	if len(outCoins) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.GetOutputCoinError, nil)
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
		return nil, rpcservice.NewRPCError(rpcservice.GetOutputCoinError, err)
	}
	paymentInfo.Amount = uint64(amount) - realFee

	inputCoins := transaction.ConvertOutputCoinToInputCoin(outCoins)

	/******* END GET output native coins(PRV), which is used to create tx *****/
	// START create tx
	// missing flag for privacy
	// false by default
	tx := transaction.Tx{}
	err = tx.Init(
		transaction.NewTxPrivacyInitParams(&senderKeySet.PrivateKey,
			paymentInfos,
			inputCoins,
			realFee,
			hasPrivacyCoin,
			*httpServer.config.Database,
			nil, // use for prv coin -> nil is valid
			meta, nil))
	// END create tx

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	return &tx, nil
}

//calculateOutputCoinsByMinValue
func (httpServer *HttpServer) calculateOutputCoinsByMinValue(outCoins []*privacy.OutputCoin, maxVal uint64) ([]*privacy.OutputCoin, uint64) {
	outCoinsTmp := make([]*privacy.OutputCoin, 0)
	amount := uint64(0)
	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.GetValue() <= maxVal {
			outCoinsTmp = append(outCoinsTmp, outCoin)
			amount += outCoin.CoinDetails.GetValue()
		}
	}
	return outCoinsTmp, amount
}

// ----------------------------- End ------------------------------------
