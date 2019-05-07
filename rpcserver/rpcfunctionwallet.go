package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"
	
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/wallet"
)

/*
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

*/
func (rpcServer RpcServer) handleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.ListAccounts{
		Accounts:   make(map[string]uint64),
		WalletName: rpcServer.config.Wallet.Name,
	}
	accounts := rpcServer.config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		constantTokenID := &common.Hash{}
		constantTokenID.SetBytes(common.ConstantID[:])
		outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, constantTokenID)
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
func (rpcServer RpcServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	for _, account := range rpcServer.config.Wallet.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(wallet.PaymentAddressType)
		if address == params.(string) {
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
func (rpcServer RpcServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetAddressesByAccount{}
	result.Addresses = rpcServer.config.Wallet.GetAddressesByAccount(params.(string))
	return result, nil
}

/*
getaccountaddress RPC returns the current coin address for receiving payments to this account. If the account doesn’t exist, it creates both the account and a new address for receiving payment. Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a constant address
*/
func (rpcServer RpcServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//if rpcServer.config.BlockChain.IsReady(false, 0) {
	activeShards := rpcServer.config.BlockChain.BestState.Beacon.ActiveShards
	randShard := rand.Int31n(int32(activeShards))
	result := rpcServer.config.Wallet.GetAccountAddress(params.(string), byte(randShard))
	return result, nil
	//}
	//return nil, NewRPCError(ErrUnexpected, errors.New("Can not get active shard"))
}

/*
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
*/
func (rpcServer RpcServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := rpcServer.config.Wallet.DumpPrivkey(params.(string))
	return result, nil
}

/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (rpcServer RpcServer) handleImportAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	accountName := arrayParams[1].(string)
	passPhrase := arrayParams[2].(string)
	account, err := rpcServer.config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return wallet.KeySerializedData{
		PaymentAddress: account.Key.Base58CheckSerialize(wallet.PaymentAddressType),
		Pubkey:         hex.EncodeToString(account.Key.KeySet.PaymentAddress.Pk),
		ReadonlyKey:    account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}, nil
}

func (rpcServer RpcServer) handleRemoveAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	accountName := arrayParams[1].(string)
	passPhrase := arrayParams[2].(string)
	err := rpcServer.config.Wallet.RemoveAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return false, NewRPCError(ErrUnexpected, err)
	}
	return true, nil
}

/*
handleGetAllPeers - return all peers which this node connected
*/
func (rpcServer RpcServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	result := jsonresult.GetAllPeersResult{}
	peersMap := []string{}
	peers := rpcServer.config.AddrMgr.AddressCache()
	for _, peer := range peers {
		for _, peerConn := range peer.PeerConns {
			peersMap = append(peersMap, peerConn.RemoteRawAddress)
		}
	}
	result.Peers = peersMap
	return result, nil
}

// handleGetBalanceByPrivatekey -  return balance of private key
func (rpcServer RpcServer) handleGetBalanceByPrivatekey(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outcoints, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, shardIDSender, constantTokenID)
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
func (rpcServer RpcServer) handleGetBalanceByPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	// all component
	arrayParams := common.InterfaceSlice(params)

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
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outcoints, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&accountWithPaymentAddress.KeySet, shardIDSender, constantTokenID)
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
func (rpcServer RpcServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	if rpcServer.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("wallet is not existed"))
	}
	if len(rpcServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != rpcServer.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("password phrase is wrong for local wallet"))
	}

	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range rpcServer.config.Wallet.MasterAccount.Child {
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, constantTokenID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, out := range outCoins {
				balance += out.CoinDetails.Value
			}
		}
	} else {
		for _, account := range rpcServer.config.Wallet.MasterAccount.Child {
			if account.Name == accountName {
				// get balance for accountName in wallet
				lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
				shardIDSender := common.GetShardIDFromLastByte(lastByte)
				outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, constantTokenID)
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
func (rpcServer RpcServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	balance := uint64(0)

	if rpcServer.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("wallet is not existed"))
	}
	if len(rpcServer.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("no account is existed"))
	}

	// convert component to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != rpcServer.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("password phrase is wrong for local wallet"))
	}

	for _, account := range rpcServer.config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			lastByte := account.Key.KeySet.PaymentAddress.Pk[len(account.Key.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			constantTokenID := &common.Hash{}
			constantTokenID.SetBytes(common.ConstantID[:])
			outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.Key.KeySet, shardIDSender, constantTokenID)
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
func (rpcServer RpcServer) handleSetTxFee(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	rpcServer.config.Wallet.GetConfig().IncrementalFee = uint64(params.(float64))
	err := rpcServer.config.Wallet.Save(rpcServer.config.Wallet.PassPhrase)
	return err == nil, NewRPCError(ErrUnexpected, err)
}

// handleListCustomToken - return list all custom token in network
func (rpcServer RpcServer) handleListCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	temps, err := rpcServer.config.BlockChain.ListCustomToken()
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

func (rpcServer RpcServer) handleListPrivacyCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	temps, listCustomTokenCrossShard, err := rpcServer.config.BlockChain.ListPrivacyCustomToken()
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
func (rpcServer RpcServer) handleGetPublicKeyFromPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddress := arrayParams[0].(string)

	key, err := wallet.Base58CheckDeserialize(paymentAddress)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	return base58.Base58Check{}.Encode(key.KeySet.PaymentAddress.Pk[:], common.ZeroByte), nil
}

