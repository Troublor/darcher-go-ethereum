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
