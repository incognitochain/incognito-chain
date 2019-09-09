package common

import "time"

// for common
const (
	EmptyString       = ""
	ZeroByte          = byte(0x00)
	DateOutputFormat  = "2006-01-02T15:04:05.999999"
	BigIntSize        = 32 // bytes
	CheckSumLen       = 4  // bytes
	AESKeySize        = 32 // bytes
	Int32Size         = 4  // bytes
	Uint32Size        = 4  // bytes
	Uint64Size        = 8  // bytes
	HashSize          = 32 // bytes
	MaxHashStringSize = HashSize * 2
	Base58Version     = 0
)

// size data for incognito key and signature
const (
	// for key size
	PrivateKeySize      = 32 // bytes
	PublicKeySize       = 33 // bytes
	TransmissionKeySize = 33 //bytes
	ReceivingKeySize    = 32 // bytes
	PaymentAddressSize  = 66 // bytes
	// for signature size
	// it is used for both privacy and no privacy
	SigPubKeySize    = 33
	SigNoPrivacySize = 64
	SigPrivacySize   = 96
)

// for exit code
const (
	ExitCodeUnknow = iota
	ExitByOs
	ExitByLogging
	ExitCodeForceUpdate
)

// For all Transaction information
const (
	TxNormalType             = "n"  // normal tx(send and receive coin)
	TxRewardType             = "s"  // reward tx
	TxReturnStakingType      = "rs" //
	TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
)

var (
	MaxTxSize    = uint64(100)  // unit KB = 100KB
	MaxBlockSize = uint64(2000) //unit kilobytes = 2 Megabyte
)

// for mining consensus
const (
	MinBeaconBlkInterval = 10 * time.Second //second
	MinShardBlkInterval  = 10 * time.Second //second => process block in
	MinShardBlkCreation  = 4 * time.Second  //second => process block in
)

// special token ids (aka. PropertyID in custom token)
var (
	PRVCoinID = Hash{4} // To send PRV in custom token
)

// centralized website's pubkey
var (
	//CentralizedWebsitePubKey         = []byte{2, 194, 130, 176, 102, 36, 183, 114, 109, 135, 49, 114, 177, 92, 214, 31, 25, 4, 72, 103, 196, 161, 36, 69, 121, 102, 159, 24, 31, 131, 101, 20, 0}
	CentralizedWebsitePaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// CentralizedWebsitePubKey = []byte{3, 159, 2, 42, 22, 163, 195, 221, 129, 31, 217, 133, 149, 16, 68, 108, 42, 192, 58, 95, 39, 204, 63, 68, 203, 132, 221, 48, 181, 131, 40, 189, 0}
)

// board addresses
const (
	DevAddress     = "1Uv2vrb74e6ScxuQiXvW9UcKoEbXnRMbuBJ6W2FBWxqhtHNGHi3sUP1D14rNEnWWzkYSMsZCmA4DKV6igmjd7qaJfj9TuMmyqz2ZG2SNx"
	BurningAddress = "1NHp2EKw7ALdXUzBfoRJvKrBBM9nkejyDcHVPvUjDcWRyG22dHHyiBKQGL1c"
)

// CONSENSUS
const (
	OFFSET = 1

	NODEMODE_RELAY  = "relay"
	NODEMODE_SHARD  = "shard"
	NODEMODE_AUTO   = "auto"
	NODEMODE_BEACON = "beacon"

	BEACON_ROLE    = "beacon"
	SHARD_ROLE     = "shard"
	PROPOSER_ROLE  = "proposer"
	VALIDATOR_ROLE = "validator"
	PENDING_ROLE   = "pending"

	MAX_SHARD_NUMBER = 8
)

