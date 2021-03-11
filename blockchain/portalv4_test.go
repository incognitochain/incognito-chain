package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/mocks"
	"github.com/incognitochain/incognito-chain/portal"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/stretchr/testify/suite"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	portalprocessv4.Logger.Init(common.NewBackend(nil).Logger("test", true))
	portaltokensv4.Logger.Init(common.NewBackend(nil).Logger("test", true))
	btcrelaying.Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("This runs before init()!")
	return
}()

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PortalTestSuiteV4 struct {
	suite.Suite
	currentPortalStateForProducer portalprocessv4.CurrentPortalStateV4
	currentPortalStateForProcess  portalprocessv4.CurrentPortalStateV4

	sdb          *statedb.StateDB
	portalParams portalv4.PortalParams
	blockChain   *BlockChain
}

const USER_BTC_ADDRESS_1 = "12ok3D39W4AZj4aF2rmgzqys3BB4uhcXVN"

func (s *PortalTestSuiteV4) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "portal_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	stateDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	s.sdb = stateDB

	s.currentPortalStateForProducer = portalprocessv4.CurrentPortalStateV4{
		UTXOs:               map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx: map[string]map[string]*statedb.ShieldingRequest{},
	}
	s.currentPortalStateForProcess = portalprocessv4.CurrentPortalStateV4{
		UTXOs:               map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx: map[string]map[string]*statedb.ShieldingRequest{},
	}
	s.portalParams = portalv4.PortalParams{
		MultiSigAddresses: map[string]string{
			portalcommonv4.PortalBTCIDStr: "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X",
		},
		MultiSigScriptHexEncode: map[string]string{
			portalcommonv4.PortalBTCIDStr: "",
		},
		PortalTokens: map[string]portaltokensv4.PortalTokenProcessor{
			portalcommonv4.PortalBTCIDStr: &portaltokensv4.PortalBTCTokenProcessor{
				&portaltokensv4.PortalToken{
					ChainID:        "Bitcoin-Testnet",
					MinTokenAmount: 10,
				},
			},
		},
		FeeUnshields: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 100000, // in nano pBTC - 10000 satoshi ~ 4 usd
		},
		BatchNumBlks:               45,
		PortalReplacementAddress:   "",
		MaxFeeForEachStep:          0,
		TimeSpaceForFeeReplacement: 0,
	}
	s.blockChain = &BlockChain{
		config: Config{
			ChainParams: &Params{
				MinBeaconBlockInterval: 40 * time.Second,
				MinShardBlockInterval:  40 * time.Second,
				Epoch:                  100,
				PortalParams: portal.PortalParams{
					PortalParamsV4: map[uint64]portalv4.PortalParams{
						0: s.portalParams,
					},
				},
			},
		},
	}
}

type portalV4InstForProducer struct {
	inst         []string
	optionalData map[string]interface{}
}

func producerPortalInstructionsV4(
	blockchain metadata.ChainRetriever,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	insts []portalV4InstForProducer,
	currentPortalState *portalprocessv4.CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	shardID byte,
	pm map[int]portalprocessv4.PortalInstructionProcessorV4,
) ([][]string, error) {
	var newInsts [][]string

	for _, item := range insts {
		inst := item.inst
		optionalData := item.optionalData

		metaType, _ := strconv.Atoi(inst[0])
		contentStr := inst[1]
		portalProcessor := pm[metaType]
		newInst, err := portalProcessor.BuildNewInsts(
			blockchain,
			contentStr,
			shardID,
			currentPortalState,
			beaconHeight,
			shardHeights,
			portalParams,
			optionalData,
		)
		if err != nil {
			Logger.log.Error(err)
			return newInsts, err
		}

		newInsts = append(newInsts, newInst...)
	}

	return newInsts, nil
}

