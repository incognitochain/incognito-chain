package debugtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"
)

// Parse from byte to AutoTxByHash
func ParseAutoTxHashFromBytes(b []byte) (*AutoTxByHash, error) {
	data := new(AutoTxByHash)
	err := json.Unmarshal(b, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Query the RPC server then return the AutoTxByHash
func (this *DebugTool) getAutoTxByHash(txHash string) (*AutoTxByHash, error) {
	if len(this.url) == 0 {
		return nil, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := fmt.Sprintf(`{
		"jsonrpc":"1.0",
		"method":"gettransactionbyhash",
		"params":["%s"],
		"id":1
	}`, txHash)
	b, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	autoTx, txError := ParseAutoTxHashFromBytes(b)
	if txError != nil {
		return nil, err
	}
	return autoTx, nil
}

// Get only the proof of transaction requiring the txHash
func (this *DebugTool) GetProofTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Proof, nil
}

// Get only the Sig of transaction requiring the txHash
func (this *DebugTool) GetSigTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Sig, nil
}

// Get only the BlockHash of transaction requiring the txHash
func (this *DebugTool) GetBlockHashTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.BlockHash, nil
}

// Get only the BlockHeight of transaction requiring the txHash
func (this *DebugTool) GetBlockHeightTransactionByHash(txHash string) (int, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return -1, err
	}
	return tx.Result.BlockHeight, nil
}

