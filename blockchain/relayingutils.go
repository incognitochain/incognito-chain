package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/pkg/errors"
	lvdbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	"strconv"
)

var btcHeaderChainInstance *btcrelaying.BlockChain = nil

type relayingChain struct {
	actions [][]string
}
type relayingBNBChain struct {
	*relayingChain
}
type relayingBTCChain struct {
	*relayingChain
}
type relayingProcessor interface{
	getActions() [][]string
	putAction(action []string)
	buildRelayingInst(
		blockchain *BlockChain,
		relayingHeaderAction metadata.RelayingHeaderAction,
		relayingState *RelayingHeaderChainState,
	) [][]string
	buildHeaderRelayingInst(
		senderAddressStr string,
		header string,
		blockHeight uint64,
		metaType int,
		shardID byte,
		txReqID common.Hash,
		status string,
	) []string
}
type portalManager struct {
	relayingChains map[int]relayingProcessor
}

func (rChain *relayingChain) getActions() [][]string {
	return rChain.actions
}
func (rChain *relayingChain) putAction(action []string) {
	rChain.actions = append(rChain.actions, action)
}
// buildHeaderRelayingInst builds a new instruction from action received from ShardToBeaconBlock
func (rChain *relayingChain) buildHeaderRelayingInst(
	senderAddressStr string,
	header string,
	blockHeight uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	headerRelayingContent := metadata.RelayingHeaderContent{
		IncogAddressStr: senderAddressStr,
		Header:          header,
		TxReqID:         txReqID,
		BlockHeight:     blockHeight,
	}
	headerRelayingContentBytes, _ := json.Marshal(headerRelayingContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(headerRelayingContentBytes),
	}
}

func (rbnbChain *relayingBNBChain) buildRelayingInst(
	blockchain *BlockChain,
	relayingHeaderAction metadata.RelayingHeaderAction,
	relayingHeaderChain *RelayingHeaderChainState,
) [][]string {
	if relayingHeaderChain == nil {
		Logger.log.Warn("WARN - [buildInstructionsForBNBHeaderRelaying]: relayingHeaderChain is null.")
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}
	meta := relayingHeaderAction.Meta
	// parse and verify header chain
	headerBytes, err := base64.StdEncoding.DecodeString(meta.Header)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot decode header string.%v\n", err)
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	var newHeader rawdbv2.BNBHeader
	err = json.Unmarshal(headerBytes, &newHeader)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot unmarshal header.%v\n", err)
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	if newHeader.Header.Height != int64(relayingHeaderAction.Meta.BlockHeight) {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Block height in metadata is unmatched with block height in new header.")
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	// if valid, create instruction with status accepted
	// if not, create instruction with status rejected
	latestBNBHeader := relayingHeaderChain.BNBHeaderChain.LatestHeader
	var isValid bool
	var err2 error
	relayingHeaderChain.BNBHeaderChain, isValid, err2 = relayingHeaderChain.BNBHeaderChain.ReceiveNewHeader(
		newHeader.Header, newHeader.LastCommit, blockchain.config.ChainParams.BNBRelayingHeaderChainID)
	if err2.(*bnbrelaying.BNBRelayingError) != nil || !isValid {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Verify new header failed. %v\n", err2)
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	// check newHeader is a header contain last commit for one of the header in unconfirmed header list or not\
	// check newLatestBNBHeader is genesis header or not
	genesisHeaderHeight := int64(0)
	genesisHeaderStr := ""
	if blockchain.config.ChainParams.BNBRelayingHeaderChainID == TestnetBNBChainID {
		genesisHeaderHeight = bnbrelaying.TestnetGenesisBlockHeight
		genesisHeaderStr = bnbrelaying.TestnetGenesisHeaderStr
	} else if blockchain.config.ChainParams.BNBRelayingHeaderChainID == MainnetBNBChainID {
		genesisHeaderHeight = bnbrelaying.MainnetGenesisBlockHeight
		genesisHeaderStr = bnbrelaying.MainnetGenesisHeaderStr
	}
	newLatestBNBHeader := relayingHeaderChain.BNBHeaderChain.LatestHeader
	if newLatestBNBHeader != nil && newLatestBNBHeader.Height == genesisHeaderHeight && latestBNBHeader == nil {
		inst1 := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			genesisHeaderStr,
			uint64(genesisHeaderHeight),
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderConfirmedAcceptedChainStatus,
		)

		inst2 := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderUnconfirmedAcceptedChainStatus,
		)
		return [][]string{inst1, inst2}
	}

	if newLatestBNBHeader != nil && latestBNBHeader != nil {
		if newLatestBNBHeader.Height == latestBNBHeader.Height + 1 {
			inst := rbnbChain.buildHeaderRelayingInst(
				relayingHeaderAction.Meta.IncogAddressStr,
				relayingHeaderAction.Meta.Header,
				relayingHeaderAction.Meta.BlockHeight,
				relayingHeaderAction.Meta.Type,
				relayingHeaderAction.ShardID,
				relayingHeaderAction.TxReqID,
				common.RelayingHeaderConfirmedAcceptedChainStatus,
			)
			return [][]string{inst}
		}
	}

	inst := rbnbChain.buildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		common.RelayingHeaderUnconfirmedAcceptedChainStatus,
	)
	return [][]string{inst}
}

