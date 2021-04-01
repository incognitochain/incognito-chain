package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
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
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/mocks"
	"github.com/incognitochain/incognito-chain/portal"
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
		NumRequiredSigs: 3,
		GeneralMultiSigAddresses: map[string]string{
			portalcommonv4.PortalBTCIDStr: "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X",
		},
		GeneralMultiSigScriptHexEncode: map[string]string{
			portalcommonv4.PortalBTCIDStr: "",
		},
		PortalTokens: map[string]portaltokensv4.PortalTokenProcessor{
			portalcommonv4.PortalBTCIDStr: &portaltokensv4.PortalBTCTokenProcessor{
				PortalToken: &portaltokensv4.PortalToken{
					ChainID:        "Bitcoin-Testnet",
					MinTokenAmount: 10,
				},
				ChainParam: &chaincfg.RegressionNetParams,
			},
		},
		DefaultFeeUnshields: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 100000, // in nano pBTC - 10000 satoshi
		},
		MinUnshieldAmts: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 1000000, // in nano pBTC - 1000000 satoshi
		},
		TinyUTXOAmount: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 1e9, // in nano pBTC - 100000 satoshi
		},
		BatchNumBlks:                45,
		MinConfirmationIncBlockNum:  3,
		PortalReplacementAddress:    "",
		MaxFeePercentageForEachStep: 20,
		TimeSpaceForFeeReplacement:  2 * time.Minute,
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

func generateUTXOKeyAndValue(tokenID string, walletAddress string, txHash string, outputIdx uint32, outputAmount uint64, publicSeed string) (string, *statedb.UTXO) {
	utxoKey := statedb.GenerateUTXOObjectKey(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx).String()
	utxoValue := statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount, publicSeed)
	return utxoKey, utxoValue
}

