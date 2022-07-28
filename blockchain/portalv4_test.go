package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"

	"github.com/incognitochain/incognito-chain/metadata/rpccaller"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/mocks"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/stretchr/testify/mock"
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

const USER_BTC_ADDRESS_1 = "mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt"
const USER_BTC_ADDRESS_2 = "n4VQ5YdHf7hLQ2gWQYYrcxoE5B7nWuDFNF"
const PORTALV4_USER_INC_ADDRESS_1 = "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
const PORTALV4_USER_INC_ADDRESS_2 = "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"
const PORTALV4_USER_INC_ADDRESS_3 = "12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy"
const PORTALV4_USER_INC_ADDRESS_4 = "12S4NL3DZ1KoprFRy1k5DdYSXUq81NtxFKdvUTP3PLqQypWzceL5fBBwXooAsX5s23j7cpb1Za37ddmfSaMpEJDPsnJGZuyWTXJSZZ5"

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
		UTXOs:                     map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx:       map[string]map[string]*statedb.ShieldingRequest{},
		WaitingUnshieldRequests:   map[string]map[string]*statedb.WaitingUnshieldRequest{},
		ProcessedUnshieldRequests: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
	}
	s.currentPortalStateForProcess = portalprocessv4.CurrentPortalStateV4{
		UTXOs:                     map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx:       map[string]map[string]*statedb.ShieldingRequest{},
		WaitingUnshieldRequests:   map[string]map[string]*statedb.WaitingUnshieldRequest{},
		ProcessedUnshieldRequests: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
	}
	s.portalParams = portalv4.PortalParams{
		MasterPubKeys: map[string][][]byte{
			portal.TestnetPortalV4BTCID: {
				{0x3, 0xb2, 0xd3, 0x16, 0x7d, 0x94, 0x9c, 0x25, 0x3, 0xe6, 0x9c, 0x9f, 0x29, 0x78, 0x7d, 0x9c, 0x8, 0x8d, 0x39, 0x17, 0x8d, 0xb4, 0x75, 0x40, 0x35, 0xf5, 0xae, 0x6a, 0xf0, 0x17, 0x12, 0x11, 0x0},
				{0x3, 0x98, 0x7a, 0x87, 0xd1, 0x99, 0x13, 0xbd, 0xe3, 0xef, 0xf0, 0x55, 0x79, 0x2, 0xb4, 0x90, 0x57, 0xed, 0x1c, 0x9c, 0x8b, 0x32, 0xf9, 0x2, 0xbb, 0xbb, 0x85, 0x71, 0x3a, 0x99, 0x1f, 0xdc, 0x41},
				{0x3, 0x73, 0x23, 0x5e, 0xb1, 0xc8, 0xf1, 0x84, 0xe7, 0x59, 0x17, 0x6c, 0xe3, 0x87, 0x37, 0xb7, 0x91, 0x19, 0x47, 0x1b, 0xba, 0x63, 0x56, 0xbc, 0xab, 0x8d, 0xcc, 0x14, 0x4b, 0x42, 0x99, 0x86, 0x1},
				{0x3, 0x29, 0xe7, 0x59, 0x31, 0x89, 0xca, 0x7a, 0xf6, 0x1, 0xb6, 0x35, 0x67, 0x3d, 0xb1, 0x53, 0xd4, 0x19, 0xd7, 0x6, 0x19, 0x3, 0x2a, 0x32, 0x94, 0x57, 0x76, 0xb2, 0xb3, 0x80, 0x65, 0xe1, 0x5d},
						},
		},
		NumRequiredSigs: 3,
		GeneralMultiSigAddresses: map[string]string{
			portal.TestnetPortalV4BTCID: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
		},
		PortalTokens: map[string]portaltokensv4.PortalTokenProcessor{
			portal.TestnetPortalV4BTCID: &portaltokensv4.PortalBTCTokenProcessor{
				PortalToken: &portaltokensv4.PortalToken{
					ChainID:             "Bitcoin-Testnet",
					MinTokenAmount:      10,
					MultipleTokenAmount: 10,
					ExternalInputSize:   130,
					ExternalOutputSize:  43,
					ExternalTxMaxSize:   1024,
				},
				ChainParam:    &chaincfg.TestNet3Params,
				PortalTokenID: portal.TestnetPortalV4BTCID,
			},
		},
		DefaultFeeUnshields: map[string]uint64{
			portal.TestnetPortalV4BTCID: 100000, // in nano pBTC - 10000 satoshi
		},
		MinShieldAmts: map[string]uint64{
			portal.TestnetPortalV4BTCID: 10, // in nano pBTC - 1 satoshi
		},
		MinUnshieldAmts: map[string]uint64{
			portal.TestnetPortalV4BTCID: 1000000, // in nano pBTC - 100000 satoshi
		},
		DustValueThreshold: map[string]uint64{
			portal.TestnetPortalV4BTCID: 1e9, // in nano pBTC - 1e8 satoshi
		},
		MinUTXOsInVault: map[string]uint64{
			portal.TestnetPortalV4BTCID: 1,
		},
		MaxUnshieldFees: map[string]uint64{
			portal.TestnetPortalV4BTCID: 1000000,
		},
		BatchNumBlks:                45,
		PortalReplacementAddress:    "",
		MaxFeePercentageForEachStep: 20,
		TimeSpaceForFeeReplacement:  2 * time.Minute,
	}

	tempPortalParam := &portal.PortalParams{
		PortalParamsV4: map[uint64]portalv4.PortalParams{
			0: s.portalParams,
		},
	}

	config.AbortParam()
	config.Param().BlockTime.MinBeaconBlockInterval = 40 * time.Second
	config.Param().BlockTime.MinShardBlockInterval = 40 * time.Second
	config.Param().EpochParam.NumberOfBlockInEpoch = 100
	portal.SetupPortalParam(tempPortalParam)
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
	err := portalprocessv4.StorePortalV4StateToDB(portalStateDB, currentPortalState, portalParams)
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

	shieldAmtInPubToken uint64
	shieldAmtInPToken   uint64
	externalTxID        string
	txOutIndex          uint32

	isValidRequest bool
}

type ExpectedResultShieldingRequest struct {
	utxos          map[string]map[string]*statedb.UTXO
	shieldRequests map[string][]*statedb.ShieldingRequest
	numBeaconInsts uint
	statusInsts    []string
}

func (s *PortalTestSuiteV4) SetupTestShieldingRequest() {
	// do nothing
}

func generateUTXOKeyAndValue(tokenID string, walletAddress string, txHash string, outputIdx uint32, outputAmount uint64, chainCodeSeed string) (string, *statedb.UTXO) {
	utxoKey := statedb.GenerateUTXOObjectKey(tokenID, walletAddress, txHash, outputIdx).String()
	utxoValue := statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount, chainCodeSeed)
	return utxoKey, utxoValue
}

