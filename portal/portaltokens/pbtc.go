package portaltokens

import (
	"errors"
	"fmt"
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalBTCTokenProcessor struct {
	*PortalToken
}

func (p *PortalBTCTokenProcessor) ParseAndVerifyProofForPorting(proof string, portingReq *statedb.WaitingPortingRequest, bc bMeta.ChainRetriever) (bool, error){
	btcChain := bc.GetBTCHeaderChain()
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
	bc bMeta.ChainRetriever,
	matchedCustodian *statedb.MatchingRedeemCustodianDetail) (bool, error){
	btcChain := bc.GetBTCHeaderChain()
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