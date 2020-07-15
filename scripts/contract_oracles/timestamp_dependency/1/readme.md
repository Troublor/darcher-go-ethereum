```solidity
contract VulContract {
    event send(uint time);
    
    constructor () public payable {}
    
    function sendWithTimestampBefore() public payable {
       // actually this is non buggy but anyway, the oracle in ContractFuzzer says this
        uint time = now;
        emit send(time);
        msg.sender.transfer(100);
    }
   
    function sendWithTimestampAfter() public payable {
        msg.sender.transfer(100);
        uint time = now;
        emit send(time);
    }
}
```