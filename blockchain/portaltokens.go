package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalTokenProcessor interface {
	ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc *BlockChain) (bool, error)
	ParseAndVerifyProofForRedeem(proof string, redeemReq *statedb.RedeemRequest, bc *BlockChain, matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error)
	IsValidRemoteAddress(address string) (bool, error)
	GetChainID() (string)
}

type PortalToken struct {
	ChainID string
}

type PortalBTCTokenProcessor struct {
	*PortalToken
}

func (p *PortalBTCTokenProcessor) ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc *BlockChain) (bool, error){
	btcChain := bc.config.BTCChain
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, errors.New("BTC relaying chain should not be null")
	}
	// parse PortingProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("PortingProof is invalid %v\n", err)
		return false, fmt.Errorf("PortingProof is invalid %v\n", err)
	}

	isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify btcTxProof failed %v", err)
		return false, fmt.Errorf("Verify btcTxProof failed %v", err)
	}

	// extract attached message from txOut's OP_RETURN
	btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
	if err != nil {
		Logger.log.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
		return false, fmt.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
	}

	encodedMsg := btcrelaying.HashAndEncodeBase58(portingReq.UniquePortingID())
	if btcAttachedMsg != encodedMsg {
		Logger.log.Errorf("PortingId in the btc attached message is not matched with portingID in metadata")
		return false, fmt.Errorf("PortingId in the btc attached message %v is not matched with portingID in metadata %v", btcAttachedMsg, encodedMsg)
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	// get list matching custodians in waitingPortingRequest
	custodians := portingReq.Custodians()
	outputs := btcTxProof.BTCTx.TxOut
	for _, cusDetail := range custodians {
		remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
		amountNeedToBeTransfer := cusDetail.Amount
		amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
			if err != nil {
				Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
				continue
			}
			if addrStr != remoteAddressNeedToBeTransfer {
				continue
			}
			if out.Value < amountNeedToBeTransferInBTC {
				Logger.log.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
				return false, fmt.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
			} else {
				isChecked = true
				break
			}
		}
		if !isChecked {
			Logger.log.Error("BTC-TxProof is invalid")
			return false, errors.New("BTC-TxProof is invalid")
		}
	}

	return true, nil
}

func (p *PortalBTCTokenProcessor) ParseAndVerifyProofForRedeem(
	proof string,
	redeemReq *statedb.RedeemRequest,
	bc *BlockChain,
	matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error){
	btcChain := bc.config.BTCChain
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, errors.New("BTC relaying chain should not be null")
	}
	// parse PortingProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("RedeemProof is invalid %v\n", err)
		return false, fmt.Errorf("RedeemProof is invalid %v\n", err)
	}

	isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify btcTxProof failed %v", err)
		return false, fmt.Errorf("Verify btcTxProof failed %v", err)
	}

	// extract attached message from txOut's OP_RETURN
	btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
	if err != nil {
		Logger.log.Errorf("Could not extract message from btc proof with error: %v", err)
		return false, fmt.Errorf("Could not extract message from btc proof with error: %v", err)
	}

	rawMsg := fmt.Sprintf("%s%s", redeemReq.GetUniqueRedeemID(), matchedCustodian.GetIncognitoAddress())
	encodedMsg := btcrelaying.HashAndEncodeBase58(rawMsg)
	if btcAttachedMsg != encodedMsg {
		Logger.log.Errorf("The hash of combination of UniqueRedeemID(%s) and CustodianAddressStr(%s) is not matched to tx's attached message",
			redeemReq.GetUniqueRedeemID(),  matchedCustodian.GetIncognitoAddress())
		return false, fmt.Errorf("The hash of combination of UniqueRedeemID(%s) and CustodianAddressStr(%s) is not matched to tx's attached message",
			redeemReq.GetUniqueRedeemID(),  matchedCustodian.GetIncognitoAddress())
	}

	// check whether amount transfer in txBNB is equal redeem amount or not
	// check receiver and amount in tx
	// get list matching custodians in matchedRedeemRequest

	outputs := btcTxProof.BTCTx.TxOut
	remoteAddressNeedToBeTransfer := redeemReq.GetRedeemerRemoteAddress()
	amountNeedToBeTransfer := matchedCustodian.GetAmount()
	amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amountNeedToBeTransfer))

	isChecked := false
	for _, out := range outputs {
		addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
		if err != nil {
			Logger.log.Warnf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
			continue
		}
		if addrStr != remoteAddressNeedToBeTransfer {
			continue
		}
		if out.Value < amountNeedToBeTransferInBTC {
			Logger.log.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
			return false, fmt.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
		} else {
			isChecked = true
			break
		}
	}

	if !isChecked {
		Logger.log.Error("BTC-TxProof is invalid")
		return false, errors.New("BTC-TxProof is invalid")
	}
	return true, nil
}

