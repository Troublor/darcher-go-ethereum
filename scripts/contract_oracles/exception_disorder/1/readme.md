This folder contains one test cases (one transaction). 

The transaction will detect the exception disorder vulnerability in the contract.

```solidity
contract VulContract {
    mapping(address=>uint256) recv;
    constructor()public payable{}
    function send(address payable target) public payable {
        (bool success, bytes memory ret) = target.call.value(100000000)("");
        require(success);
    }
    
    fallback() external payable {
        recv[msg.sender] = msg.value;
    }
}

contract Nest1{
    fallback()external payable{
        msg.sender.send(msg.value);
    }
}
```