func (s *PortalTestSuiteV4) buildTestCaseAndExpectedResultShieldingRequest() ([]TestCaseShieldingRequest, *ExpectedResultShieldingRequest) {
	// build test cases
	testcases := []TestCaseShieldingRequest{
		// valid shielding request
		{
			tokenID:                  portal.TestnetPortalV4BTCID,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_1,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzI0OSwxODEsMzAsMTA4LDI0NiwxOTUsMTQ0LDIwNiwxNjgsMSwyMzAsMTk4LDE1NiwxMjcsOTUsMTQ2LDE1NSw0Niw5MSwyMyw4MywxOTMsMSwyMzksNDYsMTkwLDIxNSw0OCwxNjIsODAsMTAyLDE5M10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyMzYsNjcsOTMsNDYsOTUsMTUwLDIxMiw1OCwyOCwxOTgsMjI3LDY2LDIzMSwyLDE5NiwxNjUsMTc2LDE1MywxNjQsMTUsMjQxLDExMSwxNjIsNzgsMjA4LDYyLDU1LDEyNiw2MiwxOSwyMzUsMTQ3XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyMzksMjMsOCwxMDcsNTcsMTQ2LDE0NCwyNTEsOCw5NywxNDIsMjAsMTUsMjAyLDE1NywzOSwxMCw4OSwxOSwxNTYsMjEsMzQsMTM1LDEzMiwxODcsMTA0LDE5NywxMjQsNTYsNzEsMjE3LDEyOF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTg5LDI2LDE3OSwyMDEsMjUyLDE3OCwyMCwzMCwyNDUsMjAyLDE5OCwxNzMsMTUxLDEzOSwyNTUsMTEsMjgsMzcsOTQsNTgsNDksMTU0LDMsNTAsODAsMTgzLDE4MiwxNjQsMTcxLDQ0LDk3LDIxN10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxNDYsNDcsMjIzLDUwLDEwMywxMzUsMTE4LDQ2LDE1MywxNDAsMTI0LDIzNCwyOCwyMzQsMTMsMTEyLDQxLDIzOSw4MSwyMzEsMjA0LDE3LDIxMiwyMzEsNzAsMTQxLDI0OSwxMjUsODEsMTQyLDE4MSwzOV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbOTksMjAsMjAsNjksMjM4LDExOSw0MywxNTQsMTA0LDE0OSwyOCw4Miw5MSw5MywxNjQsMTM2LDE0NCwyMzUsMTg4LDE4Niw4MCwyMCwxODQsNzMsNDgsMjM5LDE5MCwxNTIsMTg0LDk3LDI0LDE1NV0sIklzTGVmdCI6dHJ1ZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoyLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxMDYsMTAwLDQzLDU0LDk2LDY4LDIwMyw3MywxMjMsMTA3LDM5LDYyLDIzMCwxNzAsMjIyLDIxOSwxMjEsMTA1LDEwNCwxMzUsODUsMjgsNzUsMTg0LDY5LDI1MiwxMzUsOTAsMjA3LDcxLDEwMSwyMjVdLCJJbmRleCI6MH0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3MjkzfV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTg5OTk4NDcsIlBrU2NyaXB0IjoiQUJTK3U3eFQxSEE4NExqR0I3SURFaFNENDB5MTl3PT0ifSx7IlZhbHVlIjoxMDAwMDAwLCJQa1NjcmlwdCI6IkFDQXpZVmQzb1BDLy9UdEZMQlBSOWt6Y2NVMlU0QXJYYkRNSXpBenJQS2NvYlE9PSJ9XSwiTG9ja1RpbWUiOjE5NzAyMjd9LCJCbG9ja0hhc2giOlsxODksMzgsMTM5LDEwMiwxNTIsNDcsMTE1LDcwLDE2OCw2MSwxODQsMTc2LDE3OCw5OSwwLDE4NywxMDAsMjQ4LDIwNCwyNDAsNzUsNzMsNjcsMTIyLDE1LDAsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{1}).String(),
			isExistsInPreviousBlocks: false,

			shieldAmtInPubToken: 0.01 * 1e8,
			shieldAmtInPToken:   0.01 * 1e9,
			externalTxID:        "571e129ffbdcd8ccc5386c1814bb836d904eb09572f5e6945e0f053deb122afe",
			txOutIndex:          1,

			isValidRequest: true,
		},
		// valid shielding request: different user incognito address
		// multisig address: tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe
		{
			tokenID:                  portal.TestnetPortalV4BTCID,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_2,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6Wzg2LDE2LDI3LDE1MywzMSwyMjAsMjM0LDEyNywyNTIsNTksMTk1LDIxLDY0LDMsMTU1LDE4NSw5OCwxODMsMTc5LDIyNyw5NSwxNDIsMjcsMjM4LDI0MCwxNzksMjIyLDUsMjE5LDEwNCw4MiwyMTRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6Wzg1LDE2OSwyNDIsMTEsMTQzLDQzLDE2NSwxNjAsMjE2LDQ2LDgsNzcsNDksMTAxLDExNywxMTMsMjI4LDU1LDE1MywyMjYsMTA5LDQ0LDIxNiwxNDEsMjQ4LDI0LDEsMjMyLDkxLDE0LDE2Niw5Nl0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTYwLDE1LDEyOCw3OSwxMzIsMTQ0LDEwMiwxOTIsODgsMzcsMjM0LDgxLDMzLDE2MCwyMjAsMTAwLDExOSwyMywyMTAsMTcxLDIxLDI1LDEwOSwyNDQsMjM5LDIyMSwxNiwyOCwxOSw2NywxNTYsN10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyMTMsMTU2LDE1MSwxODAsMTA2LDEzMiw1OSwxMzUsMjM1LDE0MCwxNDMsNjAsMTMsMTk0LDE5MCwyNTUsMjEwLDE5OSw4MiwxMjgsNzAsNTcsMTY5LDE2Niw5NSw1OSwxMDIsMTQxLDEwNCwyMjUsMjE3LDE0NV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTYyLDMzLDE3NiwyMjYsMjQyLDEwMywxMzksMTE1LDIyMCwxMzgsMiw5NCwxNDksMTIsMTQ0LDE2OCwxOSw2MSwyNDAsMjQxLDEyLDU1LDIyMCw5MSwyNDksNTAsMTMxLDE4MiwzNSwyNTUsOTAsMTE1XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsxODksMTg3LDIyNCwxMTAsMjUwLDE1OCwxNTUsMTY1LDgxLDExNiw5NSw0LDIyNSwxNDQsNjQsMTQ2LDQ4LDI0NCwyMzUsMjM5LDE4NCw2LDE3NywyNiwxNTgsMjIxLDEwOSwxNzMsMTk1LDEwNiwyNywyNDldLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEwMCwxOCw3NSwyMTUsMzAsMTM1LDE5OSwxMzgsMzQsNzAsNTYsMzgsMjI1LDI0MSwxMDksMTkzLDE2LDIzOSwxNTEsMTg1LDQ3LDcwLDEwNiwyNCwxNDQsOTEsMCwxNzUsNSwyMjMsMTUyLDI0OV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbODAsMTQzLDEwMCwyMDQsMzIsMTI4LDEsMTI1LDEzMiwwLDEzMSwxOTEsMTg3LDI0MSwxNjgsMjEyLDE2MSwxNDcsNTEsMTE1LDE4NCwxOTcsMTcwLDEyMSwyMCw4OSwxODcsOTUsMCwyNCwxODksMjE1XSwiSXNMZWZ0Ijp0cnVlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjIsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzUzLDIyLDk1LDI3LDE2MiwyMywxNDEsMTM3LDE2Myw0Myw1OCwxOTcsMTQ5LDIzNiwyNiwyMzUsMzcsMTc5LDIwMyw5MSw2MSw2Miw2NCwzLDIxNiwxMTIsODEsOTUsMTk1LDIxMiw1NywyMTddLCJJbmRleCI6MH0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3MjkzfV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTY3ODI4NjQsIlBrU2NyaXB0IjoiQUJRbVNVTlVDNUhURmM3bUVnTEVqVm1UWTkxTTRBPT0ifSx7IlZhbHVlIjoyMDAwMDAsIlBrU2NyaXB0IjoiQUNDR0thWUhjS0h4VFNHQ2lPL1dDQStoajFDZTl6djRmNU10UTl3N3l5VXo1Zz09In1dLCJMb2NrVGltZSI6MTk3MDIzMH0sIkJsb2NrSGFzaCI6WzU4LDUsMjIyLDI0NSw1NywxNDgsMTg2LDEzNSwyMTAsMjM5LDk3LDEwMiw2MSw1NiwxODEsMjU1LDEwMSwxMTcsMjExLDE2NSwxNTcsNTMsMTYwLDE2NywxNDUsNjEsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{2}).String(),
			isExistsInPreviousBlocks: false,

			shieldAmtInPubToken: 0.002 * 1e8,
			shieldAmtInPToken:   0.002 * 1e9,
			externalTxID:        "3b2a29cabf5c354c5735aae98d433816af8e3de3c780db421ecfb6006400d84a",
			txOutIndex:          1,

			isValidRequest: true,
		},
		// valid shielding request: the same user incognito address
		{
			tokenID:                  portal.TestnetPortalV4BTCID,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_1,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzgsMjcsNzAsNzQsMTc5LDIyLDE4MCw2NSwyMDEsMTU4LDIwOSwyNTIsNDMsMTE0LDExMiwyNDMsMzUsMjM4LDM1LDEzMywyOSw0MiwxNjAsMTAwLDE1NiwyMzgsMjMwLDEwLDMzLDMzLDIyMywyNDFdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzI4LDIzMSwyMzQsMTEyLDEyMSwxMjcsMjUsOTAsMTYyLDg1LDMsMjQzLDE5NSw4NCwyMzcsMTMsMTM1LDY4LDQ4LDIwNiw1MSwyMTgsOSw5NywxNDMsMjE3LDE2NSwxNCwyNDAsMTQxLDQwLDE3M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNTYsMzIsMTk3LDk5LDc1LDE3MywyMTIsMjI3LDE5OCwxMyw4OSwxNjIsNjAsMTEwLDM3LDE2LDU0LDk3LDEwLDIyNCw5Myw3MSw0MiwyMzMsMTQ1LDExMiwxOTksMTgyLDE2LDM0LDE1OSwyNDZdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzQwLDEyMSwyMzUsMjExLDIyNywxMjcsMTQsMTYsMjIwLDIzNywxMTcsMTEwLDIzOCwxNDQsMSwxMzUsMTIsMjI5LDM3LDE1MSw1LDEwMywxMTcsNjYsNTMsMTYwLDM0LDI0Miw0OSwxMDMsODksMTQxXSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzY0LDQ5LDE5NiwxMjQsMTI0LDEwOSwxMTMsMTksMTc0LDI0MCwzMywyMDQsNzcsMjA0LDE3OSw1MSw4NCwxODUsMjQ5LDEwOSwxMDQsMTIwLDE2Niw3NiwyNDMsODIsMTU1LDEwNyw0LDI0OCwyNTMsMTldLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTE1LDE3NSw1LDE0NCwxMzEsMjQ4LDIzLDc0LDE3OSwxMDcsMTQ1LDI1MSwxNTIsMTg5LDMzLDIyNCwyMDIsMTIsMjIsMTU2LDE2NSwxODksMTc2LDMwLDE3NCwxMTYsMTI1LDk2LDE4NCwxMTcsMTgyLDE5NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MiwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbMTk2LDIyMSwxMDMsMTk0LDcxLDY5LDE4NCw5NCwzOCwxNTgsMjIzLDUwLDE1NCw2MywyMDMsMTQ4LDU0LDExNywyMzQsMjQwLDg2LDIzNiw0MiwxODMsMTIsMTMsMjU0LDE2NywxNTIsMTc4LDI1LDUzXSwiSW5kZXgiOjF9LCJTaWduYXR1cmVTY3JpcHQiOiIiLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5M31dLCJUeE91dCI6W3siVmFsdWUiOjMwMDAwMCwiUGtTY3JpcHQiOiJBQ0F6WVZkM29QQy8vVHRGTEJQUjlremNjVTJVNEFyWGJETUl6QXpyUEtjb2JRPT0ifSx7IlZhbHVlIjo3MzgwMCwiUGtTY3JpcHQiOiJBQlQ1UGx2VmdZS01vcmNuZ0I3UW1PSjNncG1GZkE9PSJ9XSwiTG9ja1RpbWUiOjE5NzAyMjd9LCJCbG9ja0hhc2giOlsxODksMzgsMTM5LDEwMiwxNTIsNDcsMTE1LDcwLDE2OCw2MSwxODQsMTc2LDE3OCw5OSwwLDE4NywxMDAsMjQ4LDIwNCwyNDAsNzUsNzMsNjcsMTIyLDE1LDAsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: false,

			shieldAmtInPubToken: 0.003 * 1e8,
			shieldAmtInPToken:   0.003 * 1e9,
			externalTxID:        "b515a6bf13c4f97b8c4683ea2d040443d1d83e1909d164a195de8b3b069e3b00",
			txOutIndex:          0,

			isValidRequest: true,
		},
		// invalid shielding request: duplicated shielding proof in previous blocks
		{
			tokenID:                  portal.TestnetPortalV4BTCID,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_1,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzI0OSwxODEsMzAsMTA4LDI0NiwxOTUsMTQ0LDIwNiwxNjgsMSwyMzAsMTk4LDE1NiwxMjcsOTUsMTQ2LDE1NSw0Niw5MSwyMyw4MywxOTMsMSwyMzksNDYsMTkwLDIxNSw0OCwxNjIsODAsMTAyLDE5M10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyMzYsNjcsOTMsNDYsOTUsMTUwLDIxMiw1OCwyOCwxOTgsMjI3LDY2LDIzMSwyLDE5NiwxNjUsMTc2LDE1MywxNjQsMTUsMjQxLDExMSwxNjIsNzgsMjA4LDYyLDU1LDEyNiw2MiwxOSwyMzUsMTQ3XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyMzksMjMsOCwxMDcsNTcsMTQ2LDE0NCwyNTEsOCw5NywxNDIsMjAsMTUsMjAyLDE1NywzOSwxMCw4OSwxOSwxNTYsMjEsMzQsMTM1LDEzMiwxODcsMTA0LDE5NywxMjQsNTYsNzEsMjE3LDEyOF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTg5LDI2LDE3OSwyMDEsMjUyLDE3OCwyMCwzMCwyNDUsMjAyLDE5OCwxNzMsMTUxLDEzOSwyNTUsMTEsMjgsMzcsOTQsNTgsNDksMTU0LDMsNTAsODAsMTgzLDE4MiwxNjQsMTcxLDQ0LDk3LDIxN10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxNDYsNDcsMjIzLDUwLDEwMywxMzUsMTE4LDQ2LDE1MywxNDAsMTI0LDIzNCwyOCwyMzQsMTMsMTEyLDQxLDIzOSw4MSwyMzEsMjA0LDE3LDIxMiwyMzEsNzAsMTQxLDI0OSwxMjUsODEsMTQyLDE4MSwzOV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbOTksMjAsMjAsNjksMjM4LDExOSw0MywxNTQsMTA0LDE0OSwyOCw4Miw5MSw5MywxNjQsMTM2LDE0NCwyMzUsMTg4LDE4Niw4MCwyMCwxODQsNzMsNDgsMjM5LDE5MCwxNTIsMTg0LDk3LDI0LDE1NV0sIklzTGVmdCI6dHJ1ZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoyLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxMDYsMTAwLDQzLDU0LDk2LDY4LDIwMyw3MywxMjMsMTA3LDM5LDYyLDIzMCwxNzAsMjIyLDIxOSwxMjEsMTA1LDEwNCwxMzUsODUsMjgsNzUsMTg0LDY5LDI1MiwxMzUsOTAsMjA3LDcxLDEwMSwyMjVdLCJJbmRleCI6MH0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3MjkzfV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTg5OTk4NDcsIlBrU2NyaXB0IjoiQUJTK3U3eFQxSEE4NExqR0I3SURFaFNENDB5MTl3PT0ifSx7IlZhbHVlIjoxMDAwMDAwLCJQa1NjcmlwdCI6IkFDQXpZVmQzb1BDLy9UdEZMQlBSOWt6Y2NVMlU0QXJYYkRNSXpBenJQS2NvYlE9PSJ9XSwiTG9ja1RpbWUiOjE5NzAyMjd9LCJCbG9ja0hhc2giOlsxODksMzgsMTM5LDEwMiwxNTIsNDcsMTE1LDcwLDE2OCw2MSwxODQsMTc2LDE3OCw5OSwwLDE4NywxMDAsMjQ4LDIwNCwyNDAsNzUsNzMsNjcsMTIyLDE1LDAsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{4}).String(),
			isExistsInPreviousBlocks: true,

			shieldAmtInPubToken: 0.01 * 1e8,
			shieldAmtInPToken:   0.01 * 1e9,
			externalTxID:        "571e129ffbdcd8ccc5386c1814bb836d904eb09572f5e6945e0f053deb122afe",
			txOutIndex:          1,

			isValidRequest: false,
		},
		// invalid shielding request: duplicated shielding proof in the current block
		{
			tokenID:                  portal.TestnetPortalV4BTCID,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_1,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzI0OSwxODEsMzAsMTA4LDI0NiwxOTUsMTQ0LDIwNiwxNjgsMSwyMzAsMTk4LDE1NiwxMjcsOTUsMTQ2LDE1NSw0Niw5MSwyMyw4MywxOTMsMSwyMzksNDYsMTkwLDIxNSw0OCwxNjIsODAsMTAyLDE5M10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyMzYsNjcsOTMsNDYsOTUsMTUwLDIxMiw1OCwyOCwxOTgsMjI3LDY2LDIzMSwyLDE5NiwxNjUsMTc2LDE1MywxNjQsMTUsMjQxLDExMSwxNjIsNzgsMjA4LDYyLDU1LDEyNiw2MiwxOSwyMzUsMTQ3XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyMzksMjMsOCwxMDcsNTcsMTQ2LDE0NCwyNTEsOCw5NywxNDIsMjAsMTUsMjAyLDE1NywzOSwxMCw4OSwxOSwxNTYsMjEsMzQsMTM1LDEzMiwxODcsMTA0LDE5NywxMjQsNTYsNzEsMjE3LDEyOF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTg5LDI2LDE3OSwyMDEsMjUyLDE3OCwyMCwzMCwyNDUsMjAyLDE5OCwxNzMsMTUxLDEzOSwyNTUsMTEsMjgsMzcsOTQsNTgsNDksMTU0LDMsNTAsODAsMTgzLDE4MiwxNjQsMTcxLDQ0LDk3LDIxN10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxNDYsNDcsMjIzLDUwLDEwMywxMzUsMTE4LDQ2LDE1MywxNDAsMTI0LDIzNCwyOCwyMzQsMTMsMTEyLDQxLDIzOSw4MSwyMzEsMjA0LDE3LDIxMiwyMzEsNzAsMTQxLDI0OSwxMjUsODEsMTQyLDE4MSwzOV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbOTksMjAsMjAsNjksMjM4LDExOSw0MywxNTQsMTA0LDE0OSwyOCw4Miw5MSw5MywxNjQsMTM2LDE0NCwyMzUsMTg4LDE4Niw4MCwyMCwxODQsNzMsNDgsMjM5LDE5MCwxNTIsMTg0LDk3LDI0LDE1NV0sIklzTGVmdCI6dHJ1ZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoyLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxMDYsMTAwLDQzLDU0LDk2LDY4LDIwMyw3MywxMjMsMTA3LDM5LDYyLDIzMCwxNzAsMjIyLDIxOSwxMjEsMTA1LDEwNCwxMzUsODUsMjgsNzUsMTg0LDY5LDI1MiwxMzUsOTAsMjA3LDcxLDEwMSwyMjVdLCJJbmRleCI6MH0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3MjkzfV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTg5OTk4NDcsIlBrU2NyaXB0IjoiQUJTK3U3eFQxSEE4NExqR0I3SURFaFNENDB5MTl3PT0ifSx7IlZhbHVlIjoxMDAwMDAwLCJQa1NjcmlwdCI6IkFDQXpZVmQzb1BDLy9UdEZMQlBSOWt6Y2NVMlU0QXJYYkRNSXpBenJQS2NvYlE9PSJ9XSwiTG9ja1RpbWUiOjE5NzAyMjd9LCJCbG9ja0hhc2giOlsxODksMzgsMTM5LDEwMiwxNTIsNDcsMTE1LDcwLDE2OCw2MSwxODQsMTc2LDE3OCw5OSwwLDE4NywxMDAsMjQ4LDIwNCwyNDAsNzUsNzMsNjcsMTIyLDE1LDAsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{5}).String(),
			isExistsInPreviousBlocks: false,

			shieldAmtInPubToken: 0.01 * 1e8,
			shieldAmtInPToken:   0.01 * 1e9,
			externalTxID:        "571e129ffbdcd8ccc5386c1814bb836d904eb09572f5e6945e0f053deb122afe",
			txOutIndex:          1,

			isValidRequest: false,
		},
	}

	return testcases, s.buildExpectedResultFromTestCases(testcases)
}

