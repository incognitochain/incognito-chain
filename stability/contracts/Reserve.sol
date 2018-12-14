pragma solidity ^0.5.0;

contract Reserve {
    address payable public owner;
    address payable public manager;

    event __raise(address sender, uint256 value, bytes coinReceiver, bytes32 offchain);
    event __spend(uint256 amount, bytes32 offchain);

    modifier onlyManager() {
        require(msg.sender == manager, "only managers are authorized");
        _;
    }

    modifier managerOrOwner() {
        require(msg.sender == manager || msg.sender == owner, "only managers and owner are authorized");
        _;
    }

    constructor(address payable _manager, address payable _owner) public {
        owner = _owner;
        manager = _manager;
    }

    function raise(bytes memory coinReceiver, bytes32 offchain) public payable {
        emit __raise(msg.sender, msg.value, coinReceiver, offchain);
    }

    function spend(uint256 amount, bytes32 offchain) public managerOrOwner {
        owner.transfer(amount); // TODO: transfer to somewhere else?
        emit __spend(amount, offchain);
    }
}
