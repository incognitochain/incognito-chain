// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package incognitoproxy

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
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IncognitoProxyCommittee is an auto generated low-level Go binding around an user-defined struct.
type IncognitoProxyCommittee struct {
	Pubkeys    []common.Address
	StartBlock *big.Int
}

// IncognitoproxyABI is the input ABI used to generate the binding from.
const IncognitoproxyABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_admin\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"beaconCommittee\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startHeight\",\"type\":\"uint256\"}],\"name\":\"BeaconCommitteeSwapped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"claimer\",\"type\":\"address\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"ndays\",\"type\":\"uint256\"}],\"name\":\"Extend\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"pauser\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"pauser\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"beaconCommittees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"expire\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"n\",\"type\":\"uint256\"}],\"name\":\"extend\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"inst\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"numVals\",\"type\":\"uint256\"}],\"name\":\"extractCommitteeFromInstruction\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"inst\",\"type\":\"bytes\"}],\"name\":\"extractMetaFromInstruction\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blkHeight\",\"type\":\"uint256\"}],\"name\":\"findBeaconCommitteeFromHeight\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"getBeaconCommittee\",\"outputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"pubkeys\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"}],\"internalType\":\"structIncognitoProxy.Committee\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"isBeacon\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"instHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"blkHeight\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"instPath\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool[]\",\"name\":\"instPathIsLeft\",\"type\":\"bool[]\"},{\"internalType\":\"bytes32\",\"name\":\"instRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"blkData\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"sigIdx\",\"type\":\"uint256[]\"},{\"internalType\":\"uint8[]\",\"name\":\"sigV\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"sigR\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"sigS\",\"type\":\"bytes32[]\"}],\"name\":\"instructionApproved\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"leaf\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"root\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"path\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool[]\",\"name\":\"left\",\"type\":\"bool[]\"}],\"name\":\"instructionInMerkleTree\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_successor\",\"type\":\"address\"}],\"name\":\"retire\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"successor\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"inst\",\"type\":\"bytes\"},{\"internalType\":\"bytes32[]\",\"name\":\"instPath\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool[]\",\"name\":\"instPathIsLeft\",\"type\":\"bool[]\"},{\"internalType\":\"bytes32\",\"name\":\"instRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"blkData\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"sigIdx\",\"type\":\"uint256[]\"},{\"internalType\":\"uint8[]\",\"name\":\"sigV\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"sigR\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"sigS\",\"type\":\"bytes32[]\"}],\"name\":\"swapBeaconCommittee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"committee\",\"type\":\"address[]\"},{\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8[]\",\"name\":\"v\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"r\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"s\",\"type\":\"bytes32[]\"}],\"name\":\"verifySig\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// IncognitoproxyBin is the compiled bytecode used for deploying new contracts.
var IncognitoproxyBin = "0x60806040523480156200001157600080fd5b5060405162002a4c38038062002a4c83398181016040528101906200003791906200028c565b816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600160146101000a81548160ff0219169083151502179055506301e13380420160028190555060036040518060400160405280838152602001600081525090806001815401808255809150506001900390600052602060002090600202016000909190919091506000820151816000019080519060200190620000fe92919062000113565b5060208201518160010155505050506200038b565b8280548282559060005260206000209081019282156200018f579160200282015b828111156200018e5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055509160200191906001019062000134565b5b5090506200019e9190620001a2565b5090565b620001e591905b80821115620001e157600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101620001a9565b5090565b90565b600081519050620001f98162000371565b92915050565b600082601f8301126200021157600080fd5b815162000228620002228262000314565b620002e6565b915081818352602084019350602081019050838560208402820111156200024e57600080fd5b60005b83811015620002825781620002678882620001e8565b84526020840193506020830192505060018101905062000251565b5050505092915050565b60008060408385031215620002a057600080fd5b6000620002b085828601620001e8565b925050602083015167ffffffffffffffff811115620002ce57600080fd5b620002dc85828601620001ff565b9150509250929050565b6000604051905081810181811067ffffffffffffffff821117156200030a57600080fd5b8060405250919050565b600067ffffffffffffffff8211156200032c57600080fd5b602082029050602081019050919050565b60006200034a8262000351565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6200037c816200033d565b81146200038857600080fd5b50565b6126b1806200039b6000396000f3fe608060405234801561001057600080fd5b50600436106101165760003560e01c806390500bae116100a2578063e41be77511610071578063e41be775146102bf578063f203a5ed146102db578063f65d21161461030b578063f851a4401461033b578063faea31671461035957610116565b806390500bae146102235780639714378c146102565780639e6371ba14610272578063b600ffdb1461028e57610116565b80635c975abb116100e95780635c975abb1461018f5780636ff968c3146101ad57806379599f96146101cb5780638456cb59146101e95780638eb60066146101f357610116565b80633aacfdad1461011b5780633f4ba83a1461014b57806347c4b328146101555780634e71d92d14610185575b600080fd5b610135600480360381019061013091906118ae565b610389565b60405161014291906121f3565b60405180910390f35b6101536104a9565b005b61016f600480360381019061016a9190611b05565b6105db565b60405161017c91906121f3565b60405180910390f35b61018d610705565b005b610197610895565b6040516101a491906121f3565b60405180910390f35b6101b56108a8565b6040516101c2919061216b565b60405180910390f35b6101d36108ce565b6040516101e09190612355565b60405180910390f35b6101f16108d4565b005b61020d60048036038101906102089190611d47565b610a4a565b60405161021a91906121a1565b60405180910390f35b61023d60048036038101906102389190611b98565b610b26565b60405161024d9493929190612399565b60405180910390f35b610270600480360381019061026b9190611d9b565b610ba1565b005b61028c60048036038101906102879190611885565b610d02565b005b6102a860048036038101906102a39190611d9b565b610e19565b6040516102b69291906121c3565b60405180910390f35b6102d960048036038101906102d49190611bd9565b610f47565b005b6102f560048036038101906102f09190611d9b565b61114a565b6040516103029190612355565b60405180910390f35b61032560048036038101906103209190611985565b611175565b60405161033291906121f3565b60405180910390f35b610343611360565b604051610350919061216b565b60405180910390f35b610373600480360381019061036e9190611d9b565b611385565b6040516103809190612333565b60405180910390f35b6000825184511461039957600080fd5b81518451146103a757600080fd5b60008090505b845181101561049a578681815181106103c257fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff166001878784815181106103ef57fe5b602002602001015187858151811061040357fe5b602002602001015187868151811061041757fe5b60200260200101516040516000815260200160405260405161043c949392919061220e565b6020604051602081039080840390855afa15801561045e573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff161461048d5760009150506104a0565b80806001019150506103ad565b50600190505b95945050505050565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610538576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161052f90612313565b60405180910390fd5b600160149054906101000a900460ff16610587576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161057e90612253565b60405180910390fd5b6000600160146101000a81548160ff0219169083151502179055507f5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa336040516105d19190612186565b60405180910390a1565b600082518251146105eb57600080fd5b600085905060008090505b84518110156106f65783818151811061060b57fe5b60200260200101511561065b5784818151811061062457fe5b60200260200101518260405160200161063e92919061213f565b6040516020818303038152906040528051906020012091506106e9565b6000801b85828151811061066b57fe5b602002602001015114156106a957818260405160200161068c92919061213f565b6040516020818303038152906040528051906020012091506106e8565b818582815181106106b657fe5b60200260200101516040516020016106cf92919061213f565b6040516020818303038152906040528051906020012091505b5b80806001019150506105f6565b50848114915050949350505050565b6002544210610749576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161074090612273565b60405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146107d9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107d0906122b3565b60405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f0c7ef932d3b91976772937f18d5ef9b39a9930bef486b576c374f047c4b512dc6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1660405161088b919061216b565b60405180910390a1565b600160149054906101000a900460ff1681565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60025481565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610963576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161095a90612313565b60405180910390fd5b600160149054906101000a900460ff16156109b3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109aa906122f3565b60405180910390fd5b60025442106109f7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109ee90612273565b60405180910390fd5b60018060146101000a81548160ff0219169083151502179055507f62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a25833604051610a409190612186565b60405180910390a1565b606060208202604201835114610a5f57600080fd5b60608267ffffffffffffffff81118015610a7857600080fd5b50604051908082528060200260200182016040528015610aa75781602001602082028036833780820191505090505b509050600080600090505b84811015610b1a5760208102606287010151915081838281518110610ad357fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610ab2565b50819250505092915050565b600080600080604285511015610b3b57600080fd5b600085600081518110610b4a57fe5b602001015160f81c60f81b60f81c9050600086600181518110610b6957fe5b602001015160f81c60f81b60f81c90506000806022890151915060428901519050838383839750975097509750505050509193509193565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c30576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c2790612313565b60405180910390fd5b6002544210610c74576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c6b90612273565b60405180910390fd5b61016e8110610cb8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610caf90612293565b60405180910390fd5b620151808102600254016002819055507f02ef6561d311451dadc920679eb21192a61d96ee8ead94241b8ff073029ca6e881604051610cf79190612355565b60405180910390a150565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610d91576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610d8890612313565b60405180910390fd5b6002544210610dd5576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610dcc90612273565b60405180910390fd5b80600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6060600080600090506000600380549050905060008111610e3957600080fd5b6001810390505b808214610e95576000600260018385010181610e5857fe5b0490508560038281548110610e6957fe5b90600052602060002090600202016001015411610e8857809250610e8f565b6001810391505b50610e40565b60038281548110610ea257fe5b90600052602060002090600202016000018281805480602002602001604051908101604052809291908181526020018280548015610f3557602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610eeb575b50505050509150935093505050915091565b600160149054906101000a900460ff1615610f97576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f8e906122f3565b60405180910390fd5b600089805190602001209050610fdd600182600360016003805490500381548110610fbe57fe5b9060005260206000209060020201600101548c8c8c8c8c8c8c8c611175565b610fe657600080fd5b600080600080610ff58e610b26565b935093509350935060468460ff16148015611013575060018360ff16145b61101c57600080fd5b60036001600380549050038154811061103157fe5b9060005260206000209060020201600101548211611084576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161107b906122d3565b60405180910390fd5b60606110908f83610a4a565b90506003604051806040016040528083815260200185815250908060018154018082558091505060019003906000526020600020906002020160009091909190915060008201518160000190805190602001906110ee929190611452565b506020820151816001015550507fe15e1a9dec6ad906dd5985b062bfa5ee8bc5d5738e46e4deb8a2df2fbbbb59d160038054905084604051611131929190612370565b60405180910390a1505050505050505050505050505050565b6003818154811061115757fe5b90600052602060002090600202016000915090508060010154905081565b6000606060006111848c610e19565b8092508193505050865186511461119a57600080fd5b84518651146111a857600080fd5b83518651146111b657600080fd5b60008090505b87518110156112a8576000811180156111fe57508760018203815181106111df57fe5b60200260200101518882815181106111f357fe5b602002602001015111155b8061121d5750825188828151811061121257fe5b602002602001015110155b1561122e5760009350505050611351565b8288828151811061123b57fe5b60200260200101518151811061124d57fe5b602002602001015183828151811061126157fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505080806001019150506111bc565b506000888a6040516020016112be92919061213f565b604051602081830303815290604052805190602001206040516020016112e49190612124565b604051602081830303815290604052805190602001209050600360028451028161130a57fe5b0488511161131e5760009350505050611351565b61132b8382898989610389565b61133457600080fd5b6113408e8b8e8e6105db565b61134957600080fd5b600193505050505b9b9a5050505050505050505050565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b61138d6114dc565b6003828154811061139a57fe5b90600052602060002090600202016040518060400160405290816000820180548060200260200160405190810160405280929190818152602001828054801561143857602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190600101908083116113ee575b505050505081526020016001820154815250509050919050565b8280548282559060005260206000209081019282156114cb579160200282015b828111156114ca5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190611472565b5b5090506114d891906114f6565b5090565b604051806040016040528060608152602001600081525090565b61153691905b8082111561153257600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff0219169055506001016114fc565b5090565b90565b60008135905061154881612608565b92915050565b600082601f83011261155f57600080fd5b813561157261156d8261240b565b6123de565b9150818183526020840193506020810190508385602084028201111561159757600080fd5b60005b838110156115c757816115ad8882611539565b84526020840193506020830192505060018101905061159a565b5050505092915050565b600082601f8301126115e257600080fd5b81356115f56115f082612433565b6123de565b9150818183526020840193506020810190508385602084028201111561161a57600080fd5b60005b8381101561164a578161163088826117dd565b84526020840193506020830192505060018101905061161d565b5050505092915050565b600082601f83011261166557600080fd5b81356116786116738261245b565b6123de565b9150818183526020840193506020810190508385602084028201111561169d57600080fd5b60005b838110156116cd57816116b388826117f2565b8452602084019350602083019250506001810190506116a0565b5050505092915050565b600082601f8301126116e857600080fd5b81356116fb6116f682612483565b6123de565b9150818183526020840193506020810190508385602084028201111561172057600080fd5b60005b838110156117505781611736888261185b565b845260208401935060208301925050600181019050611723565b5050505092915050565b600082601f83011261176b57600080fd5b813561177e611779826124ab565b6123de565b915081818352602084019350602081019050838560208402820111156117a357600080fd5b60005b838110156117d357816117b98882611870565b8452602084019350602083019250506001810190506117a6565b5050505092915050565b6000813590506117ec8161261f565b92915050565b60008135905061180181612636565b92915050565b600082601f83011261181857600080fd5b813561182b611826826124d3565b6123de565b9150808252602083016020830185838301111561184757600080fd5b6118528382846125ef565b50505092915050565b60008135905061186a8161264d565b92915050565b60008135905061187f81612664565b92915050565b60006020828403121561189757600080fd5b60006118a584828501611539565b91505092915050565b600080600080600060a086880312156118c657600080fd5b600086013567ffffffffffffffff8111156118e057600080fd5b6118ec8882890161154e565b95505060206118fd888289016117f2565b945050604086013567ffffffffffffffff81111561191a57600080fd5b6119268882890161175a565b935050606086013567ffffffffffffffff81111561194357600080fd5b61194f88828901611654565b925050608086013567ffffffffffffffff81111561196c57600080fd5b61197888828901611654565b9150509295509295909350565b60008060008060008060008060008060006101608c8e0312156119a757600080fd5b60006119b58e828f016117dd565b9b505060206119c68e828f016117f2565b9a505060406119d78e828f0161185b565b99505060608c013567ffffffffffffffff8111156119f457600080fd5b611a008e828f01611654565b98505060808c013567ffffffffffffffff811115611a1d57600080fd5b611a298e828f016115d1565b97505060a0611a3a8e828f016117f2565b96505060c0611a4b8e828f016117f2565b95505060e08c013567ffffffffffffffff811115611a6857600080fd5b611a748e828f016116d7565b9450506101008c013567ffffffffffffffff811115611a9257600080fd5b611a9e8e828f0161175a565b9350506101208c013567ffffffffffffffff811115611abc57600080fd5b611ac88e828f01611654565b9250506101408c013567ffffffffffffffff811115611ae657600080fd5b611af28e828f01611654565b9150509295989b509295989b9093969950565b60008060008060808587031215611b1b57600080fd5b6000611b29878288016117f2565b9450506020611b3a878288016117f2565b935050604085013567ffffffffffffffff811115611b5757600080fd5b611b6387828801611654565b925050606085013567ffffffffffffffff811115611b8057600080fd5b611b8c878288016115d1565b91505092959194509250565b600060208284031215611baa57600080fd5b600082013567ffffffffffffffff811115611bc457600080fd5b611bd084828501611807565b91505092915050565b60008060008060008060008060006101208a8c031215611bf857600080fd5b60008a013567ffffffffffffffff811115611c1257600080fd5b611c1e8c828d01611807565b99505060208a013567ffffffffffffffff811115611c3b57600080fd5b611c478c828d01611654565b98505060408a013567ffffffffffffffff811115611c6457600080fd5b611c708c828d016115d1565b9750506060611c818c828d016117f2565b9650506080611c928c828d016117f2565b95505060a08a013567ffffffffffffffff811115611caf57600080fd5b611cbb8c828d016116d7565b94505060c08a013567ffffffffffffffff811115611cd857600080fd5b611ce48c828d0161175a565b93505060e08a013567ffffffffffffffff811115611d0157600080fd5b611d0d8c828d01611654565b9250506101008a013567ffffffffffffffff811115611d2b57600080fd5b611d378c828d01611654565b9150509295985092959850929598565b60008060408385031215611d5a57600080fd5b600083013567ffffffffffffffff811115611d7457600080fd5b611d8085828601611807565b9250506020611d918582860161185b565b9150509250929050565b600060208284031215611dad57600080fd5b6000611dbb8482850161185b565b91505092915050565b6000611dd08383611deb565b60208301905092915050565b611de5816125b9565b82525050565b611df48161255a565b82525050565b611e038161255a565b82525050565b6000611e148261250f565b611e1e8185612527565b9350611e29836124ff565b8060005b83811015611e5a578151611e418882611dc4565b9750611e4c8361251a565b925050600181019050611e2d565b5085935050505092915050565b6000611e728261250f565b611e7c8185612538565b9350611e87836124ff565b8060005b83811015611eb8578151611e9f8882611dc4565b9750611eaa8361251a565b925050600181019050611e8b565b5085935050505092915050565b611ece8161256c565b82525050565b611edd81612578565b82525050565b611ef4611eef82612578565b6125fe565b82525050565b6000611f07601483612549565b91507f6e6f7420706175736564207269676874206e6f770000000000000000000000006000830152602082019050919050565b6000611f47600783612549565b91507f65787069726564000000000000000000000000000000000000000000000000006000830152602082019050919050565b6000611f87601a83612549565b91507f63616e6e6f7420657874656e6420666f7220746f6f206c6f6e670000000000006000830152602082019050919050565b6000611fc7600c83612549565b91507f756e617574686f72697a656400000000000000000000000000000000000000006000830152602082019050919050565b6000612007601b83612549565b91507f63616e6e6f74206368616e6765206f6c6420636f6d6d697474656500000000006000830152602082019050919050565b6000612047601083612549565b91507f706175736564207269676874206e6f77000000000000000000000000000000006000830152602082019050919050565b6000612087600983612549565b91507f6e6f742061646d696e00000000000000000000000000000000000000000000006000830152602082019050919050565b600060408301600083015184820360008601526120d78282611e09565b91505060208301516120ec60208601826120f7565b508091505092915050565b612100816125a2565b82525050565b61210f816125a2565b82525050565b61211e816125ac565b82525050565b60006121308284611ee3565b60208201915081905092915050565b600061214b8285611ee3565b60208201915061215b8284611ee3565b6020820191508190509392505050565b60006020820190506121806000830184611dfa565b92915050565b600060208201905061219b6000830184611ddc565b92915050565b600060208201905081810360008301526121bb8184611e67565b905092915050565b600060408201905081810360008301526121dd8185611e67565b90506121ec6020830184612106565b9392505050565b60006020820190506122086000830184611ec5565b92915050565b60006080820190506122236000830187611ed4565b6122306020830186612115565b61223d6040830185611ed4565b61224a6060830184611ed4565b95945050505050565b6000602082019050818103600083015261226c81611efa565b9050919050565b6000602082019050818103600083015261228c81611f3a565b9050919050565b600060208201905081810360008301526122ac81611f7a565b9050919050565b600060208201905081810360008301526122cc81611fba565b9050919050565b600060208201905081810360008301526122ec81611ffa565b9050919050565b6000602082019050818103600083015261230c8161203a565b9050919050565b6000602082019050818103600083015261232c8161207a565b9050919050565b6000602082019050818103600083015261234d81846120ba565b905092915050565b600060208201905061236a6000830184612106565b92915050565b60006040820190506123856000830185612106565b6123926020830184612106565b9392505050565b60006080820190506123ae6000830187612115565b6123bb6020830186612115565b6123c86040830185612106565b6123d56060830184612106565b95945050505050565b6000604051905081810181811067ffffffffffffffff8211171561240157600080fd5b8060405250919050565b600067ffffffffffffffff82111561242257600080fd5b602082029050602081019050919050565b600067ffffffffffffffff82111561244a57600080fd5b602082029050602081019050919050565b600067ffffffffffffffff82111561247257600080fd5b602082029050602081019050919050565b600067ffffffffffffffff82111561249a57600080fd5b602082029050602081019050919050565b600067ffffffffffffffff8211156124c257600080fd5b602082029050602081019050919050565b600067ffffffffffffffff8211156124ea57600080fd5b601f19601f8301169050602081019050919050565b6000819050602082019050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600061256582612582565b9050919050565b60008115159050919050565b6000819050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600060ff82169050919050565b60006125c4826125cb565b9050919050565b60006125d6826125dd565b9050919050565b60006125e882612582565b9050919050565b82818337600083830152505050565b6000819050919050565b6126118161255a565b811461261c57600080fd5b50565b6126288161256c565b811461263357600080fd5b50565b61263f81612578565b811461264a57600080fd5b50565b612656816125a2565b811461266157600080fd5b50565b61266d816125ac565b811461267857600080fd5b5056fea2646970667358221220e15ddd95d931027a09fc827c16ebb8bbf25cd37695edb51f0e5dba0ac7db108f64736f6c63430006060033"

