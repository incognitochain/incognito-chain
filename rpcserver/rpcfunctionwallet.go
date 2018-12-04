package rpcserver

import (
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/common"
	"errors"
	"log"
)

/*
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

*/
func (self RpcServer) handleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.ListAccounts{
		Accounts:   make(map[string]uint64),
		WalletName: self.config.Wallet.Name,
	}
	accounts := self.config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&account.Key.KeySet, transaction.NoSort, false)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		amount := uint64(0)
		for _, txs := range txsMap {
			for _, tx := range txs {
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					for _, note := range notes {
						amount += note.Value
					}
				}
			}
		}
		result.Accounts[accountName] = amount
	}

	return result, nil
}

/*
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
*/
func (self RpcServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	for _, account := range self.config.Wallet.MasterAccount.Child {
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
func (self RpcServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetAddressesByAccount{}
	result.Addresses = self.config.Wallet.GetAddressesByAccount(params.(string))
	return result, nil
}

/*
getaccountaddress RPC returns the current coin address for receiving payments to this account. If the account doesn’t exist, it creates both the account and a new address for receiving payment. Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a constant address
*/
func (self RpcServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := self.config.Wallet.GetAccountAddress(params.(string))
	return result, nil
}

/*
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
*/
func (self RpcServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := self.config.Wallet.DumpPrivkey(params.(string))
	return result, nil
}

/*
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
*/
func (self RpcServer) handleImportAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	accountName := arrayParams[1].(string)
	passPhrase := arrayParams[2].(string)
	account, err := self.config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return wallet.KeySerializedData{
		PaymentAddress: account.Key.Base58CheckSerialize(wallet.PaymentAddressType),
		ReadonlyKey:    account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}, nil
}

func (self RpcServer) handleRemoveAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	accountName := arrayParams[1].(string)
	passPhrase := arrayParams[2].(string)
	err := self.config.Wallet.RemoveAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return false, NewRPCError(ErrUnexpected, err)
	}
	return true, nil
}

/*
handleGetAllPeers - return all peers which this node connected
*/
func (self RpcServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonresult.GetAllPeersResult{}
	peersMap := []string{}
	peers := self.config.AddrMgr.AddressCache()
	for _, peer := range peers {
		for _, peerConn := range peer.PeerConns {
			peersMap = append(peersMap, peerConn.RemoteRawAddress)
		}
	}
	result.Peers = peersMap
	return result, nil
}

// handleGetBalanceByPrivatekey -  return balance of private key
func (self RpcServer) handleGetBalanceByPrivatekey(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	balance := uint64(0)

	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)

	// get balance for accountName in wallet
	txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&senderKey.KeySet, transaction.NoSort, false)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, txs := range txsMap {
		for _, tx := range txs {
			for _, desc := range tx.Descs {
				notes := desc.GetNote()
				for _, note := range notes {
					balance += note.Value
				}
			}
		}
	}

	return balance, nil
}

// handleGetBalanceByPrivatekey -  return balance of private key
func (self RpcServer) handleGetBalanceByPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	paymentAddressParam := arrayParams[0]
	accountWithPaymentAddress, err := wallet.Base58CheckDeserialize(paymentAddressParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// get balance for accountName in wallet
	txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&accountWithPaymentAddress.KeySet, transaction.NoSort, false)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, txs := range txsMap {
		for _, tx := range txs {
			for _, desc := range tx.Descs {
				notes := desc.GetNote()
				for _, note := range notes {
					balance += note.Value
				}
			}
		}
	}

	return balance, nil
}

/*
handleGetBalance - RPC gets the balances in decimal
*/
func (self RpcServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("Wallet is not existed"))
	}
	if len(self.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("No account is existed"))
	}

	// convert params to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != self.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("Password phrase is wrong for local wallet"))
	}

	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range self.config.Wallet.MasterAccount.Child {
			txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&account.Key.KeySet, transaction.NoSort, false)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, txs := range txsMap {
				for _, tx := range txs {
					for _, desc := range tx.Descs {
						notes := desc.GetNote()
						for _, note := range notes {
							balance += note.Value
						}
					}
				}
			}
		}
	} else {
		for _, account := range self.config.Wallet.MasterAccount.Child {
			if account.Name == accountName {
				// get balance for accountName in wallet
				txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&account.Key.KeySet, transaction.NoSort, false)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				for _, txs := range txsMap {
					for _, tx := range txs {
						for _, desc := range tx.Descs {
							notes := desc.GetNote()
							for _, note := range notes {
								balance += note.Value
							}
						}
					}
				}
				break
			}
		}
	}

	return balance, nil
}

/*
handleGetReceivedByAccount -  RPC returns the total amount received by addresses in a particular account from transactions with the specified number of confirmations. It does not count salary transactions.
*/
func (self RpcServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("Wallet is not existed"))
	}
	if len(self.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("No account is existed"))
	}

	// convert params to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != self.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("Password phrase is wrong for local wallet"))
	}

	for _, account := range self.config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&account.Key.KeySet, transaction.NoSort, false)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, txs := range txsMap {
				for _, tx := range txs {
					if self.config.BlockChain.IsSalaryTx(&tx) {
						continue
					}
					for _, desc := range tx.Descs {
						notes := desc.GetNote()
						for _, note := range notes {
							balance += note.Value
						}
					}
				}
			}
			break
		}
	}
	return balance, nil
}

/*
handleSetTxFee - RPC sets the transaction fee per kilobyte paid more by transactions created by this wallet. default is 1 coin per 1 kb
*/
func (self RpcServer) handleSetTxFee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	self.config.Wallet.Config.IncrementalFee = uint64(params.(float64))
	err := self.config.Wallet.Save(self.config.Wallet.PassPhrase)
	return err == nil, NewRPCError(ErrUnexpected, err)
}

// handleListCustomToken - return list all custom token in network
func (self RpcServer) handleListCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	temps, err := self.config.BlockChain.ListCustomToken()
	if err != nil {
		return nil, err
	}
	result := jsonresult.ListCustomToken{ListCustomToken: []jsonresult.CustomToken{}}
	for _, token := range temps {
		item := jsonresult.CustomToken{}
		item.Init(token)
		result.ListCustomToken = append(result.ListCustomToken, item)
	}
	return result, nil
}
