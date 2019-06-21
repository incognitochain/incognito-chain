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
const BridgeABI = "[{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBool\",\"inputs\":[{\"type\":\"bool\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyUint256\",\"inputs\":[{\"type\":\"uint256\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"_beaconCommRoot\"},{\"type\":\"bytes32\",\"name\":\"_bridgeCommRoot\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":1536},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[3]\",\"name\":\"path\"},{\"type\":\"bool[3]\",\"name\":\"left\"},{\"type\":\"int128\",\"name\":\"length\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":14841},{\"name\":\"verifyInst\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"commRoot\"},{\"type\":\"bytes32\",\"name\":\"instHash\"},{\"type\":\"bytes32[3]\",\"name\":\"instPath\"},{\"type\":\"bool[3]\",\"name\":\"instPathIsLeft\"},{\"type\":\"int128\",\"name\":\"instPathLen\"},{\"type\":\"bytes32\",\"name\":\"instRoot\"},{\"type\":\"bytes32\",\"name\":\"blkHash\"},{\"type\":\"bytes\",\"name\":\"signerPubkeys\"},{\"type\":\"int128\",\"name\":\"signerCount\"},{\"type\":\"bytes32\",\"name\":\"signerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"signerPaths\"},{\"type\":\"bool[24]\",\"name\":\"signerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"signerPathLen\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":315230},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[3]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[3]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"beaconSignerCount\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconSignerPathLen\"},{\"type\":\"bytes32[3]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[3]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"bridgeSignerCount\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"bridgeSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeSignerPathLen\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":1055451},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":603},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":633}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05260406125446101403934156100a157600080fd5b610140516000556101605160015561252c56600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526357bff278600051141561012f57602060046101403734156100b457600080fd5b60846004356004016101603760646004356004013511156100d457600080fd5b60206003602060208206610320016101605182840111156100f457600080fd5b606480610340826020602088068803016101600160006004601cf15050818152809050905090500151610220526102205160005260206000f3005b630cf6c71060005114156103e057610120600461014037341561015157600080fd5b60a4356002811061016157600080fd5b5060c4356002811061017257600080fd5b5060e4356002811061018357600080fd5b50606051610104358060405190131561019b57600080fd5b80919012156101a957600080fd5b50610140516102605261028060006003818352015b6102405161028051121515610217576102805160008112156101df57600080fd5b6102a0526102a0516102c0527f8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd60206102c0a16103cb565b610180610280516003811061022b57600080fd5b60200201516102e0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206102e0a16101e0610280516003811061026f57600080fd5b6020020151156102d2576000610180610280516003811061028f57600080fd5b602002015160208261040001015260208101905061026051602082610400010152602081019050806104005261040090508051602082012090506102605261038b565b61018061028051600381106102e657600080fd5b602002015115156103355760006102605160208261038001015260208101905061026051602082610380010152602081019050806103805261038090508051602082012090506102605261038a565b600061026051602082610300010152602081019050610180610280516003811061035e57600080fd5b602002015160208261030001015260208101905080610300526103009050805160208201209050610260525b5b61026051610480527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020610480a15b81516001018083528114156101be575b505061016051610260511460005260206000f3005b63612477c16000511415610e02576107e0600461014037341561040257600080fd5b60a4356002811061041257600080fd5b5060c4356002811061042357600080fd5b5060e4356002811061043457600080fd5b50606051610104358060405190131561044c57600080fd5b809190121561045a57600080fd5b5061012861016435600401610920376101086101643560040135111561047f57600080fd5b606051610184358060405190131561049657600080fd5b80919012156104a457600080fd5b506104c435600281106104b657600080fd5b506104e435600281106104c857600080fd5b5061050435600281106104da57600080fd5b5061052435600281106104ec57600080fd5b5061054435600281106104fe57600080fd5b50610564356002811061051057600080fd5b50610584356002811061052257600080fd5b506105a4356002811061053457600080fd5b506105c4356002811061054657600080fd5b506105e4356002811061055857600080fd5b50610604356002811061056a57600080fd5b50610624356002811061057c57600080fd5b50610644356002811061058e57600080fd5b5061066435600281106105a057600080fd5b5061068435600281106105b257600080fd5b506106a435600281106105c457600080fd5b506106c435600281106105d657600080fd5b506106e435600281106105e857600080fd5b5061070435600281106105fa57600080fd5b50610724356002811061060c57600080fd5b50610744356002811061061e57600080fd5b50610764356002811061063057600080fd5b50610784356002811061064257600080fd5b506107a4356002811061065457600080fd5b506060516107c4358060405190131561066c57600080fd5b809190121561067a57600080fd5b5060026102c05110156107c357600e610a80527f6e6f7420656e6f75676820736967000000000000000000000000000000000000610aa052610a80805160200180610ae0828460006004600a8704601201f16106d557600080fd5b50506020610b6052610b6051610ba052610ae0805160200180610b6051610ba001828460006004600a8704601201f161070d57600080fd5b5050610b6051610ba0015160206001820306601f8201039050610b6051610ba001610b4081516020818352015b83610b405110151561074b57610768565b6000610b40516020850101535b815160010180835281141561073a575b505050506020610b6051610ba0015160206001820306601f8201039050610b60510101610b60527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610b6051610ba0a1600060005260206000f35b6020610d40610124630cf6c710610bc05261016051610be05261026051610c0052610c206101808060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610c806101e0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152505061024051610ce052610bdc6000305af161086d57600080fd5b610d405115156109d8576021610d60527f696e737472756374696f6e206973206e6f7420696e206d65726b6c6520747265610d80527f6500000000000000000000000000000000000000000000000000000000000000610da052610d60805160200180610de0828460006004600a8704601201f16108ea57600080fd5b50506020610e8052610e8051610ec052610de0805160200180610e8051610ec001828460006004600a8704601201f161092257600080fd5b5050610e8051610ec0015160206001820306601f8201039050610e8051610ec001610e6081516040818352015b83610e60511015156109605761097d565b6000610e60516020850101535b815160010180835281141561094f575b505050506020610e8051610ec0015160206001820306601f8201039050610e80510101610e80527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610e8051610ec0a1600060005260206000f35b610ee060006008818352015b6102c051610ee0511215156109f857610df4565b6060516021610ee0510280604051901315610a1257600080fd5b8091901215610a2057600080fd5b602160208206610f8001610920518284011115610a3c57600080fd5b61010880610fa0826020602088068803016109200160006004602cf1505081815280905090509050805160200180610f00828460006004600a8704601201f1610a8457600080fd5b5050610f00805160208201209050611100526111e060006003818352015b610900516111e0511315610ab557610bc2565b6103006060516111e05160605161090051610ee0510280604051901315610adb57600080fd5b8091901215610ae957600080fd5b0180604051901315610afa57600080fd5b8091901215610b0857600080fd5b60188110610b1557600080fd5b60200201516111206111e05160038110610b2e57600080fd5b60200201526106006060516111e05160605161090051610ee0510280604051901315610b5957600080fd5b8091901215610b6757600080fd5b0180604051901315610b7857600080fd5b8091901215610b8657600080fd5b60188110610b9357600080fd5b60200201516111806111e05160038110610bac57600080fd5b60200201525b8151600101808352811415610aa2575b505061110051611200527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611200a160206113a0610124630cf6c710611220526111005161124052610140516112605261128061112080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506112e06111808060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610900516113405261123c6000305af1610c9d57600080fd5b6113a0511515610de35760196113c0527f7075626b6579206e6f7420696e206d65726b6c652074726565000000000000006113e0526113c0805160200180611420828460006004600a8704601201f1610cf557600080fd5b505060206114a0526114a0516114e0526114208051602001806114a0516114e001828460006004600a8704601201f1610d2d57600080fd5b50506114a0516114e0015160206001820306601f82010390506114a0516114e00161148081516020818352015b8361148051101515610d6b57610d88565b6000611480516020850101535b8151600101808352811415610d5a575b5050505060206114a0516114e0015160206001820306601f82010390506114a05101016114a0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96114a0516114e0a1600060005260206000f35b5b81516001018083528114156109e4575b5050600160005260206000f3005b63d2031555600051141561242657610fa06004610140373415610e2457600080fd5b60846004356004016110e0376064600435600401351115610e4457600080fd5b60843560028110610e5457600080fd5b5060a43560028110610e6557600080fd5b5060c43560028110610e7657600080fd5b5060605160e43580604051901315610e8d57600080fd5b8091901215610e9b57600080fd5b50610128610164356004016111a03761010861016435600401351115610ec057600080fd5b6060516101843580604051901315610ed757600080fd5b8091901215610ee557600080fd5b506104c43560028110610ef757600080fd5b506104e43560028110610f0957600080fd5b506105043560028110610f1b57600080fd5b506105243560028110610f2d57600080fd5b506105443560028110610f3f57600080fd5b506105643560028110610f5157600080fd5b506105843560028110610f6357600080fd5b506105a43560028110610f7557600080fd5b506105c43560028110610f8757600080fd5b506105e43560028110610f9957600080fd5b506106043560028110610fab57600080fd5b506106243560028110610fbd57600080fd5b506106443560028110610fcf57600080fd5b506106643560028110610fe157600080fd5b506106843560028110610ff357600080fd5b506106a4356002811061100557600080fd5b506106c4356002811061101757600080fd5b506106e4356002811061102957600080fd5b50610704356002811061103b57600080fd5b50610724356002811061104d57600080fd5b50610744356002811061105f57600080fd5b50610764356002811061107157600080fd5b50610784356002811061108357600080fd5b506107a4356002811061109557600080fd5b506060516107c435806040519013156110ad57600080fd5b80919012156110bb57600080fd5b5061084435600281106110cd57600080fd5b5061086435600281106110df57600080fd5b5061088435600281106110f157600080fd5b506060516108a4358060405190131561110957600080fd5b809190121561111757600080fd5b5061012861092435600401611300376101086109243560040135111561113c57600080fd5b606051610944358060405190131561115357600080fd5b809190121561116157600080fd5b50610c84356002811061117357600080fd5b50610ca4356002811061118557600080fd5b50610cc4356002811061119757600080fd5b50610ce435600281106111a957600080fd5b50610d0435600281106111bb57600080fd5b50610d2435600281106111cd57600080fd5b50610d4435600281106111df57600080fd5b50610d6435600281106111f157600080fd5b50610d84356002811061120357600080fd5b50610da4356002811061121557600080fd5b50610dc4356002811061122757600080fd5b50610de4356002811061123957600080fd5b50610e04356002811061124b57600080fd5b50610e24356002811061125d57600080fd5b50610e44356002811061126f57600080fd5b50610e64356002811061128157600080fd5b50610e84356002811061129357600080fd5b50610ea435600281106112a557600080fd5b50610ec435600281106112b757600080fd5b50610ee435600281106112c957600080fd5b50610f0435600281106112db57600080fd5b50610f2435600281106112ed57600080fd5b50610f4435600281106112ff57600080fd5b50610f64356002811061131157600080fd5b50606051610f84358060405190131561132957600080fd5b809190121561133757600080fd5b506110e0805160208201209050611460526000610260516020826114a0010152602081019050610240516020826114a0010152602081019050806114a0526114a090508051602082012090506114805261028051611480511415156115a957602e611520527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069611540527f6e20626561636f6e20626c6f636b000000000000000000000000000000000000611560526115208051602001806115a0828460006004600a8704601201f161140957600080fd5b505060206116405261164051611680526115a08051602001806116405161168001828460006004600a8704601201f161144157600080fd5b505061164051611680015160206001820306601f8201039050611640516116800161162081516040818352015b836116205110151561147f5761149c565b6000611620516020850101535b815160010180835281141561146e575b50505050602061164051611680015160206001820306601f8201039050611640510101611640527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961164051611680a1610240516116a0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206116a0a1610260516116c0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206116c0a1611460516116e0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206116e0a161148051611700527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611700a15b60206120a06109246107e063612477c16117205260005461174052611460516117605261178061016080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506117e06101c08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610220516118405261024051611860526102805161188052806118a0526111a0808051602001808461174001828460006004600a8704601201f161167f57600080fd5b50508051820160206001820306601f82010390506020019150506102c0516118c0526102e0516118e0526119006103008060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050611c006106008060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152505061090051611f005261173c90506000305af16119d457600080fd5b6120a0511515611b355760236120c0527f6661696c656420766572696679696e6720626561636f6e20696e7374727563746120e0527f696f6e0000000000000000000000000000000000000000000000000000000000612100526120c0805160200180612140828460006004600a8704601201f1611a5157600080fd5b505060206121e0526121e051612220526121408051602001806121e05161222001828460006004600a8704601201f1611a8957600080fd5b50506121e051612220015160206001820306601f82010390506121e051612220016121c081516040818352015b836121c051101515611ac757611ae4565b60006121c0516020850101535b8151600101808352811415611ab6575b5050505060206121e051612220015160206001820306601f82010390506121e05101016121e0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96121e051612220a15b6000610a2051602082612240010152602081019050610a00516020826122400101526020810190508061224052612240905080516020820120905061148052610a405161148051141515611cda57602e6122c0527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f7420696122e0527f6e2062726964676520626c6f636b000000000000000000000000000000000000612300526122c0805160200180612340828460006004600a8704601201f1611bf657600080fd5b505060206123e0526123e051612420526123408051602001806123e05161242001828460006004600a8704601201f1611c2e57600080fd5b50506123e051612420015160206001820306601f82010390506123e051612420016123c081516040818352015b836123c051101515611c6c57611c89565b60006123c0516020850101535b8151600101808352811415611c5b575b5050505060206123e051612420015160206001820306601f82010390506123e05101016123e0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96123e051612420a15b6020612dc06109246107e063612477c1612440526001546124605261146051612480526124a0610920806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152505061250061098080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506109e05161256052610a005161258052610a40516125a052806125c052611300808051602001808461246001828460006004600a8704601201f1611db057600080fd5b50508051820160206001820306601f8201039050602001915050610a80516125e052610aa05161260052612620610ac08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050612920610dc08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f6020020152806010602002015182601060200201528060116020020151826011602002015280601260200201518260126020020152806013602002015182601360200201528060146020020151826014602002015280601560200201518260156020020152806016602002015182601660200201528060176020020151826017602002015250506110c051612c205261245c90506000305af161210557600080fd5b612dc0511515612241576020612de0527f6661696c6564207665726966792062726964676520696e737472756374696f6e612e0052612de0805160200180612e40828460006004600a8704601201f161215d57600080fd5b50506020612ec052612ec051612f0052612e40805160200180612ec051612f0001828460006004600a8704601201f161219557600080fd5b5050612ec051612f00015160206001820306601f8201039050612ec051612f0001612ea081516020818352015b83612ea0511015156121d3576121f0565b6000612ea0516020850101535b81516001018083528114156121c2575b505050506020612ec051612f00015160206001820306601f8201039050612ec0510101612ec0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9612ec051612f00a15b602061306060c460206357bff278612f405280612f60526110e08080516020018084612f6001828460006004600a8704601201f161227e57600080fd5b50508051820160206001820306601f8201039050602001915050612f5c90506000305af16122ab57600080fd5b61306051612f2052612f2051600055612f2051613080527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020613080a160106130a0527f6e6f2065786563657074696f6e2e2e2e000000000000000000000000000000006130c0526130a0805160200180613100828460006004600a8704601201f161233757600080fd5b5050602061318052613180516131c052613100805160200180613180516131c001828460006004600a8704601201f161236f57600080fd5b5050613180516131c0015160206001820306601f8201039050613180516131c00161316081516020818352015b83613160511015156123ad576123ca565b6000613160516020850101535b815160010180835281141561239c575b505050506020613180516131c0015160206001820306601f8201039050613180510101613180527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9613180516131c0a1600160005260206000f3005b63c4ed3f08600051141561244c57341561243f57600080fd5b60005460005260206000f3005b630f7b9ca1600051141561247257341561246557600080fd5b60015460005260206000f3005b60006000fd5b6100b461252c036100b46000396100b461252c036000f3`

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
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCaller) ParseSwapBeaconInst(opts *bind.CallOpts, inst []byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "parseSwapBeaconInst", inst)
	return *ret0, err
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeSession) ParseSwapBeaconInst(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// ParseSwapBeaconInst is a free data retrieval call binding the contract method 0x57bff278.
//
// Solidity: function parseSwapBeaconInst(bytes inst) constant returns(bytes32 out)
func (_Bridge *BridgeCallerSession) ParseSwapBeaconInst(inst []byte) ([32]byte, error) {
	return _Bridge.Contract.ParseSwapBeaconInst(&_Bridge.CallOpts, inst)
}

// VerifyInst is a free data retrieval call binding the contract method 0x612477c1.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes signerPubkeys, int128 signerCount, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft, int128 signerPathLen) constant returns(bool out)
func (_Bridge *BridgeCaller) VerifyInst(opts *bind.CallOpts, commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys []byte, signerCount *big.Int, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool, signerPathLen *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Bridge.contract.Call(opts, out, "verifyInst", commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerCount, signerSig, signerPaths, signerPathIsLeft, signerPathLen)
	return *ret0, err
}

// VerifyInst is a free data retrieval call binding the contract method 0x612477c1.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes signerPubkeys, int128 signerCount, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft, int128 signerPathLen) constant returns(bool out)
func (_Bridge *BridgeSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys []byte, signerCount *big.Int, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool, signerPathLen *big.Int) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerCount, signerSig, signerPaths, signerPathIsLeft, signerPathLen)
}

// VerifyInst is a free data retrieval call binding the contract method 0x612477c1.
//
// Solidity: function verifyInst(bytes32 commRoot, bytes32 instHash, bytes32[3] instPath, bool[3] instPathIsLeft, int128 instPathLen, bytes32 instRoot, bytes32 blkHash, bytes signerPubkeys, int128 signerCount, bytes32 signerSig, bytes32[24] signerPaths, bool[24] signerPathIsLeft, int128 signerPathLen) constant returns(bool out)
func (_Bridge *BridgeCallerSession) VerifyInst(commRoot [32]byte, instHash [32]byte, instPath [3][32]byte, instPathIsLeft [3]bool, instPathLen *big.Int, instRoot [32]byte, blkHash [32]byte, signerPubkeys []byte, signerCount *big.Int, signerSig [32]byte, signerPaths [24][32]byte, signerPathIsLeft [24]bool, signerPathLen *big.Int) (bool, error) {
	return _Bridge.Contract.VerifyInst(&_Bridge.CallOpts, commRoot, instHash, instPath, instPathIsLeft, instPathLen, instRoot, blkHash, signerPubkeys, signerCount, signerSig, signerPaths, signerPathIsLeft, signerPathLen)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xd2031555.
//
// Solidity: function swapBeacon(bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeTransactor) SwapBeacon(opts *bind.TransactOpts, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "swapBeacon", inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xd2031555.
//
// Solidity: function swapBeacon(bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeSession) SwapBeacon(inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0xd2031555.
//
// Solidity: function swapBeacon(bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeTransactorSession) SwapBeacon(inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
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
