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

