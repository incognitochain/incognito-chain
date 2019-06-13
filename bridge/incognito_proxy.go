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
const BridgeABI = "[{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"_beaconCommRoot\"},{\"type\":\"bytes32\",\"name\":\"_bridgeCommRoot\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"get\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":283},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32[2]\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":498},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[1]\",\"name\":\"path\"},{\"type\":\"bool[1]\",\"name\":\"left\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":943},{\"name\":\"getHash\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":615},{\"name\":\"getHash256\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":2710},{\"name\":\"verifyInst\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"commRoot\"},{\"type\":\"bytes32\",\"name\":\"instHash\"},{\"type\":\"bytes32[1]\",\"name\":\"instPath\"},{\"type\":\"bool[1]\",\"name\":\"instPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"instRoot\"},{\"type\":\"bytes32\",\"name\":\"blkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"signerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"signerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"signerPaths\"},{\"type\":\"bool[2]\",\"name\":\"signerPathIsLeft\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":10955},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"newCommRoot\"},{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[1]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[1]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[2]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"bytes32[1]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[1]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes32[2]\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[2]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[2]\",\"name\":\"bridgeSignerPathIsLeft\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":19254},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":693},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":723}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526040610a916101403934156100a157600080fd5b6101405160005561016051600155610a7956600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a052636d4ce63c60005114156100b85734156100ac57600080fd5b600260005260206000f3005b6357bff278600051141561010157602060046101403734156100d957600080fd5b60846004356004016101603760646004356004013511156100f957600080fd5b6040610220f3005b639cba167e6000511415610239576080600461014037341561012257600080fd5b6064356002811061013257600080fd5b50610140516101c0526101e060006001818352015b6101a06101e0516001811061015b57600080fd5b6020020151156101be5760006101806101e0516001811061017b57600080fd5b60200201516020826102800101526020810190506101c051602082610280010152602081019050806102805261028090508051602082012090506101c052610213565b60006101c0516020826102000101526020810190506101806101e051600181106101e757600080fd5b6020020151602082610200010152602081019050806102005261020090508051602082012090506101c0525b5b8151600101808352811415610147575b5050610160516101c0511460005260206000f3005b63b00140aa6000511415610290576020600461014037341561025a57600080fd5b608460043560040161016037606460043560040135111561027a57600080fd5b61016080516020820120905060005260206000f3005b63e6d9c5f260005114156102fc57602060046101403734156102b157600080fd5b60846004356004016101603760646004356004013511156102d157600080fd5b610160602060c0825160208401600060025af16102ed57600080fd5b60c051905060005260206000f3005b63c4f5463460005114156106a7576101a0600461014037341561031e57600080fd5b6064356002811061032e57600080fd5b50610164356002811061034057600080fd5b50610184356002811061035257600080fd5b5060206103c06084639cba167e6102e05261016051610300526101c051610320526103406101808060006020020151826000602002015250506103606101a08060006020020151826000602002015250506102fc6000305af16103b457600080fd5b6103c051151561042f576308c379a06103e0526020610400526021610420527f696e737472756374696f6e206973206e6f7420696e206d65726b6c6520747265610440527f65000000000000000000000000000000000000000000000000000000000000006104605261042050600061042e5760a46103fcfd5b5b60006104a0526104c060006002818352015b6102006104c0516002811061045557600080fd5b6020020151151561046557610631565b60016105205261054060006001818352015b61026060605161054051606051610520516104c051028060405190131561049d57600080fd5b80919012156104ab57600080fd5b01806040519013156104bc57600080fd5b80919012156104ca57600080fd5b600281106104d757600080fd5b60200201516104e061054051600181106104f057600080fd5b60200201526102a060605161054051606051610520516104c051028060405190131561051b57600080fd5b809190121561052957600080fd5b018060405190131561053a57600080fd5b809190121561054857600080fd5b6002811061055557600080fd5b6020020151610500610540516001811061056e57600080fd5b60200201525b8151600101808352811415610477575b505060206106406084639cba167e610560526102006104c051600281106105aa57600080fd5b602002015161058052610140516105a0526105c06104e08060006020020151826000602002015250506105e061050080600060200201518260006020020152505061057c6000305af16105fc57600080fd5b6106405115156106135760011561061257600080fd5b5b6104a080516001825101101561062857600080fd5b60018151018152505b8151600101808352811415610441575b505060026104a051101561069b576308c379a06106605260206106805260146106a0527f6e6f7420656e6f756768207369676e61747572650000000000000000000000006106c0526106a050600061069a57608461067cfd5b5b600160005260206000f3005b630e9e539d60005114156109735761034060046101403734156106c957600080fd5b60846024356004016104803760646024356004013511156106e957600080fd5b606435600281106106f957600080fd5b50610184356002811061070b57600080fd5b506101a4356002811061071d57600080fd5b506101e4356002811061072f57600080fd5b50610304356002811061074157600080fd5b50610324356002811061075357600080fd5b506104808051602082012090506105405260006101c0516020826105800101526020810190506101e0516020826105800101526020810190508061058052610580905080516020820120905061056052610200516105605114151561096757602e610600527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069610620527f6e20626561636f6e20626c6f636b00000000000000000000000000000000000061064052610600805160200180610680828460006004600a8704601201f161082557600080fd5b505060206107205261072051610760526106808051602001806107205161076001828460006004600a8704601201f161085d57600080fd5b505061072051610760015160206001820306601f8201039050610720516107600161070081516040818352015b836107005110151561089b576108b8565b6000610700516020850101535b815160010180835281141561088a575b50505050602061072051610760015160206001820306601f8201039050610720510101610720527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961072051610760a161054051610780527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020610780a1610560516107a0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206107a0a15b600160005260206000f3005b63c4ed3f08600051141561099957341561098c57600080fd5b60005460005260206000f3005b630f7b9ca160005114156109bf5734156109b257600080fd5b60015460005260206000f3005b60006000fd5b6100b4610a79036100b46000396100b4610a79036000f3`

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
