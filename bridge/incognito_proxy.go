// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bridge

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BridgeABI is the input ABI used to generate the binding from.
const BridgeABI = "[{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBool\",\"inputs\":[{\"type\":\"bool\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"_beaconCommRoot\"},{\"type\":\"bytes32\",\"name\":\"_bridgeCommRoot\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32[2]\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":468},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[1]\",\"name\":\"path\"},{\"type\":\"bool[1]\",\"name\":\"left\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":913},{\"name\":\"verifyInst\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"commRoot\"},{\"type\":\"bytes32\",\"name\":\"instHash\"},{\"type\":\"bytes32[1]\",\"name\":\"instPath\"},{\"type\":\"bool[1]\",\"name\":\"instPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"instRoot\"},{\"type\":\"bytes32\",\"name\":\"blkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"signerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"signerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"signerPaths\"},{\"type\":\"bool[2]\",\"name\":\"signerPathIsLeft\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":55301},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"newCommRoot\"},{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[1]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[1]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[2]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"bytes32[1]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[1]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[2]\",\"name\":\"bridgeSignerPathIsLeft\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":228835},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":603},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":633}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05260406115af6101403934156100a157600080fd5b610140516000556101605160015561159756600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526357bff27860005114156100dc57602060046101403734156100b457600080fd5b60846004356004016101603760646004356004013511156100d457600080fd5b6040610220f3005b639cba167e600051141561021457608060046101403734156100fd57600080fd5b6064356002811061010d57600080fd5b50610140516101c0526101e060006001818352015b6101a06101e0516001811061013657600080fd5b6020020151156101995760006101806101e0516001811061015657600080fd5b60200201516020826102800101526020810190506101c051602082610280010152602081019050806102805261028090508051602082012090506101c0526101ee565b60006101c0516020826102000101526020810190506101806101e051600181106101c257600080fd5b6020020151602082610200010152602081019050806102005261020090508051602082012090506101c0525b5b8151600101808352811415610122575b5050610160516101c0511460005260206000f3005b63c4f5463460005114156108ce576101a0600461014037341561023657600080fd5b6064356002811061024657600080fd5b50610164356002811061025857600080fd5b50610184356002811061026a57600080fd5b5060206103c06084639cba167e6102e05261016051610300526101c051610320526103406101808060006020020151826000602002015250506103606101a08060006020020151826000602002015250506102fc6000305af16102cc57600080fd5b6103c05115156104375760216103e0527f696e737472756374696f6e206973206e6f7420696e206d65726b6c6520747265610400527f6500000000000000000000000000000000000000000000000000000000000000610420526103e0805160200180610460828460006004600a8704601201f161034957600080fd5b505060206105005261050051610540526104608051602001806105005161054001828460006004600a8704601201f161038157600080fd5b505061050051610540015160206001820306601f820103905061050051610540016104e081516040818352015b836104e0511015156103bf576103dc565b60006104e0516020850101535b81516001018083528114156103ae575b50505050602061050051610540015160206001820306601f8201039050610500510101610500527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961050051610540a1600060005260206000f35b60006105605261058060006002818352015b610200610580516002811061045d57600080fd5b6020020151151561046d57610768565b60016105e05261060060006001818352015b610260606051610600516060516105e0516105805102806040519013156104a557600080fd5b80919012156104b357600080fd5b01806040519013156104c457600080fd5b80919012156104d257600080fd5b600281106104df57600080fd5b60200201516105a061060051600181106104f857600080fd5b60200201526102a0606051610600516060516105e05161058051028060405190131561052357600080fd5b809190121561053157600080fd5b018060405190131561054257600080fd5b809190121561055057600080fd5b6002811061055d57600080fd5b60200201516105c0610600516001811061057657600080fd5b60200201525b815160010180835281141561047f575b505060206107006084639cba167e6106205261020061058051600281106105b257600080fd5b60200201516106405261014051610660526106806105a08060006020020151826000602002015250506106a06105c080600060200201518260006020020152505061063c6000305af161060457600080fd5b61070051151561074a576019610720527f7075626b6579206e6f7420696e206d65726b6c6520747265650000000000000061074052610720805160200180610780828460006004600a8704601201f161065c57600080fd5b505060206108005261080051610840526107808051602001806108005161084001828460006004600a8704601201f161069457600080fd5b505061080051610840015160206001820306601f820103905061080051610840016107e081516020818352015b836107e0511015156106d2576106ef565b60006107e0516020850101535b81516001018083528114156106c1575b50505050602061080051610840015160206001820306601f8201039050610800510101610800527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961080051610840a1600060005260206000f35b61056080516001825101101561075f57600080fd5b60018151018152505b8151600101808352811415610449575b505060026105605110156108c257600e610860527f6e6f7420656e6f75676820736967000000000000000000000000000000000000610880526108608051602001806108c0828460006004600a8704601201f16107d457600080fd5b505060206109405261094051610980526108c08051602001806109405161098001828460006004600a8704601201f161080c57600080fd5b505061094051610980015160206001820306601f8201039050610940516109800161092081516020818352015b836109205110151561084a57610867565b6000610920516020850101535b8151600101808352811415610839575b50505050602061094051610980015160206001820306601f8201039050610940510101610940527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961094051610980a1600060005260206000f35b600160005260206000f3005b630e9e539d60005114156114915761034060046101403734156108f057600080fd5b608460243560040161048037606460243560040135111561091057600080fd5b6064356002811061092057600080fd5b50610184356002811061093257600080fd5b506101a4356002811061094457600080fd5b506101e4356002811061095657600080fd5b50610304356002811061096857600080fd5b50610324356002811061097a57600080fd5b506104808051602082012090506105405260006101c0516020826105800101526020810190506101e05160208261058001015260208101905080610580526105809050805160208201209050610560526102005161056051141515610bfe57602e610600527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069610620527f6e20626561636f6e20626c6f636b00000000000000000000000000000000000061064052610600805160200180610680828460006004600a8704601201f1610a4c57600080fd5b505060206107205261072051610760526106808051602001806107205161076001828460006004600a8704601201f1610a8457600080fd5b505061072051610760015160206001820306601f8201039050610720516107600161070081516040818352015b8361070051101515610ac257610adf565b6000610700516020850101535b8151600101808352811415610ab1575b50505050602061072051610760015160206001820306601f8201039050610720510101610720527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961072051610760a161054051610780527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020610780a1610560516107a0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206107a0a16308c379a06107c05260206107e052602e610800527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069610820527f6e20626561636f6e20626c6f636b00000000000000000000000000000000000061084052610800506000610bfd5760a46107dcfd5b5b6020610a806101a463c4f54634610880526000546108a052610540516108c0526108e06101808060006020020151826000602002015250506109006101a08060006020020151826000602002015250506101c05161092052610200516109405261096061022080600060200201518260006020020152806001602002015182600160200201525050610260516109a0526109c061028080600060200201518260006020020152806001602002015182600160200201525050610a006102c08060006020020151826000602002015280600160200201518260016020020152505061089c6000305af1610cef57600080fd5b610a80511515610ec0576023610aa0527f6661696c656420766572696679696e6720626561636f6e20696e737472756374610ac0527f696f6e0000000000000000000000000000000000000000000000000000000000610ae052610aa0805160200180610b20828460006004600a8704601201f1610d6c57600080fd5b50506020610bc052610bc051610c0052610b20805160200180610bc051610c0001828460006004600a8704601201f1610da457600080fd5b5050610bc051610c00015160206001820306601f8201039050610bc051610c0001610ba081516040818352015b83610ba051101515610de257610dff565b6000610ba0516020850101535b8151600101808352811415610dd1575b505050506020610bc051610c00015160206001820306601f8201039050610bc0510101610bc0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610bc051610c00a16308c379a0610c20526020610c40526023610c60527f6661696c656420766572696679696e6720626561636f6e20696e737472756374610c80527f696f6e0000000000000000000000000000000000000000000000000000000000610ca052610c60506000610ebf5760a4610c3cfd5b5b600061034051602082610ce001015260208101905061036051602082610ce001015260208101905080610ce052610ce090508051602082012090506105605261038051610560511415156110d557602e610d60527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069610d80527f6e2062726964676520626c6f636b000000000000000000000000000000000000610da052610d60805160200180610de0828460006004600a8704601201f1610f8157600080fd5b50506020610e8052610e8051610ec052610de0805160200180610e8051610ec001828460006004600a8704601201f1610fb957600080fd5b5050610e8051610ec0015160206001820306601f8201039050610e8051610ec001610e6081516040818352015b83610e6051101515610ff757611014565b6000610e60516020850101535b8151600101808352811415610fe6575b505050506020610e8051610ec0015160206001820306601f8201039050610e80510101610e80527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610e8051610ec0a16308c379a0610ee0526020610f0052602e610f20527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069610f40527f6e2062726964676520626c6f636b000000000000000000000000000000000000610f6052610f205060006110d45760a4610efcfd5b5b60206111a06101a463c4f54634610fa052600154610fc05261054051610fe052611000610300806000602002015182600060200201525050611020610320806000602002015182600060200201525050610340516110405261038051611060526110806103a0806000602002015182600060200201528060016020020151826001602002015250506103e0516110c0526110e06104008060006020020151826000602002015280600160200201518260016020020152505061112061044080600060200201518260006020020152806001602002015182600160200201525050610fbc6000305af16111c657600080fd5b6111a051151561134d5760206111c0527f6661696c6564207665726966792062726964676520696e737472756374696f6e6111e0526111c0805160200180611220828460006004600a8704601201f161121e57600080fd5b505060206112a0526112a0516112e0526112208051602001806112a0516112e001828460006004600a8704601201f161125657600080fd5b50506112a0516112e0015160206001820306601f82010390506112a0516112e00161128081516020818352015b8361128051101515611294576112b1565b6000611280516020850101535b8151600101808352811415611283575b5050505060206112a0516112e0015160206001820306601f82010390506112a05101016112a0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96112a0516112e0a16308c379a0611300526020611320526020611340527f6661696c6564207665726966792062726964676520696e737472756374696f6e6113605261134050600061134c57608461131cfd5b5b6101405160005560106113a0527f6e6f2065786563657074696f6e2e2e2e000000000000000000000000000000006113c0526113a0805160200180611400828460006004600a8704601201f16113a257600080fd5b5050602061148052611480516114c052611400805160200180611480516114c001828460006004600a8704601201f16113da57600080fd5b5050611480516114c0015160206001820306601f8201039050611480516114c00161146081516020818352015b836114605110151561141857611435565b6000611460516020850101535b8151600101808352811415611407575b505050506020611480516114c0015160206001820306601f8201039050611480510101611480527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9611480516114c0a1600160005260206000f3005b63c4ed3f0860005114156114b75734156114aa57600080fd5b60005460005260206000f3005b630f7b9ca160005114156114dd5734156114d057600080fd5b60015460005260206000f3005b60006000fd5b6100b4611597036100b46000396100b4611597036000f3`

// DeployBridge deploys a new Ethereum contract, binding an instance of Bridge to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend, _beaconCommRoot [32]byte, _bridgeCommRoot [32]byte) (common.Address, *types.Transaction, *Bridge, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BridgeBin), backend, _beaconCommRoot, _bridgeCommRoot)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// BeaconCommRoot is a free data retrieval call binding the contract method 0xc4ed3f08.
//
// Solidity: function beaconCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeCaller) BeaconCommRoot(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "beaconCommRoot")
	return *ret0, err
}

// BeaconCommRoot is a free data retrieval call binding the contract method 0xc4ed3f08.
//
// Solidity: function beaconCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeSession) BeaconCommRoot() ([32]byte, error) {
	return _Bridge.Contract.BeaconCommRoot(&_Bridge.CallOpts)
}

// BeaconCommRoot is a free data retrieval call binding the contract method 0xc4ed3f08.
//
// Solidity: function beaconCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeCallerSession) BeaconCommRoot() ([32]byte, error) {
	return _Bridge.Contract.BeaconCommRoot(&_Bridge.CallOpts)
}

// BridgeCommRoot is a free data retrieval call binding the contract method 0x0f7b9ca1.
//
// Solidity: function bridgeCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeCaller) BridgeCommRoot(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "bridgeCommRoot")
	return *ret0, err
}

// BridgeCommRoot is a free data retrieval call binding the contract method 0x0f7b9ca1.
//
// Solidity: function bridgeCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeSession) BridgeCommRoot() ([32]byte, error) {
	return _Bridge.Contract.BridgeCommRoot(&_Bridge.CallOpts)
}

// BridgeCommRoot is a free data retrieval call binding the contract method 0x0f7b9ca1.
//
// Solidity: function bridgeCommRoot() constant returns(bytes32 out)
func (_Bridge *BridgeCallerSession) BridgeCommRoot() ([32]byte, error) {
	return _Bridge.Contract.BridgeCommRoot(&_Bridge.CallOpts)
}

// InMerkleTree is a free data retrieval call binding the contract method 0x9cba167e.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[1] path, bool[1] left) constant returns(bool out)
func (_Bridge *BridgeCaller) InMerkleTree(opts *bind.CallOpts, leaf [32]byte, root [32]byte, path [1][32]byte, left [1]bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "inMerkleTree", leaf, root, path, left)
	return *ret0, err
}

// InMerkleTree is a free data retrieval call binding the contract method 0x9cba167e.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[1] path, bool[1] left) constant returns(bool out)
func (_Bridge *BridgeSession) InMerkleTree(leaf [32]byte, root [32]byte, path [1][32]byte, left [1]bool) (bool, error) {
	return _Bridge.Contract.InMerkleTree(&_Bridge.CallOpts, leaf, root, path, left)
}

// InMerkleTree is a free data retrieval call binding the contract method 0x9cba167e.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[1] path, bool[1] left) constant returns(bool out)
func (_Bridge *BridgeCallerSession) InMerkleTree(leaf [32]byte, root [32]byte, path [1][32]byte, left [1]bool) (bool, error) {
	return _Bridge.Contract.InMerkleTree(&_Bridge.CallOpts, leaf, root, path, left)
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[2] out)
func (_Bridge *BridgeCaller) ParseSwapBeaconInst(opts *bind.CallOpts, inst []byte) ([2][32]byte, error) {
	var (
		ret0 = new([2][32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "parseSwapBeaconInst", inst)
	return *ret0, err
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[2] out)
func (_Bridge *BridgeSession) ParseSwapBeaconInst(inst []byte) ([2][32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[2] out)
func (_Bridge *BridgeCallerSession) ParseSwapBeaconInst(inst []byte) ([2][32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// VerifyInst is a free data retrieval call binding the contract method 0xc4f54634.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[1] instPath, bool[1] instPathIsLeft, bytes32 instRoot, bytes32 blkHash, bytes32[2] signerPubkeys, bytes32 signerSig, bytes32[2] signerPaths, bool[2] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeCaller) VerifyInst(opts *bind.CallOpts, commRoot [32]byte, instHash [32]byte, instPath [1][32]byte, instPathIsLeft [1]bool, instRoot [32]byte, blkHash [32]byte, signerPubkeys [2][32]byte, signerSig [32]byte, signerPaths [2][32]byte, signerPathIsLeft [2]bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "verifyInst", commRoot, instHash, instPath, instPathIsLeft, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
	return *ret0, err
}

// VerifyInst is a free data retrieval call binding the contract method 0xc4f54634.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[1] instPath, bool[1] instPathIsLeft, bytes32 instRoot, bytes32 blkHash, bytes32[2] signerPubkeys, bytes32 signerSig, bytes32[2] signerPaths, bool[2] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [1][32]byte, instPathIsLeft [1]bool, instRoot [32]byte, blkHash [32]byte, signerPubkeys [2][32]byte, signerSig [32]byte, signerPaths [2][32]byte, signerPathIsLeft [2]bool) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
}

// VerifyInst is a free data retrieval call binding the contract method 0xc4f54634.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[1] instPath, bool[1] instPathIsLeft, bytes32 instRoot, bytes32 blkHash, bytes32[2] signerPubkeys, bytes32 signerSig, bytes32[2] signerPaths, bool[2] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeCallerSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [1][32]byte, instPathIsLeft [1]bool, instRoot [32]byte, blkHash [32]byte, signerPubkeys [2][32]byte, signerSig [32]byte, signerPaths [2][32]byte, signerPathIsLeft [2]bool) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0x0e9e539d.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconInstPathIsLeft, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[2] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[2] beaconSignerPaths, bool[2] beaconSignerPathIsLeft, bytes32[1] bridgeInstPath, bool[1] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[2] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[2] bridgeSignerPaths, bool[2] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeTransactor) SwapBeacon(opts *bind.TransactOpts, newCommRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconInstPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [2][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [2][32]byte, beaconSignerPathIsLeft [2]bool, bridgeInstPath [1][32]byte, bridgeInstPathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [2][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [2][32]byte, bridgeSignerPathIsLeft [2]bool) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "swapBeacon", newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0x0e9e539d.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconInstPathIsLeft, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[2] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[2] beaconSignerPaths, bool[2] beaconSignerPathIsLeft, bytes32[1] bridgeInstPath, bool[1] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[2] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[2] bridgeSignerPaths, bool[2] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconInstPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [2][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [2][32]byte, beaconSignerPathIsLeft [2]bool, bridgeInstPath [1][32]byte, bridgeInstPathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [2][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [2][32]byte, bridgeSignerPathIsLeft [2]bool) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0x0e9e539d.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconInstPathIsLeft, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[2] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[2] beaconSignerPaths, bool[2] beaconSignerPathIsLeft, bytes32[1] bridgeInstPath, bool[1] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[2] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[2] bridgeSignerPaths, bool[2] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeTransactorSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconInstPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [2][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [2][32]byte, beaconSignerPathIsLeft [2]bool, bridgeInstPath [1][32]byte, bridgeInstPathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [2][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [2][32]byte, bridgeSignerPathIsLeft [2]bool) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
}

// BridgeNotifyBoolIterator is returned from FilterNotifyBool and is used to iterate over the raw logs and unpacked data for NotifyBool events raised by the Bridge contract.
type BridgeNotifyBoolIterator struct {
	Event *BridgeNotifyBool // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeNotifyBoolIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeNotifyBool)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeNotifyBool)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeNotifyBoolIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeNotifyBoolIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeNotifyBool represents a NotifyBool event raised by the Bridge contract.
type BridgeNotifyBool struct {
	Content bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyBool is a free log retrieval operation binding the contract event 0x6c8f06ff564112a969115be5f33d4a0f87ba918c9c9bc3090fe631968e818be4.
//
// Solidity: event NotifyBool(bool content)
func (_Bridge *BridgeFilterer) FilterNotifyBool(opts *bind.FilterOpts) (*BridgeNotifyBoolIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "NotifyBool")
	if err != nil {
		return nil, err
	}
	return &BridgeNotifyBoolIterator{contract: _Bridge.contract, event: "NotifyBool", logs: logs, sub: sub}, nil
}

// WatchNotifyBool is a free log subscription operation binding the contract event 0x6c8f06ff564112a969115be5f33d4a0f87ba918c9c9bc3090fe631968e818be4.
//
// Solidity: event NotifyBool(bool content)
func (_Bridge *BridgeFilterer) WatchNotifyBool(opts *bind.WatchOpts, sink chan<- *BridgeNotifyBool) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "NotifyBool")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeNotifyBool)
				if err := _Bridge.contract.UnpackLog(event, "NotifyBool", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// BridgeNotifyBytes32Iterator is returned from FilterNotifyBytes32 and is used to iterate over the raw logs and unpacked data for NotifyBytes32 events raised by the Bridge contract.
type BridgeNotifyBytes32Iterator struct {
	Event *BridgeNotifyBytes32 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeNotifyBytes32Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeNotifyBytes32)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeNotifyBytes32)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeNotifyBytes32Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeNotifyBytes32Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeNotifyBytes32 represents a NotifyBytes32 event raised by the Bridge contract.
type BridgeNotifyBytes32 struct {
	Content [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyBytes32 is a free log retrieval operation binding the contract event 0xb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b.
//
// Solidity: event NotifyBytes32(bytes32 content)
func (_Bridge *BridgeFilterer) FilterNotifyBytes32(opts *bind.FilterOpts) (*BridgeNotifyBytes32Iterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "NotifyBytes32")
	if err != nil {
		return nil, err
	}
	return &BridgeNotifyBytes32Iterator{contract: _Bridge.contract, event: "NotifyBytes32", logs: logs, sub: sub}, nil
}

// WatchNotifyBytes32 is a free log subscription operation binding the contract event 0xb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b.
//
// Solidity: event NotifyBytes32(bytes32 content)
func (_Bridge *BridgeFilterer) WatchNotifyBytes32(opts *bind.WatchOpts, sink chan<- *BridgeNotifyBytes32) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "NotifyBytes32")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeNotifyBytes32)
				if err := _Bridge.contract.UnpackLog(event, "NotifyBytes32", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// BridgeNotifyStringIterator is returned from FilterNotifyString and is used to iterate over the raw logs and unpacked data for NotifyString events raised by the Bridge contract.
type BridgeNotifyStringIterator struct {
	Event *BridgeNotifyString // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeNotifyStringIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeNotifyString)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeNotifyString)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeNotifyStringIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeNotifyStringIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeNotifyString represents a NotifyString event raised by the Bridge contract.
type BridgeNotifyString struct {
	Content string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyString is a free log retrieval operation binding the contract event 0x8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9.
//
// Solidity: event NotifyString(string content)
func (_Bridge *BridgeFilterer) FilterNotifyString(opts *bind.FilterOpts) (*BridgeNotifyStringIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "NotifyString")
	if err != nil {
		return nil, err
	}
	return &BridgeNotifyStringIterator{contract: _Bridge.contract, event: "NotifyString", logs: logs, sub: sub}, nil
}

// WatchNotifyString is a free log subscription operation binding the contract event 0x8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9.
//
// Solidity: event NotifyString(string content)
func (_Bridge *BridgeFilterer) WatchNotifyString(opts *bind.WatchOpts, sink chan<- *BridgeNotifyString) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "NotifyString")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeNotifyString)
				if err := _Bridge.contract.UnpackLog(event, "NotifyString", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