//todo:
func (p *PortalBTCTokenProcessor) IsValidRemoteAddress(address string) (bool, error) {
	return true, nil
}

func (p *PortalBTCTokenProcessor) GetChainID() string {
	return p.ChainID
}

type PortalBNBTokenProcessor struct {
	*PortalToken
}

func (p *PortalBNBTokenProcessor) ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc *BlockChain) (bool, error) {
	// parse PortingProof in meta
	txProofBNB, err := bnb.ParseBNBProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("PortingProof is invalid %v\n", err)
		return false, fmt.Errorf("PortingProof is invalid %v\n", err)
	}

	// check minimum confirmations block of bnb proof
	latestBNBBlockHeight, err2 := bc.GetLatestBNBBlkHeight()
	if err2 != nil {
		Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
		return false, fmt.Errorf("Can not get latest relaying bnb block height %v\n", err)
	}

	if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
		Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
			bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
		return false, fmt.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
			bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
	}
	dataHash, err2 := bc.GetBNBDataHash(txProofBNB.BlockHeight)
	if err2 != nil {
		Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
			txProofBNB.BlockHeight, err2)
		return false, fmt.Errorf("Error when get data hash in blockHeight %v - %v\n",
			txProofBNB.BlockHeight, err2)
	}

	isValid, err := txProofBNB.Verify(dataHash)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify txProofBNB failed %v", err)
		return false, fmt.Errorf("Verify txProofBNB failed %v", err)
	}

	// parse Tx from Data in txProofBNB
	txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
	if err != nil {
		Logger.log.Errorf("Data in PortingProof is invalid %v", err)
		return false, fmt.Errorf("Data in PortingProof is invalid %v", err)
	}

	// check memo attach portingID req:
	memo := txBNB.Memo
	memoBytes, err2 := base64.StdEncoding.DecodeString(memo)
	if err2 != nil {
		Logger.log.Errorf("Can not decode memo in tx bnb proof", err2)
		return false, fmt.Errorf("Data in PortingProof is invalid %v", err)
	}

	var portingMemo PortingMemoBNB
	err2 = json.Unmarshal(memoBytes, &portingMemo)
	if err2 != nil {
		Logger.log.Errorf("Can not unmarshal memo in tx bnb proof", err2)
		return false, fmt.Errorf("Data in PortingProof is invalid %v", err)
	}

	if portingMemo.PortingID != portingReq.UniquePortingID() {
		Logger.log.Errorf("PortingId in memoTx is not matched with portingID in metadata")
		return false, errors.New("PortingId in memoTx is not matched with portingID in metadata")
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	// get list matching custodians in waitingPortingRequest
	custodians := portingReq.Custodians()
	outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs
	for _, cusDetail := range custodians {
		remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
		amountNeedToBeTransfer := cusDetail.Amount
		amountNeedToBeTransferInBNB := convertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addr, _ := bnb.GetAccAddressString(&out.Address, bc.config.ChainParams.BNBRelayingHeaderChainID)
			if addr != remoteAddressNeedToBeTransfer {
				Logger.log.Warnf("[portal] remoteAddressNeedToBeTransfer: %v - addr: %v\n", remoteAddressNeedToBeTransfer, addr)
				continue
			}

			// calculate amount that was transferred to custodian's remote address
			amountTransfer := int64(0)
			for _, coin := range out.Coins {
				if coin.Denom == bnb.DenomBNB {
					amountTransfer += coin.Amount
				}
			}
			if amountTransfer < amountNeedToBeTransferInBNB {
				Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
					addr, amountNeedToBeTransferInBNB, amountTransfer)
				return false, fmt.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
					addr, amountNeedToBeTransferInBNB, amountTransfer)
			} else {
				isChecked = true
				break
			}
		}
		if !isChecked {
			Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
				remoteAddressNeedToBeTransfer)
			return false, fmt.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
				remoteAddressNeedToBeTransfer)
		}
	}
	return true, nil
}