//TODO: update shielding proof
func buildTestCaseAndExpectedResultShieldingRequest() ([]TestCaseShieldingRequest, *ExpectedResultShieldingRequest) {
	// build test cases
	testcases := []TestCaseShieldingRequest{
		// valid shielding request
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_1,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{1}).String(),
			isExistsInPreviousBlocks: false,
		},
		// valid shielding request: different user incognito address
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_2,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzIwOCw0Niw2OCwxOSw1NSwyMDksMjM1LDIyMCw0OSwxMTEsNjQsMjEsODEsMTA5LDI0NiwyMTAsMTUyLDEzMywxMDksNjgsNjMsMTU5LDI0MSwxNjUsMTk4LDE2LDUsMjYsMjA0LDIzNCw3NCwxODhdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTM2LDQzLDIyNCwxLDE3NiwxMTQsMTQxLDIxMywzMywxMTcsNjAsNzYsNjcsMzgsMjAsNDksMTE4LDE5NSwyNTMsMjMyLDE1MCw4MiwxNDksMTY1LDE2OCwxNDIsMjA3LDI1NSwxNiw1NCw3Miw1MF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTgwLDM2LDE4LDgyLDIxMywzOSwxMDksMTc1LDIwNiwxMjgsMjUwLDYsMjM4LDM2LDE2MiwyMTAsMjMyLDEzNCwxNDYsMTI0LDksNTgsMTA0LDEzNSwxNDgsMTI5LDE4OCwxNDIsMjM5LDE5LDE4MiwzXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2NywxMDUsMTE3LDI4LDg3LDExOCwxMSwxNCwxNzgsMTE0LDk4LDExOCwxNDcsMTA3LDEwNyw5NSw0MywyMzEsNTEsMjEsMTYwLDQwLDk1LDEwLDIyNSwyNTUsMTQ5LDIzMiwyMjEsMjM1LDI0OCwzMF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyNDgsMjQ0LDE3MCw0NywzMSwxMTIsMTE4LDI0MSw0OCwxOTIsMzAsMTc2LDE1MSw4LDQ5LDYsNDIsMTE0LDE1NSwxMjIsMjEwLDIxMSw4NSwyMTQsODQsNDgsMjQ0LDE4MCw4NSw2NCwyNCw4M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTI3LDEyNCw5MywyMjEsMjQ4LDE3MywxOTQsMTYsMTU3LDUwLDYwLDE4MCwyNDAsMTMxLDQzLDExNCwxNDQsMTI4LDIwMSw0NSwxNjEsMjAsMjIxLDY3LDgwLDk4LDE4LDExMSwyNTIsMjE3LDMyLDU1XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsyMDgsNDYsNjgsMTksNTUsMjA5LDIzNSwyMjAsNDksMTExLDY0LDIxLDgxLDEwOSwyNDYsMjEwLDE1MiwxMzMsMTA5LDY4LDYzLDE1OSwyNDEsMTY1LDE5OCwxNiw1LDI2LDIwNCwyMzQsNzQsMTg4XSwiSW5kZXgiOjJ9LCJTaWduYXR1cmVTY3JpcHQiOiJTREJGQWlFQXhzOWRTMlE4YWVrMHkvRjJmOUdkejB5R0VFREVzWjNYQXpKZVFYamRiK1VDSUNlZmwwREZWV0tPZUk2Rm9FQzRsS01jWE5hRnFlL0o4ZFpTQnZNUnQyRkRBU0VEenlBVFQxWkk0dmJ4ZDZaVkt5VzZsK1JiRklWT1R4TXFTdkQrZmFrL3hQdz0iLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjAsIlBrU2NyaXB0IjoiYWt4clVGTXhMVEV5VXpWTWNuTXhXR1ZSVEdKeFRqUjVVM2xMZEdwQmFtUXlaRGR6UWxBeWRHcEdhV3A2YlhBMllYWnljbXRSUTA1R1RYQnJXRzB6UmxCNmFqSlhZM1V5V2s1eFNrVnRhRGxLY21sV2RWSkZjbFozYUhWUmJreHRWMU5oWjJkdllrVlhjMEpGWTJrPSJ9LHsiVmFsdWUiOjgwMCwiUGtTY3JpcHQiOiJxUlFuSjZkdjh2bzVYY1VsWktqcktxdU0vbEhJZG9jPSJ9LHsiVmFsdWUiOjc2MjcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOls2NCw2NywyNDMsMjIyLDI0NCw5LDE4MywyMzQsMjM4LDE3MCwxNjYsMjI2LDEzLDI1NCwzNSw3OCwyMjIsNjksMjQxLDIyNCwxMDMsNjMsMTI5LDIwNCwxNCwwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{2}).String(),
			isExistsInPreviousBlocks: false,
		},
		// valid shielding request: the same user incognito address
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_2,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzIwOCw0Niw2OCwxOSw1NSwyMDksMjM1LDIyMCw0OSwxMTEsNjQsMjEsODEsMTA5LDI0NiwyMTAsMTUyLDEzMywxMDksNjgsNjMsMTU5LDI0MSwxNjUsMTk4LDE2LDUsMjYsMjA0LDIzNCw3NCwxODhdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTM2LDQzLDIyNCwxLDE3NiwxMTQsMTQxLDIxMywzMywxMTcsNjAsNzYsNjcsMzgsMjAsNDksMTE4LDE5NSwyNTMsMjMyLDE1MCw4MiwxNDksMTY1LDE2OCwxNDIsMjA3LDI1NSwxNiw1NCw3Miw1MF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTgwLDM2LDE4LDgyLDIxMywzOSwxMDksMTc1LDIwNiwxMjgsMjUwLDYsMjM4LDM2LDE2MiwyMTAsMjMyLDEzNCwxNDYsMTI0LDksNTgsMTA0LDEzNSwxNDgsMTI5LDE4OCwxNDIsMjM5LDE5LDE4MiwzXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2NywxMDUsMTE3LDI4LDg3LDExOCwxMSwxNCwxNzgsMTE0LDk4LDExOCwxNDcsMTA3LDEwNyw5NSw0MywyMzEsNTEsMjEsMTYwLDQwLDk1LDEwLDIyNSwyNTUsMTQ5LDIzMiwyMjEsMjM1LDI0OCwzMF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyNDgsMjQ0LDE3MCw0NywzMSwxMTIsMTE4LDI0MSw0OCwxOTIsMzAsMTc2LDE1MSw4LDQ5LDYsNDIsMTE0LDE1NSwxMjIsMjEwLDIxMSw4NSwyMTQsODQsNDgsMjQ0LDE4MCw4NSw2NCwyNCw4M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTI3LDEyNCw5MywyMjEsMjQ4LDE3MywxOTQsMTYsMTU3LDUwLDYwLDE4MCwyNDAsMTMxLDQzLDExNCwxNDQsMTI4LDIwMSw0NSwxNjEsMjAsMjIxLDY3LDgwLDk4LDE4LDExMSwyNTIsMjE3LDMyLDU1XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsyMDgsNDYsNjgsMTksNTUsMjA5LDIzNSwyMjAsNDksMTExLDY0LDIxLDgxLDEwOSwyNDYsMjEwLDE1MiwxMzMsMTA5LDY4LDYzLDE1OSwyNDEsMTY1LDE5OCwxNiw1LDI2LDIwNCwyMzQsNzQsMTg4XSwiSW5kZXgiOjJ9LCJTaWduYXR1cmVTY3JpcHQiOiJTREJGQWlFQXhzOWRTMlE4YWVrMHkvRjJmOUdkejB5R0VFREVzWjNYQXpKZVFYamRiK1VDSUNlZmwwREZWV0tPZUk2Rm9FQzRsS01jWE5hRnFlL0o4ZFpTQnZNUnQyRkRBU0VEenlBVFQxWkk0dmJ4ZDZaVkt5VzZsK1JiRklWT1R4TXFTdkQrZmFrL3hQdz0iLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjAsIlBrU2NyaXB0IjoiYWt4clVGTXhMVEV5VXpWTWNuTXhXR1ZSVEdKeFRqUjVVM2xMZEdwQmFtUXlaRGR6UWxBeWRHcEdhV3A2YlhBMllYWnljbXRSUTA1R1RYQnJXRzB6UmxCNmFqSlhZM1V5V2s1eFNrVnRhRGxLY21sV2RWSkZjbFozYUhWUmJreHRWMU5oWjJkdllrVlhjMEpGWTJrPSJ9LHsiVmFsdWUiOjgwMCwiUGtTY3JpcHQiOiJxUlFuSjZkdjh2bzVYY1VsWktqcktxdU0vbEhJZG9jPSJ9LHsiVmFsdWUiOjc2MjcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOls2NCw2NywyNDMsMjIyLDI0NCw5LDE4MywyMzQsMjM4LDE3MCwxNjYsMjI2LDEzLDI1NCwzNSw3OCwyMjIsNjksMjQxLDIyNCwxMDMsNjMsMTI5LDIwNCwxNCwwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: false,
		},
		// invalid shielding request: duplicated shielding proof in previous blocks
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_2,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{4}).String(),
			isExistsInPreviousBlocks: true,
		},
		// invalid shielding request: duplicated shielding proof in the current block
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            PORTALV4_USER_INC_ADDRESS_2,
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{5}).String(),
			isExistsInPreviousBlocks: false,
		},
	}

	walletAddress := "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X"

	// build expected results
	var txHash string
	var outputIdx uint32
	var outputAmount uint64

	//todo: update
	txHash = "bc4aeacc1a0510c6a5f19f3f446d8598d2f66d5115406f31dcebd13713442ed0"
	outputIdx = 1
	outputAmount = 400

	key1, value1 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount, USER_INC_ADDRESS_1)

	txHash = "48ea0dcdea31ed9b869a6e5914029df0089994ea3299be9161c6264e333284a5"
	outputIdx = 1
	outputAmount = 800

	key2, value2 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount, USER_INC_ADDRESS_2)

	txHash = "48ea0dcdea31ed9b869a6e5914029df0089994ea3299be9161c6264e333284a5"
	outputIdx = 1
	outputAmount = 800

	key3, value3 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount, USER_INC_ADDRESS_2)

	expectedRes := &ExpectedResultShieldingRequest{
		utxos: map[string]map[string]*statedb.UTXO{
			portalcommonv4.PortalBTCIDStr: {
				key1: value1,
				key2: value2,
				key3: value3,
			},
		},
		numBeaconInsts: 5,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
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
	genesisBlockHeight := 1940329
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
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 1 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{1}).String(),
			isExisted:      false,
		},
		// valid unshield request
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 0.5 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_2,
			txId:           common.HashH([]byte{2}).String(),
			isExisted:      false,
		},
		// invalid unshield request - invalid unshield amount
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 999999,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{3}).String(),
			isExisted:      false,
		},
		// invalid unshield request - existed unshield ID
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 1 * 1e9,
			incAddressStr:  PORTALV4_USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{1}).String(),
			isExisted:      true,
		},
	}

	// build expected results
	// waiting unshielding requests
	waitingUnshieldReqKey1 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, common.HashH([]byte{1}).String()).String()
	waitingUnshieldReq1 := statedb.NewWaitingUnshieldRequestStateWithValue(
		USER_BTC_ADDRESS_1, 1*1e9, common.HashH([]byte{1}).String(), beaconHeight)
	waitingUnshieldReqKey2 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, common.HashH([]byte{2}).String()).String()
	waitingUnshieldReq2 := statedb.NewWaitingUnshieldRequestStateWithValue(
		USER_BTC_ADDRESS_2, 0.5*1e9, common.HashH([]byte{2}).String(), beaconHeight)

	expectedRes := &ExpectedResultUnshieldRequest{
		waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
			portalcommonv4.PortalBTCIDStr: {
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
			Type: metadata.PortalV4UnshieldingRequestMeta,
		},
		IncAddressStr:  incAddressStr,
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
	return []string{strconv.Itoa(metadata.PortalV4UnshieldingRequestMeta), actionContentBase64Str}
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
	statusInsts            []string
}

