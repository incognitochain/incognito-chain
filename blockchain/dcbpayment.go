package blockchain

import (
	"fmt"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type producerTool struct {
	key     *privacy.PrivateKey
	db      database.DatabaseInterface
	shardID byte
}

// initVouts cache and return a list of UTXO for a specific token to prevent double spending in a single block
func (blockgen *BlkTmplGenerator) initVouts(unspentTokens map[string][]transaction.TxTokenVout, tokenID *common.Hash) []transaction.TxTokenVout {
	dcbWallet, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	key := tokenID.String()
	if _, ok := unspentTokens[key]; !ok {
		unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(dcbWallet.KeySet, tokenID)
		if err == nil {
			unspentTokens[key] = unspentTxTokenOuts
		} else {
			unspentTokens[key] = []transaction.TxTokenVout{}
		}
	}
	return unspentTokens[key]
}

func transferTxToken(
	amount uint64,
	vouts []transaction.TxTokenVout,
	tokenID common.Hash,
	receiver privacy.PaymentAddress,
	meta metadata.Metadata,
	tool producerTool,
) (*transaction.TxCustomToken, int, error) {
	sumTokens := uint64(0)
	usedID := 0
	// Choose input token UTXO
	for _, out := range vouts {
		usedID += 1
		sumTokens += out.Value
		if sumTokens >= amount {
			break
		}
	}

	if sumTokens < amount {
		return nil, 0, errors.New("not enough tokens to pay in this block")
	}

	// Build list of inputs and outputs
	txTokenIns := []transaction.TxTokenVin{}
	for i := 0; i < usedID; i += 1 {
		out := vouts[i]

		// Sign dummy signature using miner's key
		keySet := &cashec.KeySet{PrivateKey: *tool.key}
		signature, err := keySet.Sign(out.Hash()[:])
		if err != nil {
			return nil, 0, err
		}

		item := transaction.TxTokenVin{
			PaymentAddress:  out.PaymentAddress,
			TxCustomTokenID: out.GetTxCustomTokenID(),
			VoutIndex:       out.GetIndex(),
			Signature:       base58.Base58Check{}.Encode(signature, 0),
		}

		// No need for signature to spend tokens in DCB's account
		txTokenIns = append(txTokenIns, item)
	}
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: receiver,
			Value:          amount,
		},
	}
	if sumTokens > amount {
		dcbWallet, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		txTokenOuts = append(txTokenOuts, transaction.TxTokenVout{
			PaymentAddress: dcbWallet.KeySet.PaymentAddress,
			Value:          sumTokens - amount,
		})
	}

	// Build token params
	params := &transaction.CustomTokenParamTx{
		PropertyID:  tokenID.String(),
		TokenTxType: transaction.CustomTokenTransfer,
		Amount:      sumTokens,
		Receiver:    txTokenOuts,
		Mintable:    true,
	}
	params.SetVins(txTokenIns)
	params.SetVinsAmount(sumTokens)

	// Build TxCustomToken from token params
	txToken, err := initTxCustomToken(params, meta, tool)
	if err != nil {
		return nil, 0, err
	}

	return txToken, usedID, nil
}

func initTxCustomToken(
	params *transaction.CustomTokenParamTx,
	meta metadata.Metadata,
	tool producerTool,
) (*transaction.TxCustomToken, error) {
	tx := &transaction.TxCustomToken{}
	err := tx.Init(
		tool.key,
		nil,
		nil,
		0,
		params,
		tool.db,
		meta,
		false,
		tool.shardID,
	)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return tx, nil
}

func mintDCBToken(
	receiver privacy.PaymentAddress,
	amount uint64,
	meta metadata.Metadata,
	tool producerTool,
) (*transaction.TxCustomToken, error) {
	params := &transaction.CustomTokenParamTx{
		PropertyName:   common.DCBTokenName,
		PropertySymbol: common.DCBTokenSymbol,
		PropertyID:     common.DCBTokenID.String(),
		TokenTxType:    transaction.CustomTokenInit,
		Amount:         amount,
		Receiver: []transaction.TxTokenVout{
			transaction.TxTokenVout{
				Value:          amount,
				PaymentAddress: receiver,
			},
		},
		Mintable: true,
	}
	return initTxCustomToken(params, meta, tool)
}

func mintTxToken(
	receiver privacy.PaymentAddress,
	amount uint64,
	tokenID common.Hash,
	meta metadata.Metadata,
	tool producerTool,
) (*transaction.TxCustomToken, error) {
	// Ignore Name and Symbol
	params := &transaction.CustomTokenParamTx{
		PropertyID:  tokenID.String(),
		TokenTxType: transaction.CustomTokenInit,
		Amount:      amount,
		Receiver: []transaction.TxTokenVout{
			transaction.TxTokenVout{
				Value:          amount,
				PaymentAddress: receiver,
			},
		},
		Mintable: true,
	}
	return initTxCustomToken(params, meta, tool)
}
