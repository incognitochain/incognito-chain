package portalrelaying

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	"github.com/incognitochain/incognito-chain/common"
)

// RelayingHeaderChainState is state of relaying header chains
// include btc and bnb header chain
type RelayingHeaderChainState struct {
	BNBHeaderChain *bnbrelaying.BNBChainState
	BTCHeaderChain *btcrelaying.BlockChain
}

/*
Portal relaying Producer
*/

func buildRelayingInstsFromActions(
	rc RelayingProcessor,
	relayingState *RelayingHeaderChainState,
	bc metadata.ChainRetriever,
) [][]string {
	actions := rc.GetActions()
	Logger.log.Infof("[Blocks Relaying] - Processing buildRelayingInstsFromActions for %d actions", len(actions))
	// sort push header relaying inst
	actionsGroupByBlockHeight := make(map[uint64][]metadata.RelayingHeaderAction)

	var blockHeightArr []uint64

	for _, action := range actions {
		// parse inst
		var relayingHeaderAction metadata.RelayingHeaderAction
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
			actionsGroupByBlockHeight[blockHeight] = []metadata.RelayingHeaderAction{relayingHeaderAction}
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
			inst := rc.BuildRelayingInst(bc, action, relayingState)
			relayingInsts = append(relayingInsts, inst...)
		}
	}
	return relayingInsts
}

/*
HandleRelayingInsts: beacon producers handle these relaying actions
*/
func HandleRelayingInsts(
	bc metadata.ChainRetriever,
	relayingState *RelayingHeaderChainState,
	relayingProcessors map[int]RelayingProcessor,
) [][]string {
	Logger.log.Info("[Blocks Relaying] - Processing handleRelayingInsts...")
	newInsts := [][]string{}
	// sort RelayingChainsProcessors map to make it consistent for every run
	var metaTypes []int
	for metaType := range relayingProcessors {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		rc := relayingProcessors[metaType]
		insts := buildRelayingInstsFromActions(rc, relayingState, bc)
		newInsts = append(newInsts, insts...)
	}
	return newInsts
}

/*
Portal relaying Process
*/

func ProcessRelayingInstructions(instructions [][]string, relayingState *RelayingHeaderChainState) error {
	// because relaying instructions in received beacon block were sorted already as desired so dont need to do sorting again over here
	for _, inst := range instructions {
		if len(inst) < 4 {
			continue // Not error, just not relaying instruction
		}
		var err error
		switch inst[0] {
		//case strconv.Itoa(metadata.RelayingBNBHeaderMeta):
		//	err = blockchain.processRelayingBNBHeaderInst(inst, relayingState)
		case strconv.Itoa(metadata.RelayingBTCHeaderMeta):
			err = ProcessRelayingBTCHeaderInst(inst, relayingState)
		}
		if err != nil {
			Logger.log.Error(err)
		}
	}

	// store updated relayingState to leveldb with new beacon height
	//err = relayingState.BNBHeaderChain.StoreBNBChainState()
	//if err != nil {
	//	Logger.log.Error(err)
	//}
	return nil
}

func ProcessRelayingBTCHeaderInst(
	instruction []string,
	relayingState *RelayingHeaderChainState,
) error {
	Logger.log.Info("[BTC Relaying] - Processing processRelayingBTCHeaderInst...")
	btcHeaderChain := relayingState.BTCHeaderChain
	if btcHeaderChain == nil {
		return errors.New("[processRelayingBTCHeaderInst] BTC Header chain instance should not be nil")
	}

	if len(instruction) != 4 {
		return nil // skip the instruction
	}

	var relayingHeaderContent metadata.RelayingHeaderContent
	err := json.Unmarshal([]byte(instruction[3]), &relayingHeaderContent)
	if err != nil {
		return err
	}

	headerBytes, err := base64.StdEncoding.DecodeString(relayingHeaderContent.Header)
	if err != nil {
		return err
	}
	var msgBlk *wire.MsgBlock
	err = json.Unmarshal(headerBytes, &msgBlk)
	if err != nil {
		return err
	}
	block := btcutil.NewBlock(msgBlk)
	isMainChain, isOrphan, err := btcHeaderChain.ProcessBlockV2(block, btcrelaying.BFNone)
	if err != nil {
		Logger.log.Errorf("ProcessBlock fail with error: %v", err)
		return err
	}
	Logger.log.Infof("ProcessBlock (%s) success with result: isMainChain: %v, isOrphan: %v", block.Hash(), isMainChain, isOrphan)
	return nil
}
