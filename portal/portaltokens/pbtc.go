package portaltokens

import (
	"errors"
	"fmt"
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalBTCTokenProcessor struct {
	*PortalToken
}

func (p *PortalBTCTokenProcessor) GetExpectedMemoForPorting(portingID string) string {
	return btcrelaying.HashAndEncodeBase58(portingID)
}

func (p *PortalBTCTokenProcessor) GetExpectedMemoForRedeem(redeemID string, custodianIncAddress string) string {
	rawMsg := fmt.Sprintf("%s%s", redeemID, custodianIncAddress)
	encodedMsg := btcrelaying.HashAndEncodeBase58(rawMsg)
	return encodedMsg
}

func (p *PortalBTCTokenProcessor) ParseAndVerifyProof(
	proof string, bc bMeta.ChainRetriever, expectedMemo string, expectedPaymentInfos map[string]uint64) (bool, error) {
	btcChain := bc.GetBTCHeaderChain()
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, errors.New("BTC relaying chain should not be null")
	}
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("PortingProof is invalid %v\n", err)
		return false, fmt.Errorf("PortingProof is invalid %v\n", err)
	}

	// verify tx with merkle proofs
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
	if btcAttachedMsg != expectedMemo {
		Logger.log.Errorf("PortingId in the btc attached message is not matched with portingID in metadata")
		return false, fmt.Errorf("PortingId in the btc attached message %v is not matched with portingID in metadata %v", btcAttachedMsg, expectedMemo)
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	outputs := btcTxProof.BTCTx.TxOut
	for receiverAddress, amount := range expectedPaymentInfos {
		amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amount))
		isChecked := false
		for _, out := range outputs {
			addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
			if err != nil {
				Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
				continue
			}
			if addrStr != receiverAddress {
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

//todo:
func (p *PortalBTCTokenProcessor) IsValidRemoteAddress(address string) (bool, error) {
	return true, nil
}

func (p *PortalBTCTokenProcessor) GetChainID() string {
	return p.ChainID
}
