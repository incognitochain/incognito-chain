contract Incognito_proxy:
    def instructionApproved(instHash: bytes32, beaconInstPath: bytes32[4], beaconInstPathIsLeft: bool[4], beaconInstPathLen: int128, beaconInstRoot: bytes32, beaconBlkData: bytes32, beaconBlkHash: bytes32, beaconSignerPubkeys: bytes[528], beaconSignerCount: int128, beaconSignerSig: bytes32, beaconSignerPaths: bytes32[64], beaconSignerPathIsLeft: bool[64], beaconSignerPathLen: int128, bridgeInstPath: bytes32[4], bridgeInstPathIsLeft: bool[4], bridgeInstPathLen: int128, bridgeInstRoot: bytes32, bridgeBlkData: bytes32, bridgeBlkHash: bytes32, bridgeSignerPubkeys: bytes[528], bridgeSignerCount: int128, bridgeSignerSig: bytes32, bridgeSignerPaths: bytes32[64], bridgeSignerPathIsLeft: bool[64], bridgeSignerPathLen: int128) -> bool: modifying


# All these constants must mimic incognity_proxy
MAX_PATH: constant(uint256) = 4
COMM_SIZE: constant(uint256) = 2 ** MAX_PATH
TOTAL_PUBKEY: constant(uint256) = COMM_SIZE * MAX_PATH
PUBKEY_SIZE: constant(int128) = 33
PUBKEY_LENGTH: constant(int128) = PUBKEY_SIZE * COMM_SIZE
INST_LENGTH: constant(uint256) = 120

Deposit: event({_from: indexed(address), _incognito_address: string[64], _amount: wei_value})
Withdraw: event({_to: indexed(address), _amount: wei_value})

withdrawed: public(map(bytes32, bool))
incognito: public(Incognito_proxy)

@public
def __init__(incognitoProxyAddress: address):
    self.incognito = Incognito_proxy(incognitoProxyAddress)

@public
@payable
def deposit(incognito_address: string[64]):
    log.Deposit(msg.sender, incognito_address, msg.value)

@constant
@public
def parseBurnInst(inst: bytes[INST_LENGTH]) -> (uint256, bytes32, address, uint256):
    type: uint256 = convert(slice(inst, start=0, len=3), uint256)
    tokenID: bytes32 = extract32(inst, 3, type=bytes32)
    to: address = extract32(inst, 35, type=address)
    amount: uint256 = extract32(inst, 55, type=uint256)
    # tokenID: bytes32 = convert(slice(inst, start=3, len=32), bytes32)
    # to: address = convert(slice(inst, start=35, len=20), address)
    # amount: uint256 = convert(slice(inst, start=55, len=32), uint256)
    return type, tokenID, to, amount

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
    amount: uint256 = 0
    type, tokenID, to, amount = self.parseBurnInst(inst)

    # TODO: check type and tokenID

    # Each instruction can only by redeemed once
    instHash: bytes32 = keccak256(inst)
    assert self.withdrawed[instHash] == False

    # Check if instruction is approved on Incognito
    if self.incognito.instructionApproved(
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
    ):
        self.withdrawed[instHash] = True

    # Check if balance is enough
    assert self.balance >= amount

    # Send and notify
    self.withdrawed[instHash] = True
    send(to, amount)
    log.Withdraw(to, amount)

