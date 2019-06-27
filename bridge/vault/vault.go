// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault

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

// VaultABI is the input ABI used to generate the binding from.
const VaultABI = "[{\"name\":\"Deposit\",\"inputs\":[{\"type\":\"address\",\"name\":\"_from\",\"indexed\":true},{\"type\":\"string\",\"name\":\"_incognito_address\",\"indexed\":false},{\"type\":\"uint256\",\"name\":\"_amount\",\"indexed\":false,\"unit\":\"wei\"}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"Withdraw\",\"inputs\":[{\"type\":\"address\",\"name\":\"_to\",\"indexed\":true},{\"type\":\"uint256\",\"name\":\"_amount\",\"indexed\":false,\"unit\":\"wei\"}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"address\",\"name\":\"incognitoProxyAddress\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"deposit\",\"outputs\":[],\"inputs\":[{\"type\":\"string\",\"name\":\"incognito_address\"}],\"constant\":false,\"payable\":true,\"type\":\"function\",\"gas\":14608},{\"name\":\"parseBurnInst\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"},{\"type\":\"bytes32\",\"name\":\"out\"},{\"type\":\"address\",\"name\":\"out\"},{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":2527},{\"name\":\"withdraw\",\"outputs\":[],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[4]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[4]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"beaconSignerCount\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[64]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[64]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconSignerPathLen\"},{\"type\":\"bytes32[4]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[4]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"bridgeSignerCount\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[64]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[64]\",\"name\":\"bridgeSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeSignerPathLen\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":154099},{\"name\":\"withdrawed\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"arg0\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":706},{\"name\":\"incognito\",\"outputs\":[{\"type\":\"address\",\"unit\":\"Incognito_proxy\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":603}]"