func (rbtcChain *relayingBTCChain) buildRelayingInst(
	blockchain *BlockChain,
	relayingHeaderAction metadata.RelayingHeaderAction,
	relayingState *RelayingHeaderChainState,
) [][]string {
	inst := rbtcChain.buildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		common.RelayingHeaderConsideringChainStatus,
	)
	return [][]string{inst}
}

func NewPortalManager() *portalManager {
	rbnbChain := &relayingBNBChain{
		relayingChain: &relayingChain{
			actions: [][]string{},
		},
	}
	rbtcChain := &relayingBTCChain{
		relayingChain: &relayingChain{
			actions: [][]string{},
		},
	}
	return &portalManager{
		relayingChains: map[int]relayingProcessor{
			metadata.RelayingBNBHeaderMeta: rbnbChain,
			metadata.RelayingBTCHeaderMeta: rbtcChain,
		},
	}
}


type RelayingHeaderChainState struct{
	BNBHeaderChain *bnbrelaying.LatestHeaderChain
	BTCHeaderChain *btcrelaying.BlockChain
}

func (bc *BlockChain) InitRelayingHeaderChainStateFromDB(
	db incdb.Database,
	beaconHeight uint64,
) (*RelayingHeaderChainState, error) {
	bnbHeaderChainState, err := getBNBHeaderChainState(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	btcHeaderChain, err := bc.getBTCHeaderChain()
	if err != nil {
		Logger.log.Errorf("Could not get BTC chain instance with error: %v", err)
		return nil, err
	}

	return &RelayingHeaderChainState{
		BNBHeaderChain: bnbHeaderChainState,
		BTCHeaderChain: btcHeaderChain,
	}, nil
}


// getBNBHeaderChainState gets bnb header chain state at beaconHeight
func getBNBHeaderChainState(
	db incdb.Database,
	beaconHeight uint64,
) (*bnbrelaying.LatestHeaderChain, error) {
	relayingStateKey := rawdbv2.NewBNBHeaderRelayingStateKey(beaconHeight)

	relayingStateValueBytes, err := db.Get([]byte(relayingStateKey))
	if err != nil && err != lvdbErrors.ErrNotFound {
		Logger.log.Errorf("getBNBHeaderChainState - Can not get relaying bnb header state from db %v\n", err)
		return nil, err
	}

	var hc bnbrelaying.LatestHeaderChain
	if len(relayingStateValueBytes) > 0 {
		err = json.Unmarshal(relayingStateValueBytes, &hc)
		if err != nil {
			Logger.log.Errorf("getBNBHeaderChainState - Can not unmarshal relaying bnb header state %v\n", err)
			return nil, err
		}
	}
	return &hc, nil
}

// getBTCHeaderChain gets btc header chain as a singleton
func (bc *BlockChain) getBTCHeaderChain() (*btcrelaying.BlockChain, error) {
	btcChainID := bc.config.ChainParams.BNBRelayingHeaderChainID
	relayingChainParams := map[string]*chaincfg.Params{
		TestnetBTCChainID: &chaincfg.TestNet3Params,
		MainnetBTCChainID: &chaincfg.MainNetParams,
	}

	if btcHeaderChainInstance == nil {
		instance, err := btcrelaying.GetChain("btc-blocks", relayingChainParams[btcChainID])
		if err != nil {
			return nil, err
		}
		btcHeaderChainInstance = instance
	}
	return btcHeaderChainInstance, nil
}

// storeBNBHeaderChainState stores bnb header chain state at beaconHeight
func storeBNBHeaderChainState(db incdb.Database,
	beaconHeight uint64,
	bnbHeaderRelaying *bnbrelaying.LatestHeaderChain) error {
	key := rawdbv2.NewBNBHeaderRelayingStateKey(beaconHeight)
	value, err := json.Marshal(bnbHeaderRelaying)
	if err != nil {
		return err
	}
	err = db.Put([]byte(key), value)
	if err != nil {
		return rawdbv2.NewRawdbError(rawdbv2.StoreRelayingBNBHeaderError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func storeRelayingHeaderStateToDB(
	db incdb.Database,
	beaconHeight uint64,
	relayingHeaderState *RelayingHeaderChainState,
) error {
	err := storeBNBHeaderChainState(db, beaconHeight, relayingHeaderState.BNBHeaderChain)
	if err != nil {
		return err
	}
	return nil
}