func (s *PortalTestSuiteV4) buildExpectedResultFromTestCases(testcases []TestCaseShieldingRequest) *ExpectedResultShieldingRequest {
	utxos := map[string]map[string]*statedb.UTXO{
		portal.TestnetPortalV4BTCID: {},
	}
	shieldRequests := map[string][]*statedb.ShieldingRequest{
		portal.TestnetPortalV4BTCID: {},
	}
	numBeaconInsts := len(testcases)
	statusInsts := []string{}

	for _, tc := range testcases {
		portalToken := s.portalParams.PortalTokens[tc.tokenID]
		if tc.isValidRequest {
			// add utxos
			_, otm, _ := portalToken.GenerateOTMultisigAddress(
				s.portalParams.MasterPubKeys[tc.tokenID],
				int(s.portalParams.NumRequiredSigs), tc.incAddressStr)
			key, value := generateUTXOKeyAndValue(tc.tokenID, otm, tc.externalTxID, tc.txOutIndex, tc.shieldAmtInPubToken, tc.incAddressStr)
			utxos[tc.tokenID][key] = value

			// add shield request
			shieldRequests[tc.tokenID] = append(shieldRequests[tc.tokenID], statedb.NewShieldingRequestWithValue(tc.externalTxID, tc.incAddressStr, tc.shieldAmtInPToken))
			statusInsts = append(statusInsts, portalcommonv4.PortalV4RequestAcceptedChainStatus)
		} else {
			statusInsts = append(statusInsts, portalcommonv4.PortalV4RequestRejectedChainStatus)
		}
	}

	res := &ExpectedResultShieldingRequest{
		utxos:          utxos,
		shieldRequests: shieldRequests,
		numBeaconInsts: uint(numBeaconInsts),
		statusInsts:    statusInsts,
	}
	return res
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
			Type: metadataCommon.PortalV4ShieldingRequestMeta,
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
	return []string{strconv.Itoa(metadataCommon.PortalV4ShieldingRequestMeta), actionContentBase64Str}
}

// hashProof returns the hash of shielding proof (include tx proof and user inc address)
func hashProof(proof string, chainCode string) string {
	type shieldingProof struct {
		Proof      string
		IncAddress string
	}

	shieldProof := shieldingProof{
		Proof:      proof,
		IncAddress: chainCode,
	}
	shieldProofBytes, _ := json.Marshal(shieldProof)
	hash := common.HashB(shieldProofBytes)
	return fmt.Sprintf("%x", hash[:])
}

func (s *PortalTestSuiteV4) buildShieldingRequestActionsFromTcs(tcs []TestCaseShieldingRequest, shardID byte,
	shardHeight uint64) []portalV4InstForProducer {
	insts := []portalV4InstForProducer{}

	for _, tc := range tcs {
		inst := buildPortalShieldingRequestAction(
			tc.tokenID, tc.incAddressStr, tc.shieldingProof, tc.txID, shardID)

		portalTokenProcessor := s.portalParams.PortalTokens[tc.tokenID]
		shieldTxHash, _ := portalTokenProcessor.GetTxHashFromProof(tc.shieldingProof)
		proofHash := hashProof(shieldTxHash, tc.incAddressStr)
		insts = append(insts, portalV4InstForProducer{
			inst: inst,
			optionalData: map[string]interface{}{
				"isExistProofTxHash": tc.isExistsInPreviousBlocks,
				"proofHash":          proofHash,
			},
		})
	}

	return insts
}

func getBlockCypherAPI(networkName string) gobcy.API {
	//explicitly
	bc := gobcy.API{}
	bc.Token = "cbaa6f3dc69b42079f5bab8c31c50bdf"
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
	genesisBlockHeight := 1970227
	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	dbName := "btc-blocks-test"
	btcChain, err := btcrelaying.GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	defer os.RemoveAll(dbName)

	if err != nil {
		s.FailNow(fmt.Sprintf("Could not get chain instance with err: %v", err), nil)
		return
	}

	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+11; i++ {
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
	testcases, expectedResult := s.buildTestCaseAndExpectedResultShieldingRequest()

	// build actions from testcases
	instsForProducer := s.buildShieldingRequestActionsFromTcs(testcases, shardID, shardHeight)

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

func (s *PortalTestSuiteV4) TestShieldingProof() {
	fmt.Println("Running TestShieldingProof ...")

	networkName := "test3"
	genesisBlockHeight := 2092170
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
	proof := "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzAsMjMyLDE0NiwxNDgsMTgxLDIzLDExNCwyNTMsMTkzLDEyMCw2NiwxNTYsODEsMjAsMTA3LDI0LDE3OSw5MiwxMTYsMTE1LDExLDk4LDIyNiwyMDAsMzIsNDksMTM4LDEwMyw3NSw2MiwxNTAsMTIwXSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzE1MSwxNTUsMTc4LDE4NiwyMDcsNDMsMTk4LDIwMiwzMywzMywxMDMsMjcsNTEsMjAzLDE2NSwyNDMsMjE2LDkwLDg4LDE3OSwxNTMsMjI1LDY4LDQ5LDIwNSwxODksNjQsMTM5LDE4NiwxNTgsMzYsMTMwXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls5MiwyMzgsNjcsMTg1LDg5LDE5MiwyMjksMTcyLDEyMSwyMTIsMTQ3LDksMjUwLDIwLDEyMiwzNCw3MSwxOTAsMTY2LDYsOTEsMTAxLDUyLDE0NywyMTAsMTEsMTAyLDY0LDExMiwzNiwyMjcsMl0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxOTMsMTczLDg2LDExOSwxNDMsMTIyLDE1MCwxODgsMTE3LDIsNDAsMTMwLDIzMSwxMTEsMTczLDgwLDEwNCwyMDMsNDUsMjUyLDExNywxNzcsMjgsMTgzLDE4NSw0OSwxMjIsODYsODcsMTQxLDY3LDEyOV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTAwLDc0LDM5LDI1LDIxMSw4NiwxNDksMjQyLDE4LDUwLDEyMSwyMTQsMTAsNzgsMTk0LDE5NiwxOTQsMzYsNDMsMjQ4LDE4OCwyMzYsMTAxLDUzLDExNiwyMTAsMjMxLDEwNywxODMsNDcsNjgsMTgxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls1OSwxNTYsMjIzLDEyMSw0MSwxNzYsNTYsMzAsNDksNDYsNzYsNDUsMTgsMTM5LDQ2LDE5NSw4Myw3NiwxMzEsMzAsMjIxLDEyNyw2NywxOTAsNDIsNTQsMjM0LDIxLDIyOCw0OSw4NywyNTRdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzQ4LDg0LDE3NSw0MCwxMjgsNTgsMTg3LDU3LDQxLDE2NSwyMzMsNjgsMjE0LDIyMSwxNDEsOTksMzEsOTAsNTMsMjIwLDYzLDIyMSwxMjgsMTA3LDI1NCwxNzcsMTg2LDQ2LDExMiwxNzMsOTYsMTI4XSwiSW5kZXgiOjB9LCJTaWduYXR1cmVTY3JpcHQiOiIiLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjI4OTgyNTMsIlBrU2NyaXB0IjoiQUJRRkJHYXhpN0I5dUh1SWFjUUtGWFlLQW9BbURnPT0ifSx7IlZhbHVlIjoxMDAwMDAsIlBrU2NyaXB0IjoiQUNEcDdyWDM5K1lKNkw1Vk5wbGlaZ0Z0N3VpNElRNkk3Um81Y3FRNi9TbUtvZz09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6bnVsbH0="
	expectedReceivedAddress := "tb1qa8httalhucy730j4x6vkyespdhhw3wppp6yw6x3ew2jr4lff323qu2tkkv"
	chainCodeSeed := "12si2KgWLGuhXACeqHGquGpyQy7JZiA5qRTCWW7YTYrEzZBuZC2eGBfckc2NRXkQXiw7XwK2WVfKxC8AcwKGCsyRVr9SR8bN9vTcnk2PPbymztCWadgr9JMP1UY6oSk9XZb56EAKvK1fJd5S8ptY"
	minShieldAmt := uint64(10000)

	isValid, utxos, err := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].ParseAndVerifyShieldProof(proof,
		bc, expectedReceivedAddress, chainCodeSeed, minShieldAmt)

	fmt.Printf("isValid %v\n", isValid)
	fmt.Printf("err %v\n", err)
	fmt.Printf("utxos %v\n", utxos)
}

/*
	Users unshield request
*/
type TestCaseUnshieldRequest struct {
	tokenID        string
	unshieldAmount uint64
	incAddressStr  string
	remoteAddress  string
	txId           string
	isExisted      bool
}

type ExpectedResultUnshieldRequest struct {
	waitingUnshieldReqs map[string]map[string]*statedb.WaitingUnshieldRequest
	numBeaconInsts      uint
	statusInsts         []string
}

func (s *PortalTestSuiteV4) SetupTestUnshieldRequest() {
	// do nothing
}

func buildTestCaseAndExpectedResultUnshieldRequest() ([]TestCaseUnshieldRequest, *ExpectedResultUnshieldRequest) {
	beaconHeight := uint64(1003)
	// build test cases
	testcases := []TestCaseUnshieldRequest{
		// valid unshield request
		{
			tokenID:        portal.TestnetPortalV4BTCID,
			unshieldAmount: 1 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{1}).String(),
			isExisted:      false,
		},
		// valid unshield request
		{
			tokenID:        portal.TestnetPortalV4BTCID,
			unshieldAmount: 0.5 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_2,
			txId:           common.HashH([]byte{2}).String(),
			isExisted:      false,
		},
		// invalid unshield request - invalid unshield amount
		{
			tokenID:        portal.TestnetPortalV4BTCID,
			unshieldAmount: 999999,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{3}).String(),
			isExisted:      false,
		},
		// invalid unshield request - existed unshield ID
		{
			tokenID:        portal.TestnetPortalV4BTCID,
			unshieldAmount: 1 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{1}).String(),
			isExisted:      true,
		},
	}

	// build expected results
	// waiting unshielding requests
	waitingUnshieldReqKey1 := statedb.GenerateWaitingUnshieldRequestObjectKey(portal.TestnetPortalV4BTCID, common.HashH([]byte{1}).String()).String()
	waitingUnshieldReq1 := statedb.NewWaitingUnshieldRequestStateWithValue(
		USER_BTC_ADDRESS_1, 1*1e9, common.HashH([]byte{1}).String(), beaconHeight)
	waitingUnshieldReqKey2 := statedb.GenerateWaitingUnshieldRequestObjectKey(portal.TestnetPortalV4BTCID, common.HashH([]byte{2}).String()).String()
	waitingUnshieldReq2 := statedb.NewWaitingUnshieldRequestStateWithValue(
		USER_BTC_ADDRESS_2, 0.5*1e9, common.HashH([]byte{2}).String(), beaconHeight)

	expectedRes := &ExpectedResultUnshieldRequest{
		waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
			portal.TestnetPortalV4BTCID: {
				waitingUnshieldReqKey1: waitingUnshieldReq1,
				waitingUnshieldReqKey2: waitingUnshieldReq2,
			},
		},
		numBeaconInsts: 4,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRefundedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildPortalUnshieldRequestAction(
	tokenID string,
	unshieldAmount uint64,
	incAddressStr string,
	remoteAddress string,
	txID string,
	shardID byte,
) []string {
	data := metadata.PortalUnshieldRequest{
		MetadataBase: metadata.MetadataBase{
			Type: metadataCommon.PortalV4UnshieldingRequestMeta,
		},
		OTAPubKeyStr:   incAddressStr,
		RemoteAddress:  remoteAddress,
		TokenID:        tokenID,
		UnshieldAmount: unshieldAmount,
	}
	txIDHash, _ := common.Hash{}.NewHashFromStr(txID)
	actionContent := metadata.PortalUnshieldRequestAction{
		Meta:    data,
		TxReqID: *txIDHash,
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadataCommon.PortalV4UnshieldingRequestMeta), actionContentBase64Str}
}

