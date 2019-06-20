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
const BridgeABI = "[{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBool\",\"inputs\":[{\"type\":\"bool\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyUint256\",\"inputs\":[{\"type\":\"uint256\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"_beaconCommRoot\"},{\"type\":\"bytes32\",\"name\":\"_bridgeCommRoot\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"parseSwapBeaconInst\",\"outputs\":[{\"type\":\"bytes32[8]\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":487},{\"name\":\"inMerkleTree\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"leaf\"},{\"type\":\"bytes32\",\"name\":\"root\"},{\"type\":\"bytes32[3]\",\"name\":\"path\"},{\"type\":\"bool[3]\",\"name\":\"left\"},{\"type\":\"int128\",\"name\":\"length\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":14841},{\"name\":\"verifyInst\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"commRoot\"},{\"type\":\"bytes32\",\"name\":\"instHash\"},{\"type\":\"bytes32[3]\",\"name\":\"instPath\"},{\"type\":\"bool[3]\",\"name\":\"instPathIsLeft\"},{\"type\":\"int128\",\"name\":\"instPathLen\"},{\"type\":\"bytes32\",\"name\":\"instRoot\"},{\"type\":\"bytes32\",\"name\":\"blkHash\"},{\"type\":\"bytes\",\"name\":\"signerPubkeys\"},{\"type\":\"int128\",\"name\":\"signerCount\"},{\"type\":\"bytes32\",\"name\":\"signerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"signerPaths\"},{\"type\":\"bool[24]\",\"name\":\"signerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"signerPathLen\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":315230},{\"name\":\"swapBeacon\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"newCommRoot\"},{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[3]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[3]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"beaconSignerCount\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconSignerPathLen\"},{\"type\":\"bytes32[3]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[3]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"bridgeSignerCount\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[24]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[24]\",\"name\":\"bridgeSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeSignerPathLen\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":1025652},{\"name\":\"beaconCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":603},{\"name\":\"bridgeCommRoot\",\"outputs\":[{\"type\":\"bytes32\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":633}]"