// DeployIncognitoproxy deploys a new Ethereum contract, binding an instance of Incognitoproxy to it.
func DeployIncognitoproxy(auth *bind.TransactOpts, backend bind.ContractBackend, _admin common.Address, beaconCommittee []common.Address) (common.Address, *types.Transaction, *Incognitoproxy, error) {
	parsed, err := abi.JSON(strings.NewReader(IncognitoproxyABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IncognitoproxyBin), backend, _admin, beaconCommittee)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Incognitoproxy{IncognitoproxyCaller: IncognitoproxyCaller{contract: contract}, IncognitoproxyTransactor: IncognitoproxyTransactor{contract: contract}, IncognitoproxyFilterer: IncognitoproxyFilterer{contract: contract}}, nil
}

// Incognitoproxy is an auto generated Go binding around an Ethereum contract.
type Incognitoproxy struct {
	IncognitoproxyCaller     // Read-only binding to the contract
	IncognitoproxyTransactor // Write-only binding to the contract
	IncognitoproxyFilterer   // Log filterer for contract events
}

// IncognitoproxyCaller is an auto generated read-only Go binding around an Ethereum contract.
type IncognitoproxyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IncognitoproxyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IncognitoproxyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IncognitoproxyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IncognitoproxyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IncognitoproxySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IncognitoproxySession struct {
	Contract     *Incognitoproxy   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IncognitoproxyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IncognitoproxyCallerSession struct {
	Contract *IncognitoproxyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// IncognitoproxyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IncognitoproxyTransactorSession struct {
	Contract     *IncognitoproxyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// IncognitoproxyRaw is an auto generated low-level Go binding around an Ethereum contract.
type IncognitoproxyRaw struct {
	Contract *Incognitoproxy // Generic contract binding to access the raw methods on
}

// IncognitoproxyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IncognitoproxyCallerRaw struct {
	Contract *IncognitoproxyCaller // Generic read-only contract binding to access the raw methods on
}

// IncognitoproxyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IncognitoproxyTransactorRaw struct {
	Contract *IncognitoproxyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIncognitoproxy creates a new instance of Incognitoproxy, bound to a specific deployed contract.
func NewIncognitoproxy(address common.Address, backend bind.ContractBackend) (*Incognitoproxy, error) {
	contract, err := bindIncognitoproxy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Incognitoproxy{IncognitoproxyCaller: IncognitoproxyCaller{contract: contract}, IncognitoproxyTransactor: IncognitoproxyTransactor{contract: contract}, IncognitoproxyFilterer: IncognitoproxyFilterer{contract: contract}}, nil
}

// NewIncognitoproxyCaller creates a new read-only instance of Incognitoproxy, bound to a specific deployed contract.
func NewIncognitoproxyCaller(address common.Address, caller bind.ContractCaller) (*IncognitoproxyCaller, error) {
	contract, err := bindIncognitoproxy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyCaller{contract: contract}, nil
}

// NewIncognitoproxyTransactor creates a new write-only instance of Incognitoproxy, bound to a specific deployed contract.
func NewIncognitoproxyTransactor(address common.Address, transactor bind.ContractTransactor) (*IncognitoproxyTransactor, error) {
	contract, err := bindIncognitoproxy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyTransactor{contract: contract}, nil
}

// NewIncognitoproxyFilterer creates a new log filterer instance of Incognitoproxy, bound to a specific deployed contract.
func NewIncognitoproxyFilterer(address common.Address, filterer bind.ContractFilterer) (*IncognitoproxyFilterer, error) {
	contract, err := bindIncognitoproxy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyFilterer{contract: contract}, nil
}

// bindIncognitoproxy binds a generic wrapper to an already deployed contract.
func bindIncognitoproxy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IncognitoproxyABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Incognitoproxy *IncognitoproxyRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Incognitoproxy.Contract.IncognitoproxyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Incognitoproxy *IncognitoproxyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.IncognitoproxyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Incognitoproxy *IncognitoproxyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.IncognitoproxyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Incognitoproxy *IncognitoproxyCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Incognitoproxy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Incognitoproxy *IncognitoproxyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Incognitoproxy *IncognitoproxyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Incognitoproxy *IncognitoproxyCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Incognitoproxy *IncognitoproxySession) Admin() (common.Address, error) {
	return _Incognitoproxy.Contract.Admin(&_Incognitoproxy.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Incognitoproxy *IncognitoproxyCallerSession) Admin() (common.Address, error) {
	return _Incognitoproxy.Contract.Admin(&_Incognitoproxy.CallOpts)
}

// BeaconCommittees is a free data retrieval call binding the contract method 0xf203a5ed.
//
// Solidity: function beaconCommittees(uint256 ) view returns(uint256 startBlock)
func (_Incognitoproxy *IncognitoproxyCaller) BeaconCommittees(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "beaconCommittees", arg0)
	return *ret0, err
}

