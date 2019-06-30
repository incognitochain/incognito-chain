// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault

import (
	"math/big"
	"strings"

	ethereum "github.com/incognitochain/incognito-chain/ethrelaying"
	"github.com/incognitochain/incognito-chain/ethrelaying/accounts/abi"
	"github.com/incognitochain/incognito-chain/ethrelaying/accounts/abi/bind"
	"github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/core/types"
	"github.com/incognitochain/incognito-chain/ethrelaying/event"
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
const VaultABI = "[{\"name\":\"Deposit\",\"inputs\":[{\"type\":\"address\",\"name\":\"_from\",\"indexed\":true},{\"type\":\"string\",\"name\":\"_incognito_address\",\"indexed\":false},{\"type\":\"uint256\",\"name\":\"_amount\",\"indexed\":false,\"unit\":\"wei\"}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"Withdraw\",\"inputs\":[{\"type\":\"address\",\"name\":\"_to\",\"indexed\":true},{\"type\":\"uint256\",\"name\":\"_amount\",\"indexed\":false,\"unit\":\"wei\"}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyString\",\"inputs\":[{\"type\":\"string\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBytes32\",\"inputs\":[{\"type\":\"bytes32\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyBool\",\"inputs\":[{\"type\":\"bool\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyUint256\",\"inputs\":[{\"type\":\"uint256\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"name\":\"NotifyAddress\",\"inputs\":[{\"type\":\"address\",\"name\":\"content\",\"indexed\":false}],\"anonymous\":false,\"type\":\"event\"},{\"outputs\":[],\"inputs\":[{\"type\":\"address\",\"name\":\"incognitoProxyAddress\"}],\"constant\":false,\"payable\":false,\"type\":\"constructor\"},{\"name\":\"deposit\",\"outputs\":[],\"inputs\":[{\"type\":\"string\",\"name\":\"incognito_address\"}],\"constant\":false,\"payable\":true,\"type\":\"function\",\"gas\":25634},{\"name\":\"parseBurnInst\",\"outputs\":[{\"type\":\"uint256\",\"name\":\"out\"},{\"type\":\"bytes32\",\"name\":\"out\"},{\"type\":\"address\",\"name\":\"out\"},{\"type\":\"uint256\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":2543},{\"name\":\"testExtract\",\"outputs\":[{\"type\":\"address\",\"name\":\"out\"},{\"type\":\"uint256\",\"unit\":\"wei\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes\",\"name\":\"a\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":743},{\"name\":\"withdraw\",\"outputs\":[],\"inputs\":[{\"type\":\"bytes\",\"name\":\"inst\"},{\"type\":\"bytes32[4]\",\"name\":\"beaconInstPath\"},{\"type\":\"bool[4]\",\"name\":\"beaconInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"beaconInstRoot\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkData\"},{\"type\":\"bytes32\",\"name\":\"beaconBlkHash\"},{\"type\":\"bytes\",\"name\":\"beaconSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"beaconSignerCount\"},{\"type\":\"bytes32\",\"name\":\"beaconSignerSig\"},{\"type\":\"bytes32[64]\",\"name\":\"beaconSignerPaths\"},{\"type\":\"bool[64]\",\"name\":\"beaconSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"beaconSignerPathLen\"},{\"type\":\"bytes32[4]\",\"name\":\"bridgeInstPath\"},{\"type\":\"bool[4]\",\"name\":\"bridgeInstPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeInstPathLen\"},{\"type\":\"bytes32\",\"name\":\"bridgeInstRoot\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkData\"},{\"type\":\"bytes32\",\"name\":\"bridgeBlkHash\"},{\"type\":\"bytes\",\"name\":\"bridgeSignerPubkeys\"},{\"type\":\"int128\",\"name\":\"bridgeSignerCount\"},{\"type\":\"bytes32\",\"name\":\"bridgeSignerSig\"},{\"type\":\"bytes32[64]\",\"name\":\"bridgeSignerPaths\"},{\"type\":\"bool[64]\",\"name\":\"bridgeSignerPathIsLeft\"},{\"type\":\"int128\",\"name\":\"bridgeSignerPathLen\"}],\"constant\":false,\"payable\":false,\"type\":\"function\",\"gas\":118634},{\"name\":\"withdrawed\",\"outputs\":[{\"type\":\"bool\",\"name\":\"out\"}],\"inputs\":[{\"type\":\"bytes32\",\"name\":\"arg0\"}],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":736},{\"name\":\"incognito\",\"outputs\":[{\"type\":\"address\",\"unit\":\"Incognito_proxy\",\"name\":\"out\"}],\"inputs\":[],\"constant\":true,\"payable\":false,\"type\":\"function\",\"gas\":633}]"

// VaultBin is the compiled bytecode used for deploying new contracts.
const VaultBin = `0x740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05260206125056101403934156100a157600080fd5b602061250560c03960c05160205181106100ba57600080fd5b50610140516001556124ed56600035601c52740100000000000000000000000000000000000000006020526f7fffffffffffffffffffffffffffffff6040527fffffffffffffffffffffffffffffffff8000000000000000000000000000000060605274012a05f1fffffffffffffffffffffffffdabf41c006080527ffffffffffffffffffffffffed5fa0e000000000000000000000000000000000060a05263a26e118660005114156101b257602060046101403760a06004356004016101603760806004356004013511156100c957600080fd5b346102a05260406102405261024051610280526101608051602001806102405161028001828460006004600a8704601201f161010457600080fd5b505061024051610280015160206001820306601f8201039050610240516102800161022081516080818352015b83610220511015156101425761015f565b6000610220516020850101535b8151600101808352811415610131575b50505050602061024051610280015160206001820306601f820103905061024051010161024052337f2d4b597935f3cd67fb2eebf1db4debc934cee5c7baa7153f980fdbeb2e74084e61024051610280a2005b637e16e6e160005114156103f357602060046101403734156101d357600080fd5b60b66004356004016101603760966004356004013511156101f357600080fd5b60006003602082066103600161016051828401111561021157600080fd5b6096806103808260206020880688030161016001600060046021f150508181528090509050905080602001516000825180602090131561025057600080fd5b809190121561025e57600080fd5b806020036101000a820490509050905061024052610160805160036020820381131561028957600080fd5b602081066020820481156102c357816020036101000a6001820160200260200186015104826101000a8260200260200187015102016102ce565b806020026020018501515b905090509050905090506104605261016080516023602082038113156102f357600080fd5b6020810660208204811561032d57816020036101000a6001820160200260200186015104826101000a826020026020018701510201610338565b806020026020018501515b90509050905090509050602051811061035057600080fd5b61048052610160805160436020820381131561036b57600080fd5b602081066020820481156103a557816020036101000a6001820160200260200186015104826101000a8260200260200187015102016103b0565b806020026020018501515b905090509050905090506104a05260806104c0526104e06102405181526104605181602001526104805181604001526104a0518160600152506104c0516104e0f3005b63ac68730860005114156104a2576020600461014037341561041457600080fd5b60b660043560040161016037609660043560040135111561043457600080fd5b61016060206000602083510381131561044c57600080fd5b046020026020018101519050602051811061046657600080fd5b6102405261303961026052633b9aca0061026051026102805260406102a0526102c0610240518152610280518160200152506102a0516102c0f3005b63922041c760005114156123bd5761242060046101403734156104c457600080fd5b60b66004356004016125603760966004356004013511156104e457600080fd5b60a435600281106104f457600080fd5b5060c4356002811061050557600080fd5b5060e4356002811061051657600080fd5b50610104356002811061052857600080fd5b50606051610124358060405190131561054057600080fd5b809190121561054e57600080fd5b506102306101a435600401612640376102106101a43560040135111561057357600080fd5b6060516101c4358060405190131561058a57600080fd5b809190121561059857600080fd5b50610a0435600281106105aa57600080fd5b50610a2435600281106105bc57600080fd5b50610a4435600281106105ce57600080fd5b50610a6435600281106105e057600080fd5b50610a8435600281106105f257600080fd5b50610aa4356002811061060457600080fd5b50610ac4356002811061061657600080fd5b50610ae4356002811061062857600080fd5b50610b04356002811061063a57600080fd5b50610b24356002811061064c57600080fd5b50610b44356002811061065e57600080fd5b50610b64356002811061067057600080fd5b50610b84356002811061068257600080fd5b50610ba4356002811061069457600080fd5b50610bc435600281106106a657600080fd5b50610be435600281106106b857600080fd5b50610c0435600281106106ca57600080fd5b50610c2435600281106106dc57600080fd5b50610c4435600281106106ee57600080fd5b50610c64356002811061070057600080fd5b50610c84356002811061071257600080fd5b50610ca4356002811061072457600080fd5b50610cc4356002811061073657600080fd5b50610ce4356002811061074857600080fd5b50610d04356002811061075a57600080fd5b50610d24356002811061076c57600080fd5b50610d44356002811061077e57600080fd5b50610d64356002811061079057600080fd5b50610d8435600281106107a257600080fd5b50610da435600281106107b457600080fd5b50610dc435600281106107c657600080fd5b50610de435600281106107d857600080fd5b50610e0435600281106107ea57600080fd5b50610e2435600281106107fc57600080fd5b50610e44356002811061080e57600080fd5b50610e64356002811061082057600080fd5b50610e84356002811061083257600080fd5b50610ea4356002811061084457600080fd5b50610ec4356002811061085657600080fd5b50610ee4356002811061086857600080fd5b50610f04356002811061087a57600080fd5b50610f24356002811061088c57600080fd5b50610f44356002811061089e57600080fd5b50610f6435600281106108b057600080fd5b50610f8435600281106108c257600080fd5b50610fa435600281106108d457600080fd5b50610fc435600281106108e657600080fd5b50610fe435600281106108f857600080fd5b50611004356002811061090a57600080fd5b50611024356002811061091c57600080fd5b50611044356002811061092e57600080fd5b50611064356002811061094057600080fd5b50611084356002811061095257600080fd5b506110a4356002811061096457600080fd5b506110c4356002811061097657600080fd5b506110e4356002811061098857600080fd5b50611104356002811061099a57600080fd5b5061112435600281106109ac57600080fd5b5061114435600281106109be57600080fd5b5061116435600281106109d057600080fd5b5061118435600281106109e257600080fd5b506111a435600281106109f457600080fd5b506111c43560028110610a0657600080fd5b506111e43560028110610a1857600080fd5b506060516112043580604051901315610a3057600080fd5b8091901215610a3e57600080fd5b506112a43560028110610a5057600080fd5b506112c43560028110610a6257600080fd5b506112e43560028110610a7457600080fd5b506113043560028110610a8657600080fd5b506060516113243580604051901315610a9e57600080fd5b8091901215610aac57600080fd5b506102306113a4356004016128a0376102106113a435600401351115610ad157600080fd5b6060516113c43580604051901315610ae857600080fd5b8091901215610af657600080fd5b50611c043560028110610b0857600080fd5b50611c243560028110610b1a57600080fd5b50611c443560028110610b2c57600080fd5b50611c643560028110610b3e57600080fd5b50611c843560028110610b5057600080fd5b50611ca43560028110610b6257600080fd5b50611cc43560028110610b7457600080fd5b50611ce43560028110610b8657600080fd5b50611d043560028110610b9857600080fd5b50611d243560028110610baa57600080fd5b50611d443560028110610bbc57600080fd5b50611d643560028110610bce57600080fd5b50611d843560028110610be057600080fd5b50611da43560028110610bf257600080fd5b50611dc43560028110610c0457600080fd5b50611de43560028110610c1657600080fd5b50611e043560028110610c2857600080fd5b50611e243560028110610c3a57600080fd5b50611e443560028110610c4c57600080fd5b50611e643560028110610c5e57600080fd5b50611e843560028110610c7057600080fd5b50611ea43560028110610c8257600080fd5b50611ec43560028110610c9457600080fd5b50611ee43560028110610ca657600080fd5b50611f043560028110610cb857600080fd5b50611f243560028110610cca57600080fd5b50611f443560028110610cdc57600080fd5b50611f643560028110610cee57600080fd5b50611f843560028110610d0057600080fd5b50611fa43560028110610d1257600080fd5b50611fc43560028110610d2457600080fd5b50611fe43560028110610d3657600080fd5b506120043560028110610d4857600080fd5b506120243560028110610d5a57600080fd5b506120443560028110610d6c57600080fd5b506120643560028110610d7e57600080fd5b506120843560028110610d9057600080fd5b506120a43560028110610da257600080fd5b506120c43560028110610db457600080fd5b506120e43560028110610dc657600080fd5b506121043560028110610dd857600080fd5b506121243560028110610dea57600080fd5b506121443560028110610dfc57600080fd5b506121643560028110610e0e57600080fd5b506121843560028110610e2057600080fd5b506121a43560028110610e3257600080fd5b506121c43560028110610e4457600080fd5b506121e43560028110610e5657600080fd5b506122043560028110610e6857600080fd5b506122243560028110610e7a57600080fd5b506122443560028110610e8c57600080fd5b506122643560028110610e9e57600080fd5b506122843560028110610eb057600080fd5b506122a43560028110610ec257600080fd5b506122c43560028110610ed457600080fd5b506122e43560028110610ee657600080fd5b506123043560028110610ef857600080fd5b506123243560028110610f0a57600080fd5b506123443560028110610f1c57600080fd5b506123643560028110610f2e57600080fd5b506123843560028110610f4057600080fd5b506123a43560028110610f5257600080fd5b506123c43560028110610f6457600080fd5b506123e43560028110610f7657600080fd5b506060516124043580604051901315610f8e57600080fd5b8091901215610f9c57600080fd5b506000612b00526000612b60526080612cc060e46020637e16e6e1612b805280612ba0526125608080516020018084612ba001828460006004600a8704601201f1610fe657600080fd5b50508051820160206001820306601f8201039050602001915050612b9c90506000305af161101357600080fd5b612cc08051612b00526020810151612b20526040810151612b40526060810151612b60525062373230612b00511461104a57600080fd5b7f0500000000000000000000000000000000000000000000000000000000000000612b20511461107957600080fd5b612560805160208201209050612d40526000612d405160e05260c052604060c02054156110a557600080fd5b6001543b6110b257600080fd5b60015430186110c057600080fd5b60206156606128a46124206354b578c0612d6052612d4051612d8052612da0610160806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201525050612e206101e080600060200201518260006020020152806001602002015182600160200201528060026020020151826002602002015280600360200201518260036020020152505061026051612ea05261028051612ec0526102a051612ee0526102c051612f005280612f20526126408080516020018084612d8001828460006004600a8704601201f16111b757600080fd5b50508051820160206001820306601f820103905060200191505061030051612f405261032051612f6052612f806103408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f60200201525050613780610b408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f6020020152505061134051613f8052613fa06113608060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015250506140206113e0806000602002015182600060200201528060016020020151826001602002015280600260200201518260026020020152806003602002015182600360200201525050611460516140a052611480516140c0526114a0516140e0526114c0516141005280614120526128a08080516020018084612d8001828460006004600a8704601201f1611ad457600080fd5b50508051820160206001820306601f8201039050602001915050611500516141405261152051614160526141806115408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f60200201525050614980611d408060006020020151826000602002015280600160200201518260016020020152806002602002015182600260200201528060036020020151826003602002015280600460200201518260046020020152806005602002015182600560200201528060066020020151826006602002015280600760200201518260076020020152806008602002015182600860200201528060096020020151826009602002015280600a602002015182600a602002015280600b602002015182600b602002015280600c602002015182600c602002015280600d602002015182600d602002015280600e602002015182600e602002015280600f602002015182600f60200201528060106020020151826010602002015280601160200201518260116020020152806012602002015182601260200201528060136020020151826013602002015280601460200201518260146020020152806015602002015182601560200201528060166020020151826016602002015280601760200201518260176020020152806018602002015182601860200201528060196020020151826019602002015280601a602002015182601a602002015280601b602002015182601b602002015280601c602002015182601c602002015280601d602002015182601d602002015280601e602002015182601e602002015280601f602002015182601f60200201528060206020020151826020602002015280602160200201518260216020020152806022602002015182602260200201528060236020020151826023602002015280602460200201518260246020020152806025602002015182602560200201528060266020020151826026602002015280602760200201518260276020020152806028602002015182602860200201528060296020020151826029602002015280602a602002015182602a602002015280602b602002015182602b602002015280602c602002015182602c602002015280602d602002015182602d602002015280602e602002015182602e602002015280602f602002015182602f60200201528060306020020151826030602002015280603160200201518260316020020152806032602002015182603260200201528060336020020151826033602002015280603460200201518260346020020152806035602002015182603560200201528060366020020151826036602002015280603760200201518260376020020152806038602002015182603860200201528060396020020151826039602002015280603a602002015182603a602002015280603b602002015182603b602002015280603c602002015182603c602002015280603d602002015182603d602002015280603e602002015182603e602002015280603f602002015182603f602002015250506125405161518052612d7c90506001545afa61232957600080fd5b6000506156605161233957600080fd5b633b9aca00612b60510261568052615680513031101561235857600080fd5b60016000612d405160e05260c052604060c02055600060006000600061568051612b40516000f161238857600080fd5b615680516156a052612b40517f884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a942436460206156a0a2005b63dca40d9e60005114156123fa57602060046101403734156123de57600080fd5b60006101405160e05260c052604060c0205460005260206000f3005b638a984538600051141561242057341561241357600080fd5b60015460005260206000f3005b60006000fd5b6100c76124ed036100c76000396100c76124ed036000f3`

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

// TestExtract is a free data retrieval call binding the contract method 0xac687308.
//
// Solidity: function testExtract(bytes a) constant returns(address out, uint256 out)
func (_Vault *VaultCaller) TestExtract(opts *bind.CallOpts, a []byte) (common.Address, *big.Int, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Vault.contract.Call(opts, out, "testExtract", a)
	return *ret0, *ret1, err
}

// TestExtract is a free data retrieval call binding the contract method 0xac687308.
//
// Solidity: function testExtract(bytes a) constant returns(address out, uint256 out)
func (_Vault *VaultSession) TestExtract(a []byte) (common.Address, *big.Int, error) {
	return _Vault.Contract.TestExtract(&_Vault.CallOpts, a)
}

// TestExtract is a free data retrieval call binding the contract method 0xac687308.
//
// Solidity: function testExtract(bytes a) constant returns(address out, uint256 out)
func (_Vault *VaultCallerSession) TestExtract(a []byte) (common.Address, *big.Int, error) {
	return _Vault.Contract.TestExtract(&_Vault.CallOpts, a)
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

// VaultNotifyAddressIterator is returned from FilterNotifyAddress and is used to iterate over the raw logs and unpacked data for NotifyAddress events raised by the Vault contract.
type VaultNotifyAddressIterator struct {
	Event *VaultNotifyAddress // Event containing the contract specifics and raw log

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
func (it *VaultNotifyAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNotifyAddress)
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
		it.Event = new(VaultNotifyAddress)
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
func (it *VaultNotifyAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNotifyAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNotifyAddress represents a NotifyAddress event raised by the Vault contract.
type VaultNotifyAddress struct {
	Content common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyAddress is a free log retrieval operation binding the contract event 0x0ac6e167e94338a282ec23bdd86f338fc787bd67f48b3ade098144aac3fcd86e.
//
// Solidity: event NotifyAddress(address content)
func (_Vault *VaultFilterer) FilterNotifyAddress(opts *bind.FilterOpts) (*VaultNotifyAddressIterator, error) {

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NotifyAddress")
	if err != nil {
		return nil, err
	}
	return &VaultNotifyAddressIterator{contract: _Vault.contract, event: "NotifyAddress", logs: logs, sub: sub}, nil
}

// WatchNotifyAddress is a free log subscription operation binding the contract event 0x0ac6e167e94338a282ec23bdd86f338fc787bd67f48b3ade098144aac3fcd86e.
//
// Solidity: event NotifyAddress(address content)
func (_Vault *VaultFilterer) WatchNotifyAddress(opts *bind.WatchOpts, sink chan<- *VaultNotifyAddress) (event.Subscription, error) {

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NotifyAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNotifyAddress)
				if err := _Vault.contract.UnpackLog(event, "NotifyAddress", log); err != nil {
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

// VaultNotifyBoolIterator is returned from FilterNotifyBool and is used to iterate over the raw logs and unpacked data for NotifyBool events raised by the Vault contract.
type VaultNotifyBoolIterator struct {
	Event *VaultNotifyBool // Event containing the contract specifics and raw log

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
func (it *VaultNotifyBoolIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNotifyBool)
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
		it.Event = new(VaultNotifyBool)
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
func (it *VaultNotifyBoolIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNotifyBoolIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNotifyBool represents a NotifyBool event raised by the Vault contract.
type VaultNotifyBool struct {
	Content bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyBool is a free log retrieval operation binding the contract event 0x6c8f06ff564112a969115be5f33d4a0f87ba918c9c9bc3090fe631968e818be4.
//
// Solidity: event NotifyBool(bool content)
func (_Vault *VaultFilterer) FilterNotifyBool(opts *bind.FilterOpts) (*VaultNotifyBoolIterator, error) {

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NotifyBool")
	if err != nil {
		return nil, err
	}
	return &VaultNotifyBoolIterator{contract: _Vault.contract, event: "NotifyBool", logs: logs, sub: sub}, nil
}

// WatchNotifyBool is a free log subscription operation binding the contract event 0x6c8f06ff564112a969115be5f33d4a0f87ba918c9c9bc3090fe631968e818be4.
//
// Solidity: event NotifyBool(bool content)
func (_Vault *VaultFilterer) WatchNotifyBool(opts *bind.WatchOpts, sink chan<- *VaultNotifyBool) (event.Subscription, error) {

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NotifyBool")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNotifyBool)
				if err := _Vault.contract.UnpackLog(event, "NotifyBool", log); err != nil {
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

// VaultNotifyBytes32Iterator is returned from FilterNotifyBytes32 and is used to iterate over the raw logs and unpacked data for NotifyBytes32 events raised by the Vault contract.
type VaultNotifyBytes32Iterator struct {
	Event *VaultNotifyBytes32 // Event containing the contract specifics and raw log

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
func (it *VaultNotifyBytes32Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNotifyBytes32)
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
		it.Event = new(VaultNotifyBytes32)
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
func (it *VaultNotifyBytes32Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNotifyBytes32Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNotifyBytes32 represents a NotifyBytes32 event raised by the Vault contract.
type VaultNotifyBytes32 struct {
	Content [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyBytes32 is a free log retrieval operation binding the contract event 0xb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b.
//
// Solidity: event NotifyBytes32(bytes32 content)
func (_Vault *VaultFilterer) FilterNotifyBytes32(opts *bind.FilterOpts) (*VaultNotifyBytes32Iterator, error) {

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NotifyBytes32")
	if err != nil {
		return nil, err
	}
	return &VaultNotifyBytes32Iterator{contract: _Vault.contract, event: "NotifyBytes32", logs: logs, sub: sub}, nil
}

// WatchNotifyBytes32 is a free log subscription operation binding the contract event 0xb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b.
//
// Solidity: event NotifyBytes32(bytes32 content)
func (_Vault *VaultFilterer) WatchNotifyBytes32(opts *bind.WatchOpts, sink chan<- *VaultNotifyBytes32) (event.Subscription, error) {

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NotifyBytes32")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNotifyBytes32)
				if err := _Vault.contract.UnpackLog(event, "NotifyBytes32", log); err != nil {
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

// VaultNotifyStringIterator is returned from FilterNotifyString and is used to iterate over the raw logs and unpacked data for NotifyString events raised by the Vault contract.
type VaultNotifyStringIterator struct {
	Event *VaultNotifyString // Event containing the contract specifics and raw log

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
func (it *VaultNotifyStringIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNotifyString)
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
		it.Event = new(VaultNotifyString)
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
func (it *VaultNotifyStringIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNotifyStringIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNotifyString represents a NotifyString event raised by the Vault contract.
type VaultNotifyString struct {
	Content string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyString is a free log retrieval operation binding the contract event 0x8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9.
//
// Solidity: event NotifyString(string content)
func (_Vault *VaultFilterer) FilterNotifyString(opts *bind.FilterOpts) (*VaultNotifyStringIterator, error) {

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NotifyString")
	if err != nil {
		return nil, err
	}
	return &VaultNotifyStringIterator{contract: _Vault.contract, event: "NotifyString", logs: logs, sub: sub}, nil
}

// WatchNotifyString is a free log subscription operation binding the contract event 0x8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9.
//
// Solidity: event NotifyString(string content)
func (_Vault *VaultFilterer) WatchNotifyString(opts *bind.WatchOpts, sink chan<- *VaultNotifyString) (event.Subscription, error) {

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NotifyString")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNotifyString)
				if err := _Vault.contract.UnpackLog(event, "NotifyString", log); err != nil {
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

// VaultNotifyUint256Iterator is returned from FilterNotifyUint256 and is used to iterate over the raw logs and unpacked data for NotifyUint256 events raised by the Vault contract.
type VaultNotifyUint256Iterator struct {
	Event *VaultNotifyUint256 // Event containing the contract specifics and raw log

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
func (it *VaultNotifyUint256Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNotifyUint256)
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
		it.Event = new(VaultNotifyUint256)
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
func (it *VaultNotifyUint256Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNotifyUint256Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNotifyUint256 represents a NotifyUint256 event raised by the Vault contract.
type VaultNotifyUint256 struct {
	Content *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNotifyUint256 is a free log retrieval operation binding the contract event 0x8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd.
//
// Solidity: event NotifyUint256(uint256 content)
func (_Vault *VaultFilterer) FilterNotifyUint256(opts *bind.FilterOpts) (*VaultNotifyUint256Iterator, error) {

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NotifyUint256")
	if err != nil {
		return nil, err
	}
	return &VaultNotifyUint256Iterator{contract: _Vault.contract, event: "NotifyUint256", logs: logs, sub: sub}, nil
}

// WatchNotifyUint256 is a free log subscription operation binding the contract event 0x8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd.
//
// Solidity: event NotifyUint256(uint256 content)
func (_Vault *VaultFilterer) WatchNotifyUint256(opts *bind.WatchOpts, sink chan<- *VaultNotifyUint256) (event.Subscription, error) {

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NotifyUint256")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNotifyUint256)
				if err := _Vault.contract.UnpackLog(event, "NotifyUint256", log); err != nil {
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
