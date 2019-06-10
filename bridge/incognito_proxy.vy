MAX_COMM_HEIGHT: constant(uint256) = 3 # support up to 2 ** 3 = 8 committee members
INST_LENGTH: constant(uint256) = 100
BEACON_BLOCK_LENGTH: constant(uint256) = 1000
BRIDGE_BLOCK_LENGTH: constant(uint256) = 1000

Transfer: event({_from: indexed(address), _to: indexed(address), _value: uint256})
Approve: event({_owner: indexed(address), _spender: indexed(address), _value: uint256})

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

@public
def swapBeacon(
    newComRoot: bytes32,
    beaconInst: bytes[INST_LENGTH], # content of swap instruction in beacon block
    beaconInstPath: bytes32[MAX_COMM_HEIGHT],
    beaconPathIsLeft: uint256[MAX_COMM_HEIGHT],
    beaconInstRoot: bytes32,
    beaconBlkData: bytes[BEACON_BLOCK_LENGTH], # the rest of the beacon block
    beaconBlkHash: bytes32,
    beaconSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    beaconSignerSigs: bytes32, # aggregated signature of some committee members
    beaconSignerPaths: bytes32[MAX_COMM_HEIGHT],
    bridgeInst: bytes[INST_LENGTH],
    bridgeInstPath: bytes32[MAX_COMM_HEIGHT],
    bridgePathIsLeft: uint256[MAX_COMM_HEIGHT],
    bridgeInstRoot: bytes32,
    bridgeBlkData: bytes[BRIDGE_BLOCK_LENGTH], # the rest of the bridge block
    bridgeBlkHash: bytes32,
    bridgeSignerPubkeys: bytes32[MAX_COMM_HEIGHT],
    bridgeSignerSigs: bytes32,
    bridgeSignerPaths: bytes32[MAX_COMM_HEIGHT],
) -> bool:
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