// BeaconCommittees is a free data retrieval call binding the contract method 0xf203a5ed.
//
// Solidity: function beaconCommittees(uint256 ) view returns(uint256 startBlock)
func (_Incognitoproxy *IncognitoproxySession) BeaconCommittees(arg0 *big.Int) (*big.Int, error) {
	return _Incognitoproxy.Contract.BeaconCommittees(&_Incognitoproxy.CallOpts, arg0)
}

// BeaconCommittees is a free data retrieval call binding the contract method 0xf203a5ed.
//
// Solidity: function beaconCommittees(uint256 ) view returns(uint256 startBlock)
func (_Incognitoproxy *IncognitoproxyCallerSession) BeaconCommittees(arg0 *big.Int) (*big.Int, error) {
	return _Incognitoproxy.Contract.BeaconCommittees(&_Incognitoproxy.CallOpts, arg0)
}

// Expire is a free data retrieval call binding the contract method 0x79599f96.
//
// Solidity: function expire() view returns(uint256)
func (_Incognitoproxy *IncognitoproxyCaller) Expire(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "expire")
	return *ret0, err
}

// Expire is a free data retrieval call binding the contract method 0x79599f96.
//
// Solidity: function expire() view returns(uint256)
func (_Incognitoproxy *IncognitoproxySession) Expire() (*big.Int, error) {
	return _Incognitoproxy.Contract.Expire(&_Incognitoproxy.CallOpts)
}

