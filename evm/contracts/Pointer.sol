pragma solidity 0.4.26;

contract Pointer {
  address public getAddress;

  constructor(address _addr) public {
    getAddress = _addr;
  }
}
