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

func TestPortalSuiteV4(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV4))
}
