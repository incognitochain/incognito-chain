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
		DefaultFeeUnshields: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 100000, // in nano pBTC - 10000 satoshi ~ 4 usd
		},
		MinUnshieldAmts: map[string]uint64{
			portalcommonv4.PortalBTCIDStr: 1000000, // in nano pBTC - 1000000 satoshi ~ 4 usd
		},
		BatchNumBlks:               45,
		MinConfirmationIncBlockNum: 3,
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
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{1}).String(),
			isExistsInPreviousBlocks: false,
		},
		// valid shielding request
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzIwOCw0Niw2OCwxOSw1NSwyMDksMjM1LDIyMCw0OSwxMTEsNjQsMjEsODEsMTA5LDI0NiwyMTAsMTUyLDEzMywxMDksNjgsNjMsMTU5LDI0MSwxNjUsMTk4LDE2LDUsMjYsMjA0LDIzNCw3NCwxODhdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTM2LDQzLDIyNCwxLDE3NiwxMTQsMTQxLDIxMywzMywxMTcsNjAsNzYsNjcsMzgsMjAsNDksMTE4LDE5NSwyNTMsMjMyLDE1MCw4MiwxNDksMTY1LDE2OCwxNDIsMjA3LDI1NSwxNiw1NCw3Miw1MF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTgwLDM2LDE4LDgyLDIxMywzOSwxMDksMTc1LDIwNiwxMjgsMjUwLDYsMjM4LDM2LDE2MiwyMTAsMjMyLDEzNCwxNDYsMTI0LDksNTgsMTA0LDEzNSwxNDgsMTI5LDE4OCwxNDIsMjM5LDE5LDE4MiwzXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2NywxMDUsMTE3LDI4LDg3LDExOCwxMSwxNCwxNzgsMTE0LDk4LDExOCwxNDcsMTA3LDEwNyw5NSw0MywyMzEsNTEsMjEsMTYwLDQwLDk1LDEwLDIyNSwyNTUsMTQ5LDIzMiwyMjEsMjM1LDI0OCwzMF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyNDgsMjQ0LDE3MCw0NywzMSwxMTIsMTE4LDI0MSw0OCwxOTIsMzAsMTc2LDE1MSw4LDQ5LDYsNDIsMTE0LDE1NSwxMjIsMjEwLDIxMSw4NSwyMTQsODQsNDgsMjQ0LDE4MCw4NSw2NCwyNCw4M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTI3LDEyNCw5MywyMjEsMjQ4LDE3MywxOTQsMTYsMTU3LDUwLDYwLDE4MCwyNDAsMTMxLDQzLDExNCwxNDQsMTI4LDIwMSw0NSwxNjEsMjAsMjIxLDY3LDgwLDk4LDE4LDExMSwyNTIsMjE3LDMyLDU1XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsyMDgsNDYsNjgsMTksNTUsMjA5LDIzNSwyMjAsNDksMTExLDY0LDIxLDgxLDEwOSwyNDYsMjEwLDE1MiwxMzMsMTA5LDY4LDYzLDE1OSwyNDEsMTY1LDE5OCwxNiw1LDI2LDIwNCwyMzQsNzQsMTg4XSwiSW5kZXgiOjJ9LCJTaWduYXR1cmVTY3JpcHQiOiJTREJGQWlFQXhzOWRTMlE4YWVrMHkvRjJmOUdkejB5R0VFREVzWjNYQXpKZVFYamRiK1VDSUNlZmwwREZWV0tPZUk2Rm9FQzRsS01jWE5hRnFlL0o4ZFpTQnZNUnQyRkRBU0VEenlBVFQxWkk0dmJ4ZDZaVkt5VzZsK1JiRklWT1R4TXFTdkQrZmFrL3hQdz0iLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjAsIlBrU2NyaXB0IjoiYWt4clVGTXhMVEV5VXpWTWNuTXhXR1ZSVEdKeFRqUjVVM2xMZEdwQmFtUXlaRGR6UWxBeWRHcEdhV3A2YlhBMllYWnljbXRSUTA1R1RYQnJXRzB6UmxCNmFqSlhZM1V5V2s1eFNrVnRhRGxLY21sV2RWSkZjbFozYUhWUmJreHRWMU5oWjJkdllrVlhjMEpGWTJrPSJ9LHsiVmFsdWUiOjgwMCwiUGtTY3JpcHQiOiJxUlFuSjZkdjh2bzVYY1VsWktqcktxdU0vbEhJZG9jPSJ9LHsiVmFsdWUiOjc2MjcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOls2NCw2NywyNDMsMjIyLDI0NCw5LDE4MywyMzQsMjM4LDE3MCwxNjYsMjI2LDEzLDI1NCwzNSw3OCwyMjIsNjksMjQxLDIyNCwxMDMsNjMsMTI5LDIwNCwxNCwwLDAsMCwwLDAsMCwwXX0=",
			txID:                     common.HashH([]byte{2}).String(),
			isExistsInPreviousBlocks: false,
		},
		// invalid shielding request: duplicated shielding proof in previous blocks
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: true,
		},
		// invalid shielding request: duplicated shielding proof in the current block
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE2NSwxMzIsNTAsNTEsNzgsMzgsMTk4LDk3LDE0NSwxOTAsMTUzLDUwLDIzNCwxNDgsMTUzLDgsMjQwLDE1NywyLDIwLDg5LDExMCwxNTQsMTM0LDE1NSwyMzcsNDksMjM0LDIwNSwxMywyMzQsNzJdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzNiw0MywyMjQsMSwxNzYsMTE0LDE0MSwyMTMsMzMsMTE3LDYwLDc2LDY3LDM4LDIwLDQ5LDExOCwxOTUsMjUzLDIzMiwxNTAsODIsMTQ5LDE2NSwxNjgsMTQyLDIwNywyNTUsMTYsNTQsNzIsNTBdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE4MCwzNiwxOCw4MiwyMTMsMzksMTA5LDE3NSwyMDYsMTI4LDI1MCw2LDIzOCwzNiwxNjIsMjEwLDIzMiwxMzQsMTQ2LDEyNCw5LDU4LDEwNCwxMzUsMTQ4LDEyOSwxODgsMTQyLDIzOSwxOSwxODIsM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNjcsMTA1LDExNywyOCw4NywxMTgsMTEsMTQsMTc4LDExNCw5OCwxMTgsMTQ3LDEwNywxMDcsOTUsNDMsMjMxLDUxLDIxLDE2MCw0MCw5NSwxMCwyMjUsMjU1LDE0OSwyMzIsMjIxLDIzNSwyNDgsMzBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQ4LDI0NCwxNzAsNDcsMzEsMTEyLDExOCwyNDEsNDgsMTkyLDMwLDE3NiwxNTEsOCw0OSw2LDQyLDExNCwxNTUsMTIyLDIxMCwyMTEsODUsMjE0LDg0LDQ4LDI0NCwxODAsODUsNjQsMjQsODNdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEyNywxMjQsOTMsMjIxLDI0OCwxNzMsMTk0LDE2LDE1Nyw1MCw2MCwxODAsMjQwLDEzMSw0MywxMTQsMTQ0LDEyOCwyMDEsNDUsMTYxLDIwLDIyMSw2Nyw4MCw5OCwxOCwxMTEsMjUyLDIxNywzMiw1NV0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbNTgsMTMyLDE1MSw4Nyw5OSw5LDI0OCw0OSwxOCw5NSw3OSw3MCwxMzcsNzYsNCwyMTUsMTgyLDUwLDQ5LDk0LDE2LDE4NCwxNzMsMzUsNDAsMTU4LDcwLDMwLDE3MywyMzEsMTc0LDEzNl0sIkluZGV4IjoyfSwiU2lnbmF0dXJlU2NyaXB0IjoiUnpCRUFpQmtiQmM1VThoRjdLelkxWWZaYXgvWDJBRHVDY3FreE1mczFkTmNoMDEybFFJZ1VKTHAvaUlPd0w5R0NnSk5wZE55d0UzajV6Wjg2ZkNINkt5WlkyNjFsNXdCSVFQUElCTlBWa2ppOXZGM3BsVXJKYnFYNUZzVWhVNVBFeXBLOFA1OXFUL0UvQT09IiwiV2l0bmVzcyI6bnVsbCwiU2VxdWVuY2UiOjQyOTQ5NjcyOTV9XSwiVHhPdXQiOlt7IlZhbHVlIjowLCJQa1NjcmlwdCI6ImFreHJVRk14TFRFeVV6Vk1jbk14V0dWUlRHSnhUalI1VTNsTGRHcEJhbVF5WkRkelFsQXlkR3BHYVdwNmJYQTJZWFp5Y210UlEwNUdUWEJyV0cwelJsQjZhakpYWTNVeVdrNXhTa1Z0YURsS2NtbFdkVkpGY2xaM2FIVlJia3h0VjFOaFoyZHZZa1ZYYzBKRlkyaz0ifSx7IlZhbHVlIjo0MDAsIlBrU2NyaXB0IjoicVJRbko2ZHY4dm81WGNVbFpLanJLcXVNL2xISWRvYz0ifSx7IlZhbHVlIjo4MjA3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbNjQsNjcsMjQzLDIyMiwyNDQsOSwxODMsMjM0LDIzOCwxNzAsMTY2LDIyNiwxMywyNTQsMzUsNzgsMjIyLDY5LDI0MSwyMjQsMTAzLDYzLDEyOSwyMDQsMTQsMCwwLDAsMCwwLDAsMF19",
			txID:                     common.HashH([]byte{3}).String(),
			isExistsInPreviousBlocks: false,
		},
		// invalid shielding request: invalid proof (invalid memo)
		{
			tokenID:                  portalcommonv4.PortalBTCIDStr,
			incAddressStr:            "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			shieldingProof:           "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1OSw4LDIxMSwyNTQsMTY2LDk1LDYzLDUwLDEyMCw2Miw5MCw5Niw4NywyMzEsMjAyLDI5LDI0MCw1OCwyMTcsMTIyLDgxLDEzNiwxMzAsNTEsMTQ0LDIyNiwxNDYsNjQsNjcsMjA1LDI2LDE1N10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNDcsOTEsMSwxMjQsMTY2LDE3NSwyNDQsMywxMCw5NCw0LDg4LDg5LDQ2LDY1LDEzMCwxNjYsMjUwLDUsMTE3LDIxNCw2LDEyNiw3Niw5LDE1MiwxMzcsMTgyLDE3OSwxODksMTQ0LDY5XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzIzMyw0NCw1MCwxMDEsMTU0LDIyOSwxNTQsMjUsMjA4LDI0MCwxNjUsMTkxLDE5Nyw3LDE3OSw1OCw3NiwxOTYsNDQsMTczLDk0LDgzLDEyNiwxMTQsNDYsNDIsMTk4LDE1NCwzMiwyMTAsOTMsNDRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjQxLDE4MiwyMzQsMTMwLDE1OCw4NCw1LDY0LDI0MiwxNjMsMjU1LDEwMywyMDAsODEsMTEwLDIyNiwyMjgsMTg2LDI1MiwyMzAsMTI3LDEwOCwxODAsNDcsMTE3LDE3MSw0OCwxMjEsNjgsMjM5LDE3NywyNl0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTY0LDEwLDIzNSwyNTIsMjI4LDE1LDQ0LDM1LDEzNCw3OSwxMzUsNjEsOTAsMjUzLDI1NCwxMzYsMTY1LDEzNCw0NiwxOTUsMTM0LDE1NSwxNzMsMTAxLDc0LDIzMiwzNiwxMTUsMSwxODYsMTA3LDIyOV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlszNCw5NSwxNDksODAsMTE0LDIzNiwyNDYsNzUsMjIyLDE1OCw3MCwyNDQsMTc2LDIyNiwxNzksMjM5LDgyLDE3MCwxMTYsODcsMjMzLDE3MywyMjYsMTcyLDI0NSw3MywyMDksMTc3LDQ0LDI1MCwzNCw2NF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbNzgsOTYsMTQxLDIwOCw3OCwxNzEsMTU0LDkzLDcsMTE0LDM1LDExNSwxNjksMjA3LDE0MSwyMDYsMjMxLDE4MSwxNzYsMjQ1LDQ1LDg2LDIwNSw3MiwyNDAsNzMsNTUsMjIsMTkxLDk3LDE5OCwxNzRdLCJJc0xlZnQiOmZhbHNlfV0sIkJUQ1R4Ijp7IlZlcnNpb24iOjEsIlR4SW4iOlt7IlByZXZpb3VzT3V0UG9pbnQiOnsiSGFzaCI6WzE0Miw2NywyNDIsMTAxLDE4MCwxOCw2LDE5OCwyMjIsMTg5LDczLDE0OCwxMTUsMzEsNzAsMjQ4LDE2MiwyNCwyMDcsMzUsMTE4LDM5LDI4LDIxMCwxOTEsODAsMTMxLDE2NSw0Niw4OSwxODksODVdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlJ6QkVBaUJBNUdFdlRlcTN5T2JDNmhtT1k1Qld4T3pBM0VJcEFka3Nrd3BKb0FUb3lRSWdNdVpiWldxSElITlY4WmU2QXRISHE2L3N1eFdTdlpJY1A0cEZTUUtlOGYwQklRUFBJQk5QVmtqaTl2RjNwbFVySmJxWDVGc1VoVTVQRXlwSzhQNTlxVC9FL0E9PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJha3hyVUZNeExURXlVelZNY25NeFdHVlJUR0p4VFRSNVUzbExkR3BCYW1ReVpEZHpRbEF5ZEdwR2FXcDZiWEEyWVhaeWNtdFJRMDVHVFhCcldHMHpSbEI2YWpKWFkzVXlXazV4U2tWdGFEbEtjbWxXZFZKRmNsWjNhSFZSYmt4dFYxTmhaMmR2WWtWWGMwSkZZMms9In0seyJWYWx1ZSI6ODAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6NzI2NzIsIlBrU2NyaXB0IjoiZHFrVWd2eTZsUWkrRWlReTk3dVEybDkwQVVDbVc0aUlyQT09In1dLCJMb2NrVGltZSI6MH0sIkJsb2NrSGFzaCI6WzE0NCwzMSwxODMsMTQ3LDIwNiwyMywxMTEsMTc5LDkyLDI1MSwxODYsMTQ0LDc2LDIxNiwxNzgsNjUsMSwxMjYsMTc2LDEzNSwzNywxNjEsMTg5LDExLDE3LDAsMCwwLDAsMCwwLDBdfQ==",
			txID:                     common.HashH([]byte{4}).String(),
			isExistsInPreviousBlocks: false,
		},
	}

	walletAddress := "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X"

	// build expected results
	var txHash string
	var outputIdx uint32
	var outputAmount uint64

	txHash = "bc4aeacc1a0510c6a5f19f3f446d8598d2f66d5115406f31dcebd13713442ed0"
	outputIdx = 1
	outputAmount = 400

	key1, value1 := generateUTXOKeyAndValue(portalcommonv4.PortalBTCIDStr, walletAddress, txHash, outputIdx, outputAmount)

	txHash = "48ea0dcdea31ed9b869a6e5914029df0089994ea3299be9161c6264e333284a5"
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
			incAddressStr:  USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{1}).String(),
			isExisted:      false,
		},
		// valid unshield request
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 0.5 * 1e9,
			incAddressStr:  USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_2,
			txId:           common.HashH([]byte{2}).String(),
			isExisted:      false,
		},
		// invalid unshield request - invalid unshield amount
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 999999,
			incAddressStr:  USER_INC_ADDRESS_1,
			remoteAddress:  USER_BTC_ADDRESS_1,
			txId:           common.HashH([]byte{3}).String(),
			isExisted:      false,
		},
		// invalid unshield request - existed unshield ID
		{
			tokenID:        portalcommonv4.PortalBTCIDStr,
			unshieldAmount: 1 * 1e9,
			incAddressStr:  USER_INC_ADDRESS_1,
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
	bc.On("GetPortalV4MultiSigAddress", tokenID, bcH).Return(s.portalParams.MultiSigAddresses[portalcommonv4.PortalBTCIDStr])

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

func TestPortalSuiteV4(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV4))
}
