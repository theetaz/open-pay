// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockUSDT
 * @notice Mock USDT token for testnet deployments. Anyone can mint.
 */
contract MockUSDT is ERC20 {
    uint8 private _decimals;

    constructor() ERC20("Mock USDT", "mUSDT") {
        _decimals = 6; // USDT uses 6 decimals
    }

    function decimals() public view override returns (uint8) {
        return _decimals;
    }

    /**
     * @notice Mint tokens to any address. Testnet only.
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    /**
     * @notice Convenience: mint tokens to msg.sender.
     */
    function faucet(uint256 amount) external {
        _mint(msg.sender, amount);
    }
}
