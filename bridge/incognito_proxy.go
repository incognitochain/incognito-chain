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
const BridgeABI = "[{\"name\":\"Transfer\",\"inputs\":[{\"type\":\"address\",\"name\":\"_from\",\"indexed\":true},{\"type\":\"address\",\"name\":\"_to\",\"indexed\":true},{\"type\":\"uint256\",\"name\":\"_value\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"Approve\",\"inputs\":[{\"type\":\"address\",\"name\":\"_owner\",\"indexed\":true},{\"type\":\"address\",\"name\":\"_spender\",\"indexed\":true},{\"type\":\"uint256\",\"name\":\"_value\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"string\",\"name\":\"_name\"},{\"type\":\"string\",\"name\":\"_symbol\"},{\"type\":\"uint256\",\"name\":\"_decimals\"},{\"type\":\"uint256\",\"name\":\"_totalSupply\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"get\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":283},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32[2]\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":498},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[1]\",\"name\":\"path\"},{\"type\":\"bool[1]\",\"name\":\"left\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":943},{\"name\":\"getHash\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":615},{\"name\":\"getHash256\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":2710},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"newComRoot\"},{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[1]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[1]\",\"name\":\"beaconPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes32[1]\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[1]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bytes32[1]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[1]\",\"name\":\"bridgePathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes32[1]\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[1]\",\"name\":\"bridgeSignerPaths\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":4127},{\"name\":\"transfer\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"address\",\"name\":\"_to\"},{\"type\":\"uint256\",\"name\":\"_value\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":75475},{\"name\":\"transferFrom\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"address\",\"name\":\"_from\"},{\"type\":\"address\",\"name\":\"_to\"},{\"type\":\"uint256\",\"name\":\"_value\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":111894},{\"name\":\"approve\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"address\",\"name\":\"_spender\"},{\"type\":\"uint256\",\"name\":\"_value\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":38720},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":783},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":813},{\"name\":\"name\",\"outputs\":[{\"type\":\"string\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":3364},{\"name\":\"symbol\",\"outputs\":[{\"type\":\"string\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":3394},{\"name\":\"decimals\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":903},{\"name\":\"totalSupply\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":933},{\"name\":\"balanceOf\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"address\",\"name\":\"arg0\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":1135},{\"name\":\"allowance\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"address\",\"name\":\"arg0\"},{\"type\":\"address\",\"name\":\"arg1\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":1325}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526080610cb96101403934156100a157600080fd5b602a6020610cb960c03960c051610cb9016101c039600a6020610cb960c03960c0516004013511156100d257600080fd5b602a60206020610cb90160c03960c051610cb90161022039600a60206020610cb90160c03960c05160040135111561010957600080fd5b6101c080600260c052602060c020602082510161012060006002818352015b8261012051602002111561013b5761015d565b61012051602002850151610120518501555b8151600101808352811415610128575b50505050505061022080600360c052602060c020602082510161012060006002818352015b82610120516020021115610195576101b7565b61012051602002850151610120518501555b8151600101808352811415610182575b505050505050610180516004556101a0516005556101a05160063360e05260c052604060c02055610ca156600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a052636d4ce63c60005114156100b85734156100ac57600080fd5b600260005260206000f3005b6357bff278600051141561010157602060046101403734156100d957600080fd5b60846004356004016101603760646004356004013511156100f957600080fd5b6040610220f3005b639cba167e6000511415610239576080600461014037341561012257600080fd5b6064356002811061013257600080fd5b50610140516101c0526101e060006001818352015b6101a06101e0516001811061015b57600080fd5b6020020151156101be5760006101806101e0516001811061017b57600080fd5b60200201516020826102800101526020810190506101c051602082610280010152602081019050806102805261028090508051602082012090506101c052610213565b60006101c0516020826102000101526020810190506101806101e051600181106101e757600080fd5b6020020151602082610200010152602081019050806102005261020090508051602082012090506101c0525b5b8151600101808352811415610147575b5050610160516101c0511460005260206000f3005b63b00140aa6000511415610290576020600461014037341561025a57600080fd5b608460043560040161016037606460043560040135111561027a57600080fd5b61016080516020820120905060005260206000f3005b63e6d9c5f260005114156102fc57602060046101403734156102b157600080fd5b60846004356004016101603760646004356004013511156102d157600080fd5b610160602060c0825160208401600060025af16102ed57600080fd5b60c051905060005260206000f3005b63c02719e2600051141561047a57610240600461014037341561031e57600080fd5b608460243560040161038037606460243560040135111561033e57600080fd5b6064356002811061034e57600080fd5b5061040860a435600401610440376103e860a43560040135111561037157600080fd5b610164356002811061038257600080fd5b506104086101a435600401610880376103e86101a4356004013511156103a757600080fd5b610380805160208201209050610cc0526020610dc06084639cba167e610ce052610cc051610d00526101c051610d2052610d40610180806000602002015182600060200201525050610d606101a0806000602002015182600060200201525050610cfc6000305af161041857600080fd5b610dc051151561046e576308c379a0610de0526020610e00526013610e20527f696e737472756374696f6e20696e76616c696400000000000000000000000000610e4052610e2050600061046d576084610dfcfd5b5b600160005260206000f3005b600015610555575b6101a0526101405261016052610180526101805160066101405160e05260c052604060c0205410156104b357600080fd5b60066101405160e05260c052604060c02061018051815410156104d557600080fd5b6101805181540381555060066101605160e05260c052604060c020805461018051825401101561050457600080fd5b61018051815401815550610180516101c05261016051610140517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef60206101c0a360016000526000516101a0515650005b63a9059cbb60005114156105dd576040600461014037341561057657600080fd5b600435602051811061058757600080fd5b5061014051610160516330e0789e61018052336101a052610140516101c052610160516101e0526101e0516101c0516101a05160065801610482565b6102405261016052610140526102405160005260206000f3005b6323b872dd60005114156106ed57606060046101403734156105fe57600080fd5b600435602051811061060f57600080fd5b50602435602051811061062157600080fd5b506101805160076101405160e05260c052604060c0206101605160e05260c052604060c02054101561065257600080fd5b60076101405160e05260c052604060c0206101605160e05260c052604060c020610180518154101561068357600080fd5b610180518154038155506101405161016051610180516330e0789e6101a052610140516101c052610160516101e0526101805161020052610200516101e0516101c05160065801610482565b610260526101805261016052610140526102605160005260206000f3005b63095ea7b3600051141561079b576040600461014037341561070e57600080fd5b600435602051811061071f57600080fd5b5060073360e05260c052604060c0206101405160e05260c052604060c020805461016051825401101561075157600080fd5b61016051815401815550610160516101805261014051337f6e11fb1b7f119e3f2fa29896ef5fdf8b8a2d0d4df6fe90ba8668e7d8b2ffa25e6020610180a3600160005260206000f3005b63c4ed3f0860005114156107c15734156107b457600080fd5b60005460005260206000f3005b630f7b9ca160005114156107e75734156107da57600080fd5b60015460005260206000f3005b6306fdde0360005114156108ca57341561080057600080fd5b60028060c052602060c020610180602082540161012060006002818352015b8261012051602002111561083257610854565b61012051850154610120516020028501525b815160010180835281141561081f575b5050505050506101805160206001820306601f82010390506101e061018051600a818352015b826101e051111561088a576108a6565b60006101e0516101a001535b815160010180835281141561087a575b5050506020610160526040610180510160206001820306601f8201039050610160f3005b6395d89b4160005114156109ad5734156108e357600080fd5b60038060c052602060c020610180602082540161012060006002818352015b8261012051602002111561091557610937565b61012051850154610120516020028501525b8151600101808352811415610902575b5050505050506101805160206001820306601f82010390506101e061018051600a818352015b826101e051111561096d57610989565b60006101e0516101a001535b815160010180835281141561095d575b5050506020610160526040610180510160206001820306601f8201039050610160f3005b63313ce56760005114156109d35734156109c657600080fd5b60045460005260206000f3005b6318160ddd60005114156109f95734156109ec57600080fd5b60055460005260206000f3005b6370a082316000511415610a485760206004610140373415610a1a57600080fd5b6004356020518110610a2b57600080fd5b5060066101405160e05260c052604060c0205460005260206000f3005b63dd62ed3e6000511415610ab85760406004610140373415610a6957600080fd5b6004356020518110610a7a57600080fd5b506024356020518110610a8c57600080fd5b5060076101405160e05260c052604060c0206101605160e05260c052604060c0205460005260206000f3005b60006000fd5b6101e3610ca1036101e36000396101e3610ca1036000f3`

// DeployBridge deploys a new Ethereum contract, binding an instance of Bridge to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend, _name string, _symbol string, _decimals *big.Int, _totalSupply *big.Int) (common.Address, *types.Transaction, *Bridge, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BridgeBin), backend, _name, _symbol, _decimals, _totalSupply)
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

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address arg0, address arg1) constant returns(uint256 out)
func (_Bridge *BridgeCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "allowance", arg0, arg1)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address arg0, address arg1) constant returns(uint256 out)
func (_Bridge *BridgeSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Bridge.Contract.Allowance(&_Bridge.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address arg0, address arg1) constant returns(uint256 out)
func (_Bridge *BridgeCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Bridge.Contract.Allowance(&_Bridge.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address arg0) constant returns(uint256 out)
func (_Bridge *BridgeCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address arg0) constant returns(uint256 out)
func (_Bridge *BridgeSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.BalanceOf(&_Bridge.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address arg0) constant returns(uint256 out)
func (_Bridge *BridgeCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.BalanceOf(&_Bridge.CallOpts, arg0)
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

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256 out)
func (_Bridge *BridgeCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256 out)
func (_Bridge *BridgeSession) Decimals() (*big.Int, error) {
	return _Bridge.Contract.Decimals(&_Bridge.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256 out)
func (_Bridge *BridgeCallerSession) Decimals() (*big.Int, error) {
	return _Bridge.Contract.Decimals(&_Bridge.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256 out)
func (_Bridge *BridgeCaller) Get(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "get")
	return *ret0, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256 out)
func (_Bridge *BridgeSession) Get() (*big.Int, error) {
	return _Bridge.Contract.Get(&_Bridge.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256 out)
func (_Bridge *BridgeCallerSession) Get() (*big.Int, error) {
	return _Bridge.Contract.Get(&_Bridge.CallOpts)
}

// GetHash is a free data retrieval call binding the contract method 0xb00140aa.
//
// Solidity: function getHash(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCaller) GetHash(opts *bind.CallOpts, inst []byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "getHash", inst)
	return *ret0, err
}

// GetHash is a free data retrieval call binding the contract method 0xb00140aa.
//
// Solidity: function getHash(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeSession) GetHash(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.GetHash(&_Bridge.CallOpts, inst)
}

// GetHash is a free data retrieval call binding the contract method 0xb00140aa.
//
// Solidity: function getHash(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCallerSession) GetHash(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.GetHash(&_Bridge.CallOpts, inst)
}

// GetHash256 is a free data retrieval call binding the contract method 0xe6d9c5f2.
//
// Solidity: function getHash256(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCaller) GetHash256(opts *bind.CallOpts, inst []byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "getHash256", inst)
	return *ret0, err
}

// GetHash256 is a free data retrieval call binding the contract method 0xe6d9c5f2.
//
// Solidity: function getHash256(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeSession) GetHash256(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.GetHash256(&_Bridge.CallOpts, inst)
}

// GetHash256 is a free data retrieval call binding the contract method 0xe6d9c5f2.
//
// Solidity: function getHash256(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCallerSession) GetHash256(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.GetHash256(&_Bridge.CallOpts, inst)
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

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string out)
func (_Bridge *BridgeCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string out)
func (_Bridge *BridgeSession) Name() (string, error) {
	return _Bridge.Contract.Name(&_Bridge.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string out)
func (_Bridge *BridgeCallerSession) Name() (string, error) {
	return _Bridge.Contract.Name(&_Bridge.CallOpts)
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

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string out)
func (_Bridge *BridgeCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string out)
func (_Bridge *BridgeSession) Symbol() (string, error) {
	return _Bridge.Contract.Symbol(&_Bridge.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string out)
func (_Bridge *BridgeCallerSession) Symbol() (string, error) {
	return _Bridge.Contract.Symbol(&_Bridge.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256 out)
func (_Bridge *BridgeCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256 out)
func (_Bridge *BridgeSession) TotalSupply() (*big.Int, error) {
	return _Bridge.Contract.TotalSupply(&_Bridge.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256 out)
func (_Bridge *BridgeCallerSession) TotalSupply() (*big.Int, error) {
	return _Bridge.Contract.TotalSupply(&_Bridge.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool out)
func (_Bridge *BridgeSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Approve(&_Bridge.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Approve(&_Bridge.TransactOpts, _spender, _value)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xc02719e2.
//
// Solidity: function swapBeacon(bytes32 newComRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconPathIsLeft, bytes32 beaconInstRoot, bytes beaconBlkData, bytes32 beaconBlkHash, bytes32[1] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[1] beaconSignerPaths, bytes32[1] bridgeInstPath, bool[1] bridgePathIsLeft, bytes32 bridgeInstRoot, bytes bridgeBlkData, bytes32 bridgeBlkHash, bytes32[1] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[1] bridgeSignerPaths) returns(bool out)
func (_Bridge *BridgeTransactor) SwapBeacon(opts *bind.TransactOpts, newComRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData []byte, beaconBlkHash [32]byte, beaconSignerPubkeys [1][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [1][32]byte, bridgeInstPath [1][32]byte, bridgePathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData []byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [1][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [1][32]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "swapBeacon", newComRoot, inst, beaconInstPath, beaconPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, bridgeInstPath, bridgePathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xc02719e2.
//
// Solidity: function swapBeacon(bytes32 newComRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconPathIsLeft, bytes32 beaconInstRoot, bytes beaconBlkData, bytes32 beaconBlkHash, bytes32[1] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[1] beaconSignerPaths, bytes32[1] bridgeInstPath, bool[1] bridgePathIsLeft, bytes32 bridgeInstRoot, bytes bridgeBlkData, bytes32 bridgeBlkHash, bytes32[1] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[1] bridgeSignerPaths) returns(bool out)
func (_Bridge *BridgeSession) SwapBeacon(newComRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData []byte, beaconBlkHash [32]byte, beaconSignerPubkeys [1][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [1][32]byte, bridgeInstPath [1][32]byte, bridgePathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData []byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [1][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [1][32]byte) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newComRoot, inst, beaconInstPath, beaconPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, bridgeInstPath, bridgePathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xc02719e2.
//
// Solidity: function swapBeacon(bytes32 newComRoot, bytes inst, bytes32[1] beaconInstPath, bool[1] beaconPathIsLeft, bytes32 beaconInstRoot, bytes beaconBlkData, bytes32 beaconBlkHash, bytes32[1] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[1] beaconSignerPaths, bytes32[1] bridgeInstPath, bool[1] bridgePathIsLeft, bytes32 bridgeInstRoot, bytes bridgeBlkData, bytes32 bridgeBlkHash, bytes32[1] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[1] bridgeSignerPaths) returns(bool out)
func (_Bridge *BridgeTransactorSession) SwapBeacon(newComRoot [32]byte, inst []byte, beaconInstPath [1][32]byte, beaconPathIsLeft [1]bool, beaconInstRoot [32]byte, beaconBlkData []byte, beaconBlkHash [32]byte, beaconSignerPubkeys [1][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [1][32]byte, bridgeInstPath [1][32]byte, bridgePathIsLeft [1]bool, bridgeInstRoot [32]byte, bridgeBlkData []byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [1][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [1][32]byte) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newComRoot, inst, beaconInstPath, beaconPathIsLeft, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, bridgeInstPath, bridgePathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Transfer(&_Bridge.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Transfer(&_Bridge.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.TransferFrom(&_Bridge.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool out)
func (_Bridge *BridgeTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.TransferFrom(&_Bridge.TransactOpts, _from, _to, _value)
}

// BridgeApproveIterator is returned from FilterApprove and is used to iterate over the raw logs and unpacked data for Approve events raised by the Bridge contract.
type BridgeApproveIterator struct {
	Event *BridgeApprove // Event containing the contract specifics and raw log

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
func (it *BridgeApproveIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeApprove)
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
		it.Event = new(BridgeApprove)
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
func (it *BridgeApproveIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeApproveIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeApprove represents a Approve event raised by the Bridge contract.
type BridgeApprove struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApprove is a free log retrieval operation binding the contract event 0x6e11fb1b7f119e3f2fa29896ef5fdf8b8a2d0d4df6fe90ba8668e7d8b2ffa25e.
//
// Solidity: event Approve(address indexed _owner, address indexed _spender, uint256 _value)
func (_Bridge *BridgeFilterer) FilterApprove(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*BridgeApproveIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Approve", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeApproveIterator{contract: _Bridge.contract, event: "Approve", logs: logs, sub: sub}, nil
}

// WatchApprove is a free log subscription operation binding the contract event 0x6e11fb1b7f119e3f2fa29896ef5fdf8b8a2d0d4df6fe90ba8668e7d8b2ffa25e.
//
// Solidity: event Approve(address indexed _owner, address indexed _spender, uint256 _value)
func (_Bridge *BridgeFilterer) WatchApprove(opts *bind.WatchOpts, sink chan<- *BridgeApprove, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Approve", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeApprove)
				if err := _Bridge.contract.UnpackLog(event, "Approve", log); err != nil {
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

// BridgeTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Bridge contract.
type BridgeTransferIterator struct {
	Event *BridgeTransfer // Event containing the contract specifics and raw log

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
func (it *BridgeTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeTransfer)
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
		it.Event = new(BridgeTransfer)
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
func (it *BridgeTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeTransfer represents a Transfer event raised by the Bridge contract.
type BridgeTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_Bridge *BridgeFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*BridgeTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &BridgeTransferIterator{contract: _Bridge.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_Bridge *BridgeFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BridgeTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeTransfer)
				if err := _Bridge.contract.UnpackLog(event, "Transfer", log); err != nil {
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