// VaultBin is the compiled bytecode used for deploying new contracts.
const VaultBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a052602061241a6101403934156100a157600080fd5b602061241a60c03960c05160205181106100ba57600080fd5b506101405160015561240256600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05263a26e118660005114156101b257602060046101403760606004356004016101603760406004356004013511156100c957600080fd5b346102605260406102005261020051610240526101608051602001806102005161024001828460006004600a8704601201f161010457600080fd5b505061020051610240015160206001820306601f820103905061020051610240016101e081516040818352015b836101e0511015156101425761015f565b60006101e0516020850101535b8151600101808352811415610131575b50505050602061020051610240015160206001820306601f820103905061020051010161020052337f2d4b597935f3cd67fb2eebf1db4debc934cee5c7baa7153f980fdbeb2e74084e61020051610240a2005b637e16e6e160005114156103f357602060046101403734156101d357600080fd5b60986004356004016101603760786004356004013511156101f357600080fd5b60006003602082066103200161016051828401111561021157600080fd5b607880610340826020602088068803016101600160006004601ef150508181528090509050905080602001516000825180602090131561025057600080fd5b809190121561025e57600080fd5b806020036101000a820490509050905061022052610160805160036020820381131561028957600080fd5b602081066020820481156102c357816020036101000a6001820160200260200186015104826101000a8260200260200187015102016102ce565b806020026020018501515b905090509050905090506104005261016080516023602082038113156102f357600080fd5b6020810660208204811561032d57816020036101000a6001820160200260200186015104826101000a826020026020018701510201610338565b806020026020018501515b90509050905090509050602051811061035057600080fd5b61042052610160805160376020820381131561036b57600080fd5b602081066020820481156103a557816020036101000a6001820160200260200186015104826101000a8260200260200187015102016103b0565b806020026020018501515b90509050905090509050610440526080610460526104806102205181526104005181602001526104205181604001526104405181606001525061046051610480f3005b63922041c760005114156122d257612420600461014037341561041557600080fd5b609860043560040161256037607860043560040135111561043557600080fd5b60a4356002811061044557600080fd5b5060c4356002811061045657600080fd5b5060e4356002811061046757600080fd5b50610104356002811061047957600080fd5b50606051610124358060405190131561049157600080fd5b809190121561049f57600080fd5b506102306101a435600401612620376102106101a4356004013511156104c457600080fd5b6060516101c435806040519013156104db57600080fd5b80919012156104e957600080fd5b50610a0435600281106104fb57600080fd5b50610a24356002811061050d57600080fd5b50610a44356002811061051f57600080fd5b50610a64356002811061053157600080fd5b50610a84356002811061054357600080fd5b50610aa4356002811061055557600080fd5b50610ac4356002811061056757600080fd5b50610ae4356002811061057957600080fd5b50610b04356002811061058b57600080fd5b50610b24356002811061059d57600080fd5b50610b4435600281106105af57600080fd5b50610b6435600281106105c157600080fd5b50610b8435600281106105d357600080fd5b50610ba435600281106105e557600080fd5b50610bc435600281106105f757600080fd5b50610be4356002811061060957600080fd5b50610c04356002811061061b57600080fd5b50610c24356002811061062d57600080fd5b50610c44356002811061063f57600080fd5b50610c64356002811061065157600080fd5b50610c84356002811061066357600080fd5b50610ca4356002811061067557600080fd5b50610cc4356002811061068757600080fd5b50610ce4356002811061069957600080fd5b50610d0435600281106106ab57600080fd5b50610d2435600281106106bd57600080fd5b50610d4435600281106106cf57600080fd5b50610d6435600281106106e157600080fd5b50610d8435600281106106f357600080fd5b50610da4356002811061070557600080fd5b50610dc4356002811061071757600080fd5b50610de4356002811061072957600080fd5b50610e04356002811061073b57600080fd5b50610e24356002811061074d57600080fd5b50610e44356002811061075f57600080fd5b50610e64356002811061077157600080fd5b50610e84356002811061078357600080fd5b50610ea4356002811061079557600080fd5b50610ec435600281106107a757600080fd5b50610ee435600281106107b957600080fd5b50610f0435600281106107cb57600080fd5b50610f2435600281106107dd57600080fd5b50610f4435600281106107ef57600080fd5b50610f64356002811061080157600080fd5b50610f84356002811061081357600080fd5b50610fa4356002811061082557600080fd5b50610fc4356002811061083757600080fd5b50610fe4356002811061084957600080fd5b50611004356002811061085b57600080fd5b50611024356002811061086d57600080fd5b50611044356002811061087f57600080fd5b50611064356002811061089157600080fd5b5061108435600281106108a357600080fd5b506110a435600281106108b557600080fd5b506110c435600281106108c757600080fd5b506110e435600281106108d957600080fd5b5061110435600281106108eb57600080fd5b5061112435600281106108fd57600080fd5b50611144356002811061090f57600080fd5b50611164356002811061092157600080fd5b50611184356002811061093357600080fd5b506111a4356002811061094557600080fd5b506111c4356002811061095757600080fd5b506111e4356002811061096957600080fd5b50606051611204358060405190131561098157600080fd5b809190121561098f57600080fd5b506112a435600281106109a157600080fd5b506112c435600281106109b357600080fd5b506112e435600281106109c557600080fd5b5061130435600281106109d757600080fd5b5060605161132435806040519013156109ef57600080fd5b80919012156109fd57600080fd5b506102306113a435600401612880376102106113a435600401351115610a2257600080fd5b6060516113c43580604051901315610a3957600080fd5b8091901215610a4757600080fd5b50611c043560028110610a5957600080fd5b50611c243560028110610a6b57600080fd5b50611c443560028110610a7d57600080fd5b50611c643560028110610a8f57600080fd5b50611c843560028110610aa157600080fd5b50611ca43560028110610ab357600080fd5b50611cc43560028110610ac557600080fd5b50611ce43560028110610ad757600080fd5b50611d043560028110610ae957600080fd5b50611d243560028110610afb57600080fd5b50611d443560028110610b0d57600080fd5b50611d643560028110610b1f57600080fd5b50611d843560028110610b3157600080fd5b50611da43560028110610b4357600080fd5b50611dc43560028110610b5557600080fd5b50611de43560028110610b6757600080fd5b50611e043560028110610b7957600080fd5b50611e243560028110610b8b57600080fd5b50611e443560028110610b9d57600080fd5b50611e643560028110610baf57600080fd5b50611e843560028110610bc157600080fd5b50611ea43560028110610bd357600080fd5b50611ec43560028110610be557600080fd5b50611ee43560028110610bf757600080fd5b50611f043560028110610c0957600080fd5b50611f243560028110610c1b57600080fd5b50611f443560028110610c2d57600080fd5b50611f643560028110610c3f57600080fd5b50611f843560028110610c5157600080fd5b50611fa43560028110610c6357600080fd5b50611fc43560028110610c7557600080fd5b50611fe43560028110610c8757600080fd5b506120043560028110610c9957600080fd5b506120243560028110610cab57600080fd5b506120443560028110610cbd57600080fd5b506120643560028110610ccf57600080fd5b506120843560028110610ce157600080fd5b506120a43560028110610cf357600080fd5b506120c43560028110610d0557600080fd5b506120e43560028110610d1757600080fd5b506121043560028110610d2957600080fd5b506121243560028110610d3b57600080fd5b506121443560028110610d4d57600080fd5b506121643560028110610d5f57600080fd5b506121843560028110610d7157600080fd5b506121a43560028110610d8357600080fd5b506121c43560028110610d9557600080fd5b506121e43560028110610da757600080fd5b506122043560028110610db957600080fd5b506122243560028110610dcb57600080fd5b506122443560028110610ddd57600080fd5b506122643560028110610def57600080fd5b506122843560028110610e0157600080fd5b506122a43560028110610e1357600080fd5b506122c43560028110610e2557600080fd5b506122e43560028110610e3757600080fd5b506123043560028110610e4957600080fd5b506123243560028110610e5b57600080fd5b506123443560028110610e6d57600080fd5b506123643560028110610e7f57600080fd5b506123843560028110610e9157600080fd5b506123a43560028110610ea357600080fd5b506123c43560028110610eb557600080fd5b506123e43560028110610ec757600080fd5b506060516124043580604051901315610edf57600080fd5b8091901215610eed57600080fd5b506000612ae0526000612b40526080612c8060c46020637e16e6e1612b605280612b80526125608080516020018084612b8001828460006004600a8704601201f1610f3757600080fd5b50508051820160206001820306601f8201039050602001915050612b7c90506000305af1610f6457600080fd5b612c808051612ae0526020810151612b00526040810151612b20526060810151612b405250612560805160208201209050612d00526000612d005160e05260c052604060c0205415610fb557600080fd5b6001543b610fc257600080fd5b6001543018610fd057600080fd5b60206156206128a46124206354b578c0612d2052612d0051612d4052612d60610160806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201525050612de06101e080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015280600360200201518260036020020152505061026051612e605261028051612e80526102a051612ea0526102c051612ec05280612ee0526126208080516020018084612d4001828460006004600a8704601201f16110c757600080fd5b50508051820160206001820306601f820103905060200191505061030051612f005261032051612f2052612f406103408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f60200201525050613740610b408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f6020020152505061134051613f4052613f60611360806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201525050613fe06113e0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201525050611460516140605261148051614080526114a0516140a0526114c0516140c052806140e0526128808080516020018084612d4001828460006004600a8704601201f16119e457600080fd5b50508051820160206001820306601f8201039050602001915050611500516141005261152051614120526141406115408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f60200201525050614940611d408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f602002015250506125405161514052612d3c905060006001545af161223b57600080fd5b600050615620511561225c5760016000612d005160e05260c052604060c020555b612b40513031101561226d57600080fd5b60016000612d005160e05260c052604060c020556000600060006000612b4051612b20516000f161229d57600080fd5b612b405161564052612b20517f884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a94243646020615640a2005b63dca40d9e600051141561230f57602060046101403734156122f357600080fd5b60006101405160e05260c052604060c0205460005260206000f3005b638a984538600051141561233557341561232857600080fd5b60015460005260206000f3005b60006000fd5b6100c7612402036100c76000396100c7612402036000f3`

// DeployVault deploys a new Ethereum contract, binding an instance of Vault to it.
func DeployVault(auth *bind.TransactOpts, backend bind.ContractBackend, incognitoProxyAddress common.Address) (common.Address, *types.Transaction, *Vault, error) {
	parsed, err := abi.JSON(strings.NewReader(VaultABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(VaultBin), backend, incognitoProxyAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Vault{VaultCaller: VaultCaller{contract: contract}, VaultTransactor: VaultTransactor{contract: contract}, VaultFilterer: VaultFilterer{contract: contract}}, nil
}

// Vault is an auto generated Go binding around an Ethereum contract.
type Vault struct {
	VaultCaller     // Read-only binding to the contract
	VaultTransactor // Write-only binding to the contract
	VaultFilterer   // Log filterer for contract events
}

// VaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type VaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VaultSession struct {
	Contract     *Vault            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VaultCallerSession struct {
	Contract *VaultCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// VaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VaultTransactorSession struct {
	Contract     *VaultTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type VaultRaw struct {
	Contract *Vault // Generic contract binding to access the raw methods on
}

// VaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VaultCallerRaw struct {
	Contract *VaultCaller // Generic read-only contract binding to access the raw methods on
}

// VaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VaultTransactorRaw struct {
	Contract *VaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVault creates a new instance of Vault, bound to a specific deployed contract.
func NewVault(address common.Address, backend bind.ContractBackend) (*Vault, error) {
	contract, err := bindVault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Vault{VaultCaller: VaultCaller{contract: contract}, VaultTransactor: VaultTransactor{contract: contract}, VaultFilterer: VaultFilterer{contract: contract}}, nil
}

// NewVaultCaller creates a new read-only instance of Vault, bound to a specific deployed contract.
func NewVaultCaller(address common.Address, caller bind.ContractCaller) (*VaultCaller, error) {
	contract, err := bindVault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VaultCaller{contract: contract}, nil
}

// NewVaultTransactor creates a new write-only instance of Vault, bound to a specific deployed contract.
func NewVaultTransactor(address common.Address, transactor bind.ContractTransactor) (*VaultTransactor, error) {
	contract, err := bindVault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VaultTransactor{contract: contract}, nil
}

// NewVaultFilterer creates a new log filterer instance of Vault, bound to a specific deployed contract.
func NewVaultFilterer(address common.Address, filterer bind.ContractFilterer) (*VaultFilterer, error) {
	contract, err := bindVault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VaultFilterer{contract: contract}, nil
}

// bindVault binds a generic wrapper to an already deployed contract.
func bindVault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(VaultABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.VaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transact(opts, method, params...)
}

// Incognito is a free data retrieval call binding the contract method 0x8a984538.
//
// Solidity: function incognito() constant returns(address out)
func (_Vault *VaultCaller) Incognito(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Vault.contract.Call(opts, out, "incognito")
	return *ret0, err
}

// Incognito is a free data retrieval call binding the contract method 0x8a984538.
//
// Solidity: function incognito() constant returns(address out)
func (_Vault *VaultSession) Incognito() (common.Address, error) {
	return _Vault.Contract.Incognito(&_Vault.CallOpts)
}

// Incognito is a free data retrieval call binding the contract method 0x8a984538.
//
// Solidity: function incognito() constant returns(address out)
func (_Vault *VaultCallerSession) Incognito() (common.Address, error) {
	return _Vault.Contract.Incognito(&_Vault.CallOpts)
}

// ParseBurnInst is a free data retrieval call binding the contract method 0x7e16e6e1.
//
// Solidity: function parseBurnInst(bytes inst) constant returns(uint256 out, bytes32 out, address out, uint256 out)
func (_Vault *VaultCaller) ParseBurnInst(opts *bind.CallOpts, inst []byte) (*big.Int, [32]byte, common.Address, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new([32]byte)
		ret2 = new(common.Address)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Vault.contract.Call(opts, out, "parseBurnInst", inst)
	return *ret0, *ret1, *ret2, *ret3, err
}

// ParseBurnInst is a free data retrieval call binding the contract method 0x7e16e6e1.
//
// Solidity: function parseBurnInst(bytes inst) constant returns(uint256 out, bytes32 out, address out, uint256 out)
func (_Vault *VaultSession) ParseBurnInst(inst []byte) (*big.Int, [32]byte, common.Address, *big.Int, error) {
	return _Vault.Contract.ParseBurnInst(&_Vault.CallOpts, inst)
}

// ParseBurnInst is a free data retrieval call binding the contract method 0x7e16e6e1.
//
// Solidity: function parseBurnInst(bytes inst) constant returns(uint256 out, bytes32 out, address out, uint256 out)
func (_Vault *VaultCallerSession) ParseBurnInst(inst []byte) (*big.Int, [32]byte, common.Address, *big.Int, error) {
	return _Vault.Contract.ParseBurnInst(&_Vault.CallOpts, inst)
}

// Withdrawed is a free data retrieval call binding the contract method 0xdca40d9e.
//
// Solidity: function withdrawed(bytes32 arg0) constant returns(bool out)
func (_Vault *VaultCaller) Withdrawed(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Vault.contract.Call(opts, out, "withdrawed", arg0)
	return *ret0, err
}

// Withdrawed is a free data retrieval call binding the contract method 0xdca40d9e.
//
// Solidity: function withdrawed(bytes32 arg0) constant returns(bool out)
func (_Vault *VaultSession) Withdrawed(arg0 [32]byte) (bool, error) {
	return _Vault.Contract.Withdrawed(&_Vault.CallOpts, arg0)
}

// Withdrawed is a free data retrieval call binding the contract method 0xdca40d9e.
//
// Solidity: function withdrawed(bytes32 arg0) constant returns(bool out)
func (_Vault *VaultCallerSession) Withdrawed(arg0 [32]byte) (bool, error) {
	return _Vault.Contract.Withdrawed(&_Vault.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0xa26e1186.
//
// Solidity: function deposit(string incognito_address) returns()
func (_Vault *VaultTransactor) Deposit(opts *bind.TransactOpts, incognito_address string) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "deposit", incognito_address)
}

// Deposit is a paid mutator transaction binding the contract method 0xa26e1186.
//
// Solidity: function deposit(string incognito_address) returns()
func (_Vault *VaultSession) Deposit(incognito_address string) (*types.Transaction, error) {
	return _Vault.Contract.Deposit(&_Vault.TransactOpts, incognito_address)
}

// Deposit is a paid mutator transaction binding the contract method 0xa26e1186.
//
// Solidity: function deposit(string incognito_address) returns()
func (_Vault *VaultTransactorSession) Deposit(incognito_address string) (*types.Transaction, error) {
	return _Vault.Contract.Deposit(&_Vault.TransactOpts, incognito_address)
}

// Withdraw is a paid mutator transaction binding the contract method 0x922041c7.
//
// Solidity: function withdraw(bytes inst, bytes32[4] beaconInstPath, bool[4] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[64] beaconSignerPaths, bool[64] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[4] bridgeInstPath, bool[4] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[64] bridgeSignerPaths, bool[64] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns()
func (_Vault *VaultTransactor) Withdraw(opts *bind.TransactOpts, inst []byte, beaconInstPath [4][32]byte, beaconInstPathIsLeft [4]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [64][32]byte, beaconSignerPathIsLeft [64]bool, beaconSignerPathLen *big.Int, bridgeInstPath [4][32]byte, bridgeInstPathIsLeft [4]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [64][32]byte, bridgeSignerPathIsLeft [64]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "withdraw", inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// Withdraw is a paid mutator transaction binding the contract method 0x922041c7.
//
// Solidity: function withdraw(bytes inst, bytes32[4] beaconInstPath, bool[4] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[64] beaconSignerPaths, bool[64] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[4] bridgeInstPath, bool[4] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[64] bridgeSignerPaths, bool[64] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns()
func (_Vault *VaultSession) Withdraw(inst []byte, beaconInstPath [4][32]byte, beaconInstPathIsLeft [4]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [64][32]byte, beaconSignerPathIsLeft [64]bool, beaconSignerPathLen *big.Int, bridgeInstPath [4][32]byte, bridgeInstPathIsLeft [4]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [64][32]byte, bridgeSignerPathIsLeft [64]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Vault.Contract.Withdraw(&_Vault.TransactOpts, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// Withdraw is a paid mutator transaction binding the contract method 0x922041c7.
//
// Solidity: function withdraw(bytes inst, bytes32[4] beaconInstPath, bool[4] beaconInstPathIsLeft, int128 beaconInstPathLen, bytes32 beaconInstRoot, bytes32 beaconBlkData, bytes32 beaconBlkHash, bytes beaconSignerPubkeys, int128 beaconSignerCount, bytes32 beaconSignerSig, bytes32[64] beaconSignerPaths, bool[64] beaconSignerPathIsLeft, int128 beaconSignerPathLen, bytes32[4] bridgeInstPath, bool[4] bridgeInstPathIsLeft, int128 bridgeInstPathLen, bytes32 bridgeInstRoot, bytes32 bridgeBlkData, bytes32 bridgeBlkHash, bytes bridgeSignerPubkeys, int128 bridgeSignerCount, bytes32 bridgeSignerSig, bytes32[64] bridgeSignerPaths, bool[64] bridgeSignerPathIsLeft, int128 bridgeSignerPathLen) returns()
func (_Vault *VaultTransactorSession) Withdraw(inst []byte, beaconInstPath [4][32]byte, beaconInstPathIsLeft [4]bool, beaconInstPathLen *big.Int, beaconInstRoot [32]byte, beaconBlkData [32]byte, beaconBlkHash [32]byte, beaconSignerPubkeys []byte, beaconSignerCount *big.Int, beaconSignerSig [32]byte, beaconSignerPaths [64][32]byte, beaconSignerPathIsLeft [64]bool, beaconSignerPathLen *big.Int, bridgeInstPath [4][32]byte, bridgeInstPathIsLeft [4]bool, bridgeInstPathLen *big.Int, bridgeInstRoot [32]byte, bridgeBlkData [32]byte, bridgeBlkHash [32]byte, bridgeSignerPubkeys []byte, bridgeSignerCount *big.Int, bridgeSignerSig [32]byte, bridgeSignerPaths [64][32]byte, bridgeSignerPathIsLeft [64]bool, bridgeSignerPathLen *big.Int) (*types.Transaction, error) {
	return _Vault.Contract.Withdraw(&_Vault.TransactOpts, inst, beaconInstPath, beaconInstPathIsLeft, beaconInstPathLen, beaconInstRoot, beaconBlkData, beaconBlkHash, beaconSignerPubkeys, beaconSignerCount, beaconSignerSig, beaconSignerPaths, beaconSignerPathIsLeft, beaconSignerPathLen, bridgeInstPath, bridgeInstPathIsLeft, bridgeInstPathLen, bridgeInstRoot, bridgeBlkData, bridgeBlkHash, bridgeSignerPubkeys, bridgeSignerCount, bridgeSignerSig, bridgeSignerPaths, bridgeSignerPathIsLeft, bridgeSignerPathLen)
}

// VaultDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Vault contract.
type VaultDepositIterator struct {
	Event *VaultDeposit // Event containing the contract specifics and raw log

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
func (it *VaultDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultDeposit)
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
		it.Event = new(VaultDeposit)
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
func (it *VaultDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultDeposit represents a Deposit event raised by the Vault contract.
type VaultDeposit struct {
	From             common.Address
	IncognitoAddress string
	Amount           *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x2d4b597935f3cd67fb2eebf1db4debc934cee5c7baa7153f980fdbeb2e74084e.
//
// Solidity: event Deposit(address indexed _from, string _incognito_address, uint256 _amount)
func (_Vault *VaultFilterer) FilterDeposit(opts *bind.FilterOpts, _from []common.Address) (*VaultDepositIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "Deposit", _fromRule)
	if err != nil {
		return nil, err
	}
	return &VaultDepositIterator{contract: _Vault.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x2d4b597935f3cd67fb2eebf1db4debc934cee5c7baa7153f980fdbeb2e74084e.
//
// Solidity: event Deposit(address indexed _from, string _incognito_address, uint256 _amount)
func (_Vault *VaultFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *VaultDeposit, _from []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "Deposit", _fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultDeposit)
				if err := _Vault.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// VaultWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the Vault contract.
type VaultWithdrawIterator struct {
	Event *VaultWithdraw // Event containing the contract specifics and raw log

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
func (it *VaultWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultWithdraw)
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
		it.Event = new(VaultWithdraw)
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
func (it *VaultWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultWithdraw represents a Withdraw event raised by the Vault contract.
type VaultWithdraw struct {
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed _to, uint256 _amount)
func (_Vault *VaultFilterer) FilterWithdraw(opts *bind.FilterOpts, _to []common.Address) (*VaultWithdrawIterator, error) {

	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "Withdraw", _toRule)
	if err != nil {
		return nil, err
	}
	return &VaultWithdrawIterator{contract: _Vault.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed _to, uint256 _amount)
func (_Vault *VaultFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *VaultWithdraw, _to []common.Address) (event.Subscription, error) {

	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "Withdraw", _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultWithdraw)
				if err := _Vault.contract.UnpackLog(event, "Withdraw", log); err != nil {
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