func (p *PortalBNBTokenProcessor) ParseAndVerifyProofForRedeem(proof string, redeemReq *statedb.RedeemRequest, bc *BlockChain, matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error){
	// parse RedeemProof in meta
	txProofBNB, err := bnb.ParseBNBProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("RedeemProof is invalid %v\n", err)
		return false, fmt.Errorf("RedeemProof is invalid %v\n", err)
	}

	// check minimum confirmations block of bnb proof
	latestBNBBlockHeight, err2 := bc.GetLatestBNBBlkHeight()
	if err2 != nil {
		Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
		return false, fmt.Errorf("Can not get latest relaying bnb block height %v\n", err)
	}

	if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
		Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
			bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
		return false, fmt.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
			bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
	}

	dataHash, err2 := bc.GetBNBDataHash(txProofBNB.BlockHeight)
	if err2 != nil {
		Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
			txProofBNB.BlockHeight, err2)
		return false, fmt.Errorf("Error when get data hash in blockHeight %v - %v\n",
			txProofBNB.BlockHeight, err2)
	}

	isValid, err := txProofBNB.Verify(dataHash)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify txProofBNB failed %v", err)
		return false, fmt.Errorf("Verify txProofBNB failed %v", err)
	}

	// parse Tx from Data in txProofBNB
	txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
	if err != nil {
		Logger.log.Errorf("Data in RedeemProof is invalid %v", err)
		return false, fmt.Errorf("Data in RedeemProof is invalid %v", err)
	}

	// check memo attach redeemID req (compare hash memo)
	memo := txBNB.Memo
	memoHashBytes, err2 := base64.StdEncoding.DecodeString(memo)
	if err2 != nil {
		Logger.log.Errorf("Can not decode memo in tx bnb proof %v", err2)
		return false, fmt.Errorf("Can not decode memo in tx bnb proof %v", err2)
	}

	expectedRedeemMemo := RedeemMemoBNB{
		RedeemID:                  redeemReq.GetUniqueRedeemID(),
		CustodianIncognitoAddress: matchedCustodian.GetIncognitoAddress()}
	expectedRedeemMemoBytes, _ := json.Marshal(expectedRedeemMemo)
	expectedRedeemMemoHashBytes := common.HashB(expectedRedeemMemoBytes)

	if !bytes.Equal(memoHashBytes, expectedRedeemMemoHashBytes) {
		Logger.log.Errorf("Memo redeem is invalid")
		return false, fmt.Errorf("Memo redeem is invalid")
	}

	// check whether amount transfer in txBNB is equal redeem amount or not
	// check receiver and amount in tx
	// get list matching custodians in matchedRedeemRequest

	outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs

	remoteAddressNeedToBeTransfer := redeemReq.GetRedeemerRemoteAddress()
	amountNeedToBeTransfer := matchedCustodian.GetAmount()
	amountNeedToBeTransferInBNB := convertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

	isChecked := false
	for _, out := range outputs {
		addr, _ := bnb.GetAccAddressString(&out.Address, bc.config.ChainParams.BNBRelayingHeaderChainID)
		if addr != remoteAddressNeedToBeTransfer {
			continue
		}

		// calculate amount that was transferred to custodian's remote address
		amountTransfer := int64(0)
		for _, coin := range out.Coins {
			if coin.Denom == bnb.DenomBNB {
				amountTransfer += coin.Amount
				// note: log error for debug
				Logger.log.Infof("TxProof-BNB coin.Amount %d",
					coin.Amount)
			}
		}
		if amountTransfer < amountNeedToBeTransferInBNB {
			Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
				addr, amountNeedToBeTransferInBNB, amountTransfer)
			return false, fmt.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
				addr, amountNeedToBeTransferInBNB, amountTransfer)
		} else {
			isChecked = true
			break
		}
	}

	if !isChecked {
		Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
			remoteAddressNeedToBeTransfer)
		return false, fmt.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
			remoteAddressNeedToBeTransfer)
	}
	return true, nil
}

func (p *PortalBNBTokenProcessor) IsValidRemoteAddress(address string) (bool, error) {
	return bnb.IsValidBNBAddress(address, p.ChainID), nil
}

func (p *PortalBNBTokenProcessor) GetChainID() string {
	return p.ChainID
}
