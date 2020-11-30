package instructions

import (
	"encoding/base64"
	"encoding/json"
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
)

func buildRelayingInstsFromActions(
	rc relayingProcessor,
	relayingState *RelayingHeaderChainState,
	blockchain *BlockChain,
) [][]string {
	actions := rc.getActions()
	Logger.log.Infof("[Blocks Relaying] - Processing buildRelayingInstsFromActions for %d actions", len(actions))
	// sort push header relaying inst
	actionsGroupByBlockHeight := make(map[uint64][]metadata2.RelayingHeaderAction)

	var blockHeightArr []uint64

	for _, action := range actions {
		// parse inst
		var relayingHeaderAction metadata2.RelayingHeaderAction
		relayingHeaderActionBytes, err := base64.StdEncoding.DecodeString(action[1])
		if err != nil {
			continue
		}
		err = json.Unmarshal(relayingHeaderActionBytes, &relayingHeaderAction)
		if err != nil {
			continue
		}

		// get blockHeight in action
		blockHeight := relayingHeaderAction.Meta.BlockHeight

		// add to blockHeightArr
		if isExist, _ := common.SliceExists(blockHeightArr, blockHeight); !isExist {
			blockHeightArr = append(blockHeightArr, blockHeight)
		}

		// add to actionsGroupByBlockHeight
		if actionsGroupByBlockHeight[blockHeight] != nil {
			actionsGroupByBlockHeight[blockHeight] = append(actionsGroupByBlockHeight[blockHeight], relayingHeaderAction)
		} else {
			actionsGroupByBlockHeight[blockHeight] = []metadata2.RelayingHeaderAction{relayingHeaderAction}
		}
	}

	// sort blockHeightArr
	sort.Slice(blockHeightArr, func(i, j int) bool {
		return blockHeightArr[i] < blockHeightArr[j]
	})

	relayingInsts := [][]string{}
	for _, value := range blockHeightArr {
		blockHeight := uint64(value)
		actions := actionsGroupByBlockHeight[blockHeight]
		for _, action := range actions {
			inst := rc.buildRelayingInst(blockchain, action, relayingState)
			relayingInsts = append(relayingInsts, inst...)
		}
	}
	return relayingInsts
}

func (blockchain *BlockChain) handleRelayingInsts(
	relayingState *RelayingHeaderChainState,
	pm *portalManager,
) [][]string {
	Logger.log.Info("[Blocks Relaying] - Processing handleRelayingInsts...")
	newInsts := [][]string{}
	// sort relayingChains map to make it consistent for every run
	var metaTypes []int
	for metaType := range pm.relayingChains {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		rc := pm.relayingChains[metaType]
		insts := buildRelayingInstsFromActions(rc, relayingState, blockchain)
		newInsts = append(newInsts, insts...)
	}
	return newInsts
}