func buildUnshieldRequestActionsFromTcs(tcs []TestCaseUnshieldRequest, shardID byte, shardHeight uint64) []portalV4InstForProducer {
	insts := []portalV4InstForProducer{}

	for _, tc := range tcs {
		inst := buildPortalUnshieldRequestAction(
			tc.tokenID, tc.unshieldAmount, tc.incAddressStr, tc.remoteAddress, tc.txId, shardID)
		insts = append(insts, portalV4InstForProducer{
			inst: inst,
			optionalData: map[string]interface{}{
				"isExistUnshieldID": tc.isExisted,
			},
		})
	}

	return insts
}

func (s *PortalTestSuiteV4) TestUnshieldRequest() {
	fmt.Println("Running TestUnshieldRequest - beacon height 1003 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1003)
	shardHeight := uint64(1003)
	shardHeights := map[byte]uint64{
		0: uint64(1003),
	}
	shardID := byte(0)

	s.SetupTestUnshieldRequest()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultUnshieldRequest()

	// build actions from testcases
	instsForProducer := buildUnshieldRequestActionsFromTcs(testcases, shardID, shardHeight)

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

	s.Equal(expectedResult.waitingUnshieldReqs, s.currentPortalStateForProducer.WaitingUnshieldRequests)
	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Batch unshield process
*/
type TestCaseBatchUnshieldProcess struct {
	waitingUnshieldReqs map[string]map[string]*statedb.WaitingUnshieldRequest
	utxos               map[string]map[string]*statedb.UTXO
}

type ExpectedResultBatchUnshieldProcess struct {
	waitingUnshieldReqs    map[string]map[string]*statedb.WaitingUnshieldRequest
	batchUnshieldProcesses map[string]map[string]*statedb.ProcessedUnshieldRequestBatch
	utxos                  map[string]map[string]*statedb.UTXO
	numBeaconInsts         uint
}

func (s *PortalTestSuiteV4) SetupTestBatchUnshieldProcess() {
	// do nothing
}

func (s *PortalTestSuiteV4) buildTestCaseAndExpectedResultBatchUnshieldProcess() ([]TestCaseBatchUnshieldProcess,
	[]ExpectedResultBatchUnshieldProcess) {
	// prepare waiting unshielding requests
	unshieldReqsData := []struct {
		tokenID       string
		unshieldID    string
		unshieldAmt   uint64
		remoteAddress string
		beaconHeight  uint64
	}{
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{1}).String(), uint64(0.1 * 1e9), USER_BTC_ADDRESS_1, uint64(1)},
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{2}).String(), uint64(6 * 1e9), USER_BTC_ADDRESS_2, uint64(2)},
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{3}).String(), uint64(0.5 * 1e9), USER_BTC_ADDRESS_2, uint64(3)},
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{4}).String(), uint64(0.004 * 1e9), USER_BTC_ADDRESS_2, uint64(4)},
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{5}).String(), uint64(2.4 * 1e9), USER_BTC_ADDRESS_2, uint64(1)},
		{portal.TestnetPortalV4BTCID, common.HashH([]byte{6}).String(), uint64(0.3 * 1e9), USER_BTC_ADDRESS_2, uint64(43)},
	}
	type waitingUnshieldReqTmp struct {
		key   string
		value *statedb.WaitingUnshieldRequest
	}
	wUnshieldReqs := map[string]*statedb.WaitingUnshieldRequest{}
	wUnshieldReqArrs := []waitingUnshieldReqTmp{}
	for _, u := range unshieldReqsData {
		key := statedb.GenerateWaitingUnshieldRequestObjectKey(u.tokenID, u.unshieldID).String()
		value := statedb.NewWaitingUnshieldRequestStateWithValue(u.remoteAddress, u.unshieldAmt, u.unshieldID, u.beaconHeight)
		wUnshieldReqs[key] = value
		wUnshieldReqArrs = append(wUnshieldReqArrs, waitingUnshieldReqTmp{
			key:   key,
			value: value,
		})
	}

	// utxosArr
	utxosData := []struct {
		tokenID    string
		incAddress string
		outputHash string
		outputIdx  int
		outputAmt  uint64
	}{
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_1,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 0, uint64(5 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_1,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 1, uint64(1 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_2,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 2, uint64(0.1 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_2,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 3, uint64(0.01 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_2,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 4, uint64(0.02 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_2,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 5, uint64(3 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_2,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 6, uint64(0.5 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_1,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 7, uint64(2 * 1e8)},
		{portal.TestnetPortalV4BTCID, PORTALV4_USER_INC_ADDRESS_1,
			"251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355", 8, uint64(0.003 * 1e8)},
	}

	type utxoTmp struct {
		key   string
		value *statedb.UTXO
	}
	utxosArr := []utxoTmp{}
	utxosMap := map[string]*statedb.UTXO{}
	for _, u := range utxosData {
		masterPubKeys := s.portalParams.MasterPubKeys[u.tokenID]
		numReq := s.portalParams.NumRequiredSigs
		_, multisigAddress, _ := s.portalParams.PortalTokens[u.tokenID].GenerateOTMultisigAddress(masterPubKeys, int(numReq), u.incAddress)
		key, value := generateUTXOKeyAndValue(u.tokenID, multisigAddress, u.outputHash, uint32(u.outputIdx), u.outputAmt, u.incAddress)
		utxosArr = append(utxosArr, utxoTmp{key: key, value: value})
		utxosMap[key] = value
	}

	// build test cases
	testcases := []TestCaseBatchUnshieldProcess{
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portal.TestnetPortalV4BTCID: wUnshieldReqs,
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portal.TestnetPortalV4BTCID: utxosMap,
			},
		},
	}

	// build expected results
	// batch unshielding process
	currentBeaconHeight := uint64(45)
	externalFee := map[uint64]statedb.ExternalFeeInfo{
		currentBeaconHeight: {
			100000, "",
		},
	}

	batchDatas := []struct {
		processedUnshieldIDs []string
		spendUtxos           []*statedb.UTXO
	}{
		{
			[]string{
				common.HashH([]byte{5}).String(),
				common.HashH([]byte{1}).String(),
				common.HashH([]byte{2}).String(),
				common.HashH([]byte{3}).String(),
			},
			[]*statedb.UTXO{
				utxosArr[5].value,
				utxosArr[2].value,
				utxosArr[0].value,
				utxosArr[1].value,
				utxosArr[6].value,
				utxosArr[8].value,
			},
		},
		{
			[]string{
				common.HashH([]byte{4}).String(),
			},
			[]*statedb.UTXO{
				utxosArr[3].value,
				utxosArr[4].value,
			}},
	}
	batchUnshields := map[string]*statedb.ProcessedUnshieldRequestBatch{}
	for _, b := range batchDatas {
		batchID := portalprocessv4.GetBatchID(currentBeaconHeight, b.processedUnshieldIDs)
		key := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, batchID).String()
		value := statedb.NewProcessedUnshieldRequestBatchWithValue(
			batchID, b.processedUnshieldIDs, b.spendUtxos, externalFee)
		batchUnshields[key] = value
	}

	expectedRes := []ExpectedResultBatchUnshieldProcess{
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portal.TestnetPortalV4BTCID: {
					wUnshieldReqArrs[5].key: wUnshieldReqArrs[5].value,
				},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
				portal.TestnetPortalV4BTCID: batchUnshields,
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portal.TestnetPortalV4BTCID: {
					utxosArr[7].key: utxosArr[7].value,
				},
			},
			numBeaconInsts: 2,
		},
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV4) TestBatchUnshieldProcess() {
	fmt.Println("Running TestBatchUnshieldProcess - beacon height 45 ...")
	//bc := s.blockChain
	// mock test

	pm := portal.NewPortalManager()
	beaconHeight := uint64(45)
	shardHeights := map[byte]uint64{
		0: uint64(1003),
	}
	shardID := byte(0)

	bc := new(mocks.ChainRetriever)
	bc.On("GetBTCChainParams").Return(&chaincfg.TestNet3Params)
	bc.On("GetFinalBeaconHeight").Return(uint64(42))
	tokenID := portal.TestnetPortalV4BTCID
	bc.On("GetPortalV4GeneralMultiSigAddress", tokenID, beaconHeight-1).Return(s.portalParams.
		GeneralMultiSigAddresses[portal.TestnetPortalV4BTCID])

	s.SetupTestBatchUnshieldProcess()

	// build test cases and expected results
	testcases, expectedResults := s.buildTestCaseAndExpectedResultBatchUnshieldProcess()
	if len(testcases) != len(expectedResults) {
		fmt.Printf("Testcases and expected results is invalid")
		return
	}

	for i := 0; i < len(testcases); i++ {
		tc := testcases[i]
		expectedRes := expectedResults[i]

		s.currentPortalStateForProducer.UTXOs = tc.utxos
		s.currentPortalStateForProducer.WaitingUnshieldRequests = tc.waitingUnshieldReqs
		s.currentPortalStateForProcess.UTXOs = tc.utxos
		s.currentPortalStateForProcess.WaitingUnshieldRequests = tc.waitingUnshieldReqs
		s.currentPortalStateForProducer.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
		s.currentPortalStateForProcess.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}

		// beacon producer instructions
		newInsts, err := pm.PortalInstProcessorsV4[metadataCommon.PortalV4UnshieldBatchingMeta].BuildNewInsts(bc, "",
			shardID, &s.currentPortalStateForProducer, beaconHeight-1, shardHeights, s.portalParams, nil)
		s.Equal(nil, err)

		// process new instructions
		err = processPortalInstructionsV4(
			bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV4)

		// check results
		s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)), "FAILED AT TESTCASE %v", i)
		s.Equal(nil, err, "FAILED AT TESTCASE %v", i)

		s.Equal(expectedRes.waitingUnshieldReqs, s.currentPortalStateForProducer.WaitingUnshieldRequests, "FAILED AT TESTCASE %v", i)
		s.Equal(expectedRes.batchUnshieldProcesses, s.currentPortalStateForProducer.ProcessedUnshieldRequests, "FAILED AT TESTCASE %v", i)
		s.Equal(expectedRes.utxos, s.currentPortalStateForProducer.UTXOs, "FAILED AT TESTCASE %v", i)
		s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer, "FAILED AT TESTCASE %v", i)
	}
}