// Expire is a free data retrieval call binding the contract method 0x79599f96.
//
// Solidity: function expire() view returns(uint256)
func (_Incognitoproxy *IncognitoproxyCallerSession) Expire() (*big.Int, error) {
	return _Incognitoproxy.Contract.Expire(&_Incognitoproxy.CallOpts)
}

// ExtractCommitteeFromInstruction is a free data retrieval call binding the contract method 0x8eb60066.
//
// Solidity: function extractCommitteeFromInstruction(bytes inst, uint256 numVals) pure returns(address[])
func (_Incognitoproxy *IncognitoproxyCaller) ExtractCommitteeFromInstruction(opts *bind.CallOpts, inst []byte, numVals *big.Int) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "extractCommitteeFromInstruction", inst, numVals)
	return *ret0, err
}

// ExtractCommitteeFromInstruction is a free data retrieval call binding the contract method 0x8eb60066.
//
// Solidity: function extractCommitteeFromInstruction(bytes inst, uint256 numVals) pure returns(address[])
func (_Incognitoproxy *IncognitoproxySession) ExtractCommitteeFromInstruction(inst []byte, numVals *big.Int) ([]common.Address, error) {
	return _Incognitoproxy.Contract.ExtractCommitteeFromInstruction(&_Incognitoproxy.CallOpts, inst, numVals)
}

