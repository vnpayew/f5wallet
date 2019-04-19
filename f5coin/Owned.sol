/*
    @author: DuHD
    @version: 1.0
    @date: 09/04/2019
*/

pragma solidity ^0.5.7;

contract Owned {
    address owner;
    mapping(address => bool)  internal MemberApi;
    address[] internal memberApiIdx;

    constructor() public{
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, 'CHI OWNER CONTRACT MOI DUOC GOI HAM');
        _;
    }

    function getOwner() view public returns (address) {
        return owner;
    }

    function changeOwner(address _newOwner) onlyMember public {
        owner = _newOwner;
    }

    function registerMemberApi(address _newMember) onlyMember public {
        MemberApi[_newMember] = true;
        memberApiIdx.push(_newMember);
    }

    function getMemberApiIdxLenght() view public returns (int16)
    {
        return int16(memberApiIdx.length);
    }

    modifier onlyMember() {
        require(MemberApi[msg.sender], 'CHI CAC ACC ETH DA DANG KY MOI GOI DUOC HAM');
        _;
    }
}