/*
	Feature 7: fee replacement
*/

const BatchID1 = "batch1"
const BatchID2 = "batch2"
const BatchID3 = "batch3"
const BatchID4 = "batch4"

var keyBatchShield1 = statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, BatchID1).String()
var keyBatchShield2 = statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, BatchID2).String()

type OutPut struct {
	externalAddress string
	amount          uint64
}

type TestCaseFeeReplacement struct {
	custodianIncAddress string
	batchID             string
	fee                 uint
	tokenID             string
	outputs             []OutPut
}

type ExpectedResultFeeReplacement struct {
	processedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch
	numBeaconInsts            uint
	statusInsts               []string
}

func (s *PortalTestSuiteV4) SetupTestFeeReplacement() {
	_, otm1, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_1)

	_, otm2, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_2)

	_, otm3, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_3)

	_, otm4, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_4)

	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1", "txid2", "txid3"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm1, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 1000000, PORTALV4_USER_INC_ADDRESS_1),
			statedb.NewUTXOWithValue(otm2, "49491148bd2f7b5432a26472af97724e114f22a74d9d2fb20c619b4f79f19fd9", 0, 2000000, PORTALV4_USER_INC_ADDRESS_2),
			statedb.NewUTXOWithValue(otm3, "b751ff30df21ad84ce3f509ee3981c348143bd6a5aa30f4256ecb663fab14fd1", 1, 3000000, PORTALV4_USER_INC_ADDRESS_3),
		},
		map[uint64]statedb.ExternalFeeInfo{
			900: {
				100000, "",
			},
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm4, "163a6cc24df4efbd5c997aa623d4e319f1b7671be83a86bb0fa27bc701ae4a76", 1, 1000000, PORTALV4_USER_INC_ADDRESS_4),
		},
		map[uint64]statedb.ExternalFeeInfo{
			1000: {
				100000, "",
			},
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portal.TestnetPortalV4BTCID: {
			keyBatchShield1: processUnshield1,
			keyBatchShield2: processUnshield2,
		},
	}

	s.currentPortalStateForProducer.ProcessedUnshieldRequests = processedUnshieldRequests
	s.currentPortalStateForProcess.ProcessedUnshieldRequests = CloneUnshieldBatchRequests(processedUnshieldRequests)

}

func buildFeeReplacementActionsFromTcs(tcs []TestCaseFeeReplacement, shardID byte) []portalV4InstForProducer {
	insts := []portalV4InstForProducer{}

	for _, tc := range tcs {
		inst := buildPortalFeeReplacementAction(
			tc.tokenID,
			tc.batchID,
			tc.fee,
			shardID,
		)
		optionalData := make(map[string]interface{})
		outputs := make([]*portaltokensv4.OutputTx, 0)
		for _, v := range tc.outputs {
			outputs = append(outputs, &portaltokensv4.OutputTx{ReceiverAddress: v.externalAddress, Amount: v.amount})
		}
		optionalData["outputs"] = outputs
		insts = append(insts, portalV4InstForProducer{
			inst:         inst,
			optionalData: optionalData,
		})
	}

	return insts
}

func buildPortalFeeReplacementAction(
	tokenID string,
	batchID string,
	fee uint,
	shardID byte,
) []string {
	data := metadata.PortalReplacementFeeRequest{
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				Type: metadataCommon.PortalV4FeeReplacementRequestMeta,
			},
			Sig: []byte{},
		},
		TokenID: tokenID,
		BatchID: batchID,
		Fee:     fee,
	}

	actionContent := metadata.PortalReplacementFeeRequestAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadataCommon.PortalV4FeeReplacementRequestMeta), actionContentBase64Str}
}