// ExtractCommitteeFromInstruction is a free data retrieval call binding the contract method 0x8eb60066.
//
// Solidity: function extractCommitteeFromInstruction(bytes inst, uint256 numVals) pure returns(address[])
func (_Incognitoproxy *IncognitoproxyCallerSession) ExtractCommitteeFromInstruction(inst []byte, numVals *big.Int) ([]common.Address, error) {
	return _Incognitoproxy.Contract.ExtractCommitteeFromInstruction(&_Incognitoproxy.CallOpts, inst, numVals)
}

// ExtractMetaFromInstruction is a free data retrieval call binding the contract method 0x90500bae.
//
// Solidity: function extractMetaFromInstruction(bytes inst) pure returns(uint8, uint8, uint256, uint256)
func (_Incognitoproxy *IncognitoproxyCaller) ExtractMetaFromInstruction(opts *bind.CallOpts, inst []byte) (uint8, uint8, *big.Int, *big.Int, error) {
	var (
		ret0 = new(uint8)
		ret1 = new(uint8)
		ret2 = new(*big.Int)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Incognitoproxy.contract.Call(opts, out, "extractMetaFromInstruction", inst)
	return *ret0, *ret1, *ret2, *ret3, err
}

// ExtractMetaFromInstruction is a free data retrieval call binding the contract method 0x90500bae.
//
// Solidity: function extractMetaFromInstruction(bytes inst) pure returns(uint8, uint8, uint256, uint256)
func (_Incognitoproxy *IncognitoproxySession) ExtractMetaFromInstruction(inst []byte) (uint8, uint8, *big.Int, *big.Int, error) {
	return _Incognitoproxy.Contract.ExtractMetaFromInstruction(&_Incognitoproxy.CallOpts, inst)
}

// ExtractMetaFromInstruction is a free data retrieval call binding the contract method 0x90500bae.
//
// Solidity: function extractMetaFromInstruction(bytes inst) pure returns(uint8, uint8, uint256, uint256)
func (_Incognitoproxy *IncognitoproxyCallerSession) ExtractMetaFromInstruction(inst []byte) (uint8, uint8, *big.Int, *big.Int, error) {
	return _Incognitoproxy.Contract.ExtractMetaFromInstruction(&_Incognitoproxy.CallOpts, inst)
}

// FindBeaconCommitteeFromHeight is a free data retrieval call binding the contract method 0xb600ffdb.
//
// Solidity: function findBeaconCommitteeFromHeight(uint256 blkHeight) view returns(address[], uint256)
func (_Incognitoproxy *IncognitoproxyCaller) FindBeaconCommitteeFromHeight(opts *bind.CallOpts, blkHeight *big.Int) ([]common.Address, *big.Int, error) {
	var (
		ret0 = new([]common.Address)
		ret1 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Incognitoproxy.contract.Call(opts, out, "findBeaconCommitteeFromHeight", blkHeight)
	return *ret0, *ret1, err
}

// FindBeaconCommitteeFromHeight is a free data retrieval call binding the contract method 0xb600ffdb.
//
// Solidity: function findBeaconCommitteeFromHeight(uint256 blkHeight) view returns(address[], uint256)
func (_Incognitoproxy *IncognitoproxySession) FindBeaconCommitteeFromHeight(blkHeight *big.Int) ([]common.Address, *big.Int, error) {
	return _Incognitoproxy.Contract.FindBeaconCommitteeFromHeight(&_Incognitoproxy.CallOpts, blkHeight)
}

// FindBeaconCommitteeFromHeight is a free data retrieval call binding the contract method 0xb600ffdb.
//
// Solidity: function findBeaconCommitteeFromHeight(uint256 blkHeight) view returns(address[], uint256)
func (_Incognitoproxy *IncognitoproxyCallerSession) FindBeaconCommitteeFromHeight(blkHeight *big.Int) ([]common.Address, *big.Int, error) {
	return _Incognitoproxy.Contract.FindBeaconCommitteeFromHeight(&_Incognitoproxy.CallOpts, blkHeight)
}

// GetBeaconCommittee is a free data retrieval call binding the contract method 0xfaea3167.
//
// Solidity: function getBeaconCommittee(uint256 i) view returns((address[],uint256))
func (_Incognitoproxy *IncognitoproxyCaller) GetBeaconCommittee(opts *bind.CallOpts, i *big.Int) (IncognitoProxyCommittee, error) {
	var (
		ret0 = new(IncognitoProxyCommittee)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "getBeaconCommittee", i)
	return *ret0, err
}

// GetBeaconCommittee is a free data retrieval call binding the contract method 0xfaea3167.
//
// Solidity: function getBeaconCommittee(uint256 i) view returns((address[],uint256))
func (_Incognitoproxy *IncognitoproxySession) GetBeaconCommittee(i *big.Int) (IncognitoProxyCommittee, error) {
	return _Incognitoproxy.Contract.GetBeaconCommittee(&_Incognitoproxy.CallOpts, i)
}

// GetBeaconCommittee is a free data retrieval call binding the contract method 0xfaea3167.
//
// Solidity: function getBeaconCommittee(uint256 i) view returns((address[],uint256))
func (_Incognitoproxy *IncognitoproxyCallerSession) GetBeaconCommittee(i *big.Int) (IncognitoProxyCommittee, error) {
	return _Incognitoproxy.Contract.GetBeaconCommittee(&_Incognitoproxy.CallOpts, i)
}

// InstructionApproved is a free data retrieval call binding the contract method 0xf65d2116.
//
// Solidity: function instructionApproved(bool isBeacon, bytes32 instHash, uint256 blkHeight, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) view returns(bool)
func (_Incognitoproxy *IncognitoproxyCaller) InstructionApproved(opts *bind.CallOpts, isBeacon bool, instHash [32]byte, blkHeight *big.Int, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "instructionApproved", isBeacon, instHash, blkHeight, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
	return *ret0, err
}

// InstructionApproved is a free data retrieval call binding the contract method 0xf65d2116.
//
// Solidity: function instructionApproved(bool isBeacon, bytes32 instHash, uint256 blkHeight, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) view returns(bool)
func (_Incognitoproxy *IncognitoproxySession) InstructionApproved(isBeacon bool, instHash [32]byte, blkHeight *big.Int, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (bool, error) {
	return _Incognitoproxy.Contract.InstructionApproved(&_Incognitoproxy.CallOpts, isBeacon, instHash, blkHeight, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
}

// InstructionApproved is a free data retrieval call binding the contract method 0xf65d2116.
//
// Solidity: function instructionApproved(bool isBeacon, bytes32 instHash, uint256 blkHeight, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) view returns(bool)
func (_Incognitoproxy *IncognitoproxyCallerSession) InstructionApproved(isBeacon bool, instHash [32]byte, blkHeight *big.Int, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (bool, error) {
	return _Incognitoproxy.Contract.InstructionApproved(&_Incognitoproxy.CallOpts, isBeacon, instHash, blkHeight, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
}

// InstructionInMerkleTree is a free data retrieval call binding the contract method 0x47c4b328.
//
// Solidity: function instructionInMerkleTree(bytes32 leaf, bytes32 root, bytes32[] path, bool[] left) pure returns(bool)
func (_Incognitoproxy *IncognitoproxyCaller) InstructionInMerkleTree(opts *bind.CallOpts, leaf [32]byte, root [32]byte, path [][32]byte, left []bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "instructionInMerkleTree", leaf, root, path, left)
	return *ret0, err
}

// InstructionInMerkleTree is a free data retrieval call binding the contract method 0x47c4b328.
//
// Solidity: function instructionInMerkleTree(bytes32 leaf, bytes32 root, bytes32[] path, bool[] left) pure returns(bool)
func (_Incognitoproxy *IncognitoproxySession) InstructionInMerkleTree(leaf [32]byte, root [32]byte, path [][32]byte, left []bool) (bool, error) {
	return _Incognitoproxy.Contract.InstructionInMerkleTree(&_Incognitoproxy.CallOpts, leaf, root, path, left)
}

// InstructionInMerkleTree is a free data retrieval call binding the contract method 0x47c4b328.
//
// Solidity: function instructionInMerkleTree(bytes32 leaf, bytes32 root, bytes32[] path, bool[] left) pure returns(bool)
func (_Incognitoproxy *IncognitoproxyCallerSession) InstructionInMerkleTree(leaf [32]byte, root [32]byte, path [][32]byte, left []bool) (bool, error) {
	return _Incognitoproxy.Contract.InstructionInMerkleTree(&_Incognitoproxy.CallOpts, leaf, root, path, left)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Incognitoproxy *IncognitoproxyCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "paused")
	return *ret0, err
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Incognitoproxy *IncognitoproxySession) Paused() (bool, error) {
	return _Incognitoproxy.Contract.Paused(&_Incognitoproxy.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Incognitoproxy *IncognitoproxyCallerSession) Paused() (bool, error) {
	return _Incognitoproxy.Contract.Paused(&_Incognitoproxy.CallOpts)
}

// Successor is a free data retrieval call binding the contract method 0x6ff968c3.
//
// Solidity: function successor() view returns(address)
func (_Incognitoproxy *IncognitoproxyCaller) Successor(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "successor")
	return *ret0, err
}

// Successor is a free data retrieval call binding the contract method 0x6ff968c3.
//
// Solidity: function successor() view returns(address)
func (_Incognitoproxy *IncognitoproxySession) Successor() (common.Address, error) {
	return _Incognitoproxy.Contract.Successor(&_Incognitoproxy.CallOpts)
}

// Successor is a free data retrieval call binding the contract method 0x6ff968c3.
//
// Solidity: function successor() view returns(address)
func (_Incognitoproxy *IncognitoproxyCallerSession) Successor() (common.Address, error) {
	return _Incognitoproxy.Contract.Successor(&_Incognitoproxy.CallOpts)
}

// VerifySig is a free data retrieval call binding the contract method 0x3aacfdad.
//
// Solidity: function verifySig(address[] committee, bytes32 msgHash, uint8[] v, bytes32[] r, bytes32[] s) pure returns(bool)
func (_Incognitoproxy *IncognitoproxyCaller) VerifySig(opts *bind.CallOpts, committee []common.Address, msgHash [32]byte, v []uint8, r [][32]byte, s [][32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Incognitoproxy.contract.Call(opts, out, "verifySig", committee, msgHash, v, r, s)
	return *ret0, err
}

// VerifySig is a free data retrieval call binding the contract method 0x3aacfdad.
//
// Solidity: function verifySig(address[] committee, bytes32 msgHash, uint8[] v, bytes32[] r, bytes32[] s) pure returns(bool)
func (_Incognitoproxy *IncognitoproxySession) VerifySig(committee []common.Address, msgHash [32]byte, v []uint8, r [][32]byte, s [][32]byte) (bool, error) {
	return _Incognitoproxy.Contract.VerifySig(&_Incognitoproxy.CallOpts, committee, msgHash, v, r, s)
}

// VerifySig is a free data retrieval call binding the contract method 0x3aacfdad.
//
// Solidity: function verifySig(address[] committee, bytes32 msgHash, uint8[] v, bytes32[] r, bytes32[] s) pure returns(bool)
func (_Incognitoproxy *IncognitoproxyCallerSession) VerifySig(committee []common.Address, msgHash [32]byte, v []uint8, r [][32]byte, s [][32]byte) (bool, error) {
	return _Incognitoproxy.Contract.VerifySig(&_Incognitoproxy.CallOpts, committee, msgHash, v, r, s)
}

// Claim is a paid mutator transaction binding the contract method 0x4e71d92d.
//
// Solidity: function claim() returns()
func (_Incognitoproxy *IncognitoproxyTransactor) Claim(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "claim")
}

// Claim is a paid mutator transaction binding the contract method 0x4e71d92d.
//
// Solidity: function claim() returns()
func (_Incognitoproxy *IncognitoproxySession) Claim() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Claim(&_Incognitoproxy.TransactOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x4e71d92d.
//
// Solidity: function claim() returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) Claim() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Claim(&_Incognitoproxy.TransactOpts)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 n) returns()
func (_Incognitoproxy *IncognitoproxyTransactor) Extend(opts *bind.TransactOpts, n *big.Int) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "extend", n)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 n) returns()
func (_Incognitoproxy *IncognitoproxySession) Extend(n *big.Int) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Extend(&_Incognitoproxy.TransactOpts, n)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 n) returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) Extend(n *big.Int) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Extend(&_Incognitoproxy.TransactOpts, n)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Incognitoproxy *IncognitoproxyTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Incognitoproxy *IncognitoproxySession) Pause() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Pause(&_Incognitoproxy.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) Pause() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Pause(&_Incognitoproxy.TransactOpts)
}

