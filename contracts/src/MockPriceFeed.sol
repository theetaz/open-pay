// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/**
 * @title MockPriceFeed
 * @notice Simulates a Chainlink AggregatorV3 for local / testnet testing.
 *         Owner can set price and updatedAt to exercise staleness logic.
 */
contract MockPriceFeed {
    int256 private _price;
    uint8 private _decimals;
    uint256 private _updatedAt;

    constructor(int256 price_, uint8 decimals_) {
        _price = price_;
        _decimals = decimals_;
        _updatedAt = block.timestamp;
    }

    function latestRoundData()
        external
        view
        returns (
            uint80 roundId,
            int256 answer,
            uint256 startedAt,
            uint256 updatedAt,
            uint80 answeredInRound
        )
    {
        return (1, _price, _updatedAt, _updatedAt, 1);
    }

    function decimals() external view returns (uint8) {
        return _decimals;
    }

    /// @notice Change the reported price (for test scenarios).
    function setPrice(int256 price_) external {
        _price = price_;
        _updatedAt = block.timestamp;
    }

    /// @notice Manually set updatedAt to simulate stale data.
    function setUpdatedAt(uint256 updatedAt_) external {
        _updatedAt = updatedAt_;
    }
}
