MAX_COMM_HEIGHT: constant(uint256) = 1 # support up to 2 ** 3 = 8 committee members
MAX_COMM_SIZE: constant(uint256) = 2 ** MAX_COMM_HEIGHT
PUBKEY_LENGTH: constant(uint256) = MAX_COMM_SIZE * MAX_COMM_HEIGHT
INST_LENGTH: constant(uint256) = 100
BEACON_BLOCK_LENGTH: constant(uint256) = 1000
BRIDGE_BLOCK_LENGTH: constant(uint256) = 1000
MIN_SIGN: constant(uint256) = 1

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
    return MAX_COMM_SIZE

@constant
@public
def parseSwapBeaconInst(inst: bytes[INST_LENGTH]) -> bytes32[MAX_COMM_SIZE]:
    comm: bytes32[MAX_COMM_SIZE]
    return comm

@constant
@public
def inMerkleTree(leaf: bytes32, root: bytes32, path: bytes32[MAX_COMM_HEIGHT], left: bool[MAX_COMM_HEIGHT]) -> bool:
    hash: bytes32 = leaf
    for i in range(MAX_COMM_HEIGHT):
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

@public
def swapBeacon(
    newCommRoot: bytes32,
    inst: bytes[INST_LENGTH], # content of swap instruction
    beaconInstPath: bytes32[MAX_COMM_HEIGHT],
    beaconInstPathIsLeft: bool[MAX_COMM_HEIGHT],
    beaconInstRoot: bytes32,
    beaconBlkData: bytes[BEACON_BLOCK_LENGTH], # the rest of the beacon block
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    beaconSignerSig: bytes32, # aggregated signature of some committee members
    beaconSignerPaths: bytes32[PUBKEY_LENGTH],
    beaconSignerPathIsLeft: bool[PUBKEY_LENGTH],
    bridgeInstPath: bytes32[MAX_COMM_HEIGHT],
    bridgeInstPathIsLeft: bool[MAX_COMM_HEIGHT],
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes[BRIDGE_BLOCK_LENGTH], # the rest of the bridge block
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    bridgeSignerSig: bytes32,
    bridgeSignerPaths: bytes32[PUBKEY_LENGTH],
    bridgeSignerPathIsLeft: bool[PUBKEY_LENGTH]
) -> bool:
    # Check if beaconInst is in merkle tree with root beaconInstRoot
    instHash: bytes32 = sha3(inst)
    if not self.inMerkleTree(instHash, beaconInstRoot, beaconInstPath, beaconInstPathIsLeft):
        raise "instruction is not in merkle tree"

    # Check if beaconInstRoot is in beaconBlkHash
    blkHash: bytes32 = sha3(concat(instHash, beaconBlkData))
    if not blkHash == beaconBlkHash:
        raise "instruction merkle root is not in block"

    # TODO: Check if beaconSignerSig is valid

    # Check if beaconSignerPubkeys are in beaconCommRoot
    count: uint256 = 0
    for i in range(MAX_COMM_SIZE):
        if convert(beaconSignerPubkeys[i], uint256) == 0:
            continue

        path: bytes32[MAX_COMM_HEIGHT]
        left: bool[MAX_COMM_HEIGHT]
        h: int128 = convert(MAX_COMM_HEIGHT, int128)
        for j in range(MAX_COMM_HEIGHT):
            path[j] = beaconSignerPaths[i * h + j]
            left[j] = beaconSignerPathIsLeft[i * h + j]

        if not self.inMerkleTree(beaconSignerPubkeys[i], self.beaconCommRoot, path, left):
            assert 1 == 0

        count += 1

    # Check if enough validators signed this block
    if count < MIN_SIGN:
        raise "not enough signature"

    # Check if bridgeInst is in merkle tree with root bridgeInstRoot
    # Check if bridgeInstRoot is in bridgeBlkHash
    # Check if bridgeSignerSig is valid
    # Check if bridgeSignerPubkeys are in bridgeCommRoot

    # Update beacon committee merkle root
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