// Retire is a paid mutator transaction binding the contract method 0x9e6371ba.
//
// Solidity: function retire(address _successor) returns()
func (_Incognitoproxy *IncognitoproxyTransactor) Retire(opts *bind.TransactOpts, _successor common.Address) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "retire", _successor)
}

// Retire is a paid mutator transaction binding the contract method 0x9e6371ba.
//
// Solidity: function retire(address _successor) returns()
func (_Incognitoproxy *IncognitoproxySession) Retire(_successor common.Address) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Retire(&_Incognitoproxy.TransactOpts, _successor)
}

// Retire is a paid mutator transaction binding the contract method 0x9e6371ba.
//
// Solidity: function retire(address _successor) returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) Retire(_successor common.Address) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Retire(&_Incognitoproxy.TransactOpts, _successor)
}

// SwapBeaconCommittee is a paid mutator transaction binding the contract method 0xe41be775.
//
// Solidity: function swapBeaconCommittee(bytes inst, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) returns()
func (_Incognitoproxy *IncognitoproxyTransactor) SwapBeaconCommittee(opts *bind.TransactOpts, inst []byte, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "swapBeaconCommittee", inst, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
}

// SwapBeaconCommittee is a paid mutator transaction binding the contract method 0xe41be775.
//
// Solidity: function swapBeaconCommittee(bytes inst, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) returns()
func (_Incognitoproxy *IncognitoproxySession) SwapBeaconCommittee(inst []byte, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.SwapBeaconCommittee(&_Incognitoproxy.TransactOpts, inst, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
}

// SwapBeaconCommittee is a paid mutator transaction binding the contract method 0xe41be775.
//
// Solidity: function swapBeaconCommittee(bytes inst, bytes32[] instPath, bool[] instPathIsLeft, bytes32 instRoot, bytes32 blkData, uint256[] sigIdx, uint8[] sigV, bytes32[] sigR, bytes32[] sigS) returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) SwapBeaconCommittee(inst []byte, instPath [][32]byte, instPathIsLeft []bool, instRoot [32]byte, blkData [32]byte, sigIdx []*big.Int, sigV []uint8, sigR [][32]byte, sigS [][32]byte) (*types.Transaction, error) {
	return _Incognitoproxy.Contract.SwapBeaconCommittee(&_Incognitoproxy.TransactOpts, inst, instPath, instPathIsLeft, instRoot, blkData, sigIdx, sigV, sigR, sigS)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Incognitoproxy *IncognitoproxyTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Incognitoproxy.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Incognitoproxy *IncognitoproxySession) Unpause() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Unpause(&_Incognitoproxy.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Incognitoproxy *IncognitoproxyTransactorSession) Unpause() (*types.Transaction, error) {
	return _Incognitoproxy.Contract.Unpause(&_Incognitoproxy.TransactOpts)
}

