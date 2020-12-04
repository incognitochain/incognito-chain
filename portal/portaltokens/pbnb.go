package portaltokens

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/binance-chain/go-sdk/types/msg"
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
)

type PortalBNBTokenProcessor struct {
	*PortalToken
}

func (p *PortalBNBTokenProcessor) ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc bMeta.ChainRetriever) (bool, error) {
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

	type PortingMemoBNB struct {
		PortingID string `json:"PortingID"`
	}
	var portingMemo PortingMemoBNB
	err2 = json.Unmarshal(memoBytes, &portingMemo)
	if err2 != nil {
		Logger.log.Errorf("Can not unmarshal memo in tx bnb proof", err2)
		return false, fmt.Errorf("Data in PortingProof is invalid %v", err)
	}

	if portingMemo.PortingID != portingReq.UniquePortingID() {
		Logger.log.Errorf("PortingId in memoTx %v is not matched with portingID in metadata %v", portingMemo.PortingID, portingReq.UniquePortingID())
		return false, fmt.Errorf("PortingId in memoTx %v is not matched with portingID in metadata %v", portingMemo.PortingID, portingReq.UniquePortingID())
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	// get list matching custodians in waitingPortingRequest
	custodians := portingReq.Custodians()
	outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs
	for _, cusDetail := range custodians {
		remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
		amountNeedToBeTransfer := cusDetail.Amount
		amountNeedToBeTransferInBNB := ConvertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addr, _ := bnb.GetAccAddressString(&out.Address, bc.GetBNBChainID(0))
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

func (p *PortalBNBTokenProcessor) ParseAndVerifyProofForRedeem(proof string, redeemReq *statedb.RedeemRequest, bc bMeta.ChainRetriever, matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error){
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

	type RedeemMemoBNB struct {
		RedeemID                  string `json:"RedeemID"`
		CustodianIncognitoAddress string `json:"CustodianIncognitoAddress"`
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
	amountNeedToBeTransferInBNB := ConvertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

	isChecked := false
	for _, out := range outputs {
		addr, _ := bnb.GetAccAddressString(&out.Address, bc.GetBNBChainID(0))
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

// ConvertIncPBNBAmountToExternalBNBAmount converts amount in inc chain (decimal 9) to amount in bnb chain (decimal 8)
func ConvertIncPBNBAmountToExternalBNBAmount(incPBNBAmount int64) int64 {
	return incPBNBAmount / 10 // incPBNBAmount / 1^9 * 1^8
}



