package portalprocess

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"strconv"
)

// PortalSig defines sigs of one beacon validator on unshield external tx
type PortalSig struct {
	TokenID   string
	RawTxHash string
	Sigs      [][]byte // array of sigs for all TxIn
}

// CheckAndSignPortalUnshieldExternalTx checks portal instructions need beacons sign on
func CheckAndSignPortalUnshieldExternalTx(seedKey []byte, insts [][]string, portalParam portalv4.PortalParams) ([]*PortalSig, error) {
	var pSigs []*PortalSig
	var tokenID string
	var hexRawExternalTx string
	var utxos []*statedb.UTXO

	for _, inst := range insts {
		metaType := inst[0]
		switch metaType {
		case strconv.Itoa(metadata.PortalV4UnshieldBatchingMeta):
			{
				// unmarshal instructions content
				var actionData metadata.PortalUnshieldRequestBatchContent
				err := json.Unmarshal([]byte(inst[3]), &actionData)
				if err != nil {
					return nil, fmt.Errorf("[checkAndSignPortalV4] Can not unmarshal instruction content %v - Error %v\n", inst[3], err)
				}
				tokenID = actionData.TokenID
				hexRawExternalTx = actionData.RawExternalTx
				utxos = actionData.UTXOs
			}
		case strconv.Itoa(metadata.PortalV4FeeReplacementRequestMeta):
			{
				reqStatus := inst[2]
				if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
					continue
				}
				// unmarshal instructions content
				var actionData metadata.PortalReplacementFeeRequestContent
				err := json.Unmarshal([]byte(inst[3]), &actionData)
				if err != nil {
					return nil, fmt.Errorf("[checkAndSignPortalV4] Can not unmarshal instruction content %v - Error %v\n", inst[3], err)
				}
				tokenID = actionData.TokenID
				hexRawExternalTx = actionData.ExternalRawTx
				utxos = actionData.UTXOs
			}
		// other cases
		default:
			continue
		}

		rawTxBytes, err := hex.DecodeString(hexRawExternalTx)
		if err != nil {
			return nil, fmt.Errorf("[checkAndSignPortalV4] Error when decoding raw tx string: %v", err)
		}

		portalTokenProcessor := portalParam.PortalTokens[tokenID]
		if portalTokenProcessor == nil {
			return nil, fmt.Errorf("[checkAndSignPortalV4] portalTokenProcessor is nil")
		}
		sigs, txHash, err := portalTokenProcessor.PartSignOnRawExternalTx(seedKey, portalParam.MasterPubKeys[tokenID], int(portalParam.NumRequiredSigs), rawTxBytes, utxos)
		if err != nil {
			return nil, fmt.Errorf("[checkAndSignPortalV4] Error when signing raw tx bytes: %v", err)
		}
		pSigs = append(pSigs, &PortalSig{
			TokenID:   tokenID,
			RawTxHash: txHash,
			Sigs:      sigs,
		})
	}

	return pSigs, nil
}
