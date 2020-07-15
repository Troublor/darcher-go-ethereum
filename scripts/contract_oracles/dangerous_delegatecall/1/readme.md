```solidity
struct Data {
    bool changed;
}

contract ValContract {
    Data data;
    function delegateCall(address target) public {
        (bool success, bytes memory ret ) = target.delegatecall(abi.encodeWithSignature("action()"));
        require(success);
    }
    
    function getDataChanged () view public returns (bool) {
        return data.changed;
    }
}

contract Attacker {
    Data data;
    function action() public {
        data.changed = true;
    }
    function getDataChanged () view public returns (bool) {
        return data.changed;
    }
}
```