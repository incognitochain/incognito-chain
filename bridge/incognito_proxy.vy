COMM_PATH_LENGTH: constant(uint256) = 1 # support up to 2 ** COMM_PATH_LENGTH committee members
COMM_SIZE: constant(uint256) = 2 ** COMM_PATH_LENGTH
PUBKEY_LENGTH: constant(uint256) = COMM_SIZE * COMM_PATH_LENGTH

INST_PATH_LENGTH: constant(uint256) = 1 # up to 2 ** INST_PATH_LENGTH instructions
INST_LENGTH: constant(uint256) = 100

MIN_SIGN: constant(uint256) = 2

Transfer: event({_from: indexed(address), _to: indexed(address), _value: uint256})
Approve: event({_owner: indexed(address), _spender: indexed(address), _value: uint256})

beaconCommRoot: public(bytes32)
bridgeCommRoot: public(bytes32)

name: public(string[10])
symbol: public(string[10])
decimals: public(uint256)
totalSupply: public(uint256)
balanceOf: public(map(address, uint256))
allowance: public(map(address, map(address, uint256)))

@public
def __init__(_name: string[10], _symbol: string[10], _decimals: uint256, _totalSupply: uint256):
    self.name = _name
    self.symbol = _symbol
    self.decimals = _decimals
    self.totalSupply = _totalSupply
    self.balanceOf[msg.sender] = _totalSupply

@constant
@public
def get() -> uint256:
    return COMM_SIZE

@constant
@public
def parseSwapBeaconInst(inst: bytes[INST_LENGTH]) -> bytes32[COMM_SIZE]:
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
def getHash(inst: bytes[INST_LENGTH]) -> bytes32:
    return keccak256(inst)

@constant
@public
def getHash256(inst: bytes[INST_LENGTH]) -> bytes32:
    return sha256(inst)

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
        raise "instruction is not in merkle tree"

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
            assert 1 == 0

        count += 1

    # Check if enough validators signed this block
    if count < MIN_SIGN:
        raise "not enough signature"

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
    # Check if beaconInstRoot is in block with hash beaconBlkHash
    instHash: bytes32 = sha3(inst)
    blk: bytes32 = sha3(concat(instHash, beaconBlkData))
    if not blk == beaconBlkHash:
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
        raise "failed verify beacon instruction"

    # Check if bridgeInstRoot is in block with hash bridgeBlkHash
    blk = sha3(concat(instHash, bridgeBlkData))
    if not blk == bridgeBlkHash:
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
        raise "failed verify bridge instruction"

    # Update beacon committee merkle root
    self.beaconCommRoot = newCommRoot
    return True

@private
def _transfer(_from: address, _to: address, _value: uint256) -> bool:
    assert self.balanceOf[_from] >= _value
    self.balanceOf[_from] -= _value
    self.balanceOf[_to] += _value
    log.Transfer(_from, _to, _value)
    return True

@public
def transfer(_to: address, _value: uint256) -> bool:
    return self._transfer(msg.sender, _to, _value)

@public
def transferFrom(_from: address, _to: address, _value: uint256) -> bool:
    assert self.allowance[_from][_to] >= _value
    self.allowance[_from][_to] -= _value
    return self._transfer(_from, _to, _value)

@public
def approve(_spender: address, _value: uint256) -> bool:
    self.allowance[msg.sender][_spender] += _value
    log.Approve(msg.sender, _spender, _value)
    return True