func processPortalInstructionsV4(
	blockchain metadata.ChainRetriever,
	beaconHeight uint64,
	insts [][]string,
	portalStateDB *statedb.StateDB,
	currentPortalState *portalprocessv4.CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	pm map[int]portalprocessv4.PortalInstructionProcessorV4,
) error {
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
	for _, inst := range insts {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		metaType, _ := strconv.Atoi(inst[0])
		processor := pm[metaType]
		if processor != nil {
			err = processor.ProcessInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
			if err != nil {
				Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
			}
			continue
		}
	}
	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err := portalprocessv4.StorePortalV4StateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

/*
	Shielding Request
*/
type TestCaseShieldingRequest struct {
	tokenID                  string
	incAddressStr            string
	shieldingProof           string
	txID                     string
	isExistsInPreviousBlocks bool
}

type ExpectedResultShieldingRequest struct {
	utxos          map[string]map[string]*statedb.UTXO
	numBeaconInsts uint
	statusInsts    []string
}

func (s *PortalTestSuiteV4) SetupTestShieldingRequest() {
	// do nothing
}

func generateUTXOKeyAndValue(tokenID string, walletAddress string, txHash string, outputIdx uint32, outputAmount uint64) (string, *statedb.UTXO) {
	utxoKey := statedb.GenerateUTXOObjectKey(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx).String()
	utxoValue := statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount)
	return utxoKey, utxoValue
}

