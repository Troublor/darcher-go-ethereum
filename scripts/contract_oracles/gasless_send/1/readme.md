This folder contains four test cases (four transactions). 

- The first one call a contract function to send nonZero value to another contract
with cheap fallback function.
    - expect not to have gasless send vulnerability
    
- The second one call a contract function to send zero value to another contract
with cheap fallback function.
    - expect not to have gasless send vulnerability

- The third one call a contract function to send nonZero value to another contract
with expensive fallback function.
    - expect to **have** gasless send vulnerability

- The fourth one call a contract function to send zero value to another contract
with expensive fallback function.
    - expect not to have gasless send vulnerability (since no value is send)

```solidity
contract GaslessTest {
    constructor() public payable{
        
    }
    function sendZero(address payable target) payable public {
        target.send(0);
    }
    function sendNonZero(address payable target) payable public {
        target.transfer(10000000000);
    }
}

contract Cheap {
    event Hello(uint a);
     fallback () external payable{
        emit Hello(msg.value + uint256(10000000000));
    }
}

contract Expensive{
    mapping(address=>uint256) recv;
    
    fallback() external payable {
        recv[msg.sender] = msg.value;
    }
}
```