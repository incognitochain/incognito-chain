COMM_PATH_LENGTH: constant(uint256) = 1 # support up to 2 ** COMM_PATH_LENGTH committee members
COMM_SIZE: constant(uint256) = 2 ** COMM_PATH_LENGTH
PUBKEY_LENGTH: constant(uint256) = COMM_SIZE * COMM_PATH_LENGTH

INST_PATH_LENGTH: constant(uint256) = 1 # up to 2 ** INST_PATH_LENGTH instructions
INST_LENGTH: constant(uint256) = 100

MIN_SIGN: constant(uint256) = 2

NotifyString: event({content: string[100]})
NotifyBytes32: event({content: bytes32})
NotifyBool: event({content: bool})

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
def inMerkleTree(leaf: bytes32, root: bytes32, path: bytes32[COMM_PATH_LENGTH], left: bool[COMM_PATH_LENGTH]) -> bool:
    hash: bytes32 = leaf
    for i in range(COMM_PATH_LENGTH):
        if left[i]:
            hash = keccak256(concat(path[i], hash))
        else:
            hash = keccak256(concat(hash, path[i]))
    return hash == root

@constant
@public
def verifyInst(
    commRoot: bytes32,
    instHash: bytes32,
    instPath: bytes32[INST_PATH_LENGTH],
    instPathIsLeft: bool[INST_PATH_LENGTH],
    instRoot: bytes32,
    blkHash: bytes32,
    signerPubkeys: bytes32[COMM_SIZE],
    signerSig: bytes32,
    signerPaths: bytes32[PUBKEY_LENGTH],
    signerPathIsLeft: bool[PUBKEY_LENGTH]
) -> bool:
    # Check if inst is in merkle tree with root instRoot
    if not self.inMerkleTree(instHash, instRoot, instPath, instPathIsLeft):
        log.NotifyString("instruction is not in merkle tree")
        return False

    # TODO: Check if signerSig is valid

    # Check if signerPubkeys are in merkle tree with root commRoot
    count: uint256 = 0
    for i in range(COMM_SIZE):
        if convert(signerPubkeys[i], uint256) == 0:
            continue

        path: bytes32[COMM_PATH_LENGTH]
        left: bool[COMM_PATH_LENGTH]
        h: int128 = convert(COMM_PATH_LENGTH, int128)
        for j in range(COMM_PATH_LENGTH):
            path[j] = signerPaths[i * h + j]
            left[j] = signerPathIsLeft[i * h + j]

        if not self.inMerkleTree(signerPubkeys[i], commRoot, path, left):
            log.NotifyString("pubkey not in merkle tree")
            return False

        count += 1

    # Check if enough validators signed this block
    if count < MIN_SIGN:
        log.NotifyString("not enough sig")
        return False

    return True

@public
def swapBeacon(
    newCommRoot: bytes32,
    inst: bytes[INST_LENGTH], # content of swap instruction
    beaconInstPath: bytes32[INST_PATH_LENGTH],
    beaconInstPathIsLeft: bool[INST_PATH_LENGTH],
    beaconInstRoot: bytes32,
    beaconBlkData: bytes32, # hash of the rest of the beacon block
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes32[COMM_SIZE],
    beaconSignerSig: bytes32, # aggregated signature of some committee members
    beaconSignerPaths: bytes32[PUBKEY_LENGTH],
    beaconSignerPathIsLeft: bool[PUBKEY_LENGTH],
    bridgeInstPath: bytes32[INST_PATH_LENGTH],
    bridgeInstPathIsLeft: bool[INST_PATH_LENGTH],
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes32, # hash of the rest of the bridge block
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes32[COMM_SIZE],
    bridgeSignerSig: bytes32,
    bridgeSignerPaths: bytes32[PUBKEY_LENGTH],
    bridgeSignerPathIsLeft: bool[PUBKEY_LENGTH]
) -> bool:
    # TODO: remove newCommRoot, parse from inst instead

    # Check if beaconInstRoot is in block with hash beaconBlkHash
    instHash: bytes32 = keccak256(inst)
    blk: bytes32 = keccak256(concat(beaconInstRoot, beaconBlkData))
    if not blk == beaconBlkHash:
        log.NotifyString("instruction merkle root is not in beacon block")
        log.NotifyBytes32(instHash)
        log.NotifyBytes32(blk)
        raise "instruction merkle root is not in beacon block"

    # Check that inst is in beacon block
    if not self.verifyInst(
        self.beaconCommRoot,
        instHash,
        beaconInstPath,
        beaconInstPathIsLeft,
        beaconInstRoot,
        beaconBlkHash,
        beaconSignerPubkeys,
        beaconSignerSig,
        beaconSignerPaths,
        beaconSignerPathIsLeft
    ):
        log.NotifyString("failed verifying beacon instruction")
        raise "failed verifying beacon instruction"

    # Check if bridgeInstRoot is in block with hash bridgeBlkHash
    blk = keccak256(concat(bridgeInstRoot, bridgeBlkData))
    if not blk == bridgeBlkHash:
        log.NotifyString("instruction merkle root is not in bridge block")
        raise "instruction merkle root is not in bridge block"

    # Check that inst is in bridge block
    if not self.verifyInst(
        self.bridgeCommRoot,
        instHash,
        bridgeInstPath,
        bridgeInstPathIsLeft,
        bridgeInstRoot,
        bridgeBlkHash,
        bridgeSignerPubkeys,
        bridgeSignerSig,
        bridgeSignerPaths,
        bridgeSignerPathIsLeft
    ):
        log.NotifyString("failed verify bridge instruction")
        raise "failed verify bridge instruction"

    # # Update beacon committee merkle root
    self.beaconCommRoot = newCommRoot
    log.NotifyString("no exeception...")
    return True

