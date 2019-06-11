MAX_COMM_HEIGHT: constant(uint256) = 3 # support up to 2 ** 3 = 8 committee members
MAX_COMM_SIZE: constant(uint256) = 2 ** MAX_COMM_HEIGHT
INST_LENGTH: constant(uint256) = 100
BEACON_BLOCK_LENGTH: constant(uint256) = 1000
BRIDGE_BLOCK_LENGTH: constant(uint256) = 1000

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

@public
def swapBeacon(
    newComRoot: bytes32,
    inst: bytes[INST_LENGTH], # content of swap instruction
    beaconInstPath: bytes32[MAX_COMM_HEIGHT],
    beaconPathIsLeft: uint256[MAX_COMM_HEIGHT],
    beaconInstRoot: bytes32,
    beaconBlkData: bytes[BEACON_BLOCK_LENGTH], # the rest of the beacon block
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    beaconSignerSig: bytes32, # aggregated signature of some committee members
    beaconSignerPaths: bytes32[MAX_COMM_HEIGHT],
    bridgeInstPath: bytes32[MAX_COMM_HEIGHT],
    bridgePathIsLeft: uint256[MAX_COMM_HEIGHT],
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes[BRIDGE_BLOCK_LENGTH], # the rest of the bridge block
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    bridgeSignerSig: bytes32,
    bridgeSignerPaths: bytes32[MAX_COMM_HEIGHT],
) -> bool:
    # Check if beaconInst is in beaconInstRoot
    # Check if beaconInstRoot is in beaconBlkHash
    # Check if beaconSignerSig is valid
    # Check if beaconSignerPubkeys are in beaconCommRoot

    # Check if bridgeInst is in bridgeInstRoot
    # Check if bridgeInstRoot is in bridgeBlkHash
    # Check if bridgeSignerSig is valid
    # Check if bridgeSignerPubkeys are in bridgeCommRoot
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