// Get the whole result of rpc call 'gettransactionbyhash'
func (this *DebugTool) GetTransactionByHash(txHash string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := fmt.Sprintf(`{
		"jsonrpc":"1.0",
		"method":"gettransactionbyhash",
		"params":["%s"],
		"id":1
	}`, txHash)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) CreateAndSendTransaction() ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := `{
		"jsonrpc": "1.0",
		"method": "createandsendtransaction",
		"params": [
			"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or", 
			{
				"12RuhVZQtGgYmCVzVi49zFZD7gR8SQx8Uuz8oHh6eSZ8PwB2MwaNE6Kkhd6GoykfkRnHNSHz1o2CzMiQBCyFPikHmjvvrZkLERuhcVE":2000000000,
				"12RxDSnQVjPojzf7uju6dcgC2zkKkg85muvQh347S76wKSSsKPAqXkvfpSeJzyEH3PREHZZ6SKsXLkDZbs3BSqwEdxqprqih4VzANK9":2000000000,
				"12S6m2LpzN17jorYnLb2ApNKaV2EVeZtd6unvrPT1GH8yHGCyjYzKbywweQDZ7aAkhD31gutYAgfQizb2JhJTgBb3AJ8aB4hyppm2ax":2000000000,
				"12S42y9fq8xWXx1YpZ6KVDLGx6tLjeFWqbSBo6zGxwgVnPe1rMGxtLs87PyziCzYPEiRGdmwU1ewWFXwjLwog3X71K87ApNUrd3LQB3":2000000000,
				"12S3yvTvWUJfubx3whjYLv23NtaNSwQMGWWScSaAkf3uQg8xdZjPFD4fG8vGvXjpRgrRioS5zuyzZbkac44rjBfs7mEdgoL4pwKu87u":2000000000,
				"12S6mGbnS3Df5bGBaUfBTh56NRax4PvFPDhUnxvP9D6cZVjnTx9T4FsVdFT44pFE8KXTGYaHSAmb2MkpnUJzkrAe49EPHkBULM8N2ZJ":2000000000,
				"12Rs5tQTYkWGzEdPNo2GRA1tjZ5aDCTYUyzXf6SJFq89QnY3US3ZzYSjWHVmmLUa6h8bdHHUuVYoR3iCVRoYDCNn1AfP6pxTz5YL8Aj":2000000000,
				"12S33dTF3aVsuSxY7iniK3UULUYyLMZumExKm6DPfsqnNepGjgDZqkQCDp1Z7Te9dFKQp7G2WeeYqCr5vcDCfrA3id4x5UvL4yyLrrT":2000000000,
				"12RysvT327ju2GTJ1RmquyRAQzPK1wezr7CCpraBLQ2VoHwz91n3rMeDFpcDiipBfmawrzTGrheEz4TrtZ5WFt8pYXkqiM5oXqucRRE":2000000000
			}, 
			1,   
			1
		],
		"id": 1
	}`
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) CreateAndSendTransactionFromAToB(privKeyA string, privKeyB string, amount string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	keyWallet, _ := wallet.Base58CheckDeserialize(privKeyB)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "createandsendtransaction",
		"params": [
			"%s", 
			{
				"%s": %s
			}, 
			1,   
			1
		],
		"id": 1
	}`, privKeyA, paymentAddStr, amount)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) GetListOutputCoins(privKeyStr string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	keyWallet, _ := wallet.Base58CheckDeserialize(privKeyStr)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listoutputcoins",
		"params": [
			0,
			999999,
			[
				{
			  "PaymentAddress": "%s",
			  "ReadonlyKey": "%s",
			  "StartHeight": 0
				}
			]
		  ],
		"id": 1
	}`, paymentAddStr, viewingKeyStr)

	//fmt.Println("==============")

	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) ListUnspentOutputCoins(privKeyStr string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	query := fmt.Sprintf(`{  
	   "jsonrpc":"1.0",
	   "method":"listunspentoutputcoins",
	   "params":[  
		  0,
		  999999,
		  [  
			 {  
				"PrivateKey":"%s",
				"StartHeight": 0
			 }
			 
		  ]
	   ],
	   "id":1
	}`, privKeyStr)

	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) GetBalanceByPrivatekey(privKeyStr string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	query := fmt.Sprintf(`{  
	   "jsonrpc":"1.0",
	   "method":"getbalancebyprivatekey",
	   "params":["%s"],
	   "id":1
	}`, privKeyStr)

	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) CreateAndSendPrivacyCustomTokenTransaction(privKeyStrA string) ([]byte, error) {
	keyWallet, _ := wallet.Base58CheckDeserialize(privKeyStrA)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)

	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "createandsendprivacycustomtokentransaction",
		"params": [
			"%s",
			{},
			5,
			1,
			{
				"Privacy": true,
				"TokenID": "",
				"TokenName": "abcxyz",
				"TokenSymbol": "pTTT",
				"TokenFee": 0,
				"TokenTxType": 0,
				"TokenAmount": 1000000000000000000,
				"TokenReceivers": {
					"%s": 1000000000000000000
				}
			}
			]
	}`, privKeyStrA, paymentAddStr)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) ListPrivacyCustomToken() ([]byte, error) {
	query := `{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "listprivacycustomtoken",
		"params": []
	}`
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) TransferPrivacyCustomToken(privKeyStrA string, privKeyStrB string, tokenID string, amount string) ([]byte, error) {
	keyWallet, _ := wallet.Base58CheckDeserialize(privKeyStrB)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)

	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "createandsendprivacycustomtokentransaction",
		"params": [
			"%s",
			{},
			10,
			1,
			{
				"Privacy": true,
				"TokenID": "%s",
				"TokenName": "",
				"TokenSymbol": "",
				"TokenFee": 0,
				"TokenTxType": 1,
				"TokenAmount": 0,
				"TokenReceivers": {
					"%s": %s
				}
			}
			]
	}`, privKeyStrA, tokenID, paymentAddStr, amount)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) GetBalancePrivacyCustomToken(privKeyStr string, tokenID string) ([]byte, error) {
	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "getbalanceprivacycustomtoken",
		"params": [
			"%s",
			"%s"
		]
	}`, privKeyStr, tokenID)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) SwitchCoinVersion(privKey string) ([]byte, error) {
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "createconvertcoinver1tover2transaction",
		"params": [
			"%s", 
			1
		],
		"id": 1
	}`, privKey)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) Stake(privKey string, seed string) ([]byte, error) {
	keyWallet, _ := wallet.Base58CheckDeserialize(privKey)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	query := fmt.Sprintf(`{
	  "jsonrpc":"1.0",
	  "method":"createandsendstakingtransaction",
	  "params":[
			"%s",
			{
				"15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs": 1750000000000
			},
			5,
			0,
			{
				"StakingType": 63,
				"CandidatePaymentAddress": "%s",
				"PrivateSeed": "%s",
				"RewardReceiverPaymentAddress": "%s",
				"AutoReStaking": true
			}
	  ],
	  "id":1
	}`, privKey, paymentAddStr, seed, paymentAddStr)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) Unstake(privKey string, seed string) ([]byte, error) {
	keyWallet, _ := wallet.Base58CheckDeserialize(privKey)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	query := fmt.Sprintf(`{
		"id":1,
		"jsonrpc":"1.0",
		"method":"createandsendstopautostakingtransaction",
		"params": [
			"%s",
			{
				"15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs": 0
			},
			10,
			0,
			{
				"StopAutoStakingType" : 127,
				"CandidatePaymentAddress" : "%s",
				"PrivateSeed":"%s"
			}
		]
	}`, privKey, paymentAddStr, seed)
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) WithdrawReward(privKey string, tokenID string) ([]byte, error) {
	keyWallet, _ := wallet.Base58CheckDeserialize(privKey)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	query := fmt.Sprintf(`{
    "jsonrpc": "1.0",
    "method": "withdrawreward",
    "params": [
        "%s",
        0,
			10,
			0,
        {
            "PaymentAddress": "%s",
            "TokenID": "%s"
        }
    ],
    "id": 1
	}`, privKey, paymentAddStr, tokenID)
	return this.SendPostRequestWithQuery(query)
}