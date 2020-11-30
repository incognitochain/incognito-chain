package instructions

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/metadata"

	//"github.com/binance-chain/go-sdk/types/msg"
	//"github.com/incognitochain/incognito-chain/relaying/bnb"
	//btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	"github.com/incognitochain/incognito-chain/common"
	bMeta "github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

type portalExchangeRateProcessor struct {
	*portalInstProcessor
}

func (p *portalExchangeRateProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalExchangeRateProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalExchangeRateProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func (p *portalExchangeRateProcessor) buildNewInsts(
	bc bMeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalExchangeRatesAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}
	metaType := actionData.Meta.Type

	//check key from db
	if currentPortalState.ExchangeRatesRequests != nil {
		_, ok := currentPortalState.ExchangeRatesRequests[actionData.TxReqID.String()]
		if ok {
			Logger.log.Errorf("ERROR: exchange rates key is duplicated")

			portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
				SenderAddress: actionData.Meta.SenderAddress,
				Rates:         actionData.Meta.Rates,
				TxReqID:       actionData.TxReqID,
				LockTime:      actionData.LockTime,
			}

			portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

			inst := []string{
				strconv.Itoa(metaType),
				strconv.Itoa(int(shardID)),
				common.PortalExchangeRatesRejectedChainStatus,
				string(portalExchangeRatesContentBytes),
			}

			return [][]string{inst}, nil
		}
	}

	//success
	portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
		SenderAddress: actionData.Meta.SenderAddress,
		Rates:         actionData.Meta.Rates,
		TxReqID:       actionData.TxReqID,
		LockTime:      actionData.LockTime,
	}

	portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

	inst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PortalExchangeRatesAcceptedChainStatus,
		string(portalExchangeRatesContentBytes),
	}

	//update E-R request
	if currentPortalState.ExchangeRatesRequests != nil {
		currentPortalState.ExchangeRatesRequests[actionData.TxReqID.String()] = metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)
	} else {
		//new object
		newExchangeRatesRequest := make(map[string]*metadata.ExchangeRatesRequestStatus)
		newExchangeRatesRequest[actionData.TxReqID.String()] = metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)

		currentPortalState.ExchangeRatesRequests = newExchangeRatesRequest
	}

	return [][]string{inst}, nil
}