func buildExpectedResultFeeReplacement(s *PortalTestSuiteV4) ([]TestCaseFeeReplacement, *ExpectedResultFeeReplacement) {

	testcases := []TestCaseFeeReplacement{
		// request replace fee higher than max step
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID1,
			fee:     200000,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          1000001,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          1000002,
				},
			},
		},
		// request replace lower than latest request
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID1,
			fee:     90000,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          1000001,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          1000002,
				},
			},
		},
		// request replace fee successfully
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID1,
			fee:     102000,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          1000001,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          1000002,
				},
			},
		},
		// request replace fee with beacon height lower than next acceptable beacon height
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID1,
			fee:     103000,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          1000001,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          1000002,
				},
			},
		},
		// request replace fee new batch id
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID2,
			fee:     110000,
			outputs: []OutPut{
				{
					externalAddress: "2NBf3uA9wMJRT2eM7AyXkM6RXcPfDi24rPA",
					amount:          1000020,
				},
			},
		},
		// request replace fee with non exist batch id
		{
			tokenID: portal.TestnetPortalV4BTCID,
			batchID: BatchID3,
			fee:     110000,
			outputs: []OutPut{
				{
					externalAddress: "2N8mFbLG59ugUJM9ZBP292i6nXZHmfAw5Lk",
					amount:          1000050,
				},
			},
		},
	}

	_, otm1, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_1)

	_, otm2, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_2)

	_, otm3, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_3)

	_, otm4, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_4)

	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1", "txid2", "txid3"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm1, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 1000000, PORTALV4_USER_INC_ADDRESS_1),
			statedb.NewUTXOWithValue(otm2, "49491148bd2f7b5432a26472af97724e114f22a74d9d2fb20c619b4f79f19fd9", 0, 2000000, PORTALV4_USER_INC_ADDRESS_2),
			statedb.NewUTXOWithValue(otm3, "b751ff30df21ad84ce3f509ee3981c348143bd6a5aa30f4256ecb663fab14fd1", 1, 3000000, PORTALV4_USER_INC_ADDRESS_3),
		},
		map[uint64]statedb.ExternalFeeInfo{
			900: {
				100000, "",
			},
			1501: {
				102000, common.Hash{}.String(),
			},
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm4, "163a6cc24df4efbd5c997aa623d4e319f1b7671be83a86bb0fa27bc701ae4a76", 1, 1000000, PORTALV4_USER_INC_ADDRESS_4),
		},
		map[uint64]statedb.ExternalFeeInfo{
			1000: {
				100000, "",
			},
			1501: {
				110000, common.Hash{}.String(),
			},
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portal.TestnetPortalV4BTCID: {
			keyBatchShield1: processUnshield1,
			keyBatchShield2: processUnshield2,
		},
	}

	// build expected results
	expectedRes := &ExpectedResultFeeReplacement{
		processedUnshieldRequests: processedUnshieldRequests,
		numBeaconInsts:            6,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV4) TestFeeReplacement() {
	fmt.Println("Running TestCaseFeeReplacement - beacon height 1501 ...")
	networkName := "test3"
	genesisBlockHeight := 1938974
	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	if err != nil {
		s.FailNow(fmt.Sprintf("Could not setGenesisBlockToChainParams with err: %v", err), nil)
		return
	}
	dbName := "btc-blocks-test"
	btcChain, err := btcrelaying.GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	defer os.RemoveAll(dbName)

	if err != nil {
		s.FailNow(fmt.Sprintf("Could not get chain instance with err: %v", err), nil)
		return
	}

	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+1; i++ {
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
	bc.On("CheckBlockTimeIsReachedByBeaconHeight", mock.AnythingOfTypeArgument("uint64"), mock.AnythingOfTypeArgument("uint64"), mock.AnythingOfTypeArgument("time.Duration")).Return(
		func(recentBeaconHeight, beaconHeight uint64, duration time.Duration) bool {
			{
				return (recentBeaconHeight+1)-beaconHeight >= s.blockChain.convertDurationTimeToBeaconBlocks(duration)
			}
		})
	bc.On("GetBTCChainParams").Return(&chaincfg.TestNet3Params)
	tokenID := portal.TestnetPortalV4BTCID
	bcH := uint64(1500)
	bc.On("GetPortalV4GeneralMultiSigAddress", tokenID, bcH).Return(s.portalParams.GeneralMultiSigAddresses[portal.TestnetPortalV4BTCID])

	beaconHeight := uint64(1501)
	shardHeights := map[byte]uint64{
		0: uint64(1501),
	}
	shardID := byte(0)
	pm := portal.NewPortalManager()

	s.SetupTestFeeReplacement()

	unshieldBatchPool := s.currentPortalStateForProducer.ProcessedUnshieldRequests
	for key, unshieldBatch := range unshieldBatchPool {
		fmt.Printf("token %v - unshield batch: %v\n", key, unshieldBatch)
	}

	testcases, expectedResult := buildExpectedResultFeeReplacement(s)

	// build actions from testcases
	instsForProducer := buildFeeReplacementActionsFromTcs(testcases, shardID)

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

	s.Equal(expectedResult.processedUnshieldRequests, s.currentPortalStateForProducer.ProcessedUnshieldRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 8: submit confirmed external transaction
*/
// btctx: c45f4286489c1e5557f9b570d07d10248aea220ae550b05ad41b25e48220c044
const confirmedTxProof1 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6Wzk4LDIzNSw5MCwyNDYsNjgsMTEzLDE3MCw3NCwyNTUsMjM0LDM0LDk2LDk3LDc3LDEyMSwyMTcsMTc2LDM3LDQ4LDE0Myw5MiwxNzMsMTU3LDgxLDMwLDc5LDk1LDExMSwxNzgsMjQ5LDY5LDE4MV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbODcsMTg5LDEwNCw4OCw2OCwyOCwyOSwxODAsNjAsNjUsMTY3LDIxMiwxNTAsMTkzLDE1OSwxOTEsMTM1LDcxLDE0MCw5OCwyNDYsNzksMjQ4LDE5MywxNDAsNDksOTcsMTA5LDIyNSw4NSwyNCw3MV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbODYsMTcyLDY4LDE0MSwxMjYsMTg4LDI1MSwyMTAsMTMyLDQxLDcwLDIwOCwxMzQsMTYsMjgsNjEsMjMxLDE4NSwxNzUsNzAsMTU3LDE0OCw1OSwxMzcsMTg5LDQsMjEyLDE5MywxODgsMzAsNDgsOF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyNDksMTc5LDI1MSwyNTIsMjA1LDE4NiwxNjUsMTIsMTQ2LDY3LDE3OCwxMTMsNCw2NSw3OSwxMzksMTU3LDIzMiwyMDAsMTMsMTY3LDY4LDE2MSwxMTAsMTk5LDE0NSwyMTIsMTc0LDE3NiwyMjMsMTUyLDI0MV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNDIsNzksMTg5LDUzLDE3NiwyNDIsMTk0LDg5LDExLDE1MCwxNSwxNDUsNjMsMTk1LDcyLDg4LDQ3LDI3LDY3LDY4LDE3OCwxMDIsMTAxLDE4NywyNDAsMjE0LDE3Myw2LDE2MiwxNzksMTIsMjUxXSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzIzLDExMywxNTMsMTU3LDY2LDE5MywxNjcsMTU4LDk5LDExOSwyMDIsMjE1LDIwMiwyMTQsNDAsOCwyNDAsMjEwLDE3NywyMDAsMTQzLDk3LDE3MSw3OCwxMTEsMjExLDc0LDI0MiwxNTQsMTcsMTY1LDE3XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzIyLDE2LDE4MSw0MywyNiwxNTYsMjA4LDg1LDIxNSwxMSwyMDgsNjgsMjI1LDYsOTIsMjQ5LDE4NiwxNzgsMTAsMywxMzQsMTM3LDE4NiwxMDYsMjA0LDE3OCwyMDgsOCwxNTIsODQsMTQ5LDI0Nl0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjgsMzUsMjU1LDEyNCw5Myw4OSw5MywxNjcsMTIsOCwyNTAsMTk2LDE5NiwxODUsMTE0LDI1NSw4MCwxNjcsMTEwLDM1LDIzNiw3LDM4LDIwOCwxMzEsMjIyLDczLDQ0LDE3NywyMDYsOTksMTExXSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxOTMsMTE0LDg1LDExMSwyOCwyMDAsMjQ0LDQ3LDk3LDI0Myw5NywyMDEsNjgsOTAsMTcsMTkyLDIzNiw2Niw1NSwzNywyNDEsMTA0LDIwNiw1NiwxMDIsMTIwLDI0MiwxNzIsMTgyLDI1MiwxNjYsMjM0XSwiSW5kZXgiOjB9LCJTaWduYXR1cmVTY3JpcHQiOiIiLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjEwMDAzMCwiUGtTY3JpcHQiOiJBQ0NHS2FZSGNLSHhUU0dDaU8vV0NBK2hqMUNlOXp2NGY1TXRROXc3eXlVejVnPT0ifSx7IlZhbHVlIjo4OTg5NzAsIlBrU2NyaXB0IjoiQUNCS0JYYTEyWjV0YW5vdytIWUc4V050Tkh6STViUjkvY2gxaXZaWjN2ei9Idz09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6WzIyOCwxMDAsMzYsMTEzLDUyLDExMCwxMjMsNTAsMTQ3LDIzMiw4NywzMiwyMDUsMjA3LDY0LDMsMTc4LDcwLDE3OSwxNzEsNzgsNDYsOTksMTE4LDE1NiwxOCwxNzYsMCwwLDAsMCwwXX0="

// btctx: b545f9b26f5f4f1e519dad5c8f3025b0d9794d616022eaff4aaa7144f65aeb62
const confirmedTxProof2 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzY4LDE5MiwzMiwxMzAsMjI4LDM3LDI3LDIxMiw5MCwxNzYsODAsMjI5LDEwLDM0LDIzNCwxMzgsMzYsMTYsMTI1LDIwOCwxMTIsMTgxLDI0OSw4Nyw4NSwzMCwxNTYsNzIsMTM0LDY2LDk1LDE5Nl0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls4NywxODksMTA0LDg4LDY4LDI4LDI5LDE4MCw2MCw2NSwxNjcsMjEyLDE1MCwxOTMsMTU5LDE5MSwxMzUsNzEsMTQwLDk4LDI0Niw3OSwyNDgsMTkzLDE0MCw0OSw5NywxMDksMjI1LDg1LDI0LDcxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls4NiwxNzIsNjgsMTQxLDEyNiwxODgsMjUxLDIxMCwxMzIsNDEsNzAsMjA4LDEzNCwxNiwyOCw2MSwyMzEsMTg1LDE3NSw3MCwxNTcsMTQ4LDU5LDEzNywxODksNCwyMTIsMTkzLDE4OCwzMCw0OCw4XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzI0OSwxNzksMjUxLDI1MiwyMDUsMTg2LDE2NSwxMiwxNDYsNjcsMTc4LDExMyw0LDY1LDc5LDEzOSwxNTcsMjMyLDIwMCwxMywxNjcsNjgsMTYxLDExMCwxOTksMTQ1LDIxMiwxNzQsMTc2LDIyMywxNTIsMjQxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0Miw3OSwxODksNTMsMTc2LDI0MiwxOTQsODksMTEsMTUwLDE1LDE0NSw2MywxOTUsNzIsODgsNDcsMjcsNjcsNjgsMTc4LDEwMiwxMDEsMTg3LDI0MCwyMTQsMTczLDYsMTYyLDE3OSwxMiwyNTFdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjMsMTEzLDE1MywxNTcsNjYsMTkzLDE2NywxNTgsOTksMTE5LDIwMiwyMTUsMjAyLDIxNCw0MCw4LDI0MCwyMTAsMTc3LDIwMCwxNDMsOTcsMTcxLDc4LDExMSwyMTEsNzQsMjQyLDE1NCwxNywxNjUsMTddLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjIsMTYsMTgxLDQzLDI2LDE1NiwyMDgsODUsMjE1LDExLDIwOCw2OCwyMjUsNiw5MiwyNDksMTg2LDE3OCwxMCwzLDEzNCwxMzcsMTg2LDEwNiwyMDQsMTc4LDIwOCw4LDE1Miw4NCwxNDksMjQ2XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2OCwzNSwyNTUsMTI0LDkzLDg5LDkzLDE2NywxMiw4LDI1MCwxOTYsMTk2LDE4NSwxMTQsMjU1LDgwLDE2NywxMTAsMzUsMjM2LDcsMzgsMjA4LDEzMSwyMjIsNzMsNDQsMTc3LDIwNiw5OSwxMTFdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzE0OSwxNDMsMjI2LDE0NiwyMTksNDIsMTg2LDExMiwyNDUsMTE3LDI1MSwxMDUsMTM2LDE4NiwyNDQsMjE4LDE3MCwyMDksNzQsNzQsMjM5LDUzLDIwMywyOCwyMTYsOTksMzIsNzgsMTkyLDIzMCwyNTEsNTldLCJJbmRleCI6MX0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTAwMDMwLCJQa1NjcmlwdCI6IkFDQ0dLYVlIY0tIeFRTR0NpTy9XQ0EraGoxQ2U5enY0ZjVNdFE5dzd5eVV6NWc9PSJ9LHsiVmFsdWUiOjEwMDAyMCwiUGtTY3JpcHQiOiJxUlFtamxCdTRya0VVUXpJK0NHemF1QTRpeThvVFljPSJ9LHsiVmFsdWUiOjEwMDAxMCwiUGtTY3JpcHQiOiJBQ0NHS2FZSGNLSHhUU0dDaU8vV0NBK2hqMUNlOXp2NGY1TXRROXc3eXlVejVnPT0ifSx7IlZhbHVlIjo2OTg5NDAsIlBrU2NyaXB0IjoiQUNCS0JYYTEyWjV0YW5vdytIWUc4V050Tkh6STViUjkvY2gxaXZaWjN2ei9Idz09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6WzIyOCwxMDAsMzYsMTEzLDUyLDExMCwxMjMsNTAsMTQ3LDIzMiw4NywzMiwyMDUsMjA3LDY0LDMsMTc4LDcwLDE3OSwxNzEsNzgsNDYsOTksMTE4LDE1NiwxOCwxNzYsMCwwLDAsMCwwXX0="

// invalid proof
const confirmedTxProof3 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE0MywyMTEsMjI2LDExNiwyNTMsNjksMjQ2LDIyNCwxMTAsMTg0LDMwLDE1Nyw4NCwyMDcsMTQyLDI1MywxMjIsNTAsMTk0LDgsMjAzLDExOSw3NSwxODMsMjUsNjUsMTU1LDIxMywxODYsMTg0LDEyNSwxMF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls0OCwxODIsMTE2LDI1MCwzOSwxMDgsMTk1LDE0NCwyMSw3OSwyMjIsNzQsMTk3LDE2MSwxMDcsMTYwLDIxLDMwLDIwNiwyNDksMTc5LDExMSwyMjMsMzIsNDcsMTM5LDE1MywyOCwxOTIsMjIwLDE0NiwyNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxMzgsNSwzOSw3NCwyNCw3NSw4MSw2MCwxNjcsNDYsMTg2LDEwNiwxNTAsNDQsMjAwLDIxLDIzOCw0MSwyMzQsMzksMjI1LDkyLDExLDIzNCwxNDAsMTA3LDI0OCwyNDQsMTQ0LDExNiwyMTksMTM2XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzE4MiwxMDYsOTEsMTYxLDE0NSwxMzMsMjQ2LDc1LDIwOSw3NCwxODEsMTgyLDkyLDI1NCw0OSwxOTMsNTEsMjMzLDE1NywxODUsNTQsNzMsNTAsMjQ0LDEwNywzMiwzMSwxODksNDMsNCwxMTIsMTI4XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsxMiwzNSwxODIsMTk3LDE5NiwxODYsMTQzLDE1MSw0MywxMDMsMjU1LDE2LDE2MSwyNDAsMTM5LDE2OCwxNzEsOTgsODYsMTA3LDk3LDIxMiw5MCwxNjUsMTQ5LDYyLDMwLDY1LDc1LDIyOCw2NywxODFdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbNDcsMTY0LDYwLDcsODAsMTQsNzQsMTY1LDE5NSwxODYsMTE2LDY0LDExOCwxMDIsMTk1LDEsMTMxLDQ0LDU5LDE3MSwyMDEsMTU3LDc2LDUzLDgyLDksMTM4LDE3OSw5MSw2LDQsNDRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzIzNSwyNDEsMTY4LDM0LDE4MywxNzgsNDEsMjUzLDEwMiwxMzYsMTg2LDg3LDE4OCwyMzQsMzgsMTU4LDExMSwyMjUsMTIyLDIzMCwyMjksNDgsMTgyLDEwNiwyNyw2NSwyMTQsNDIsMTUzLDQyLDMwLDkwXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2LDE4NCwyOCwyNDgsMTcyLDM0LDE0MywyNTEsMTcwLDEzLDIxNyw3OSwyMjcsMTA2LDIxMiw1NSw5MSwyMDMsMTAzLDkwLDkwLDIyLDI0Niw2NSw0OCwyMTMsMjU1LDE5OSwzOCwxMTMsMTkxLDIxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyOCwxOTQsMTQyLDMzLDQzLDg3LDIxLDIzNCwxOSwxOTEsMTYzLDIxMiwyMTcsMjUsNDksMTk5LDIwMywxNzIsMjUsNywxNjEsMTM2LDE2MywzMyw3OSwxODcsNDQsNzEsMTAxLDI5LDE4NSwyNTNdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzE0MywyMTEsMjI2LDExNiwyNTMsNjksMjQ2LDIyNCwxMTAsMTg0LDMwLDE1Nyw4NCwyMDcsMTQyLDI1MywxMjIsNTAsMTk0LDgsMjAzLDExOSw3NSwxODMsMjUsNjUsMTU1LDIxMywxODYsMTg0LDEyNSwxMF0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiU0RCRkFpRUE5WS9XeDZvMDh4QjAzZkdya3EyVGQ5NXV5akxiK0ZTRk13cHpWcHdkZTFNQ0lGcWJzdWlWeis5Wnhka05YWmZXQ1p5WHZMdUJrK3Y1KzZzYk1kbGUwSVkvQVNFRHp5QVRUMVpJNHZieGQ2WlZLeVc2bCtSYkZJVk9UeE1xU3ZEK2Zhay94UHc9IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFpeFNUVXBQYkN0b00wNUZibEkwTm5SaE5GSkVkQ3QwVEhObVdIRktiamswYUVaSWVDOWtUbloyU1ZCUlBRPT0ifSx7IlZhbHVlIjo2MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjoyMjQzNzIsIlBrU2NyaXB0IjoiZHFrVWd2eTZsUWkrRWlReTk3dVEybDkwQVVDbVc0aUlyQT09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6WzE3MiwyMzYsMTY4LDEwNSwxMzQsMzMsMTM1LDEzMiwxMiwyMjUsMTIzLDIxMiwzOSwyNDUsMTUsMTkzLDE0NiwxMjMsMTA1LDExMiwzNiwxODAsMTgyLDEwNSw0MiwyMDcsMTE1LDIxOCwwLDAsMCwwXX0="

// multi sender to multi receiver
// btctx: f03e36236a6fd9714d8f85f1091dea632aa4a5abb562f43bc7f04bdf9f0bf928
const confirmedTxProof4 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzYsMywyMDUsMjAzLDIzMCwxMTgsMTc2LDE1Miw1MywzNCwxNywyNCw4NCwxNTMsMjQ2LDM4LDEwNSwzNCw2OCwxNzksMTQwLDY2LDczLDE4MiwyMjMsMTMzLDE0OSwyMTYsMTMzLDE2MiwxMzksNDBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzExOCwyNDUsMTksMTczLDE4MSwxNTUsNzgsMTg4LDQ2LDIyMCwxNzMsNDMsMjM4LDEzLDE1Nyw0OCwyLDE4NCw0NCwxMDksNTQsNTUsNDgsMzQsNSw2Myw2NSw1NSw3MSwyNTUsMTE0LDYzXSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzE0LDcyLDIyOSw5OSw1NCwxNjksMTM2LDQ0LDIyMSwxMyw4MiwxNjIsMTY1LDIwMSwxMjUsMTIzLDIyNSw3OSw0MiwxNjUsMjA3LDIwOCwxODMsMTQxLDExMiwxNTksMjQzLDIyMCwxNTUsMTQ1LDQzLDcyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlszNiwyMjIsMTIzLDI1MiwxMzEsNjIsMTEsNywzMiwxNTQsOTQsMTU1LDIxLDEyMCwxODAsMTU3LDcsNzEsMTEyLDE1LDEyOCwxOSwyMjIsMTQ1LDE1MywxMDQsMTk3LDcxLDE4NCwxMTUsNDIsNF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls0Miw3OSwxODksNTMsMTc2LDI0MiwxOTQsODksMTEsMTUwLDE1LDE0NSw2MywxOTUsNzIsODgsNDcsMjcsNjcsNjgsMTc4LDEwMiwxMDEsMTg3LDI0MCwyMTQsMTczLDYsMTYyLDE3OSwxMiwyNTFdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjMsMTEzLDE1MywxNTcsNjYsMTkzLDE2NywxNTgsOTksMTE5LDIwMiwyMTUsMjAyLDIxNCw0MCw4LDI0MCwyMTAsMTc3LDIwMCwxNDMsOTcsMTcxLDc4LDExMSwyMTEsNzQsMjQyLDE1NCwxNywxNjUsMTddLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjIsMTYsMTgxLDQzLDI2LDE1NiwyMDgsODUsMjE1LDExLDIwOCw2OCwyMjUsNiw5MiwyNDksMTg2LDE3OCwxMCwzLDEzNCwxMzcsMTg2LDEwNiwyMDQsMTc4LDIwOCw4LDE1Miw4NCwxNDksMjQ2XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2OCwzNSwyNTUsMTI0LDkzLDg5LDkzLDE2NywxMiw4LDI1MCwxOTYsMTk2LDE4NSwxMTQsMjU1LDgwLDE2NywxMTAsMzUsMjM2LDcsMzgsMjA4LDEzMSwyMjIsNzMsNDQsMTc3LDIwNiw5OSwxMTFdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzIzMiwxNSwxNzQsMjQ2LDE3NSw1NSwyMTMsMTI4LDE4MSwxNTYsMzksMjE0LDY3LDEyMCw0MSwxMzMsMjUsMjE4LDE2MSwyMDQsMTA3LDM0LDUwLDE0Myw5OSwxMTQsMTcsMzcsNjUsMjM0LDIwMywxODldLCJJbmRleCI6MH0sIlNpZ25hdHVyZVNjcmlwdCI6IiIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fSx7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzI1MiwyNTUsMjMxLDIzOSwxNzYsNzEsMTgzLDYyLDk4LDExMCwyMzMsMTE1LDQ4LDg5LDQsMTg1LDExLDEyMiwxNzcsMTMwLDIwMywxNjEsMTcwLDg5LDE3NywzLDcyLDEzNSw3NywxNDEsMTgyLDcyXSwiSW5kZXgiOjB9LCJTaWduYXR1cmVTY3JpcHQiOiIiLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjMwMDAzMCwiUGtTY3JpcHQiOiJBQ0NHS2FZSGNLSHhUU0dDaU8vV0NBK2hqMUNlOXp2NGY1TXRROXc3eXlVejVnPT0ifSx7IlZhbHVlIjo1MDAwMjAsIlBrU2NyaXB0IjoicVJRbWpsQnU0cmtFVVF6SStDR3phdUE0aXk4b1RZYz0ifSx7IlZhbHVlIjoxMDAwMTAsIlBrU2NyaXB0IjoiQUNDR0thWUhjS0h4VFNHQ2lPL1dDQStoajFDZTl6djRmNU10UTl3N3l5VXo1Zz09In0seyJWYWx1ZSI6MzAwMDEwLCJQa1NjcmlwdCI6IkFCVFAwS1BGSGhLQkdPb2trTWNvakVWR0pDaEJHQT09In0seyJWYWx1ZSI6Nzk4OTMwLCJQa1NjcmlwdCI6IkFDQktCWGExMlo1dGFub3crSFlHOFdOdE5Iekk1YlI5L2NoMWl2WlozdnovSHc9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsyMjgsMTAwLDM2LDExMyw1MiwxMTAsMTIzLDUwLDE0NywyMzIsODcsMzIsMjA1LDIwNyw2NCwzLDE3OCw3MCwxNzksMTcxLDc4LDQ2LDk5LDExOCwxNTYsMTgsMTc2LDAsMCwwLDAsMF19"

type TestCaseSubmitConfirmedTx struct {
	confirmedTxProof string
	batchID          string
	tokenID          string
	outputs          []OutPut
}

type ExpectedResultSubmitConfirmedTx struct {
	utxos                     map[string]map[string]*statedb.UTXO
	processedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch
	numBeaconInsts            uint
	statusInsts               []string
}

func (s *PortalTestSuiteV4) SetupTestSubmitConfirmedTx() {
	_, otm1, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_1)

	_, otm2, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_2)

	keyUTXO, utxo1 := generateUTXOKeyAndValue(portal.TestnetPortalV4BTCID, otm1, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 100000, PORTALV4_USER_INC_ADDRESS_1)
	utxos := map[string]map[string]*statedb.UTXO{
		portal.TestnetPortalV4BTCID: {
			keyUTXO: utxo1,
		},
	}

	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm1, "eaa6fcb6acf2786638ce68f1253742ecc0115a44c961f3612ff4c81c6f5572c1", 0, 200000, PORTALV4_USER_INC_ADDRESS_1),
		},
		map[uint64]statedb.ExternalFeeInfo{
			900: {
				100000, "",
			},
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid2", "txid3", "txid4"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm1, "3bfbe6c04e2063d81ccb35ef4a4ad1aadaf4ba8869fb75f570ba2adb92e28f95", 1, 1000000, PORTALV4_USER_INC_ADDRESS_1),
		},
		map[uint64]statedb.ExternalFeeInfo{
			900: {
				100000, "",
			},
		},
	)

	processUnshield3 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID3,
		[]string{"txid5", "txid6", "txid7", "txid8"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(otm1, "bdcbea41251172638f32226bcca1da1985297843d6279cb580d537aff6ae0fe8", 0, 1000000, PORTALV4_USER_INC_ADDRESS_1),
			statedb.NewUTXOWithValue(otm2, "48b68d4d874803b159aaa1cb82b17a0bb904593073e96e623eb747b0efe7fffc", 0, 1000000, PORTALV4_USER_INC_ADDRESS_2),
		},
		map[uint64]statedb.ExternalFeeInfo{
			900: {
				100000, "",
			},
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portal.TestnetPortalV4BTCID: {
			statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, BatchID1).String(): processUnshield1,
			statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, BatchID2).String(): processUnshield2,
			statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portal.TestnetPortalV4BTCID, BatchID3).String(): processUnshield3,
		},
	}

	s.currentPortalStateForProducer.ProcessedUnshieldRequests = processedUnshieldRequests
	s.currentPortalStateForProducer.UTXOs = utxos
	s.currentPortalStateForProcess.ProcessedUnshieldRequests = CloneUnshieldBatchRequests(processedUnshieldRequests)
	s.currentPortalStateForProcess.UTXOs = CloneUTXOs(utxos)
}

