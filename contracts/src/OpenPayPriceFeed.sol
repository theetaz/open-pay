// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import "./interfaces/AggregatorV3Interface.sol";

/**
 * @title OpenPayPriceFeed
 * @notice Wraps Chainlink oracles to provide USD-denominated pricing for any
 *         supported token. Stablecoins can be configured to return $1 without
 *         an oracle. Native currency (BNB/ETH) is represented as address(0).
 */
contract OpenPayPriceFeed is Ownable {
    struct TokenConfig {
        address feed;          // Chainlink aggregator (address(0) for stablecoin 1:1)
        bool isStablecoin;     // true → return $1.00 when feed is zero
        bool isConfigured;     // guard against unconfigured tokens
    }

    mapping(address => TokenConfig) public tokenConfigs;

    /// @notice Chainlink feed for the native currency (BNB on BSC, ETH on Ethereum).
    address public nativeFeed;

    /// @notice Maximum age (seconds) of an oracle answer before it is rejected.
    uint256 public stalenessThreshold = 3600; // 1 hour

    // ─── Events ───
    event TokenConfigured(address indexed token, address feed, bool isStablecoin);
    event NativeFeedUpdated(address oldFeed, address newFeed);
    event StalenessUpdated(uint256 oldThreshold, uint256 newThreshold);

    constructor() Ownable(msg.sender) {}

    // ─── Admin ───

    /**
     * @notice Register (or update) the price-feed configuration for an ERC-20.
     * @param token  ERC-20 address
     * @param feed   Chainlink aggregator; pass address(0) for a stablecoin that
     *               should always return $1.00
     * @param isStablecoin  When true AND feed == address(0), price = $1.00
     */
    function configureToken(
        address token,
        address feed,
        bool isStablecoin
    ) external onlyOwner {
        tokenConfigs[token] = TokenConfig({
            feed: feed,
            isStablecoin: isStablecoin,
            isConfigured: true
        });
        emit TokenConfigured(token, feed, isStablecoin);
    }

    function setNativeFeed(address feed) external onlyOwner {
        emit NativeFeedUpdated(nativeFeed, feed);
        nativeFeed = feed;
    }

    function setStalenessThreshold(uint256 threshold) external onlyOwner {
        require(threshold >= 60, "Threshold too low");
        emit StalenessUpdated(stalenessThreshold, threshold);
        stalenessThreshold = threshold;
    }

    // ─── Price Queries ───

    /**
     * @notice USD price of `token` with 8 decimals (Chainlink standard).
     * @param token ERC-20 address, or address(0) for native currency.
     * @return price  USD price scaled to 8 decimals.
     * @return decimals_  Always 8 for Chainlink feeds / stablecoin fallback.
     */
    function getPrice(address token) public view returns (uint256 price, uint8 decimals_) {
        if (token == address(0)) {
            require(nativeFeed != address(0), "Native feed not set");
            return _readFeed(nativeFeed);
        }

        TokenConfig memory cfg = tokenConfigs[token];
        require(cfg.isConfigured, "Token not configured");

        // Stablecoin shortcut: $1.00 without an oracle call.
        if (cfg.isStablecoin && cfg.feed == address(0)) {
            return (1_00000000, 8);
        }

        require(cfg.feed != address(0), "Feed not set");
        return _readFeed(cfg.feed);
    }

    /**
     * @notice How many tokens does `usdAmount` buy at the current rate?
     * @param token      ERC-20 address (or address(0) for native).
     * @param usdAmount  USD value in 8 decimals ($100 = 100_00000000).
     * @return tokenAmount  Amount in the token's native decimals.
     */
    function getAmountInToken(
        address token,
        uint256 usdAmount
    ) external view returns (uint256) {
        (uint256 price, ) = getPrice(token);
        uint8 tokenDecimals = _tokenDecimals(token);
        // tokenAmount = usdAmount × 10^tokenDecimals / price
        return (usdAmount * (10 ** tokenDecimals)) / price;
    }

    /**
     * @notice USD value of `tokenAmount` at the current rate.
     * @param token        ERC-20 address (or address(0) for native).
     * @param tokenAmount  Amount in the token's native decimals.
     * @return usdAmount   USD value in 8 decimals.
     */
    function getAmountInUsd(
        address token,
        uint256 tokenAmount
    ) external view returns (uint256) {
        (uint256 price, ) = getPrice(token);
        uint8 tokenDecimals = _tokenDecimals(token);
        // usdAmount = tokenAmount × price / 10^tokenDecimals
        return (tokenAmount * price) / (10 ** tokenDecimals);
    }

    // ─── Internal ───

    function _readFeed(address feed) internal view returns (uint256, uint8) {
        AggregatorV3Interface agg = AggregatorV3Interface(feed);
        (, int256 answer, , uint256 updatedAt, ) = agg.latestRoundData();

        require(answer > 0, "Invalid price");
        require(
            block.timestamp - updatedAt <= stalenessThreshold,
            "Stale price"
        );

        return (uint256(answer), agg.decimals());
    }

    function _tokenDecimals(address token) internal view returns (uint8) {
        if (token == address(0)) return 18; // BNB / ETH
        return IERC20Metadata(token).decimals();
    }
}