// BridgeBin is the compiled bytecode used for deploying new contracts.
const BridgeBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05260406125c76101403934156100a157600080fd5b61014051600055610160516001556125af56600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a0526357bff27860005114156100dd57602060046101403734156100b457600080fd5b60846004356004016101603760646004356004013511156100d457600080fd5b610100610220f3005b630cf6c710600051141561038e5761012060046101403734156100ff57600080fd5b60a4356002811061010f57600080fd5b5060c4356002811061012057600080fd5b5060e4356002811061013157600080fd5b50606051610104358060405190131561014957600080fd5b809190121561015757600080fd5b50610140516102605261028060006003818352015b61024051610280511215156101c55761028051600081121561018d57600080fd5b6102a0526102a0516102c0527f8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd60206102c0a1610379565b61018061028051600381106101d957600080fd5b60200201516102e0527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b60206102e0a16101e0610280516003811061021d57600080fd5b602002015115610280576000610180610280516003811061023d57600080fd5b6020020151602082610400010152602081019050610260516020826104000101526020810190508061040052610400905080516020820120905061026052610339565b610180610280516003811061029457600080fd5b602002015115156102e357600061026051602082610380010152602081019050610260516020826103800101526020810190508061038052610380905080516020820120905061026052610338565b600061026051602082610300010152602081019050610180610280516003811061030c57600080fd5b602002015160208261030001015260208101905080610300526103009050805160208201209050610260525b5b61026051610480527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020610480a15b815160010180835281141561016c575b505061016051610260511460005260206000f3005b63612477c16000511415610db0576107e060046101403734156103b057600080fd5b60a435600281106103c057600080fd5b5060c435600281106103d157600080fd5b5060e435600281106103e257600080fd5b5060605161010435806040519013156103fa57600080fd5b809190121561040857600080fd5b5061012861016435600401610920376101086101643560040135111561042d57600080fd5b606051610184358060405190131561044457600080fd5b809190121561045257600080fd5b506104c4356002811061046457600080fd5b506104e4356002811061047657600080fd5b50610504356002811061048857600080fd5b50610524356002811061049a57600080fd5b5061054435600281106104ac57600080fd5b5061056435600281106104be57600080fd5b5061058435600281106104d057600080fd5b506105a435600281106104e257600080fd5b506105c435600281106104f457600080fd5b506105e4356002811061050657600080fd5b50610604356002811061051857600080fd5b50610624356002811061052a57600080fd5b50610644356002811061053c57600080fd5b50610664356002811061054e57600080fd5b50610684356002811061056057600080fd5b506106a4356002811061057257600080fd5b506106c4356002811061058457600080fd5b506106e4356002811061059657600080fd5b5061070435600281106105a857600080fd5b5061072435600281106105ba57600080fd5b5061074435600281106105cc57600080fd5b5061076435600281106105de57600080fd5b5061078435600281106105f057600080fd5b506107a4356002811061060257600080fd5b506060516107c4358060405190131561061a57600080fd5b809190121561062857600080fd5b5060026102c051101561077157600e610a80527f6e6f7420656e6f75676820736967000000000000000000000000000000000000610aa052610a80805160200180610ae0828460006004600a8704601201f161068357600080fd5b50506020610b6052610b6051610ba052610ae0805160200180610b6051610ba001828460006004600a8704601201f16106bb57600080fd5b5050610b6051610ba0015160206001820306601f8201039050610b6051610ba001610b4081516020818352015b83610b40511015156106f957610716565b6000610b40516020850101535b81516001018083528114156106e8575b505050506020610b6051610ba0015160206001820306601f8201039050610b60510101610b60527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610b6051610ba0a1600060005260206000f35b6020610d40610124630cf6c710610bc05261016051610be05261026051610c0052610c206101808060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610c806101e0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152505061024051610ce052610bdc6000305af161081b57600080fd5b610d40511515610986576021610d60527f696e737472756374696f6e206973206e6f7420696e206d65726b6c6520747265610d80527f6500000000000000000000000000000000000000000000000000000000000000610da052610d60805160200180610de0828460006004600a8704601201f161089857600080fd5b50506020610e8052610e8051610ec052610de0805160200180610e8051610ec001828460006004600a8704601201f16108d057600080fd5b5050610e8051610ec0015160206001820306601f8201039050610e8051610ec001610e6081516040818352015b83610e605110151561090e5761092b565b6000610e60516020850101535b81516001018083528114156108fd575b505050506020610e8051610ec0015160206001820306601f8201039050610e80510101610e80527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9610e8051610ec0a1600060005260206000f35b610ee060006008818352015b6102c051610ee0511215156109a657610da2565b6060516021610ee05102806040519013156109c057600080fd5b80919012156109ce57600080fd5b602160208206610f80016109205182840111156109ea57600080fd5b61010880610fa0826020602088068803016109200160006004602cf1505081815280905090509050805160200180610f00828460006004600a8704601201f1610a3257600080fd5b5050610f00805160208201209050611100526111e060006003818352015b610900516111e0511315610a6357610b70565b6103006060516111e05160605161090051610ee0510280604051901315610a8957600080fd5b8091901215610a9757600080fd5b0180604051901315610aa857600080fd5b8091901215610ab657600080fd5b60188110610ac357600080fd5b60200201516111206111e05160038110610adc57600080fd5b60200201526106006060516111e05160605161090051610ee0510280604051901315610b0757600080fd5b8091901215610b1557600080fd5b0180604051901315610b2657600080fd5b8091901215610b3457600080fd5b60188110610b4157600080fd5b60200201516111806111e05160038110610b5a57600080fd5b60200201525b8151600101808352811415610a50575b505061110051611200527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611200a160206113a0610124630cf6c710611220526111005161124052610140516112605261128061112080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506112e06111808060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610900516113405261123c6000305af1610c4b57600080fd5b6113a0511515610d915760196113c0527f7075626b6579206e6f7420696e206d65726b6c652074726565000000000000006113e0526113c0805160200180611420828460006004600a8704601201f1610ca357600080fd5b505060206114a0526114a0516114e0526114208051602001806114a0516114e001828460006004600a8704601201f1610cdb57600080fd5b50506114a0516114e0015160206001820306601f82010390506114a0516114e00161148081516020818352015b8361148051101515610d1957610d36565b6000611480516020850101535b8151600101808352811415610d08575b5050505060206114a0516114e0015160206001820306601f82010390506114a05101016114a0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96114a0516114e0a1600060005260206000f35b5b8151600101808352811415610992575b5050600160005260206000f3005b632bb750bc60005114156124a957610fc06004610140373415610dd257600080fd5b6084602435600401611100376064602435600401351115610df257600080fd5b60a43560028110610e0257600080fd5b5060c43560028110610e1357600080fd5b5060e43560028110610e2457600080fd5b506060516101043580604051901315610e3c57600080fd5b8091901215610e4a57600080fd5b50610128610184356004016111c03761010861018435600401351115610e6f57600080fd5b6060516101a43580604051901315610e8657600080fd5b8091901215610e9457600080fd5b506104e43560028110610ea657600080fd5b506105043560028110610eb857600080fd5b506105243560028110610eca57600080fd5b506105443560028110610edc57600080fd5b506105643560028110610eee57600080fd5b506105843560028110610f0057600080fd5b506105a43560028110610f1257600080fd5b506105c43560028110610f2457600080fd5b506105e43560028110610f3657600080fd5b506106043560028110610f4857600080fd5b506106243560028110610f5a57600080fd5b506106443560028110610f6c57600080fd5b506106643560028110610f7e57600080fd5b506106843560028110610f9057600080fd5b506106a43560028110610fa257600080fd5b506106c43560028110610fb457600080fd5b506106e43560028110610fc657600080fd5b506107043560028110610fd857600080fd5b506107243560028110610fea57600080fd5b506107443560028110610ffc57600080fd5b50610764356002811061100e57600080fd5b50610784356002811061102057600080fd5b506107a4356002811061103257600080fd5b506107c4356002811061104457600080fd5b506060516107e4358060405190131561105c57600080fd5b809190121561106a57600080fd5b50610864356002811061107c57600080fd5b50610884356002811061108e57600080fd5b506108a435600281106110a057600080fd5b506060516108c435806040519013156110b857600080fd5b80919012156110c657600080fd5b506101286109443560040161132037610108610944356004013511156110eb57600080fd5b606051610964358060405190131561110257600080fd5b809190121561111057600080fd5b50610ca4356002811061112257600080fd5b50610cc4356002811061113457600080fd5b50610ce4356002811061114657600080fd5b50610d04356002811061115857600080fd5b50610d24356002811061116a57600080fd5b50610d44356002811061117c57600080fd5b50610d64356002811061118e57600080fd5b50610d8435600281106111a057600080fd5b50610da435600281106111b257600080fd5b50610dc435600281106111c457600080fd5b50610de435600281106111d657600080fd5b50610e0435600281106111e857600080fd5b50610e2435600281106111fa57600080fd5b50610e44356002811061120c57600080fd5b50610e64356002811061121e57600080fd5b50610e84356002811061123057600080fd5b50610ea4356002811061124257600080fd5b50610ec4356002811061125457600080fd5b50610ee4356002811061126657600080fd5b50610f04356002811061127857600080fd5b50610f24356002811061128a57600080fd5b50610f44356002811061129c57600080fd5b50610f6435600281106112ae57600080fd5b50610f8435600281106112c057600080fd5b50606051610fa435806040519013156112d857600080fd5b80919012156112e657600080fd5b50611100805160208201209050611480526000610280516020826114c0010152602081019050610260516020826114c0010152602081019050806114c0526114c090508051602082012090506114a0526018611540527f6173646661736466617364666164736661736466616664730000000000000000611560526115408051602001806115a0828460006004600a8704601201f161138457600080fd5b505060206116205261162051611660526115a08051602001806116205161166001828460006004600a8704601201f16113bc57600080fd5b505061162051611660015160206001820306601f8201039050611620516116600161160081516020818352015b83611600511015156113fa57611417565b6000611600516020850101535b81516001018083528114156113e9575b50505050602061162051611660015160206001820306601f8201039050611620510101611620527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961162051611660a16102a0516114a05114151561168957602e611680527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f7420696116a0527f6e20626561636f6e20626c6f636b0000000000000000000000000000000000006116c052611680805160200180611700828460006004600a8704601201f16114e957600080fd5b505060206117a0526117a0516117e0526117008051602001806117a0516117e001828460006004600a8704601201f161152157600080fd5b50506117a0516117e0015160206001820306601f82010390506117a0516117e00161178081516040818352015b836117805110151561155f5761157c565b6000611780516020850101535b815160010180835281141561154e575b5050505060206117a0516117e0015160206001820306601f82010390506117a05101016117a0527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f96117a0516117e0a161026051611800527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611800a161028051611820527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611820a161148051611840527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611840a16114a051611860527fb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b6020611860a15b60206122006109246107e063612477c1611880526000546118a052611480516118c0526118e061018080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506119406101e08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610240516119a052610260516119c0526102a0516119e05280611a00526111c080805160200180846118a001828460006004600a8704601201f161175f57600080fd5b50508051820160206001820306601f82010390506020019150506102e051611a205261030051611a4052611a606103208060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050611d606106208060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050610920516120605261189c90506000305af1611ab457600080fd5b612200511515611c15576023612220527f6661696c656420766572696679696e6720626561636f6e20696e737472756374612240527f696f6e0000000000000000000000000000000000000000000000000000000000612260526122208051602001806122a0828460006004600a8704601201f1611b3157600080fd5b505060206123405261234051612380526122a08051602001806123405161238001828460006004600a8704601201f1611b6957600080fd5b505061234051612380015160206001820306601f8201039050612340516123800161232081516040818352015b8361232051101515611ba757611bc4565b6000612320516020850101535b8151600101808352811415611b96575b50505050602061234051612380015160206001820306601f8201039050612340510101612340527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961234051612380a15b6000610a40516020826123a0010152602081019050610a20516020826123a0010152602081019050806123a0526123a090508051602082012090506114a052610a60516114a051141515611dba57602e612420527f696e737472756374696f6e206d65726b6c6520726f6f74206973206e6f742069612440527f6e2062726964676520626c6f636b000000000000000000000000000000000000612460526124208051602001806124a0828460006004600a8704601201f1611cd657600080fd5b505060206125405261254051612580526124a08051602001806125405161258001828460006004600a8704601201f1611d0e57600080fd5b505061254051612580015160206001820306601f8201039050612540516125800161252081516040818352015b8361252051101515611d4c57611d69565b6000612520516020850101535b8151600101808352811415611d3b575b50505050602061254051612580015160206001820306601f8201039050612540510101612540527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961254051612580a15b6020612f206109246107e063612477c16125a0526001546125c052611480516125e05261260061094080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015250506126606109a08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201525050610a00516126c052610a20516126e052610a605161270052806127205261132080805160200180846125c001828460006004600a8704601201f1611e9057600080fd5b50508051820160206001820306601f8201039050602001915050610aa05161274052610ac05161276052612780610ae08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f602002015280601060200201518260106020020152806011602002015182601160200201528060126020020151826012602002015280601360200201518260136020020152806014602002015182601460200201528060156020020151826015602002015280601660200201518260166020020152806017602002015182601760200201525050612a80610de08060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f6020020152806010602002015182601060200201528060116020020151826011602002015280601260200201518260126020020152806013602002015182601360200201528060146020020151826014602002015280601560200201518260156020020152806016602002015182601660200201528060176020020151826017602002015250506110e051612d80526125bc90506000305af16121e557600080fd5b612f2051151561236c576020612f40527f6661696c6564207665726966792062726964676520696e737472756374696f6e612f6052612f40805160200180612fa0828460006004600a8704601201f161223d57600080fd5b50506020613020526130205161306052612fa08051602001806130205161306001828460006004600a8704601201f161227557600080fd5b505061302051613060015160206001820306601f8201039050613020516130600161300081516020818352015b83613000511015156122b3576122d0565b6000613000516020850101535b81516001018083528114156122a2575b50505050602061302051613060015160206001820306601f8201039050613020510101613020527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961302051613060a16308c379a06130805260206130a05260206130c0527f6661696c6564207665726966792062726964676520696e737472756374696f6e6130e0526130c050600061236b57608461309cfd5b5b6010613120527f6e6f2065786563657074696f6e2e2e2e0000000000000000000000000000000061314052613120805160200180613180828460006004600a8704601201f16123ba57600080fd5b505060206132005261320051613240526131808051602001806132005161324001828460006004600a8704601201f16123f257600080fd5b505061320051613240015160206001820306601f820103905061320051613240016131e081516020818352015b836131e0511015156124305761244d565b60006131e0516020850101535b815160010180835281141561241f575b50505050602061320051613240015160206001820306601f8201039050613200510101613200527f8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f961320051613240a1600160005260206000f3005b63c4ed3f0860005114156124cf5734156124c257600080fd5b60005460005260206000f3005b630f7b9ca160005114156124f55734156124e857600080fd5b60015460005260206000f3005b60006000fd5b6100b46125af036100b46000396100b46125af036000f3`

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

// SwapBeacon is a paid mutator transaction binding the contract method 0x2bb750bc.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeTransactor) SwapBeacon(opts *bind.TransactOpts, newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "swapBeacon", newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0x2bb750bc.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// SwapBeacon is a paid mutator transaction binding the contract method 0x2bb750bc.
//
// Solidity: function swapBeacon(bytes32 newCommRoot, bytes inst, bytes32[3] beaconInstPath, bool[3] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[24] beaconSignerPaths, bool[24] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[3] bridgeInstPath, bool[3] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[24] bridgeSignerPaths, bool[24] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns(bool out)
func (_Bridge *BridgeTransactorSession) SwapBeacon(newCommRoot [32]byte, inst []byte, beaconInstPath [3][32]byte, beaconInstPathIsLeft [3]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [24][32]byte, beaconSignerPathIsLeft [24]bool, beaconSignerPathLen *big.Int, bridgeInstPath [3][32]byte, bridgeInstPathIsLeft [3]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [24][32]byte, bridgeSignerPathIsLeft [24]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SwapBeacon(&_Bridge.TransactOpts, newCommRoot, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
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