func buildTestCaseAndExpectedResultShieldingRequest() ([]TestCaseShieldingRequest, *ExpectedResultShieldingRequest) {
	// build test cases
	testcases := []TestCaseShieldingRequest{
		// valid shielding request
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1MSwxNCw4NSwyMTUsMTI3LDIzOCwxMDcsMTE3LDE5NCw5MSwxNCwyMDUsMzEsMTM3LDE2MCw0MSwxOTQsMTgyLDg2LDIsMjQ3LDc0LDcyLDIzNywxMjQsMTE0LDE3MiwxNDQsMTQsNzMsMjIxLDEzMF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMjM4LDczLDY5LDQ3LDQ4LDE3NSwxNzQsMjA0LDcxLDI1NSwyNTAsMTgsODcsOTcsNDUsNzMsMjAxLDE5NywxMTUsMTM0LDIyOSw1OCwyNDEsMTcyLDI1MSwxMDUsOTYsMTg2LDMyLDIzMiwxNjQsMjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjAwLDEwOCwxNiwyNTUsMjUzLDExMSwxNzQsMTM1LDIwLDEzNSwxODYsMTg2LDc4LDE2OSwxMzIsMjAsMTI3LDI1NSwxODksMTkxLDIxNywxNDgsMTkwLDE4LDYyLDE4OSwxOSwyMiw5MCwzLDI1NSwxNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1OSw2MiwyNywwLDE4MywxNTYsNTgsMTcwLDE4MSwxMTEsMjYsMTc1LDU0LDI1LDIyNSwyMzksMTQwLDEzOSwxMjUsMTE0LDUwLDgxLDk3LDcyLDYxLDE5MCwxNDYsMjMwLDE2NCw0NywyMzYsMTA5XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0OSwyNSw3MywyMDUsNjUsMTQ4LDEwNCwxMjgsMTkxLDkyLDMsNjIsMTY4LDMyLDE2NiwxNTYsMTYxLDIxNiwxMDksMTMyLDI0NSw1MCwxNTEsMjA5LDM2LDE3OSwyNCw0NiwxNjMsMTkwLDIzMywyMTBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMiwxNDMsNTEsMjMzLDM4LDIyMSwxNDMsNDMsMTM0LDExNSwxNTgsMTI3LDkxLDE1MCw4Myw5NywzMiwxODksMjU0LDEzNiwxOTQsMjEwLDE1NywyMjksNzUsNDQsMTAyLDE1NiwyMjcsMTQ0LDEwOCwyM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxNzIsMjI0LDEwLDIwNCwyMTgsMTUyLDE3Miw5NCwyMTUsMjE4LDcxLDE3MSwxLDEyNyw4OSwyNywxMzQsNjQsMTQ4LDEsMjAyLDIxNSwyMTIsMjA4LDE2OCwxMDUsMTI5LDI1NSwxMzQsMTg0LDIyMywxNThdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBalJDS0p1RzMvNlErdVFDaTVPL1MwRVA2VkZOT3hrU25yQTN3RnZ3ZVhKb0NJSFpKVklQVmRsbE16L0JFRTVBU3BadGQ5NHQ4SWhKY0dJS1FnQUNuQjg2MEFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk1NzcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{1}).String(),
			isExistsInPreviousBlocks: false,
		},
		// valid shielding request
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzIzMiwyOSw4Nyw0Nyw1NSwyMDUsMTk1LDYyLDM1LDgwLDMwLDE0MCwxMCwyMTIsODAsMTc0LDEyMCwxNzEsNjgsMjAsMTQ2LDIwMiwyMTMsMjUsMTIxLDIxNiwyMzUsMjM2LDksMTM0LDIzNywyMTVdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzkyLDMwLDYsOTcsMTk1LDYzLDk5LDEsMzMsMjI3LDQyLDIxOCwxNzksMTM1LDE3OSwxNTcsOTAsMjM3LDIwNiwyMTYsMjI0LDI4LDI5LDEzMyw2OCwyMzksMzEsMjAsMTYxLDIwNSwyMzksMTcyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyMjAsMzUsOTksMjI0LDE2MCw3MiwxMTEsMjEzLDE0OCwxOTksMTk5LDM4LDIzNSwxOTIsNzksNTksMTQwLDEsMzgsMzQsMTg0LDExOCwxODEsNTQsODYsOTcsNDAsODgsMjQ4LDI0MSwxOTAsMTU2XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzU3LDE2OSwxNjAsMTcsMTMyLDE4OSw0MCw5Miw1MSwyMzksNjgsMTk2LDI1MywyMCwxNTgsODIsMTgxLDc3LDIwLDEwNiwxNzUsMjIyLDEzLDExOCw5MiwxNTMsMjM5LDE5MCw1NywyMTMsMTY3LDk3XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyNDUsMTIyLDE4NSwxMTQsNDcsNzUsMjI2LDExNCwxNTcsMTcxLDIwOSwyNDQsOTMsMTUzLDIzNiw5Myw2MSwwLDE5NSwxMjYsMTY0LDc0LDEzNywxODEsNjYsMTI1LDEyLDI0MywyNDEsMjMsMjA5LDEyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2MywxMzgsOTEsMTIzLDY0LDksMjI5LDIxLDE3NCwxOTksMSwyMTEsNiw2OCwxNzYsMjI3LDQ1LDIzMyw3MCwxMDMsNDgsOTQsNDEsMTQ5LDE1NSw0NCwyNDAsMTc1LDExNiw1OCw5NiwxNDRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOls4NSwyMjcsMjMsMTUsMjA5LDEzNiwyMTksMTEsOTEsMjIyLDI0LDE3NiwxMzIsMjMxLDE4OCwyNDAsMTgzLDI4LDI0OSwxODUsNTcsMTk5LDE5OCw4NCw4Miw3Miw1NiwxNTMsNjQsMzQsMzEsMzddLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBMVd5OHVRellVOUpqV3hDZXVUb0lGNERvT2RoRUFjdmNiVmwrblQ1ck9qWUNJQk9aU2VaYnVvT0FoS3dzNE1yMDMrMDNBaEhXZlhTMkVNZ2w5VWhaOHZHQUFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6ODAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk0NDcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{2}).String(),
			isExistsInPreviousBlocks: false,
		},
		// invalid shielding request: duplicated shielding proof in previous blocks
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1MSwxNCw4NSwyMTUsMTI3LDIzOCwxMDcsMTE3LDE5NCw5MSwxNCwyMDUsMzEsMTM3LDE2MCw0MSwxOTQsMTgyLDg2LDIsMjQ3LDc0LDcyLDIzNywxMjQsMTE0LDE3MiwxNDQsMTQsNzMsMjIxLDEzMF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMjM4LDczLDY5LDQ3LDQ4LDE3NSwxNzQsMjA0LDcxLDI1NSwyNTAsMTgsODcsOTcsNDUsNzMsMjAxLDE5NywxMTUsMTM0LDIyOSw1OCwyNDEsMTcyLDI1MSwxMDUsOTYsMTg2LDMyLDIzMiwxNjQsMjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjAwLDEwOCwxNiwyNTUsMjUzLDExMSwxNzQsMTM1LDIwLDEzNSwxODYsMTg2LDc4LDE2OSwxMzIsMjAsMTI3LDI1NSwxODksMTkxLDIxNywxNDgsMTkwLDE4LDYyLDE4OSwxOSwyMiw5MCwzLDI1NSwxNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1OSw2MiwyNywwLDE4MywxNTYsNTgsMTcwLDE4MSwxMTEsMjYsMTc1LDU0LDI1LDIyNSwyMzksMTQwLDEzOSwxMjUsMTE0LDUwLDgxLDk3LDcyLDYxLDE5MCwxNDYsMjMwLDE2NCw0NywyMzYsMTA5XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0OSwyNSw3MywyMDUsNjUsMTQ4LDEwNCwxMjgsMTkxLDkyLDMsNjIsMTY4LDMyLDE2NiwxNTYsMTYxLDIxNiwxMDksMTMyLDI0NSw1MCwxNTEsMjA5LDM2LDE3OSwyNCw0NiwxNjMsMTkwLDIzMywyMTBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMiwxNDMsNTEsMjMzLDM4LDIyMSwxNDMsNDMsMTM0LDExNSwxNTgsMTI3LDkxLDE1MCw4Myw5NywzMiwxODksMjU0LDEzNiwxOTQsMjEwLDE1NywyMjksNzUsNDQsMTAyLDE1NiwyMjcsMTQ0LDEwOCwyM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxNzIsMjI0LDEwLDIwNCwyMTgsMTUyLDE3Miw5NCwyMTUsMjE4LDcxLDE3MSwxLDEyNyw4OSwyNywxMzQsNjQsMTQ4LDEsMjAyLDIxNSwyMTIsMjA4LDE2OCwxMDUsMTI5LDI1NSwxMzQsMTg0LDIyMywxNThdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBalJDS0p1RzMvNlErdVFDaTVPL1MwRVA2VkZOT3hrU25yQTN3RnZ3ZVhKb0NJSFpKVklQVmRsbE16L0JFRTVBU3BadGQ5NHQ4SWhKY0dJS1FnQUNuQjg2MEFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk1NzcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: true,
		},
		// invalid shielding request: duplicated shielding proof in the current block
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1MSwxNCw4NSwyMTUsMTI3LDIzOCwxMDcsMTE3LDE5NCw5MSwxNCwyMDUsMzEsMTM3LDE2MCw0MSwxOTQsMTgyLDg2LDIsMjQ3LDc0LDcyLDIzNywxMjQsMTE0LDE3MiwxNDQsMTQsNzMsMjIxLDEzMF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMjM4LDczLDY5LDQ3LDQ4LDE3NSwxNzQsMjA0LDcxLDI1NSwyNTAsMTgsODcsOTcsNDUsNzMsMjAxLDE5NywxMTUsMTM0LDIyOSw1OCwyNDEsMTcyLDI1MSwxMDUsOTYsMTg2LDMyLDIzMiwxNjQsMjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjAwLDEwOCwxNiwyNTUsMjUzLDExMSwxNzQsMTM1LDIwLDEzNSwxODYsMTg2LDc4LDE2OSwxMzIsMjAsMTI3LDI1NSwxODksMTkxLDIxNywxNDgsMTkwLDE4LDYyLDE4OSwxOSwyMiw5MCwzLDI1NSwxNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1OSw2MiwyNywwLDE4MywxNTYsNTgsMTcwLDE4MSwxMTEsMjYsMTc1LDU0LDI1LDIyNSwyMzksMTQwLDEzOSwxMjUsMTE0LDUwLDgxLDk3LDcyLDYxLDE5MCwxNDYsMjMwLDE2NCw0NywyMzYsMTA5XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0OSwyNSw3MywyMDUsNjUsMTQ4LDEwNCwxMjgsMTkxLDkyLDMsNjIsMTY4LDMyLDE2NiwxNTYsMTYxLDIxNiwxMDksMTMyLDI0NSw1MCwxNTEsMjA5LDM2LDE3OSwyNCw0NiwxNjMsMTkwLDIzMywyMTBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMiwxNDMsNTEsMjMzLDM4LDIyMSwxNDMsNDMsMTM0LDExNSwxNTgsMTI3LDkxLDE1MCw4Myw5NywzMiwxODksMjU0LDEzNiwxOTQsMjEwLDE1NywyMjksNzUsNDQsMTAyLDE1NiwyMjcsMTQ0LDEwOCwyM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxNzIsMjI0LDEwLDIwNCwyMTgsMTUyLDE3Miw5NCwyMTUsMjE4LDcxLDE3MSwxLDEyNyw4OSwyNywxMzQsNjQsMTQ4LDEsMjAyLDIxNSwyMTIsMjA4LDE2OCwxMDUsMTI5LDI1NSwxMzQsMTg0LDIyMywxNThdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBalJDS0p1RzMvNlErdVFDaTVPL1MwRVA2VkZOT3hrU25yQTN3RnZ3ZVhKb0NJSFpKVklQVmRsbE16L0JFRTVBU3BadGQ5NHQ4SWhKY0dJS1FnQUNuQjg2MEFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk1NzcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: false,
		},
		// invalid shielding request: invalid proof (invalid memo)
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE4Miw3MCwzOSwxOTksNjksMTEzLDEyNiwyMDYsMjQxLDI1LDE1NCwyMzEsNzgsMTY4LDE3MSwxOTgsMjUyLDE5MCwyNTMsMTA3LDE3MywyMjcsOTAsMjMyLDMwLDE3MCwxOTAsMjIzLDE1LDE2MSwyMjMsNjddLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE3Myw5OCwxOTIsMTcxLDEwMSwxNDIsMTIxLDE5MiwyMDYsOTUsMTA1LDE3MSwxODUsMTQ0LDI0LDExLDIxOSwyMywxODYsMTgwLDI0NiwyMCwxMzMsNzgsMTUsNzIsNzgsMjksOSw5NSwxNTUsMjUyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls5MywyMTgsMTA3LDE1MywxMCwyMzQsMjUzLDMwLDMzLDExMywxODIsMTE1LDEzLDE3OSwyMjQsMTc3LDYsMTMwLDQ2LDEzNSw3OCwxMzUsMjQyLDI1NCwxOTUsOTAsOTUsMTQxLDk1LDIwNywxOSwyNTRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzMiw5OCwxNjMsMjA1LDY4LDExOSwxMzQsMTQsMTkwLDE5Miw4MSw2Niw3MiwxMTIsMzQsMTEyLDIzLDE3MywyMDYsMTM5LDE1MSwyNDUsNTIsNTksNTQsODIsNjcsMTg3LDEzNywyNDksMTA5LDE2M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTM4LDQwLDE1MCwzMiwxOTAsMTc1LDIxNCwzOCw1MSw3Nyw0MCwyMDAsODksMTEyLDIwNywxNywyMjcsMiw0Niw5OSwyNTAsNzksMzYsMTg2LDEwMCw4OCw0NywxMTgsMTA5LDU1LDEzNCwxNjZdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzcyLDMsMTg1LDIyNCw0OCwxMDgsMjIyLDE5LDYzLDYyLDY0LDgsMTIzLDE5MywxNjksNTAsMSwyMyw4MywxODgsNzMsMTM5LDE1MSwxMzMsMTY0LDE4NywyMTksMzYsMjMwLDgwLDE2MiwxNjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTYxLDI0Niw0MywxODAsMTEwLDE2NywyMzEsMTQ3LDg3LDkxLDc3LDYwLDIxNSwxMzIsMTA5LDIxNCw2Myw5LDE3LDQ2LDMsMjMwLDEwNSwxNDAsMTkxLDEwMCwyMTIsMTE1LDEwMiwxNzYsMjM0LDE3MF0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbMjA1LDIyOSwyNDUsMTc5LDExNCwzMSwxNzIsMTkwLDkzLDEwNiwyMiwxNzMsNDIsMTU5LDE0OSwyMjAsNjEsMyw0NSwxMTUsODQsNDksMzksNTEsMjA0LDIxLDE1MiwxNiwxNzksMTY0LDE3MCwxNDddLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlJ6QkVBaUE0OSs0QUx0dnR3VXViMGgxNmE4STQybW9jaW9teGpXaXBTUzdCdGZpMTVRSWdNTmxGaEMzTlVuUzNsRWFMMm45TmtEbEFIa25DMVhDMmRoT3Y4bThiWjZvQklRUFBJQk5QVmtqaTl2RjNwbFVySmJxWDVGc1VoVTVQRXlwSzhQNTlxVC9FL0E9PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZOTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NTAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTkyNDcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxMTgsMjMwLDkzLDE4NywyNTAsODAsMjEwLDEwOCw4NCwxOCwxMiwyMTUsMTY1LDg5LDIyMiwzNiwxODMsNzEsMywyNDUsNjcsMTAwLDI0OSw4LDYsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{4}).String(),
			isExistsInPreviousBlocks: false,
		},
	}

	walletAddress := "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X"

	// build expected results
	var txHash string
	var outputIdx uint32
	var outputAmount uint64

	txHash = "251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355"
	outputIdx = 1
	outputAmount = 400

	key1, value1 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount)

	txHash = "b44f6c7c896757abe7afd6ac083c2930f1d0f57a356887e872f3b88bba5ea0b7"
	outputIdx = 1
	outputAmount = 800

	key2, value2 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount)

	expectedRes := &ExpectedResultShieldingRequest{
		utxos: map[string]map[string]*statedb.UTXO{
			portalcommonv4.PortalBTCIDStr: {
				key1: value1,
				key2: value2,
			},
		},
		numBeaconInsts: 5,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildPortalShieldingRequestAction(
	tokenID string,
	incAddressStr string,
	shieldingProof string,
	txID string,
	shardID byte,
) []string {
	data := metadata.PortalShieldingRequest{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalV4ShieldingRequestMeta,
		},
		TokenID:         tokenID,
		IncogAddressStr: incAddressStr,
		ShieldingProof:  shieldingProof,
	}
	txIDHash, _ := common.Hash{}.NewHashFromStr(txID)
	actionContent := metadata.PortalShieldingRequestAction{
		Meta:    data,
		TxReqID: *txIDHash,
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalV4ShieldingRequestMeta), actionContentBase64Str}
}

