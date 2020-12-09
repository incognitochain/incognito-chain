package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	portalMeta "github.com/incognitochain/incognito-chain/portal/metadata"
	"sort"
	"strconv"
)

type portalExchangeRateProcessor struct {
	*portalInstProcessor
}

func (p *portalExchangeRateProcessor) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalExchangeRateProcessor) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalExchangeRateProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func (p *portalExchangeRateProcessor) BuildNewInsts(
	bc bMeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	var actionData portalMeta.PortalExchangeRatesAction
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

			portalExchangeRatesContent := portalMeta.PortalExchangeRatesContent{
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
	portalExchangeRatesContent := portalMeta.PortalExchangeRatesContent{
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
		currentPortalState.ExchangeRatesRequests[actionData.TxReqID.String()] = portalMeta.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)
	} else {
		//new object
		newExchangeRatesRequest := make(map[string]*portalMeta.ExchangeRatesRequestStatus)
		newExchangeRatesRequest[actionData.TxReqID.String()] = portalMeta.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)

		currentPortalState.ExchangeRatesRequests = newExchangeRatesRequest
	}

	return [][]string{inst}, nil
}

func (p *portalExchangeRateProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]bMeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	// parse instruction
	var portingExchangeRatesContent portalMeta.PortalExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &portingExchangeRatesContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of portal exchange rates instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]
	Logger.log.Infof("Portal exchange rates, data input: %+v, status: %+v", portingExchangeRatesContent, reqStatus)

	switch reqStatus {
	case common.PortalExchangeRatesAcceptedChainStatus:
		//save db
		newExchangeRates := portalMeta.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			portingExchangeRatesContent.SenderAddress,
			portingExchangeRatesContent.Rates,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.StorePortalExchangeRateStatus(
			stateDB,
			portingExchangeRatesContent.TxReqID.String(),
			newExchangeRatesStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return nil
		}

		currentPortalState.ExchangeRatesRequests[portingExchangeRatesContent.TxReqID.String()] = newExchangeRates

		Logger.log.Infof("Portal exchange rates, exchange rates request: total exchange rate request %v", len(currentPortalState.ExchangeRatesRequests))

	case common.PortalExchangeRatesRejectedChainStatus:
		//save db
		newExchangeRates := portalMeta.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesRejectedStatus,
			portingExchangeRatesContent.SenderAddress,
			nil,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.StorePortalExchangeRateStatus(
			stateDB,
			portingExchangeRatesContent.TxReqID.String(),
			newExchangeRatesStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return nil
		}
	}

	return nil
}


func PickExchangesRatesFinal(currentPortalState *CurrentPortalState) {
	// sort exchange rate requests by rate
	sumRates := map[string][]uint64{}

	for _, req := range currentPortalState.ExchangeRatesRequests {
		for _, rate := range req.Rates {
			sumRates[rate.PTokenID] = append(sumRates[rate.PTokenID], rate.Rate)
		}
	}

	updateFinalExchangeRates := currentPortalState.FinalExchangeRatesState.Rates()
	if updateFinalExchangeRates == nil {
		updateFinalExchangeRates = map[string]statedb.FinalExchangeRatesDetail{}
	}
	for tokenID, rates := range sumRates {
		// sort rates
		sort.SliceStable(rates, func(i, j int) bool {
			return rates[i] < rates[j]
		})

		// pick one median rate to make final rate for tokenID
		medianRate := calcMedian(rates)

		if medianRate > 0 {
			updateFinalExchangeRates[tokenID] = statedb.FinalExchangeRatesDetail{Amount: medianRate}
		}
	}
	currentPortalState.FinalExchangeRatesState = statedb.NewFinalExchangeRatesStateWithValue(updateFinalExchangeRates)
}