// Ethereum Decentralized bridge
const (
	ABIJSON               = `[{"name": "Deposit", "inputs": [{"type": "address", "name": "_token", "indexed": false}, {"type": "string", "name": "_incognito_address", "indexed": false}, {"type": "uint256", "name": "_amount", "indexed": false, "unit": "wei"}], "anonymous": false, "type": "event"}, {"name": "Withdraw", "inputs": [{"type": "address", "name": "_token", "indexed": false}, {"type": "address", "name": "_to", "indexed": false}, {"type": "uint256", "name": "_amount", "indexed": false, "unit": "wei"}], "anonymous": false, "type": "event"}, {"name": "NotifyString", "inputs": [{"type": "string", "name": "content", "indexed": false}], "anonymous": false, "type": "event"}, {"name": "NotifyBytes32", "inputs": [{"type": "bytes32", "name": "content", "indexed": false}], "anonymous": false, "type": "event"}, {"name": "NotifyBool", "inputs": [{"type": "bool", "name": "content", "indexed": false}], "anonymous": false, "type": "event"}, {"name": "NotifyUint256", "inputs": [{"type": "uint256", "name": "content", "indexed": false}], "anonymous": false, "type": "event"}, {"name": "NotifyAddress", "inputs": [{"type": "address", "name": "content", "indexed": false}], "anonymous": false, "type": "event"}, {"outputs": [], "inputs": [{"type": "address", "name": "incognitoProxyAddress"}], "constant": false, "payable": false, "type": "constructor"}, {"name": "deposit", "outputs": [], "inputs": [{"type": "string", "name": "incognito_address"}], "constant": false, "payable": true, "type": "function", "gas": 25690}, {"name": "depositERC20", "outputs": [], "inputs": [{"type": "address", "name": "token"}, {"type": "uint256", "name": "amount"}, {"type": "string", "name": "incognito_address"}], "constant": false, "payable": false, "type": "function", "gas": 27797}, {"name": "parseBurnInst", "outputs": [{"type": "uint256", "name": "out"}, {"type": "address", "name": "out"}, {"type": "address", "name": "out"}, {"type": "uint256", "name": "out"}], "inputs": [{"type": "bytes", "name": "inst"}], "constant": true, "payable": false, "type": "function", "gas": 2681}, {"name": "withdraw", "outputs": [], "inputs": [{"type": "bytes", "name": "inst"}, {"type": "uint256", "name": "beaconHeight"}, {"type": "bytes32[8]", "name": "beaconInstPath"}, {"type": "bool[8]", "name": "beaconInstPathIsLeft"}, {"type": "int128", "name": "beaconInstPathLen"}, {"type": "bytes32", "name": "beaconInstRoot"}, {"type": "bytes32", "name": "beaconBlkData"}, {"type": "bytes32", "name": "beaconBlkHash"}, {"type": "uint256", "name": "beaconSignerSig"}, {"type": "int128", "name": "beaconNumR"}, {"type": "uint256[8]", "name": "beaconXs"}, {"type": "uint256[8]", "name": "beaconYs"}, {"type": "int128[8]", "name": "beaconRIdxs"}, {"type": "int128", "name": "beaconNumSig"}, {"type": "uint256[8]", "name": "beaconSigIdxs"}, {"type": "uint256", "name": "beaconRx"}, {"type": "uint256", "name": "beaconRy"}, {"type": "bytes", "name": "beaconR"}, {"type": "uint256", "name": "bridgeHeight"}, {"type": "bytes32[8]", "name": "bridgeInstPath"}, {"type": "bool[8]", "name": "bridgeInstPathIsLeft"}, {"type": "int128", "name": "bridgeInstPathLen"}, {"type": "bytes32", "name": "bridgeInstRoot"}, {"type": "bytes32", "name": "bridgeBlkData"}, {"type": "bytes32", "name": "bridgeBlkHash"}, {"type": "uint256", "name": "bridgeSignerSig"}, {"type": "int128", "name": "bridgeNumR"}, {"type": "uint256[8]", "name": "bridgeXs"}, {"type": "uint256[8]", "name": "bridgeYs"}, {"type": "int128[8]", "name": "bridgeRIdxs"}, {"type": "int128", "name": "bridgeNumSig"}, {"type": "uint256[8]", "name": "bridgeSigIdxs"}, {"type": "uint256", "name": "bridgeRx"}, {"type": "uint256", "name": "bridgeRy"}, {"type": "bytes", "name": "bridgeR"}], "constant": false, "payable": false, "type": "function", "gas": 100530}, {"name": "withdrawed", "outputs": [{"type": "bool", "name": "out"}], "inputs": [{"type": "bytes32", "name": "arg0"}], "constant": true, "payable": false, "type": "function", "gas": 736}, {"name": "incognito", "outputs": [{"type": "address", "unit": "Incognito_proxy", "name": "out"}], "inputs": [], "constant": true, "payable": false, "type": "function", "gas": 633}]`
	BridgeShardID         = 1
	EthAddrStr            = "0x0000000000000000000000000000000000000000"
	EthContractAddressStr = "0xe97f7d3f866cb8200941082c887c7ae5eeab1f58"
)

const (
	BridgeRequestNotFoundStatus   = 0
	BridgeRequestProcessingStatus = 1
	BridgeRequestAcceptedStatus   = 2
	BridgeRequestRejectedStatus   = 3
)