func buildShieldingRequestActionsFromTcs(tcs []TestCaseShieldingRequest, shardID byte, shardHeight uint64) []portalV4InstForProducer {
	insts := []portalV4InstForProducer{}

	for _, tc := range tcs {
		inst := buildPortalShieldingRequestAction(
			tc.tokenID, tc.incAddressStr, tc.shieldingProof, tc.txID, shardID)
		insts = append(insts, portalV4InstForProducer{
			inst: inst,
			optionalData: map[string]interface{}{
				"isExistProofTxHash": tc.isExistsInPreviousBlocks,
			},
		})
	}

	return insts
}

func getBlockCypherAPI(networkName string) gobcy.API {
	//explicitly
	bc := gobcy.API{}
	bc.Token = "a8ed119b4edf4f609a83bd3fbe9a3831"
	bc.Coin = "btc"        //options: "btc","bcy","ltc","doge"
	bc.Chain = networkName //depending on coin: "main","test3","test"
	return bc
}

func buildBTCBlockFromCypher(networkName string, blkHeight int) (*btcutil.Block, error) {
	bc := getBlockCypherAPI(networkName)
	cypherBlock, err := bc.GetBlock(blkHeight, "", nil)
	if err != nil {
		return nil, err
	}
	prevBlkHash, _ := chainhash.NewHashFromStr(cypherBlock.PrevBlock)
	merkleRoot, _ := chainhash.NewHashFromStr(cypherBlock.MerkleRoot)
	msgBlk := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(cypherBlock.Ver),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  cypherBlock.Time,
			Bits:       uint32(cypherBlock.Bits),
			Nonce:      uint32(cypherBlock.Nonce),
		},
		Transactions: []*wire.MsgTx{},
	}
	blk := btcutil.NewBlock(&msgBlk)
	blk.SetHeight(int32(blkHeight))
	return blk, nil
}