func buildSubmitConfirmedTxActionsFromTcs(tcs []TestCaseSubmitConfirmedTx, shardID byte) []portalV4InstForProducer {
	insts := []portalV4InstForProducer{}

	for _, tc := range tcs {
		inst := buildPortalSubmitConfirmedTxAction(
			tc.confirmedTxProof,
			tc.tokenID,
			tc.batchID,
			shardID,
		)
		optionalData := make(map[string]interface{})
		outputs := []*portaltokens.OutputTx{}
		for _, v := range tc.outputs {
			outputs = append(outputs, &portaltokens.OutputTx{
				ReceiverAddress: v.externalAddress,
				Amount:          v.amount,
			})
		}
		optionalData["outputs"] = outputs
		insts = append(insts, portalV4InstForProducer{
			inst:         inst,
			optionalData: optionalData,
		})
	}

	return insts
}

func buildPortalSubmitConfirmedTxAction(
	unshieldProof string,
	tokenID string,
	batchID string,
	shardID byte,
) []string {
	data := metadata.PortalSubmitConfirmedTxRequest{
		MetadataBase: metadata.MetadataBase{
			Type: metadataCommon.PortalV4SubmitConfirmedTxMeta,
		},
		UnshieldProof: unshieldProof,
		TokenID:       tokenID,
		BatchID:       batchID,
	}

	actionContent := metadata.PortalSubmitConfirmedTxAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadataCommon.PortalV4SubmitConfirmedTxMeta), actionContentBase64Str}
}