func (s *PortalTestSuiteV4) SetupTestBatchUnshieldProcess() {
	// do nothing
}

func buildTestCaseAndExpectedResultBatchUnshieldProcess() ([]TestCaseBatchUnshieldProcess, []ExpectedResultBatchUnshieldProcess) {
	// waiting unshielding requests
	unshieldId1 := common.HashH([]byte{1}).String()
	unshieldAmt1 := uint64(0.6 * 1e9)
	remoteAddr1 := USER_BTC_ADDRESS_1
	beaconHeight := uint64(1)
	wUnshieldReqKey1 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId1).String()
	wUnshieldReq1 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr1, unshieldAmt1, unshieldId1, beaconHeight)

	unshieldId2 := common.HashH([]byte{2}).String()
	unshieldAmt2 := uint64(0.5 * 1e9)
	remoteAddr2 := USER_BTC_ADDRESS_2
	beaconHeight = uint64(2)
	wUnshieldReqKey2 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId2).String()
	wUnshieldReq2 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr2, unshieldAmt2, unshieldId2, beaconHeight)

	unshieldId3 := common.HashH([]byte{3}).String()
	unshieldAmt3 := uint64(0.7 * 1e9)
	remoteAddr3 := USER_BTC_ADDRESS_2
	beaconHeight = uint64(3)
	wUnshieldReqKey3 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId3).String()
	wUnshieldReq3 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr3, unshieldAmt3, unshieldId3, beaconHeight)

	unshieldId4 := common.HashH([]byte{4}).String()
	unshieldAmt4 := uint64(0.4 * 1e9)
	remoteAddr4 := USER_BTC_ADDRESS_2
	beaconHeight = uint64(4)
	wUnshieldReqKey4 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId4).String()
	wUnshieldReq4 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr4, unshieldAmt4, unshieldId4, beaconHeight)

	unshieldId5 := common.HashH([]byte{5}).String()
	unshieldAmt5 := uint64(2.4 * 1e9)
	remoteAddr5 := USER_BTC_ADDRESS_2
	beaconHeight = uint64(1)
	wUnshieldReqKey5 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId5).String()
	wUnshieldReq5 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr5, unshieldAmt5, unshieldId5, beaconHeight)

	unshieldId6 := common.HashH([]byte{6}).String()
	unshieldAmt6 := uint64(0.1 * 1e9)
	remoteAddr6 := USER_BTC_ADDRESS_2
	beaconHeight = uint64(43)
	wUnshieldReqKey6 := statedb.GenerateWaitingUnshieldRequestObjectKey(portalcommonv4.PortalBTCIDStr, unshieldId6).String()
	wUnshieldReq6 := statedb.NewWaitingUnshieldRequestStateWithValue(
		remoteAddr6, unshieldAmt6, unshieldId6, beaconHeight)

	// utxos
	walletAddress := "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X"
	var utxoTxHash1 string
	var utxoOutputIdx uint32
	var utxoAmount uint64

	utxoTxHash1 = "251f22409938485254c6c739b9f91cb7f0bce784b018de5b0bdb88d10f17e355"
	utxoOutputIdx = 1
	utxoAmount = 2 * 1e8
	keyUtxo1, valueUtxo1 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, utxoTxHash1, utxoOutputIdx, utxoAmount)

	utxoTxHash1 = "b44f6c7c896757abe7afd6ac083c2930f1d0f57a356887e872f3b88bba5ea0b7"
	utxoOutputIdx = 1
	utxoAmount = 1 * 1e8
	keyUtxo2, valueUtxo2 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, utxoTxHash1, utxoOutputIdx, utxoAmount)

	utxoTxHash1 = "93aaa4b3109815cc33273154732d033ddc959f2aad166a5dbeac1f72b3f5e5cd"
	utxoOutputIdx = 1
	utxoAmount = 0.1 * 1e8
	keyUtxo3, valueUtxo3 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, utxoTxHash1, utxoOutputIdx, utxoAmount)

	// build test cases
	testcases := []TestCaseBatchUnshieldProcess{
		// TC0 - success: there is only one waiting unshield request, one utxo
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey1: wUnshieldReq1,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo1: valueUtxo1,
				},
			},
		},
		// TC1 - success: there is only one waiting unshield request, multiple utxos (choose one to spend, and a smaller one)
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey1: wUnshieldReq1,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo1: valueUtxo1,
					keyUtxo2: valueUtxo2,
				},
			},
		},
		// TC2 - success: multiple waiting unshield requests, multiple utxos (utxo amount > unshield amount)
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey1: wUnshieldReq1,
					wUnshieldReqKey2: wUnshieldReq2,
					wUnshieldReqKey3: wUnshieldReq3,
					wUnshieldReqKey4: wUnshieldReq4,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo1: valueUtxo1,
					keyUtxo2: valueUtxo2,
					keyUtxo3: valueUtxo3,
				},
			},
		},
		// TC3 - success: multiple waiting unshield requests, multiple utxos (utxo amount < unshield amount)
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey5: wUnshieldReq5,
					wUnshieldReqKey1: wUnshieldReq1,
					wUnshieldReqKey2: wUnshieldReq2,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo1: valueUtxo1,
					keyUtxo2: valueUtxo2,
				},
			},
		},
		// TC4 - success: doesn't have enough utxos for any unshield request
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey5: wUnshieldReq5,
					wUnshieldReqKey1: wUnshieldReq1,
					wUnshieldReqKey2: wUnshieldReq2,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo3: valueUtxo3,
				},
			},
		},
		// TC5 - success: there is a waiting unshield request hasn't enough confirmation blocks
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey6: wUnshieldReq6,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo3: valueUtxo3,
				},
			},
		},
	}

	// build expected results
	// batch unshielding process
	currentBeaconHeight := uint64(45)
	var batchID string
	var processedUnshieldIDs []string
	var spendUtxos map[string][]*statedb.UTXO
	var externalFee map[uint64]uint

	processedUnshieldIDs = []string{unshieldId1}
	batchID = portalprocessv4.GetBatchID(currentBeaconHeight, processedUnshieldIDs)
	spendUtxos = map[string][]*statedb.UTXO{walletAddress: {valueUtxo1}}
	externalFee = map[uint64]uint{
		currentBeaconHeight: 100000,
	}
	batchUnshieldProcessKey1 := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portalcommonv4.PortalBTCIDStr, batchID).String()
	batchUnshieldProcess1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, processedUnshieldIDs, spendUtxos, externalFee)

	processedUnshieldIDs = []string{unshieldId1}
	batchID = portalprocessv4.GetBatchID(currentBeaconHeight, processedUnshieldIDs)
	spendUtxos = map[string][]*statedb.UTXO{walletAddress: {valueUtxo1, valueUtxo2}}
	externalFee = map[uint64]uint{
		currentBeaconHeight: 100000,
	}
	batchUnshieldProcessKey2 := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portalcommonv4.PortalBTCIDStr, batchID).String()
	batchUnshieldProcess2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, processedUnshieldIDs, spendUtxos, externalFee)

	processedUnshieldIDs = []string{unshieldId1, unshieldId2, unshieldId3}
	batchID = portalprocessv4.GetBatchID(currentBeaconHeight, processedUnshieldIDs)
	spendUtxos = map[string][]*statedb.UTXO{walletAddress: {valueUtxo1, valueUtxo3}}
	externalFee = map[uint64]uint{
		currentBeaconHeight: 100000,
	}
	batchUnshieldProcessKey3 := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portalcommonv4.PortalBTCIDStr, batchID).String()
	batchUnshieldProcess3 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, processedUnshieldIDs, spendUtxos, externalFee)

	processedUnshieldIDs = []string{unshieldId4}
	batchID = portalprocessv4.GetBatchID(currentBeaconHeight, processedUnshieldIDs)
	spendUtxos = map[string][]*statedb.UTXO{walletAddress: {valueUtxo2}}
	externalFee = map[uint64]uint{
		currentBeaconHeight: 100000,
	}
	batchUnshieldProcessKey4 := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portalcommonv4.PortalBTCIDStr, batchID).String()
	batchUnshieldProcess4 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, processedUnshieldIDs, spendUtxos, externalFee)

	processedUnshieldIDs = []string{unshieldId5, unshieldId1}
	batchID = portalprocessv4.GetBatchID(currentBeaconHeight, processedUnshieldIDs)
	spendUtxos = map[string][]*statedb.UTXO{walletAddress: {valueUtxo1, valueUtxo2}}
	externalFee = map[uint64]uint{
		currentBeaconHeight: 100000,
	}
	batchUnshieldProcessKey5 := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(portalcommonv4.PortalBTCIDStr, batchID).String()
	batchUnshieldProcess5 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, processedUnshieldIDs, spendUtxos, externalFee)

	expectedRes := []ExpectedResultBatchUnshieldProcess{
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
				portalcommonv4.PortalBTCIDStr: {
					batchUnshieldProcessKey1: batchUnshieldProcess1,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {},
			},
			numBeaconInsts: 1,
		},
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
				portalcommonv4.PortalBTCIDStr: {
					batchUnshieldProcessKey2: batchUnshieldProcess2,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {},
			},
			numBeaconInsts: 1,
		},
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
				portalcommonv4.PortalBTCIDStr: {
					batchUnshieldProcessKey3: batchUnshieldProcess3,
					batchUnshieldProcessKey4: batchUnshieldProcess4,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {},
			},
			numBeaconInsts: 2,
		},
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey2: wUnshieldReq2,
				},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
				portalcommonv4.PortalBTCIDStr: {
					batchUnshieldProcessKey5: batchUnshieldProcess5,
				},
			},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {},
			},
			numBeaconInsts: 1,
		},
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey5: wUnshieldReq5,
					wUnshieldReqKey1: wUnshieldReq1,
					wUnshieldReqKey2: wUnshieldReq2,
				},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo3: valueUtxo3,
				},
			},
			numBeaconInsts: 0,
		},
		{
			waitingUnshieldReqs: map[string]map[string]*statedb.WaitingUnshieldRequest{
				portalcommonv4.PortalBTCIDStr: {
					wUnshieldReqKey6: wUnshieldReq6,
				},
			},
			batchUnshieldProcesses: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
			utxos: map[string]map[string]*statedb.UTXO{
				portalcommonv4.PortalBTCIDStr: {
					keyUtxo3: valueUtxo3,
				},
			},
			numBeaconInsts: 0,
		},
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV4) TestBatchUnshieldProcess() {
	fmt.Println("Running TestBatchUnshieldProcess - beacon height 45 ...")
	//bc := s.blockChain
	// mock test
	bc := new(mocks.ChainRetriever)
	bc.On("GetBTCChainParams").Return(&chaincfg.TestNet3Params)
	tokenID := portalcommonv4.PortalBTCIDStr
	bcH := uint64(0)
	bc.On("GetPortalV4GeneralMultiSigAddress", tokenID, bcH).Return(s.portalParams.GeneralMultiSigAddresses[portalcommonv4.PortalBTCIDStr])

	pm := portal.NewPortalManager()
	beaconHeight := uint64(45)
	shardHeights := map[byte]uint64{
		0: uint64(1003),
	}
	shardID := byte(0)

	s.SetupTestBatchUnshieldProcess()

	// build test cases and expected results
	testcases, expectedResults := buildTestCaseAndExpectedResultBatchUnshieldProcess()
	if len(testcases) != len(expectedResults) {
		fmt.Errorf("Testcases and expected results is invalid")
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
		newInsts, err := pm.PortalInstProcessorsV4[metadata.PortalV4UnshieldBatchingMeta].BuildNewInsts(bc, "", shardID, &s.currentPortalStateForProducer, beaconHeight-1, shardHeights, s.portalParams, nil)
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
const keyBatchShield1 = "9da36f3e18071935a3d812f47e2cb86f48f49260d89b62fbf8b8c9bdc1cceb5a"
const keyBatchShield2 = "b83ad865d55f3e5399e455ad5c561ecc9b31f8cbd89b62fbf8b8c9bdc1cceb5a"
const keyBatchShield3 = "8da36f3e18071935a3d812f47e2cb86f48f49260681df4129d4538f9bfcd4cad"

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

	btcMultiSigAddress := s.portalParams.GeneralMultiSigAddresses[portalcommonv4.PortalBTCIDStr]
	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1", "txid2", "txid3"},
		map[string][]*statedb.UTXO{
			btcMultiSigAddress: {
				statedb.NewUTXOWithValue(btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 1000000),
				statedb.NewUTXOWithValue(btcMultiSigAddress, "49491148bd2f7b5432a26472af97724e114f22a74d9d2fb20c619b4f79f19fd9", 0, 2000000),
				statedb.NewUTXOWithValue(btcMultiSigAddress, "b751ff30df21ad84ce3f509ee3981c348143bd6a5aa30f4256ecb663fab14fd1", 1, 3000000),
			},
		},
		map[uint64]uint{
			900: 900,
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		map[string][]*statedb.UTXO{
			btcMultiSigAddress: {
				statedb.NewUTXOWithValue(btcMultiSigAddress, "163a6cc24df4efbd5c997aa623d4e319f1b7671be83a86bb0fa27bc701ae4a76", 1, 1000000),
			},
		},
		map[uint64]uint{
			1000: 1000,
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portalcommonv4.PortalBTCIDStr: {
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
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalV4FeeReplacementRequestMeta,
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
	return []string{strconv.Itoa(metadata.PortalV4FeeReplacementRequestMeta), actionContentBase64Str}
}

func buildExpectedResultFeeReplacement(s *PortalTestSuiteV4) ([]TestCaseFeeReplacement, *ExpectedResultFeeReplacement) {

	testcases := []TestCaseFeeReplacement{
		// request replace fee higher than max step
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID1,
			fee:     1500,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          100,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          100,
				},
			},
		},
		// request replace lower than latest request
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID1,
			fee:     800,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          100,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          100,
				},
			},
		},
		// request replace fee successfully
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID1,
			fee:     1000,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          100,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          100,
				},
			},
		},
		// request replace fee with beacon height lower than next acceptable beacon height
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID1,
			fee:     1100,
			outputs: []OutPut{
				{
					externalAddress: "2N4PN5oRh5JdizoBvESPnV2yyPBGCfUNwAr",
					amount:          100,
				},
				{
					externalAddress: "2ND6aMnSBt4jrMiGvaxKzW57rUEDihsQARK",
					amount:          100,
				},
			},
		},
		// request replace fee new batch id
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID2,
			fee:     1200,
			outputs: []OutPut{
				{
					externalAddress: "2NBf3uA9wMJRT2eM7AyXkM6RXcPfDi24rPA",
					amount:          200,
				},
			},
		},
		// request replace fee with non exist batch id
		{
			tokenID: portalcommonv4.PortalBTCIDStr,
			batchID: BatchID3,
			fee:     1000,
			outputs: []OutPut{
				{
					externalAddress: "2N8mFbLG59ugUJM9ZBP292i6nXZHmfAw5Lk",
					amount:          100,
				},
			},
		},
	}

	btcMultiSigAddress := s.portalParams.GeneralMultiSigAddresses[portalcommonv4.PortalBTCIDStr]
	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1", "txid2", "txid3"},
		map[string][]*statedb.UTXO{
			btcMultiSigAddress: {
				statedb.NewUTXOWithValue(btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 1000000, ""),
				statedb.NewUTXOWithValue(btcMultiSigAddress, "49491148bd2f7b5432a26472af97724e114f22a74d9d2fb20c619b4f79f19fd9", 0, 2000000, ""),
				statedb.NewUTXOWithValue(btcMultiSigAddress, "b751ff30df21ad84ce3f509ee3981c348143bd6a5aa30f4256ecb663fab14fd1", 1, 3000000, ""),
			},
		},
		map[uint64]uint{
			900:  900,
			1500: 1000,
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(btcMultiSigAddress, "163a6cc24df4efbd5c997aa623d4e319f1b7671be83a86bb0fa27bc701ae4a76", 1, 1000000, ""),
		},
		map[uint64]uint{
			1000: 1000,
			1500: 1200,
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portalcommonv4.PortalBTCIDStr: {
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
	bc.On("CheckBlockTimeIsReachedByBeaconHeight", mock.AnythingOfTypeArgument("uint64"), mock.AnythingOfTypeArgument("uint64"), mock.AnythingOfTypeArgument("time.Duration")).Return(
		func(recentBeaconHeight, beaconHeight uint64, duration time.Duration) bool {
			{
				return (recentBeaconHeight+1)-beaconHeight >= s.blockChain.convertDurationTimeToBeaconBlocks(duration)
			}
		})
	bc.On("GetBTCChainParams").Return(&chaincfg.TestNet3Params)

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

const confirmedTxProof1 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzUyLDk3LDgzLDI0MiwxOTMsODgsMTYzLDE1MiwxNSwzMiwxNDYsMzQsNCwxMTAsMTQsMjI5LDMyLDc5LDIwNSwxNjAsMTIsMzEsMjQ4LDU0LDE1NywxNjYsNjYsMjE4LDgsNDIsNDgsNjBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbNzYsODQsMTc4LDI1MCwyMjIsMjA3LDgyLDUsMjA4LDEwNCwxOTUsMzUsODUsMjA3LDI1MCw3OSwyMzUsMjA3LDk3LDM5LDEwMSwxNTUsODUsMTgyLDExOCwxODAsNiwxNzMsMTM1LDY4LDEwOSwxNDRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzMzLDE0MSwyMzUsMjI4LDExMiwxODUsMjQ4LDIwMywyMDYsMTEwLDczLDE5LDcwLDIzNCwzOSw0NCwxMDcsNjEsMTQwLDEwMywyNTUsOTIsMTA5LDE3NSw1OCwxOTEsMTAwLDE3NywxMzksMTIzLDMzLDFdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzM1LDUyLDc4LDIwMyw4OCw5NCw1Niw0LDE3MywyNTQsNTAsNjgsMzgsMjUsMjI2LDU1LDE4NSw0MywyMDQsMjUzLDk5LDE3Nyw1MywxNywxNjUsMTc2LDExNiwxNTIsOTcsMiwxMjUsMTZdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzM3LDIwMyw2OSw4Miw1LDIxNiwxOTYsMjI4LDYsMjI2LDMxLDI0NSwyMywyMTIsMTIwLDIyMywyNTQsMTE0LDE3OCwxNSwyNCwyMDksMTQzLDE3MiwxMjMsMjQxLDQzLDE4Nyw4NywxNywyMSwxMTldLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbNDMsNDUsMTA4LDkyLDE1MCwzNyw2NSwyMjksMTEsMTk1LDE2OCwyNDksMTY5LDc3LDQwLDIwMywyMzMsMTU0LDcwLDUyLDI0MywxMDMsMjQ4LDc5LDIwNCw4NywxNDgsODMsMjgsMTQxLDEzMiw1MV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTIsOTcsODMsMjQyLDE5Myw4OCwxNjMsMTUyLDE1LDMyLDE0NiwzNCw0LDExMCwxNCwyMjksMzIsNzksMjA1LDE2MCwxMiwzMSwyNDgsNTQsMTU3LDE2Niw2NiwyMTgsOCw0Miw0OCw2MF0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQjZmeXdwbXhvYmVRcnR3Mi9NTXhFa09Tc3JqcXNrWVZrMXhHRVlUR2VuVVFJZ0RqQzBjQ083dFptYmk0ZGF3aXV2K0RFNnhOc3hKNXB2ZVN2ZVBoZngwVHdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFnWmlZWFJqYURFPSJ9LHsiVmFsdWUiOjMwMCwiUGtTY3JpcHQiOiJxUlFuSjZkdjh2bzVYY1VsWktqcktxdU0vbEhJZG9jPSJ9LHsiVmFsdWUiOjIwNjU3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbMTQsMjMxLDkzLDQzLDUyLDU5LDIyNyw3MSwyMTgsMjEyLDE0NSwxNTksMjMzLDYsNDYsNDQsODgsMjE4LDkzLDI0OSwzMywyNDYsMjM1LDE0Niw2LDAsMCwwLDAsMCwwLDBdfQ=="
const confirmedTxProof2 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzQ0LDY2LDEwNCw0NiwxMjgsMjIwLDIxOCw2NCw3OCwxNzAsMTM5LDU2LDE4NCwyMDQsMzUsNjMsMTc0LDk5LDM1LDQ3LDE3MCwyNTEsMTU1LDIyNSw0NCw5Miw3NCwxNyw1MiwxNjYsMzcsMTYwXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls4LDY2LDIzMiwxMjUsNTksNjIsNzQsMTYyLDQsNDIsNDIsMTUwLDYzLDk5LDgzLDE0MCw3LDEyOCw2MSwyMTMsMCw0NCw5NSwxOTgsMjI1LDEyOCwyMjcsMjAzLDI3LDI0NywxNjUsMTI3XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6Wzg0LDI0NSwzOSwxNTksMTYwLDcxLDEzNSwzMiwzNiwxMTMsMjYsMTA5LDIyNCwxOTcsMTcwLDEzNiwyMzcsMjIsMTk3LDE4OSwxOTAsMTE1LDIwMCwxODMsMTY2LDg4LDYsMTc5LDEzNCw1NSwxMCwyMDNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE5OCwxNDcsMTgxLDI2LDU1LDExOSwxMjIsMTA3LDE4LDIwMSwxMjIsNDgsMTQ2LDEzOCwyNDcsMTQsMjQsNDcsMTYsMTYyLDEzOSwxMzAsMTc0LDE1Myw2NSw3Myw2NywxMDMsMjUsMTAyLDIyLDIzNV0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjMsMTEyLDIzNCwyMTQsNzcsNzMsMjEsMTg0LDE3NSwxOSwxMjgsMTYwLDIzNywxODQsNTAsMTkwLDQ3LDc5LDcwLDE3NCwzNywxMTEsMTIzLDMxLDE0OSwyNDUsMzAsNjcsMjQsMzksMTYzLDE1Ml0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxMTcsMjAzLDExOSwyMDUsMzcsMzksMTQwLDIyNiwxOTIsMyw3OSw1MiwyMDgsMjQ5LDkzLDYsMTYzLDE2NiwxODMsNjAsMzYsMTE3LDEzMiwxMzQsMTMyLDk0LDcxLDQ3LDEwMyw5NSwxNTcsMTQ0XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsxNDYsMjI5LDY3LDgyLDIzMiwxODcsNDQsMTE3LDI4LDEwMyw2MCwxODQsNjMsMTE0LDI1MywzMCw4Myw1NiwyNDksNDAsMjM4LDQxLDQwLDE0OCwxODIsMzYsMTE4LDgyLDE4MywxMTEsNzksNDFdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzQwLDIzNCw2MSwyNCwyMzUsMjksMjM3LDE1MCwyMjQsMjAwLDEzMSwyMjUsMTY1LDIzLDEyNSwxMTcsODIsNzIsMTc4LDE3OCwyMSwxMzksMTkyLDEzMiwxLDI2LDkyLDE3MSwxOTYsMTUsNjksMjA4XSwiSW5kZXgiOjJ9LCJTaWduYXR1cmVTY3JpcHQiOiJSekJFQWlCOHhlcXl0THdNalUwVEJGRDVEL2JrR0Z2TmlZREF0U1VpL1JOK3VJS3JtQUlnV2NkTElBNDBlOWpTcnNjTkNDV3BETStwakQyKzN6YW5BcjVHUnVVckV0NEJJUVBQSUJOUFZramk5dkYzcGxVckpicVg1RnNVaFU1UEV5cEs4UDU5cVQvRS9BPT0iLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjAsIlBrU2NyaXB0IjoiYWdaaVlYUmphREk9In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MjAxMTcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNDIsMTMsNDQsMTcyLDk5LDIzOCwyMzksNTIsMTAwLDIxNiwxNzEsMTYwLDIxNSwxNTYsMjUxLDQyLDg5LDE2MiwxOTIsMTk3LDIyMiwyMSwxOTMsMTUsMjIsMCwwLDAsMCwwLDAsMF19"
const confirmedTxProof3 = "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE0MywyMTEsMjI2LDExNiwyNTMsNjksMjQ2LDIyNCwxMTAsMTg0LDMwLDE1Nyw4NCwyMDcsMTQyLDI1MywxMjIsNTAsMTk0LDgsMjAzLDExOSw3NSwxODMsMjUsNjUsMTU1LDIxMywxODYsMTg0LDEyNSwxMF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls0OCwxODIsMTE2LDI1MCwzOSwxMDgsMTk1LDE0NCwyMSw3OSwyMjIsNzQsMTk3LDE2MSwxMDcsMTYwLDIxLDMwLDIwNiwyNDksMTc5LDExMSwyMjMsMzIsNDcsMTM5LDE1MywyOCwxOTIsMjIwLDE0NiwyNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxMzgsNSwzOSw3NCwyNCw3NSw4MSw2MCwxNjcsNDYsMTg2LDEwNiwxNTAsNDQsMjAwLDIxLDIzOCw0MSwyMzQsMzksMjI1LDkyLDExLDIzNCwxNDAsMTA3LDI0OCwyNDQsMTQ0LDExNiwyMTksMTM2XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzE4MiwxMDYsOTEsMTYxLDE0NSwxMzMsMjQ2LDc1LDIwOSw3NCwxODEsMTgyLDkyLDI1NCw0OSwxOTMsNTEsMjMzLDE1NywxODUsNTQsNzMsNTAsMjQ0LDEwNywzMiwzMSwxODksNDMsNCwxMTIsMTI4XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsxMiwzNSwxODIsMTk3LDE5NiwxODYsMTQzLDE1MSw0MywxMDMsMjU1LDE2LDE2MSwyNDAsMTM5LDE2OCwxNzEsOTgsODYsMTA3LDk3LDIxMiw5MCwxNjUsMTQ5LDYyLDMwLDY1LDc1LDIyOCw2NywxODFdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbNDcsMTY0LDYwLDcsODAsMTQsNzQsMTY1LDE5NSwxODYsMTE2LDY0LDExOCwxMDIsMTk1LDEsMTMxLDQ0LDU5LDE3MSwyMDEsMTU3LDc2LDUzLDgyLDksMTM4LDE3OSw5MSw2LDQsNDRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzIzNSwyNDEsMTY4LDM0LDE4MywxNzgsNDEsMjUzLDEwMiwxMzYsMTg2LDg3LDE4OCwyMzQsMzgsMTU4LDExMSwyMjUsMTIyLDIzMCwyMjksNDgsMTgyLDEwNiwyNyw2NSwyMTQsNDIsMTUzLDQyLDMwLDkwXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2LDE4NCwyOCwyNDgsMTcyLDM0LDE0MywyNTEsMTcwLDEzLDIxNyw3OSwyMjcsMTA2LDIxMiw1NSw5MSwyMDMsMTAzLDkwLDkwLDIyLDI0Niw2NSw0OCwyMTMsMjU1LDE5OSwzOCwxMTMsMTkxLDIxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyOCwxOTQsMTQyLDMzLDQzLDg3LDIxLDIzNCwxOSwxOTEsMTYzLDIxMiwyMTcsMjUsNDksMTk5LDIwMywxNzIsMjUsNywxNjEsMTM2LDE2MywzMyw3OSwxODcsNDQsNzEsMTAxLDI5LDE4NSwyNTNdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzE0MywyMTEsMjI2LDExNiwyNTMsNjksMjQ2LDIyNCwxMTAsMTg0LDMwLDE1Nyw4NCwyMDcsMTQyLDI1MywxMjIsNTAsMTk0LDgsMjAzLDExOSw3NSwxODMsMjUsNjUsMTU1LDIxMywxODYsMTg0LDEyNSwxMF0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiU0RCRkFpRUE5WS9XeDZvMDh4QjAzZkdya3EyVGQ5NXV5akxiK0ZTRk13cHpWcHdkZTFNQ0lGcWJzdWlWeis5Wnhka05YWmZXQ1p5WHZMdUJrK3Y1KzZzYk1kbGUwSVkvQVNFRHp5QVRUMVpJNHZieGQ2WlZLeVc2bCtSYkZJVk9UeE1xU3ZEK2Zhay94UHc9IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFpeFNUVXBQYkN0b00wNUZibEkwTm5SaE5GSkVkQ3QwVEhObVdIRktiamswYUVaSWVDOWtUbloyU1ZCUlBRPT0ifSx7IlZhbHVlIjo2MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjoyMjQzNzIsIlBrU2NyaXB0IjoiZHFrVWd2eTZsUWkrRWlReTk3dVEybDkwQVVDbVc0aUlyQT09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6WzE3MiwyMzYsMTY4LDEwNSwxMzQsMzMsMTM1LDEzMiwxMiwyMjUsMTIzLDIxMiwzOSwyNDUsMTUsMTkzLDE0NiwxMjMsMTA1LDExMiwzNiwxODAsMTgyLDEwNSw0MiwyMDcsMTE1LDIxOCwwLDAsMCwwXX0="

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

	btcMultiSigAddress := s.portalParams.GeneralMultiSigAddresses[portalcommonv4.PortalBTCIDStr]
	utxos := map[string]map[string]*statedb.UTXO{
		portalcommonv4.PortalBTCIDStr: {
			statedb.GenerateUTXOObjectKey(portalcommonv4.PortalBTCIDStr, btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 0).String(): statedb.NewUTXOWithValue(btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 100000),
		},
	}

	processUnshield1 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID1,
		[]string{"txid1", "txid2", "txid3"},
		map[string][]*statedb.UTXO{
			btcMultiSigAddress: {
				statedb.NewUTXOWithValue(btcMultiSigAddress, "3c302a08da42a69d36f81f0ca0cd4f20e50e6e042292200f98a358c1f2536134", 2, 211872),
			},
		},
		map[uint64]uint{
			900: 900,
		},
	)

	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		map[string][]*statedb.UTXO{
			btcMultiSigAddress: {
				statedb.NewUTXOWithValue(btcMultiSigAddress, "d0450fc4ab5c1a0184c08b15b2b24852757d17a5e183c8e096ed1deb183dea28", 2, 201572),
			},
		},
		map[uint64]uint{
			1000: 1000,
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portalcommonv4.PortalBTCIDStr: {
			keyBatchShield1: processUnshield1,
			keyBatchShield2: processUnshield2,
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
		outputs := make(map[string]uint64, 0)
		for _, v := range tc.outputs {
			outputs[v.externalAddress] = v.amount
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
			Type: metadata.PortalV4SubmitConfirmedTxMeta,
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
	return []string{strconv.Itoa(metadata.PortalV4SubmitConfirmedTxMeta), actionContentBase64Str}
}

func buildExpectedResultSubmitConfirmedTx(s *PortalTestSuiteV4) ([]TestCaseSubmitConfirmedTx, *ExpectedResultSubmitConfirmedTx) {

	testcases := []TestCaseSubmitConfirmedTx{
		// request submit external confirmed tx
		{
			batchID:          BatchID1,
			confirmedTxProof: confirmedTxProof1,
			tokenID:          portalcommonv4.PortalBTCIDStr,
			outputs: []OutPut{
				{
					externalAddress: "msTYtu7nsMiwFUtNgCSQBk26JeBf9q3GTM",
					amount:          300,
				},
			},
		},
		// submit existed proof
		{
			batchID:          BatchID1,
			confirmedTxProof: confirmedTxProof1,
			tokenID:          portalcommonv4.PortalBTCIDStr,
			outputs: []OutPut{
				{
					externalAddress: "msTYtu7nsMiwFUtNgCSQBk26JeBf9q3GTM",
					amount:          300,
				},
			},
		},
		// request submit proof with non-exist batchID
		{
			batchID:          BatchID3,
			confirmedTxProof: confirmedTxProof2,
			tokenID:          portalcommonv4.PortalBTCIDStr,
			outputs: []OutPut{
				{
					externalAddress: "msTYtu7nsMiwFUtNgCSQBk26JeBf9q3GTM",
					amount:          400,
				},
			},
		},
		// request submit wrong proof
		{
			batchID:          BatchID2,
			confirmedTxProof: confirmedTxProof3,
			tokenID:          portalcommonv4.PortalBTCIDStr,
			outputs: []OutPut{
				{
					externalAddress: "msTYtu7nsMiwFUtNgCSQBk26JeBf9q3GTM",
					amount:          400,
				},
			},
		},
	}

	btcMultiSigAddress := s.portalParams.GeneralMultiSigAddresses[portalcommonv4.PortalBTCIDStr]
	processUnshield2 := statedb.NewProcessedUnshieldRequestBatchWithValue(
		BatchID2,
		[]string{"txid4", "txid5"},
		[]*statedb.UTXO{
			statedb.NewUTXOWithValue(btcMultiSigAddress, "d0450fc4ab5c1a0184c08b15b2b24852757d17a5e183c8e096ed1deb183dea28", 2, 201572, btcMultiSigAddress),
		},
		map[uint64]uint{
			1000: 1000,
		},
	)

	processedUnshieldRequests := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{
		portalcommonv4.PortalBTCIDStr: {
			keyBatchShield2: processUnshield2,
		},
	}

	utxos := map[string]map[string]*statedb.UTXO{
		portalcommonv4.PortalBTCIDStr: {
			statedb.GenerateUTXOObjectKey(portalcommonv4.PortalBTCIDStr, btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 0).String(): statedb.NewUTXOWithValue(btcMultiSigAddress, "7a4734c33040cc93794722b29c75020a9a8364cb294a525704f33712acbb41aa", 1, 100000),
			statedb.GenerateUTXOObjectKey(portalcommonv4.PortalBTCIDStr, btcMultiSigAddress, "d0450fc4ab5c1a0184c08b15b2b24852757d17a5e183c8e096ed1deb183dea28", 1).String(): statedb.NewUTXOWithValue(btcMultiSigAddress, "d0450fc4ab5c1a0184c08b15b2b24852757d17a5e183c8e096ed1deb183dea28", 1, 300),
		},
	}

	// build expected results
	expectedRes := &ExpectedResultSubmitConfirmedTx{
		processedUnshieldRequests: processedUnshieldRequests,
		numBeaconInsts:            4,
		statusInsts: []string{
			portalcommonv4.PortalV4RequestAcceptedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
			portalcommonv4.PortalV4RequestRejectedChainStatus,
		},
		utxos: utxos,
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV4) TestSubmitConfirmedTx() {
	fmt.Println("Running TestSubmitConfirmedTx - beacon height 1501 ...")
	networkName := "test3"
	genesisBlockHeight := 1938974
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
			)
		}
		newReqs[key] = newBatch
	}
	return newReqs
}

func TestPortalSuiteV4(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV4))
}
