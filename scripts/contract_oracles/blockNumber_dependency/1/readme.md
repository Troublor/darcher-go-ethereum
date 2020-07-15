```solidity
contract VulContract {
    event send(uint number);
    
    constructor () public payable {}
    
    function sendWithBlockNumberBefore() public payable {
       // actually this is non buggy but anyway, the oracle in ContractFuzzer says this
        uint number = block.number;
        emit send(number);
        msg.sender.transfer(100);
    }
   
    function sendWithBlockNumberAfter() public payable {
        uint number = block.number;
        emit send(number);
        msg.sender.transfer(0);
    }
}
```