// IncognitoproxyBeaconCommitteeSwappedIterator is returned from FilterBeaconCommitteeSwapped and is used to iterate over the raw logs and unpacked data for BeaconCommitteeSwapped events raised by the Incognitoproxy contract.
type IncognitoproxyBeaconCommitteeSwappedIterator struct {
	Event *IncognitoproxyBeaconCommitteeSwapped // Event containing the contract specifics and raw log

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
func (it *IncognitoproxyBeaconCommitteeSwappedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IncognitoproxyBeaconCommitteeSwapped)
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
		it.Event = new(IncognitoproxyBeaconCommitteeSwapped)
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
func (it *IncognitoproxyBeaconCommitteeSwappedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IncognitoproxyBeaconCommitteeSwappedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IncognitoproxyBeaconCommitteeSwapped represents a BeaconCommitteeSwapped event raised by the Incognitoproxy contract.
type IncognitoproxyBeaconCommitteeSwapped struct {
	Id          *big.Int
	StartHeight *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBeaconCommitteeSwapped is a free log retrieval operation binding the contract event 0xe15e1a9dec6ad906dd5985b062bfa5ee8bc5d5738e46e4deb8a2df2fbbbb59d1.
//
// Solidity: event BeaconCommitteeSwapped(uint256 id, uint256 startHeight)
func (_Incognitoproxy *IncognitoproxyFilterer) FilterBeaconCommitteeSwapped(opts *bind.FilterOpts) (*IncognitoproxyBeaconCommitteeSwappedIterator, error) {

	logs, sub, err := _Incognitoproxy.contract.FilterLogs(opts, "BeaconCommitteeSwapped")
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyBeaconCommitteeSwappedIterator{contract: _Incognitoproxy.contract, event: "BeaconCommitteeSwapped", logs: logs, sub: sub}, nil
}

// WatchBeaconCommitteeSwapped is a free log subscription operation binding the contract event 0xe15e1a9dec6ad906dd5985b062bfa5ee8bc5d5738e46e4deb8a2df2fbbbb59d1.
//
// Solidity: event BeaconCommitteeSwapped(uint256 id, uint256 startHeight)
func (_Incognitoproxy *IncognitoproxyFilterer) WatchBeaconCommitteeSwapped(opts *bind.WatchOpts, sink chan<- *IncognitoproxyBeaconCommitteeSwapped) (event.Subscription, error) {

	logs, sub, err := _Incognitoproxy.contract.WatchLogs(opts, "BeaconCommitteeSwapped")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IncognitoproxyBeaconCommitteeSwapped)
				if err := _Incognitoproxy.contract.UnpackLog(event, "BeaconCommitteeSwapped", log); err != nil {
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

// ParseBeaconCommitteeSwapped is a log parse operation binding the contract event 0xe15e1a9dec6ad906dd5985b062bfa5ee8bc5d5738e46e4deb8a2df2fbbbb59d1.
//
// Solidity: event BeaconCommitteeSwapped(uint256 id, uint256 startHeight)
func (_Incognitoproxy *IncognitoproxyFilterer) ParseBeaconCommitteeSwapped(log types.Log) (*IncognitoproxyBeaconCommitteeSwapped, error) {
	event := new(IncognitoproxyBeaconCommitteeSwapped)
	if err := _Incognitoproxy.contract.UnpackLog(event, "BeaconCommitteeSwapped", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IncognitoproxyClaimIterator is returned from FilterClaim and is used to iterate over the raw logs and unpacked data for Claim events raised by the Incognitoproxy contract.
type IncognitoproxyClaimIterator struct {
	Event *IncognitoproxyClaim // Event containing the contract specifics and raw log

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
func (it *IncognitoproxyClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IncognitoproxyClaim)
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
		it.Event = new(IncognitoproxyClaim)
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
func (it *IncognitoproxyClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IncognitoproxyClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IncognitoproxyClaim represents a Claim event raised by the Incognitoproxy contract.
type IncognitoproxyClaim struct {
	Claimer common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0x0c7ef932d3b91976772937f18d5ef9b39a9930bef486b576c374f047c4b512dc.
//
// Solidity: event Claim(address claimer)
func (_Incognitoproxy *IncognitoproxyFilterer) FilterClaim(opts *bind.FilterOpts) (*IncognitoproxyClaimIterator, error) {

	logs, sub, err := _Incognitoproxy.contract.FilterLogs(opts, "Claim")
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyClaimIterator{contract: _Incognitoproxy.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0x0c7ef932d3b91976772937f18d5ef9b39a9930bef486b576c374f047c4b512dc.
//
// Solidity: event Claim(address claimer)
func (_Incognitoproxy *IncognitoproxyFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *IncognitoproxyClaim) (event.Subscription, error) {

	logs, sub, err := _Incognitoproxy.contract.WatchLogs(opts, "Claim")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IncognitoproxyClaim)
				if err := _Incognitoproxy.contract.UnpackLog(event, "Claim", log); err != nil {
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

// ParseClaim is a log parse operation binding the contract event 0x0c7ef932d3b91976772937f18d5ef9b39a9930bef486b576c374f047c4b512dc.
//
// Solidity: event Claim(address claimer)
func (_Incognitoproxy *IncognitoproxyFilterer) ParseClaim(log types.Log) (*IncognitoproxyClaim, error) {
	event := new(IncognitoproxyClaim)
	if err := _Incognitoproxy.contract.UnpackLog(event, "Claim", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IncognitoproxyExtendIterator is returned from FilterExtend and is used to iterate over the raw logs and unpacked data for Extend events raised by the Incognitoproxy contract.
type IncognitoproxyExtendIterator struct {
	Event *IncognitoproxyExtend // Event containing the contract specifics and raw log

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
func (it *IncognitoproxyExtendIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IncognitoproxyExtend)
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
		it.Event = new(IncognitoproxyExtend)
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
func (it *IncognitoproxyExtendIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IncognitoproxyExtendIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IncognitoproxyExtend represents a Extend event raised by the Incognitoproxy contract.
type IncognitoproxyExtend struct {
	Ndays *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterExtend is a free log retrieval operation binding the contract event 0x02ef6561d311451dadc920679eb21192a61d96ee8ead94241b8ff073029ca6e8.
//
// Solidity: event Extend(uint256 ndays)
func (_Incognitoproxy *IncognitoproxyFilterer) FilterExtend(opts *bind.FilterOpts) (*IncognitoproxyExtendIterator, error) {

	logs, sub, err := _Incognitoproxy.contract.FilterLogs(opts, "Extend")
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyExtendIterator{contract: _Incognitoproxy.contract, event: "Extend", logs: logs, sub: sub}, nil
}

// WatchExtend is a free log subscription operation binding the contract event 0x02ef6561d311451dadc920679eb21192a61d96ee8ead94241b8ff073029ca6e8.
//
// Solidity: event Extend(uint256 ndays)
func (_Incognitoproxy *IncognitoproxyFilterer) WatchExtend(opts *bind.WatchOpts, sink chan<- *IncognitoproxyExtend) (event.Subscription, error) {

	logs, sub, err := _Incognitoproxy.contract.WatchLogs(opts, "Extend")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IncognitoproxyExtend)
				if err := _Incognitoproxy.contract.UnpackLog(event, "Extend", log); err != nil {
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

// ParseExtend is a log parse operation binding the contract event 0x02ef6561d311451dadc920679eb21192a61d96ee8ead94241b8ff073029ca6e8.
//
// Solidity: event Extend(uint256 ndays)
func (_Incognitoproxy *IncognitoproxyFilterer) ParseExtend(log types.Log) (*IncognitoproxyExtend, error) {
	event := new(IncognitoproxyExtend)
	if err := _Incognitoproxy.contract.UnpackLog(event, "Extend", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IncognitoproxyPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the Incognitoproxy contract.
type IncognitoproxyPausedIterator struct {
	Event *IncognitoproxyPaused // Event containing the contract specifics and raw log

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
func (it *IncognitoproxyPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IncognitoproxyPaused)
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
		it.Event = new(IncognitoproxyPaused)
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
func (it *IncognitoproxyPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IncognitoproxyPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IncognitoproxyPaused represents a Paused event raised by the Incognitoproxy contract.
type IncognitoproxyPaused struct {
	Pauser common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) FilterPaused(opts *bind.FilterOpts) (*IncognitoproxyPausedIterator, error) {

	logs, sub, err := _Incognitoproxy.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyPausedIterator{contract: _Incognitoproxy.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *IncognitoproxyPaused) (event.Subscription, error) {

	logs, sub, err := _Incognitoproxy.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IncognitoproxyPaused)
				if err := _Incognitoproxy.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) ParsePaused(log types.Log) (*IncognitoproxyPaused, error) {
	event := new(IncognitoproxyPaused)
	if err := _Incognitoproxy.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IncognitoproxyUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the Incognitoproxy contract.
type IncognitoproxyUnpausedIterator struct {
	Event *IncognitoproxyUnpaused // Event containing the contract specifics and raw log

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
func (it *IncognitoproxyUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IncognitoproxyUnpaused)
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
		it.Event = new(IncognitoproxyUnpaused)
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
func (it *IncognitoproxyUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IncognitoproxyUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IncognitoproxyUnpaused represents a Unpaused event raised by the Incognitoproxy contract.
type IncognitoproxyUnpaused struct {
	Pauser common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) FilterUnpaused(opts *bind.FilterOpts) (*IncognitoproxyUnpausedIterator, error) {

	logs, sub, err := _Incognitoproxy.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &IncognitoproxyUnpausedIterator{contract: _Incognitoproxy.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *IncognitoproxyUnpaused) (event.Subscription, error) {

	logs, sub, err := _Incognitoproxy.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IncognitoproxyUnpaused)
				if err := _Incognitoproxy.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address pauser)
func (_Incognitoproxy *IncognitoproxyFilterer) ParseUnpaused(log types.Log) (*IncognitoproxyUnpaused, error) {
	event := new(IncognitoproxyUnpaused)
	if err := _Incognitoproxy.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	return event, nil
}