func buildExpectedResultSubmitConfirmedTx(s *PortalTestSuiteV4) ([]TestCaseSubmitConfirmedTx, *ExpectedResultSubmitConfirmedTx) {

	testcases := []TestCaseSubmitConfirmedTx{
		// request submit external confirmed tx 1 - 1
		{
			batchID:          BatchID1,
			confirmedTxProof: confirmedTxProof1,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001300,
				},
			},
		},
		// request submit external confirmed tx twice
		{
			batchID:          BatchID1,
			confirmedTxProof: confirmedTxProof1,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001300,
				},
			},
		},
		// request submit proof with non-exist batchID
		{
			batchID:          BatchID4,
			confirmedTxProof: confirmedTxProof2,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001300,
				},
				{
					externalAddress: "2Mvm69jhFFBRBL5mHsb8TNjMfCpStoVAqWD",
					amount:          1001200,
				},
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001100,
				},
			},
		},
		// request submit wrong proof
		{
			batchID:          BatchID2,
			confirmedTxProof: confirmedTxProof3,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001300,
				},
				{
					externalAddress: "2Mvm69jhFFBRBL5mHsb8TNjMfCpStoVAqWD",
					amount:          1001200,
				},
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001100,
				},
			},
		},
		// request submit 1 - n proof
		{
			batchID:          BatchID2,
			confirmedTxProof: confirmedTxProof2,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001300,
				},
				{
					externalAddress: "2Mvm69jhFFBRBL5mHsb8TNjMfCpStoVAqWD",
					amount:          1001200,
				},
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001100,
				},
			},
		},
		// request submit n - n proof
		{
			batchID:          BatchID3,
			confirmedTxProof: confirmedTxProof4,
			tokenID:          portal.TestnetPortalV4BTCID,
			outputs: []OutPut{
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          3001300,
				},
				{
					externalAddress: "2Mvm69jhFFBRBL5mHsb8TNjMfCpStoVAqWD",
					amount:          5001200,
				},
				{
					externalAddress: "tb1qsc56vpms58c56gvz3rhavzq05x84p8hh80u8lyedg0wrhje9x0nqd2q0qe",
					amount:          1001100,
				},
				{
					externalAddress: "tb1qelg283g7z2q3363yjrrj3rz9gcjzssgcx0yhfa",
					amount:          3001100,
				},
			},
		},
	}

	btcMultiSigAddress := s.portalParams.GeneralMultiSigAddresses[portal.TestnetPortalV4BTCID]

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portal.TestnetPortalV4BTCID: {},
	}

	_, otm1, _ := s.portalParams.PortalTokens[portal.TestnetPortalV4BTCID].GenerateOTMultisigAddress(
		s.portalParams.MasterPubKeys[portal.TestnetPortalV4BTCID],
		int(s.portalParams.NumRequiredSigs), PORTALV4_USER_INC_ADDRESS_1)

	keyUTXO1, utxo1 := generateUTXOKeyAndValue(portal.TestnetPortalV4BTCID, otm1, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 100000, PORTALV4_USER_INC_ADDRESS_1)
	keyUTXO2, utxo2 := generateUTXOKeyAndValue(portal.TestnetPortalV4BTCID, btcMultiSigAddress, "c45f4286489c1e5557f9b570d07d10248aea220ae550b05ad41b25e48220c044", 1, 898970, "")
	keyUTXO3, utxo3 := generateUTXOKeyAndValue(portal.TestnetPortalV4BTCID, btcMultiSigAddress, "b545f9b26f5f4f1e519dad5c8f3025b0d9794d616022eaff4aaa7144f65aeb62", 3, 698940, "")
	keyUTXO4, utxo4 := generateUTXOKeyAndValue(portal.TestnetPortalV4BTCID, btcMultiSigAddress, "f03e36236a6fd9714d8f85f1091dea632aa4a5abb562f43bc7f04bdf9f0bf928", 4, 798930, "")
	utxos := map[string]map[string]*statedb.UTXO{
		portal.TestnetPortalV4BTCID: {
			keyUTXO1: utxo1,
			keyUTXO2: utxo2,
			keyUTXO3: utxo3,
			keyUTXO4: utxo4,
		},
	}

	// build expected results
	expectedRes := &ExpectedResultSubmitConfirmedTx{
		processedUnshieldRequests: processedUnshieldRequests,
		numBeaconInsts:            6,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
		},
		utxos: utxos,
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV4) TestSubmitConfirmedTx() {
	fmt.Println("Running TestSubmitConfirmedTx - beacon height 1501 ...")
	networkName := "test3"
	genesisBlockHeight := 1970927
	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	dbName := "btc-blocks-test"
	btcChain, err := btcrelaying.GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	defer os.RemoveAll(dbName)

	if err != nil {
		s.FailNow(fmt.Sprintf("Could not get chain instance with err: %v", err), nil)
		return
	}

	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+9; i++ {
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
		time.Sleep(2000 * time.Millisecond) // 2s
	}

	bc := new(mocks.ChainRetriever)
	bc.On("GetBTCHeaderChain").Return(btcChain)

	beaconHeight := uint64(1501)
	shardHeights := map[byte]uint64{
		0: uint64(1501),
	}
	shardID := byte(0)
	pm := portal.NewPortalManager()

	s.SetupTestSubmitConfirmedTx()

	unshieldBatchPool := s.currentPortalStateForProducer.ProcessedUnshieldRequests
	for key, unshieldBatch := range unshieldBatchPool {
		fmt.Printf("token %v - unshield batch: %v\n", key, unshieldBatch)
	}

	testcases, expectedResult := buildExpectedResultSubmitConfirmedTx(s)

	// build actions from testcases
	instsForProducer := buildSubmitConfirmedTxActionsFromTcs(testcases, shardID)

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

	s.Equal(expectedResult.processedUnshieldRequests, s.currentPortalStateForProducer.ProcessedUnshieldRequests)
	s.Equal(expectedResult.utxos, s.currentPortalStateForProducer.UTXOs)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

// util functions
func CloneUnshieldBatchRequests(processedUnshieldRequestBatch map[string]map[string]*statedb.ProcessedUnshieldRequestBatch) map[string]map[string]*statedb.ProcessedUnshieldRequestBatch {
	newReqs := make(map[string]map[string]*statedb.ProcessedUnshieldRequestBatch, len(processedUnshieldRequestBatch))
	for key, batch := range processedUnshieldRequestBatch {
		newBatch := make(map[string]*statedb.ProcessedUnshieldRequestBatch, len(batch))
		for key2, batch2 := range batch {
			newBatch[key2] = statedb.NewProcessedUnshieldRequestBatchWithValue(
				batch2.GetBatchID(),
				batch2.GetUnshieldRequests(),
				batch2.GetUTXOs(),
				batch2.GetExternalFees(),
			)
		}
		newReqs[key] = newBatch
	}
	return newReqs
}

func CloneUTXOs(utxos map[string]map[string]*statedb.UTXO) map[string]map[string]*statedb.UTXO {
	newReqs := make(map[string]map[string]*statedb.UTXO, len(utxos))
	for key, batch := range utxos {
		newBatch := make(map[string]*statedb.UTXO, len(batch))
		for key2, batch2 := range batch {
			newBatch[key2] = statedb.NewUTXOWithValue(
				batch2.GetWalletAddress(),
				batch2.GetTxHash(),
				batch2.GetOutputIndex(),
				batch2.GetOutputAmount(),
				batch2.GetChainCodeSeed(),
			)
		}
		newReqs[key] = newBatch
	}
	return newReqs
}

// benchmark utxos on chain

type TestCaseUTXOProcess struct {
	waitingUnshieldReqs map[string]map[string]*statedb.WaitingUnshieldRequest
	utxos               map[string]map[string]*statedb.UTXO
}

func (s *PortalTestSuiteV4) SetupTestUTXOProcess() {
	// do nothing
}

type UTXOResponse struct {
	rpccaller.RPCBaseRes
	Result interface{} `json:"Result"`
}

type UnshieldRequest struct {
	testcases []map[string]*statedb.WaitingUnshieldRequest
}

func slotIndex(value float64) int {
	if value > 5 {
		return 0
	} else if value < 0.01 {
		return 3
	} else if value > 1 {
		return 2
	} else {
		return 1
	}
}

func (s *PortalTestSuiteV4) buildTestCaseAndExpectedResultUTXOProcess() (TestCaseUTXOProcess, UnshieldRequest, map[string]*statedb.WaitingUnshieldRequest) {
	// prepare waiting unshielding requests
	testSize := 10 // 619 current => 4692 old; 398 new with testsize 500
	adverageInit := 5
	testLength := testSize / adverageInit
	m := map[int]int{}
	m[0] = testSize * 10 / 100
	m[1] = testSize * 60 / 100
	m[2] = testSize * 10 / 100
	m[3] = testSize - m[0] - m[1] - m[3]

	requestUnshieldData := UnshieldRequest{}
	count := 0
	listUnshields := map[string]*statedb.WaitingUnshieldRequest{}
	totalOccupied := 0
	for i := 0; i < testLength; i++ {
		testDepth := 0
		// todo: update range here
		if i == testLength-1 {
			testDepth = testSize - totalOccupied
		} else {
			max := testSize - totalOccupied - (testLength - i + 1)
			min := 1
			testDepth = rand.Intn(max-min) + min
		}
		totalOccupied += testDepth
		temp := map[string]*statedb.WaitingUnshieldRequest{}
		for j := 0; j < testDepth; j++ {
			unshieldID := common.HashH([]byte(strconv.Itoa(count))).String()
			key := statedb.GenerateWaitingUnshieldRequestObjectKey(portal.TestnetPortalV4BTCID, unshieldID).String()
			// todo: update range here
			minAmt := 0.001
			maxAmt := float64(10)
			r := float64(0)
			for true {
				r = minAmt + rand.Float64()*(maxAmt-minAmt)
				if m[slotIndex(r)] > 0 {
					m[slotIndex(r)]--
					break
				}
			}
			value := statedb.NewWaitingUnshieldRequestStateWithValue(USER_BTC_ADDRESS_1, uint64(r*1e9), unshieldID, uint64(1))
			temp[key] = value
			listUnshields[unshieldID] = value
			count++
		}
		requestUnshieldData.testcases = append(requestUnshieldData.testcases, temp)

	}

	// utxos
	var utxosData []struct {
		tokenID    string
		incAddress string
		outputHash string
		outputIdx  int
		outputAmt  uint64
	}

	rpcClient := rpccaller.NewRPCClient()
	meta := map[string]interface{}{
		"BeaconHeight": "1470000",
	}
	params := []interface{}{
		meta,
	}

	var res UTXOResponse

	err := rpcClient.RPCCall(
		"",
		"https://mainnet.incognito.org/fullnode",
		"",
		"getportalv4state",
		params,
		&res,
	)
	if err != nil {
		return TestCaseUTXOProcess{}, UnshieldRequest{}, nil
	}
	utxosRps := res.Result.(map[string]interface{})
	utxosMap := utxosRps["UTXOs"].(map[string]interface{})[portal.MainnetPortalV4BTCID].(map[string]interface{})
	for _, v := range utxosMap {
		vMap := v.(map[string]interface{})
		utxo := struct {
			tokenID    string
			incAddress string
			outputHash string
			outputIdx  int
			outputAmt  uint64
		}{
			portal.TestnetPortalV4BTCID,
			vMap["WalletAddress"].(string),
			vMap["TxHash"].(string),
			int(vMap["OutputIdx"].(float64)),
			uint64(vMap["OutputAmount"].(float64)),
		}
		utxosData = append(utxosData, utxo)
	}

	utxos := map[string]*statedb.UTXO{}

	for _, u := range utxosData {
		masterPubKeys := s.portalParams.MasterPubKeys[u.tokenID]
		numReq := s.portalParams.NumRequiredSigs
		_, multisigAddress, _ := s.portalParams.PortalTokens[u.tokenID].GenerateOTMultisigAddress(masterPubKeys, int(numReq), u.incAddress)
		key, value := generateUTXOKeyAndValue(u.tokenID, multisigAddress, u.outputHash, uint32(u.outputIdx), u.outputAmt, u.incAddress)
		//utxos = append(utxos, utxoTmp{key: key, value: value})
		utxos[key] = value
	}

	// build test cases
	testInit := TestCaseUTXOProcess{
		utxos: map[string]map[string]*statedb.UTXO{
			portal.TestnetPortalV4BTCID: utxos,
		},
	}

	return testInit, requestUnshieldData, listUnshields
}

func (s *PortalTestSuiteV4) TestUTXOProcess() {
	fmt.Println("Running TestUTXOProcess - beacon height 1538097 ...")
	//bc := s.blockChain
	// mock test
	bc := new(mocks.ChainRetriever)
	bc.On("GetBTCChainParams").Return(&chaincfg.TestNet3Params)
	bc.On("GetFinalBeaconHeight").Return(uint64(42))
	tokenID := portal.TestnetPortalV4BTCID
	bcH := uint64(45)
	bc.On("GetPortalV4GeneralMultiSigAddress", tokenID, bcH).Return(s.portalParams.GeneralMultiSigAddresses[portal.TestnetPortalV4BTCID])

	pm := portal.NewPortalManager()
	beaconHeight := uint64(46)
	shardHeights := map[byte]uint64{
		0: uint64(1003),
	}
	shardID := byte(0)

	s.SetupTestUTXOProcess()

	// build test cases and expected results
	testInit, testcases, listUnshields := s.buildTestCaseAndExpectedResultUTXOProcess()

	s.currentPortalStateForProducer.UTXOs = testInit.utxos
	s.currentPortalStateForProducer.WaitingUnshieldRequests = testInit.waitingUnshieldReqs
	s.currentPortalStateForProcess.UTXOs = testInit.utxos
	s.currentPortalStateForProcess.WaitingUnshieldRequests = testInit.waitingUnshieldReqs
	s.currentPortalStateForProducer.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	s.currentPortalStateForProcess.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}

	for i := 0; i < len(testcases.testcases); i++ {
		s.currentPortalStateForProducer.WaitingUnshieldRequests = map[string]map[string]*statedb.WaitingUnshieldRequest{
			portal.TestnetPortalV4BTCID: testcases.testcases[i],
		}
		fmt.Printf("UTOXs length before unshield %v \n", len(s.currentPortalStateForProcess.UTXOs[portal.TestnetPortalV4BTCID]))
		fmt.Printf("Unshield request list %v \n", s.currentPortalStateForProducer.WaitingUnshieldRequests[portal.TestnetPortalV4BTCID])

		// beacon producer instructions
		newInsts, err := pm.PortalInstProcessorsV4[metadataCommon.PortalV4UnshieldBatchingMeta].BuildNewInsts(bc, "", shardID, &s.currentPortalStateForProducer, beaconHeight-1, shardHeights, s.portalParams, nil)
		s.Equal(nil, err)

		// process new instructions
		err = processPortalInstructionsV4(
			bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV4)

		// check results after unshield
		fmt.Printf("test result %v \n", s.currentPortalStateForProcess.ProcessedUnshieldRequests)
		fmt.Printf("UTOXs length after unshield %v \n", len(s.currentPortalStateForProcess.UTXOs[portal.TestnetPortalV4BTCID]))

		// process unshield requests. fee 3619
		fee := uint64(3619)
		if len(s.currentPortalStateForProcess.ProcessedUnshieldRequests) > 0 {
			processedUnshieldRequests := s.currentPortalStateForProcess.ProcessedUnshieldRequests[portal.TestnetPortalV4BTCID]
			for _, v := range processedUnshieldRequests {
				totalInput := uint64(0)
				totalOutput := uint64(0)
				for _, input := range v.GetUTXOs() {
					totalInput += input.GetOutputAmount()
				}
				for _, output := range v.GetUnshieldRequests() {
					fmt.Println(output)
					fmt.Printf("listUnshield %v \n", listUnshields)
					totalOutput += listUnshields[output].GetAmount()
				}

				if totalInput > (totalOutput + fee) {
					change := totalInput - (totalOutput + fee)
					key, value := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, USER_BTC_ADDRESS_1, common.HashH([]byte(strconv.Itoa(rand.Int()))).String(), uint32(len(v.GetUnshieldRequests())+1), change, "")
					s.currentPortalStateForProcess.UTXOs[portal.TestnetPortalV4BTCID][key] = value
				}
			}
		}
		s.currentPortalStateForProducer.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}

		// check results after unshield
		fmt.Printf("test result %v \n", s.currentPortalStateForProcess.ProcessedUnshieldRequests)
		fmt.Printf("UTOXs length after unshield successfully %v \n", len(s.currentPortalStateForProcess.UTXOs[portal.TestnetPortalV4BTCID]))
	}
}

func TestPortalSuiteV4(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV4))
}
