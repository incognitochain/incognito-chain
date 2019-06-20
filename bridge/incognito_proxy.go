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
const BridgeABI = "[{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBool\",\"inputs\":[{\"type\":\"bool\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyUint256\",\"inputs\":[{\"type\":\"uint256\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"_beaconCommRoot\"},{\"type\":\"bytes32\",\"name\":\"_bridgeCommRoot\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32[8]\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":487},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[3]\",\"name\":\"path\"},{\"type\":\"bool[3]\",\"name\":\"left\"},{\"type\":\"int128\",\"name\":\"length\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":6618},{\"name\":\"verifyInst\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"commRoot\"},{\"type\":\"bytes32\",\"name\":\"instHash\"},{\"type\":\"bytes32[3]\",\"name\":\"instPath\"},{\"type\":\"bool[3]\",\"name\":\"instPathIsLeft\"},{\"type\":\"int128\",\"name\":\"instPathLen\"},{\"type\":\"bytes32\",\"name\":\"instRoot\"},{\"type\":\"bytes32\",\"name\":\"blkHash\"},{\"type\":\"bytes32[8]\",\"name\":\"signerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"signerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"signerPaths\"},{\"type\":\"bool[24]\",\"name\":\"signerPathIsLeft\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":210207},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"newCommRoot\"},{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[3]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[3]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes32[8]\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"bytes32[3]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[3]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes32[8]\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"bridgeSignerPathIsLeft\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":335895},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":603},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":633}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526040611af26101403934156100a157600080fd5b6101405160005561016051600155611ada56600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526357bff27860005114156100dd57602060046101403734156100b457600080fd5b60846004356004016101603760646004356004013511156100d457600080fd5b610100610220f3005b630cf6c710600051141561031b5761012060046101403734156100ff57600080fd5b60a4356002811061010f57600080fd5b5060c4356002811061012057600080fd5b5060e4356002811061013157600080fd5b50606051610104358060405190131561014957600080fd5b809190121561015757600080fd5b50610140516102605261028060006003818352015b61024051610280511215156101c55761028051600081121561018d57600080fd5b6102a0526102a0516102c0527f8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd60206102c0a1610306565b6101e061028051600381106101d957600080fd5b60200201511561023c57600061018061028051600381106101f957600080fd5b60200201516020826103e0010152602081019050610260516020826103e0010152602081019050806103e0526103e09050805160208201209050610260526102f5565b610180610280516003811061025057600080fd5b6020020151151561029f576000610260516020826103600101526020810190506102605160208261036001015260208101905080610360526103609050805160208201209050610260526102f4565b6000610260516020826102e001015260208101905061018061028051600381106102c857600080fd5b60200201516020826102e0010152602081019050806102e0526102e09050805160208201209050610260525b5b5b815160010180835281141561016c575b505061016051610260511460005260206000f3005b6376c5d0236000511415610c3b57610880600461014037341561033d57600080fd5b60a4356002811061034d57600080fd5b5060c4356002811061035e57600080fd5b5060e4356002811061036f57600080fd5b50606051610104358060405190131561038757600080fd5b809190121561039557600080fd5b5061058435600281106103a757600080fd5b506105a435600281106103b957600080fd5b506105c435600281106103cb57600080fd5b506105e435600281106103dd57600080fd5b5061060435600281106103ef57600080fd5b50610624356002811061040157600080fd5b50610644356002811061041357600080fd5b50610664356002811061042557600080fd5b50610684356002811061043757600080fd5b506106a4356002811061044957600080fd5b506106c4356002811061045b57600080fd5b506106e4356002811061046d57600080fd5b50610704356002811061047f57600080fd5b50610724356002811061049157600080fd5b5061074435600281106104a357600080fd5b5061076435600281106104b557600080fd5b5061078435600281106104c757600080fd5b506107a435600281106104d957600080fd5b506107c435600281106104eb57600080fd5b506107e435600281106104fd57600080fd5b50610804356002811061050f57600080fd5b50610824356002811061052157600080fd5b50610844356002811061053357600080fd5b50610864356002811061054557600080fd5b506020610b40610124630cf6c7106109c052610160516109e05261026051610a0052610a206101808060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610a806101e0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152505061024051610ae0526109dc6000305af16105f057600080fd5b610b4051151561075b576021610b60527f696e737472756374696f6e206973206e6f7420696e206d65726b6c6520747265610b80527f6500000000000000000000000000000000000000000000000000000000000000610ba052610b60805160200180610be0828460006004600a8704601201f161066d57600080fd5b50506020610c8052610c8051610cc052610be0805160200180610c8051610cc001828460006004600a8704601201f16106a557600080fd5b5050610c8051610cc0015160206001820306601f8201039050610c8051610cc001610c6081516040818352015b83610c60511015156106e357610700565b6000610c60516020850101535b81516001018083528114156106d2575b505050506020610c8051610cc0015160206001820306601f8201039050610c80510101610c80527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610c8051610cc0a1600060005260206000f35b6000610ce052610d0060006008818352015b6102a0610d00516008811061078157600080fd5b6020020151151561079157610ad5565b6003610de052610e0060006003818352015b6103c0606051610e0051606051610de051610d005102806040519013156107c957600080fd5b80919012156107d757600080fd5b01806040519013156107e857600080fd5b80919012156107f657600080fd5b6018811061080357600080fd5b6020020151610d20610e00516003811061081c57600080fd5b60200201526106c0606051610e0051606051610de051610d0051028060405190131561084757600080fd5b809190121561085557600080fd5b018060405190131561086657600080fd5b809190121561087457600080fd5b6018811061088157600080fd5b6020020151610d80610e00516003811061089a57600080fd5b60200201525b81516001018083528114156107a3575b50506020610fa0610124630cf6c710610e20526102a0610d0051600881106108d757600080fd5b6020020151610e405261014051610e6052610e80610d208060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610ee0610d80806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152505061024051610f4052610e3c6000305af161097157600080fd5b610fa0511515610ab7576019610fc0527f7075626b6579206e6f7420696e206d65726b6c65207472656500000000000000610fe052610fc0805160200180611020828460006004600a8704601201f16109c957600080fd5b505060206110a0526110a0516110e0526110208051602001806110a0516110e001828460006004600a8704601201f1610a0157600080fd5b50506110a0516110e0015160206001820306601f82010390506110a0516110e00161108081516020818352015b8361108051101515610a3f57610a5c565b6000611080516020850101535b8151600101808352811415610a2e575b5050505060206110a0516110e0015160206001820306601f82010390506110a05101016110a0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96110a0516110e0a1600060005260206000f35b610ce0805160018251011015610acc57600080fd5b60018151018152505b815160010180835281141561076d575b50506002610ce0511015610c2f57600e611100527f6e6f7420656e6f7567682073696700000000000000000000000000000000000061112052611100805160200180611160828460006004600a8704601201f1610b4157600080fd5b505060206111e0526111e051611220526111608051602001806111e05161122001828460006004600a8704601201f1610b7957600080fd5b50506111e051611220015160206001820306601f82010390506111e051611220016111c081516020818352015b836111c051101515610bb757610bd4565b60006111c0516020850101535b8151600101808352811415610ba6575b5050505060206111e051611220015160206001820306601f82010390506111e05101016111e0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96111e051611220a1600060005260206000f35b600160005260206000f3005b63fa3f9eae60005114156119d4576110e06004610140373415610c5d57600080fd5b6084602435600401611220376064602435600401351115610c7d57600080fd5b60a43560028110610c8d57600080fd5b5060c43560028110610c9e57600080fd5b5060e43560028110610caf57600080fd5b506060516101043580604051901315610cc757600080fd5b8091901215610cd557600080fd5b506105a43560028110610ce757600080fd5b506105c43560028110610cf957600080fd5b506105e43560028110610d0b57600080fd5b506106043560028110610d1d57600080fd5b506106243560028110610d2f57600080fd5b506106443560028110610d4157600080fd5b506106643560028110610d5357600080fd5b506106843560028110610d6557600080fd5b506106a43560028110610d7757600080fd5b506106c43560028110610d8957600080fd5b506106e43560028110610d9b57600080fd5b506107043560028110610dad57600080fd5b506107243560028110610dbf57600080fd5b506107443560028110610dd157600080fd5b506107643560028110610de357600080fd5b506107843560028110610df557600080fd5b506107a43560028110610e0757600080fd5b506107c43560028110610e1957600080fd5b506107e43560028110610e2b57600080fd5b506108043560028110610e3d57600080fd5b506108243560028110610e4f57600080fd5b506108443560028110610e6157600080fd5b506108643560028110610e7357600080fd5b506108843560028110610e8557600080fd5b506109043560028110610e9757600080fd5b506109243560028110610ea957600080fd5b506109443560028110610ebb57600080fd5b50610de43560028110610ecd57600080fd5b50610e043560028110610edf57600080fd5b50610e243560028110610ef157600080fd5b50610e443560028110610f0357600080fd5b50610e643560028110610f1557600080fd5b50610e843560028110610f2757600080fd5b50610ea43560028110610f3957600080fd5b50610ec43560028110610f4b57600080fd5b50610ee43560028110610f5d57600080fd5b50610f043560028110610f6f57600080fd5b50610f243560028110610f8157600080fd5b50610f443560028110610f9357600080fd5b50610f643560028110610fa557600080fd5b50610f843560028110610fb757600080fd5b50610fa43560028110610fc957600080fd5b50610fc43560028110610fdb57600080fd5b50610fe43560028110610fed57600080fd5b506110043560028110610fff57600080fd5b50611024356002811061101157600080fd5b50611044356002811061102357600080fd5b50611064356002811061103557600080fd5b50611084356002811061104757600080fd5b506110a4356002811061105957600080fd5b506110c4356002811061106b57600080fd5b506112208051602082012090506112e0526000610280516020826113200101526020810190506102605160208261132001015260208101905080611320526113209050805160208201209050611300526102a051611300511415156112dd57602e6113a0527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f7420696113c0527f6e20626561636f6e20626c6f636b0000000000000000000000000000000000006113e0526113a0805160200180611420828460006004600a8704601201f161113d57600080fd5b505060206114c0526114c051611500526114208051602001806114c05161150001828460006004600a8704601201f161117557600080fd5b50506114c051611500015160206001820306601f82010390506114c051611500016114a081516040818352015b836114a0511015156111b3576111d0565b60006114a0516020850101535b81516001018083528114156111a2575b5050505060206114c051611500015160206001820306601f82010390506114c05101016114c0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96114c051611500a161026051611520527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611520a161028051611540527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611540a16112e051611560527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611560a161130051611580527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611580a15b6020611e806108846376c5d0236115a0526000546115c0526112e0516115e05261160061018080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506116606101e08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610240516116c052610260516116e0526102a051611700526117206102c0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201528060046020020151826004602002015280600560200201518260056020020152806006602002015182600660200201528060076020020151826007602002015250506103c051611820526118406103e08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050611b406106e08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f6020020152806010602002015182601060200201528060116020020151826011602002015280601260200201518260126020020152806013602002015182601360200201528060146020020151826014602002015280601560200201518260156020020152806016602002015182601660200201528060176020020151826017602002015250506115bc6000305af161173657600080fd5b611e80511515611897576023611ea0527f6661696c656420766572696679696e6720626561636f6e20696e737472756374611ec0527f696f6e0000000000000000000000000000000000000000000000000000000000611ee052611ea0805160200180611f20828460006004600a8704601201f16117b357600080fd5b50506020611fc052611fc05161200052611f20805160200180611fc05161200001828460006004600a8704601201f16117eb57600080fd5b5050611fc051612000015160206001820306601f8201039050611fc05161200001611fa081516040818352015b83611fa05110151561182957611846565b6000611fa0516020850101535b8151600101808352811415611818575b505050506020611fc051612000015160206001820306601f8201039050611fc0510101611fc0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9611fc051612000a15b6010612020527f6e6f2065786563657074696f6e2e2e2e0000000000000000000000000000000061204052612020805160200180612080828460006004600a8704601201f16118e557600080fd5b505060206121005261210051612140526120808051602001806121005161214001828460006004600a8704601201f161191d57600080fd5b505061210051612140015160206001820306601f820103905061210051612140016120e081516020818352015b836120e05110151561195b57611978565b60006120e0516020850101535b815160010180835281141561194a575b50505050602061210051612140015160206001820306601f8201039050612100510101612100527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961210051612140a1600160005260206000f3005b63c4ed3f0860005114156119fa5734156119ed57600080fd5b60005460005260206000f3005b630f7b9ca16000511415611a20573415611a1357600080fd5b60015460005260206000f3005b60006000fd5b6100b4611ada036100b46000396100b4611ada036000f3`

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

// InMerkleTree is a free data retrieval call binding the contract method 0x0cf6c710.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[3] path, bool[3] left, int128 length) constant returns(bool out)
func (_Bridge *BridgeCaller) InMerkleTree(opts *bind.CallOpts, leaf [32]byte, root [32]byte, path [3][32]byte, left [3]bool, length *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "inMerkleTree", leaf, root, path, left, length)
	return *ret0, err
}

// InMerkleTree is a free data retrieval call binding the contract method 0x0cf6c710.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[3] path, bool[3] left, int128 length) constant returns(bool out)
func (_Bridge *BridgeSession) InMerkleTree(leaf [32]byte, root [32]byte, path [3][32]byte, left [3]bool, length *big.Int) (bool, error) {
	return _Bridge.Contract.InMerkleTree(&_Bridge.CallOpts, leaf, root, path, left, length)
}

// InMerkleTree is a free data retrieval call binding the contract method 0x0cf6c710.
//
// Solidity: function inMerkleTree(bytes32 leaf, bytes32 root, bytes32[3] path, bool[3] left, int128 length) constant returns(bool out)
func (_Bridge *BridgeCallerSession) InMerkleTree(leaf [32]byte, root [32]byte, path [3][32]byte, left [3]bool, length *big.Int) (bool, error) {
	return _Bridge.Contract.InMerkleTree(&_Bridge.CallOpts, leaf, root, path, left, length)
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[8] out)
func (_Bridge *BridgeCaller) ParseSwapBeaconInst(opts *bind.CallOpts, inst []byte) ([8][32]byte, error) {
	var (
		ret0 = new([8][32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "parseSwapBeaconInst", inst)
	return *ret0, err
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[8] out)
func (_Bridge *BridgeSession) ParseSwapBeaconInst(inst []byte) ([8][32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32[8] out)
func (_Bridge *BridgeCallerSession) ParseSwapBeaconInst(inst []byte) ([8][32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// VerifyInst is a free data retrieval call binding the contract method 0x76c5d023.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes32[8] signerPubkeys, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeCaller) VerifyInst(opts *bind.CallOpts, commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys [8][32]byte, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "verifyInst", commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
	return *ret0, err
}

// VerifyInst is a free data retrieval call binding the contract method 0x76c5d023.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes32[8] signerPubkeys, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys [8][32]byte, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
}

// VerifyInst is a free data retrieval call binding the contract method 0x76c5d023.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes32[8] signerPubkeys, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft) constant returns(bool out)
func (_Bridge *BridgeCallerSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys [8][32]byte, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerSig, signerPaths, signerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xfa3f9eae.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[8] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[8] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeTransactor) SwapBeacon(opts *bind.TransactOpts, newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [8][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [8][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "swapBeacon", newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xfa3f9eae.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[8] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[8] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [8][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [8][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xfa3f9eae.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes32[8] beaconSignerPubkeys, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes32[8] bridgeSignerPubkeys, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft) returns(bool out)
func (_Bridge *BridgeTransactorSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys [8][32]byte, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys [8][32]byte, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft)
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

// BridgeNotifyUint256Iterator is returned from FilterNotifyUint256 and is used to iterate over the raw logs and unpacked data for NotifyUint256 events raised by the Bridge contract.
type BridgeNotifyUint256Iterator struct {
	Event *BridgeNotifyUint256 // Event containing the contract specifics and raw log

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
func (it *BridgeNotifyUint256Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeNotifyUint256)
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
		it.Event = new(BridgeNotifyUint256)
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
func (it *BridgeNotifyUint256Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeNotifyUint256Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeNotifyUint256 represents a NotifyUint256 event raised by the Bridge contract.
type BridgeNotifyUint256 struct {
	Content *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyUint256 is a free log retrieval operation binding the contract event 0x8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd.
//
// Solidity: event NotifyUint256(uint256 content)
func (_Bridge *BridgeFilterer) FilterNotifyUint256(opts *bind.FilterOpts) (*BridgeNotifyUint256Iterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "NotifyUint256")
	if err != nil {
		return nil, err
	}
	return &BridgeNotifyUint256Iterator{contract: _Bridge.contract, event: "NotifyUint256", logs: logs, sub: sub}, nil
}

// WatchNotifyUint256 is a free log subscription operation binding the contract event 0x8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd.
//
// Solidity: event NotifyUint256(uint256 content)
func (_Bridge *BridgeFilterer) WatchNotifyUint256(opts *bind.WatchOpts, sink chan<- *BridgeNotifyUint256) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "NotifyUint256")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeNotifyUint256)
				if err := _Bridge.contract.UnpackLog(event, "NotifyUint256", log); err != nil {
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
