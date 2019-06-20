MAX_PATH: constant(uint256) = 3 # support up to 2 ** MAX_PATH committee members
COMM_SIZE: constant(uint256) = 2 ** MAX_PATH

TOTAL_PUBKEY: constant(uint256) = COMM_SIZE * MAX_PATH
PUBKEY_SIZE: constant(int128) = 33 # each pubkey is 33 bytes
PUBKEY_LENGTH: constant(int128) = PUBKEY_SIZE * COMM_SIZE # length of the array storing all pubkeys

INST_LENGTH: constant(uint256) = 100

MIN_SIGN: constant(uint256) = 2

NotifyString: event({content: string[100]})
NotifyBytes32: event({content: bytes32})
NotifyBool: event({content: bool})
NotifyUint256: event({content: uint256})

beaconCommRoot: public(bytes32)
bridgeCommRoot: public(bytes32)

@public
def __init__(_beaconCommRoot: bytes32, _bridgeCommRoot: bytes32):
    self.beaconCommRoot = _beaconCommRoot
    self.bridgeCommRoot = _bridgeCommRoot

@constant
@public
def parseSwapBeaconInst(inst: bytes[INST_LENGTH]) -> bytes32[COMM_SIZE]:
    # TODO: implement
    comm: bytes32[COMM_SIZE]
    return comm

@constant
@public
def inMerkleTree(leaf: bytes32, root: bytes32, path: bytes32[MAX_PATH], left: bool[MAX_PATH], length: int128) -> bool:
    hash: bytes32 = leaf
    for i in range(MAX_PATH):
        if i >= length:
            log.NotifyUint256(convert(i, uint256))
            break
        log.NotifyBytes32(path[i])
        if left[i]:
            hash = keccak256(concat(path[i], hash))
        elif convert(path[i], uint256) == 0:
            hash = keccak256(concat(hash, hash))
        else:
            hash = keccak256(concat(hash, path[i]))
        log.NotifyBytes32(hash)
    return hash == root

@constant
@public
def verifyInst(
    commRoot: bytes32,
    instHash: bytes32,
    instPath: bytes32[MAX_PATH],
    instPathIsLeft: bool[MAX_PATH],
    instPathLen: int128,
    instRoot: bytes32,
    blkHash: bytes32,
    signerPubkeys: bytes[PUBKEY_LENGTH],
    signerCount: int128,
    signerSig: bytes32,
    signerPaths: bytes32[TOTAL_PUBKEY],
    signerPathIsLeft: bool[TOTAL_PUBKEY],
    signerPathLen: int128
) -> bool:
    # Check if enough validators signed this block
    if signerCount < MIN_SIGN:
        log.NotifyString("not enough sig")
        return False

    # Check if inst is in merkle tree with root instRoot
    if not self.inMerkleTree(instHash, instRoot, instPath, instPathIsLeft, instPathLen):
        log.NotifyString("instruction is not in merkle tree")
        return False

    # TODO: Check if signerSig is valid

    # Check if signerPubkeys are in merkle tree with root commRoot
    for i in range(COMM_SIZE):
        if i >= signerCount:
            break

        # Get hash of the pubkey
        pk: bytes[PUBKEY_SIZE] = slice(signerPubkeys, start=i * PUBKEY_SIZE, len=PUBKEY_SIZE)
        signerPubkeyHash: bytes32 = keccak256(pk)

        path: bytes32[MAX_PATH]
        left: bool[MAX_PATH]
        for j in range(MAX_PATH):
            if j > signerPathLen:
                break
            path[j] = signerPaths[i * signerPathLen + j]
            left[j] = signerPathIsLeft[i * signerPathLen + j]

        log.NotifyBytes32(signerPubkeyHash)
        if not self.inMerkleTree(signerPubkeyHash, commRoot, path, left, signerPathLen):
            log.NotifyString("pubkey not in merkle tree")
            return False

    return True

@public
def swapBeacon(
    newCommRoot: bytes32,
    inst: bytes[INST_LENGTH], # content of swap instruction
    beaconInstPath: bytes32[MAX_PATH],
    beaconInstPathIsLeft: bool[MAX_PATH],
    beaconInstPathLen: int128,
    beaconInstRoot: bytes32,
    beaconBlkData: bytes32, # hash of the rest of the beacon block
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes[PUBKEY_LENGTH],
    beaconSignerCount: int128,
    beaconSignerSig: bytes32, # aggregated signature of some committee members
    beaconSignerPaths: bytes32[TOTAL_PUBKEY],
    beaconSignerPathIsLeft: bool[TOTAL_PUBKEY],
    beaconSignerPathLen: int128,
    bridgeInstPath: bytes32[MAX_PATH],
    bridgeInstPathIsLeft: bool[MAX_PATH],
    bridgeInstPathLen: int128,
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes32, # hash of the rest of the bridge block
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes[PUBKEY_LENGTH],
    bridgeSignerCount: int128,
    bridgeSignerSig: bytes32,
    bridgeSignerPaths: bytes32[TOTAL_PUBKEY],
    bridgeSignerPathIsLeft: bool[TOTAL_PUBKEY],
    bridgeSignerPathLen: int128
) -> bool:
    # TODO: remove newCommRoot, parse from inst instead

    # Check if beaconInstRoot is in block with hash beaconBlkHash
    instHash: bytes32 = keccak256(inst)
    blk: bytes32 = keccak256(concat(beaconBlkData, beaconInstRoot))
    if not blk == beaconBlkHash:
        log.NotifyString("instruction merkle root is not in beacon block")
        log.NotifyBytes32(beaconInstRoot)
        log.NotifyBytes32(beaconBlkData)
        log.NotifyBytes32(instHash)
        log.NotifyBytes32(blk)
        # raise "instruction merkle root is not in beacon block"

    # Check that inst is in beacon block
    if not self.verifyInst(
        self.beaconCommRoot,
        instHash,
        beaconInstPath,
        beaconInstPathIsLeft,
        beaconInstPathLen,
        beaconInstRoot,
        beaconBlkHash,
        beaconSignerPubkeys,
        beaconSignerCount,
        beaconSignerSig,
        beaconSignerPaths,
        beaconSignerPathIsLeft,
        beaconSignerPathLen
    ):
        log.NotifyString("failed verifying beacon instruction")
        # raise "failed verifying beacon instruction"

    # Check if bridgeInstRoot is in block with hash bridgeBlkHash
    blk = keccak256(concat(bridgeBlkData, bridgeInstRoot))
    if not blk == bridgeBlkHash:
        log.NotifyString("instruction merkle root is not in bridge block")
        # raise "instruction merkle root is not in bridge block"

    # Check that inst is in bridge block
    if not self.verifyInst(
        self.bridgeCommRoot,
        instHash,
        bridgeInstPath,
        bridgeInstPathIsLeft,
        bridgeInstPathLen,
        bridgeInstRoot,
        bridgeBlkHash,
        bridgeSignerPubkeys,
        bridgeSignerCount,
        bridgeSignerSig,
        bridgeSignerPaths,
        bridgeSignerPathIsLeft,
        bridgeSignerPathLen
    ):
        log.NotifyString("failed verify bridge instruction")
        raise "failed verify bridge instruction"

    # # # Update beacon committee merkle root
    # self.beaconCommRoot = newCommRoot
    log.NotifyString("no exeception...")
    return True

