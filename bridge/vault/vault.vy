contract Incognito_proxy:
    def instructionApproved(instHash: bytes32, beaconInstPath: bytes32[4], beaconInstPathIsLeft: bool[4], beaconInstPathLen: int128, beaconInstRoot: bytes32, beaconBlkData: bytes32, beaconBlkHash: bytes32, beaconSignerPubkeys: bytes[528], beaconSignerCount: int128, beaconSignerSig: bytes32, beaconSignerPaths: bytes32[64], beaconSignerPathIsLeft: bool[64], beaconSignerPathLen: int128, bridgeInstPath: bytes32[4], bridgeInstPathIsLeft: bool[4], bridgeInstPathLen: int128, bridgeInstRoot: bytes32, bridgeBlkData: bytes32, bridgeBlkHash: bytes32, bridgeSignerPubkeys: bytes[528], bridgeSignerCount: int128, bridgeSignerSig: bytes32, bridgeSignerPaths: bytes32[64], bridgeSignerPathIsLeft: bool[64], bridgeSignerPathLen: int128) -> bool: modifying


# All these constants must mimic incognity_proxy
MAX_PATH: constant(uint256) = 4
COMM_SIZE: constant(uint256) = 2 ** MAX_PATH
TOTAL_PUBKEY: constant(uint256) = COMM_SIZE * MAX_PATH
PUBKEY_SIZE: constant(int128) = 33
PUBKEY_LENGTH: constant(int128) = PUBKEY_SIZE * COMM_SIZE
INST_LENGTH: constant(uint256) = 150
INC_ADDRESS_LENGTH: constant(uint256) = 128

Deposit: event({_from: indexed(address), _incognito_address: string[INC_ADDRESS_LENGTH], _amount: wei_value})
Withdraw: event({_to: indexed(address), _amount: wei_value})


NotifyString: event({content: string[128]})
NotifyBytes32: event({content: bytes32})
NotifyBool: event({content: bool})
NotifyUint256: event({content: uint256})
NotifyAddress: event({content: address})


withdrawed: public(map(bytes32, bool))
incognito: public(Incognito_proxy)

@public
def __init__(incognitoProxyAddress: address):
    self.incognito = Incognito_proxy(incognitoProxyAddress)

@public
@payable
def deposit(incognito_address: string[INC_ADDRESS_LENGTH]):
    log.Deposit(msg.sender, incognito_address, msg.value)

@constant
@public
def parseBurnInst(inst: bytes[INST_LENGTH]) -> (uint256, bytes32, address, uint256):
    type: uint256 = convert(slice(inst, start=0, len=3), uint256)
    tokenID: bytes32 = extract32(inst, 3, type=bytes32)
    to: address = extract32(inst, 35, type=address)
    amount: uint256 = extract32(inst, 67, type=uint256)
    return type, tokenID, to, amount

@constant
@public
def testExtract(a: bytes[INST_LENGTH]) -> (address, wei_value):
    x: address = extract32(a, 0, type=address)
    s: uint256 = 12345
    t: wei_value = as_wei_value(s, "gwei")
    return x, t

@public
def withdraw(
    inst: bytes[INST_LENGTH],
    beaconInstPath: bytes32[MAX_PATH],
    beaconInstPathIsLeft: bool[MAX_PATH],
    beaconInstPathLen: int128,
    beaconInstRoot: bytes32,
    beaconBlkData: bytes32,
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes[PUBKEY_LENGTH],
    beaconSignerCount: int128,
    beaconSignerSig: bytes32,
    beaconSignerPaths: bytes32[TOTAL_PUBKEY],
    beaconSignerPathIsLeft: bool[TOTAL_PUBKEY],
    beaconSignerPathLen: int128,
    bridgeInstPath: bytes32[MAX_PATH],
    bridgeInstPathIsLeft: bool[MAX_PATH],
    bridgeInstPathLen: int128,
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes32,
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes[PUBKEY_LENGTH],
    bridgeSignerCount: int128,
    bridgeSignerSig: bytes32,
    bridgeSignerPaths: bytes32[TOTAL_PUBKEY],
    bridgeSignerPathIsLeft: bool[TOTAL_PUBKEY],
    bridgeSignerPathLen: int128
):
    type: uint256 = 0
    tokenID: bytes32
    to: address
    burned: uint256 = 0
    type, tokenID, to, burned = self.parseBurnInst(inst)
    # log.NotifyUint256(type)
    # log.NotifyBytes32(tokenID)
    # log.NotifyAddress(to)
    # log.NotifyUint256(burned)

    # Check type and tokenID
    assert type == 3617328 # Burn metadata and shardID of bridge
    assert tokenID == 0x0500000000000000000000000000000000000000000000000000000000000000

    # Each instruction can only by redeemed once
    instHash: bytes32 = keccak256(inst)
    assert self.withdrawed[instHash] == False

    # Check if instruction is approved on Incognito
    assert self.incognito.instructionApproved(
        instHash,
        beaconInstPath,
        beaconInstPathIsLeft,
        beaconInstPathLen,
        beaconInstRoot,
        beaconBlkData,
        beaconBlkHash,
        beaconSignerPubkeys,
        beaconSignerCount,
        beaconSignerSig,
        beaconSignerPaths,
        beaconSignerPathIsLeft,
        beaconSignerPathLen,
        bridgeInstPath,
        bridgeInstPathIsLeft,
        bridgeInstPathLen,
        bridgeInstRoot,
        bridgeBlkData,
        bridgeBlkHash,
        bridgeSignerPubkeys,
        bridgeSignerCount,
        bridgeSignerSig,
        bridgeSignerPaths,
        bridgeSignerPathIsLeft,
        bridgeSignerPathLen
    )

    # Check if balance is enough
    amount: wei_value = as_wei_value(burned, "gwei")
    assert self.balance >= amount

    # Send and notify
    self.withdrawed[instHash] = True
    send(to, amount)
    log.Withdraw(to, amount)