func setGenesisBlockToChainParams(networkName string, genesisBlkHeight int) (*chaincfg.Params, error) {
	blk, err := buildBTCBlockFromCypher(networkName, genesisBlkHeight)
	if err != nil {
		return nil, err
	}

	chainParams := chaincfg.TestNet3Params
	chainParams.GenesisBlock = blk.MsgBlock()
	chainParams.GenesisHash = blk.Hash()
	return &chainParams, nil
}

func (s *PortalTestSuiteV4) TestShieldingRequest() {
	fmt.Println("Running TestShieldingRequest - beacon height 1003 ...")

	networkName := "test3"
	genesisBlockHeight := 1939008
	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	dbName := "btc-blocks-test"
	btcChain, err := btcrelaying.GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	defer os.RemoveAll(dbName)

	if err != nil {
		s.FailNow(fmt.Sprintf("Could not get chain instance with err: %v", err), nil)
		return
	}

	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+10; i++ {
		blk, err := buildBTCBlockFromCypher(networkName, i)
		if err != nil {
			s.FailNow(fmt.Sprintf("buildBTCBlockFromCypher fail on block %v: %v\n", i, err), nil)
			return
		}
		isMainChain, isOrphan, err := btcChain.ProcessBlockV2(blk, 0)
		if err != nil {
			s.FailNow(fmt.Sprintf("ProcessBlock fail on block %v: %v\n", i, err))
			return
		}
		if isOrphan {
			s.FailNow(fmt.Sprintf("ProcessBlock incorrectly returned block %v is an orphan\n", i))
			return
		}
		fmt.Printf("Block %s (%d) is on main chain: %t\n", blk.Hash(), blk.Height(), isMainChain)
		time.Sleep(500 * time.Millisecond)
	}

	bc := new(mocks.ChainRetriever)
	bc.On("GetBTCHeaderChain").Return(btcChain)

	pm := portal.NewPortalManager()
	beaconHeight := uint64(1003)
	shardHeight := uint64(1003)
	shardHeights := map[byte]uint64{
		0: uint64(1003),
	}
	shardID := byte(0)

	s.SetupTestShieldingRequest()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultShieldingRequest()

	// build actions from testcases
	instsForProducer := buildShieldingRequestActionsFromTcs(testcases, shardID, shardHeight)

	// producer instructions
	newInsts, err := producerPortalInstructionsV4(
		bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV4)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructionsV4(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV4)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
	}

	s.Equal(expectedResult.utxos, s.currentPortalStateForProducer.UTXOs)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

func TestPortalSuiteV4(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV4))
}