// handleGetRecentTransactionsByBlockNumber - RPC return list rencent txs by number of confirmed blocks
func (rpcServer RpcServer) handleGetRecentTransactionsByBlockNumber(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// #param 1: number of confirmed blocks
	numberOfBlock := uint64(arrayParams[0].(float64))

	// #param 2: viewing key
	senderKeySet, err := rpcServer.GetKeySetFromKeyParams(arrayParams[1].(string))
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderViewingKey, err)
	}
	readOnlyKey := senderKeySet.ReadonlyKey

	// get chain from pubkey
	shardID := common.GetShardIDFromLastByte(readOnlyKey.Pk[len(readOnlyKey.Pk)-1])

	txs, err := rpcServer.config.BlockChain.GetRecentTransactions(numberOfBlock, &readOnlyKey, shardID)
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderViewingKey, err)
	}

	result := jsonresult.GetRecentTransactions{
		Txs: make(map[string]jsonresult.TransactionDetail),
	}
	if len(txs) > 0 {
		for txId, tx := range txs {
			result.Txs[txId] = jsonresult.TransactionDetail{
				Hash:     txId,
				LockTime: time.Unix(tx.GetLockTime(), 0).Format(common.DateOutputFormat),
				Image:    common.Render([]byte(txId)),
			}
		}
	}
	return result, nil
}

// ------------------------------------ Defragment output coin of account by combine many input coin in to 1 output coin --------------------
/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (rpcServer RpcServer) handleDefragmentAccount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	data, err := rpcServer.createRawDefragmentAccountTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
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
func (rpcServer RpcServer) createRawDefragmentAccountTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := rpcServer.buildRawDefragmentAccountTransaction(params, nil)
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
func (rpcServer RpcServer) buildRawDefragmentAccountTransaction(params interface{}, meta metadata.Metadata) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 4 {
		return nil, NewRPCError(ErrRPCInvalidParams, nil)
	}
	senderKeyParam := arrayParams[0].(string)
	maxVal := uint64(arrayParams[1].(float64))
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))
	// param #4: hasPrivacy flag: 1 or -1
	hasPrivacy := int(arrayParams[3].(float64)) > 0
	/********* END Fetch all component to *******/

	// param #1: private key of sender
	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, constantTokenID)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	// remove out coin in mem pool
	outCoins, err = rpcServer.filterMemPoolOutCoinsToSpent(outCoins)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	outCoins, amount := rpcServer.calculateOutputCoinsByMinValue(outCoins, maxVal)
	if len(outCoins) == 0 {
		return nil, NewRPCError(ErrGetOutputCoin, nil)
	}
	paymentInfo := &privacy.PaymentInfo{
		Amount:         uint64(amount),
		PaymentAddress: senderKeySet.PaymentAddress,
	}
	paymentInfos := []*privacy.PaymentInfo{paymentInfo}
	// check real fee(nano constant) per tx
	realFee, _, _ := rpcServer.estimateFee(estimateFeeCoinPerKb, outCoins, paymentInfos, shardIDSender, 8, hasPrivacy, nil, nil, nil)
	if len(outCoins) == 0 {
		realFee = 0
	}

	if uint64(amount) < realFee {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	paymentInfo.Amount = uint64(amount) - realFee

	inputCoins := transaction.ConvertOutputCoinToInputCoin(outCoins)
	// build hash array for input coin
	inputCoinHs := rpcServer.makeArrayInputCoinHashHs(inputCoins)

	/******* END GET output coins constant, which is used to create tx *****/
	// START create tx
	// missing flag for privacy
	// false by default
	//fmt.Printf("#inputCoins: %d\n", len(inputCoins))
	tx := transaction.Tx{}
	err = tx.Init(
		&senderKeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacy,
		*rpcServer.config.Database,
		nil, // use for constant coin -> nil is valid
		meta,
	)
	// END create tx

	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	// pool inCoinsH
	txHash := tx.Hash()
	if txHash != nil {
		rpcServer.config.TxMemPool.PrePoolTxCoinHashH(*txHash, inputCoinHs)
	}

	return &tx, nil
}

//calculateOutputCoinsByMinValue
func (rpcServer RpcServer) calculateOutputCoinsByMinValue(outCoins []*privacy.OutputCoin, maxVal uint64) ([]*privacy.OutputCoin, uint64) {
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
