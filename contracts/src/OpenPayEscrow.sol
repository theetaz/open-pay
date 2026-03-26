// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title IOpenPayPriceFeed
 */
interface IOpenPayPriceFeed {
    function getPrice(address token) external view returns (uint256 price, uint8 decimals_);
    function getAmountInToken(address token, uint256 usdAmount) external view returns (uint256);
    function getAmountInUsd(address token, uint256 tokenAmount) external view returns (uint256);
}

/**
 * @title OpenPayEscrow
 * @notice Payment escrow for Open Pay. Holds crypto until confirmed by the
 *         platform, then releases to the merchant minus a platform fee.
 *
 * Supports:
 *   - Multiple ERC-20 tokens (USDT, USDC, …) and native BNB/ETH.
 *   - Optional on-chain price verification via Chainlink oracles.
 *   - Slippage tolerance for volatile-asset payments.
 *   - Gas-free `getQuote()` for frontend / backend price previews.
 */
contract OpenPayEscrow is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    // ─── Types ───

    enum PaymentStatus { None, Pending, Confirmed, Refunded, Expired }

    struct Payment {
        bytes32 paymentId;       // Open Pay payment UUID (keccak256)
        address merchant;        // Merchant wallet
        address payer;           // Customer wallet
        address token;           // ERC-20 address (address(0) = native)
        uint256 amount;          // Crypto amount in token decimals
        uint256 platformFee;     // Fee portion of amount
        uint256 usdAmount;       // Original USD value (8 decimals, 0 = unset)
        uint256 exchangeRate;    // Token/USD rate at payment time (8 decimals)
        uint256 createdAt;
        uint256 expiresAt;
        PaymentStatus status;
    }

    // ─── State ───

    uint256 public platformFeeBps;           // basis points (100 = 1%)
    address public feeRecipient;
    IOpenPayPriceFeed public priceFeed;      // optional oracle
    uint256 public slippageBps = 200;        // 2% default

    mapping(address => bool) public supportedTokens;
    mapping(bytes32 => Payment) public payments;

    // ─── Events ───

    event PaymentCreated(
        bytes32 indexed paymentId, address indexed merchant,
        address indexed payer, address token,
        uint256 amount, uint256 usdAmount
    );
    event PaymentConfirmed(
        bytes32 indexed paymentId, address indexed merchant,
        uint256 merchantAmount, uint256 platformFee
    );
    event PaymentRefunded(bytes32 indexed paymentId, address indexed payer, uint256 amount);
    event PaymentExpired(bytes32 indexed paymentId);
    event TokenSupported(address indexed token, bool supported);
    event PlatformFeeUpdated(uint256 oldFee, uint256 newFee);
    event PriceFeedUpdated(address oldFeed, address newFeed);
    event SlippageUpdated(uint256 oldSlippage, uint256 newSlippage);

    // ─── Constructor ───

    constructor(uint256 _platformFeeBps, address _feeRecipient) Ownable(msg.sender) {
        require(_platformFeeBps <= 1000, "Fee too high"); // max 10 %
        require(_feeRecipient != address(0), "Invalid fee recipient");
        platformFeeBps = _platformFeeBps;
        feeRecipient = _feeRecipient;
    }

    // ─── Create Payment (ERC-20) ───

    /**
     * @notice Lock an ERC-20 payment in escrow.
     * @param paymentId  Unique hash derived from the Open Pay payment UUID.
     * @param merchant   Merchant wallet that receives funds on confirmation.
     * @param token      ERC-20 token address.
     * @param amount     Token amount (in token decimals) the payer transfers.
     * @param usdAmount  Intended USD value (8 decimals). Pass 0 to skip
     *                   on-chain price verification (e.g. stablecoin 1:1).
     * @param expiresAt  Unix timestamp after which the payment can be expired.
     */
    function createPayment(
        bytes32 paymentId,
        address merchant,
        address token,
        uint256 amount,
        uint256 usdAmount,
        uint256 expiresAt
    ) external nonReentrant {
        require(payments[paymentId].status == PaymentStatus.None, "Payment exists");
        require(merchant != address(0), "Invalid merchant");
        require(amount > 0, "Invalid amount");
        require(expiresAt > block.timestamp, "Already expired");
        require(supportedTokens[token], "Token not supported");

        uint256 exchangeRate = 0;

        // On-chain price verification (when usdAmount supplied & oracle wired)
        if (usdAmount > 0 && address(priceFeed) != address(0)) {
            exchangeRate = _verifyPrice(token, amount, usdAmount);
        }

        uint256 fee = (amount * platformFeeBps) / 10000;

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        payments[paymentId] = Payment({
            paymentId: paymentId,
            merchant: merchant,
            payer: msg.sender,
            token: token,
            amount: amount,
            platformFee: fee,
            usdAmount: usdAmount,
            exchangeRate: exchangeRate,
            createdAt: block.timestamp,
            expiresAt: expiresAt,
            status: PaymentStatus.Pending
        });

        emit PaymentCreated(paymentId, merchant, msg.sender, token, amount, usdAmount);
    }

    // ─── Create Payment (Native BNB / ETH) ───

    /**
     * @param usdAmount  Intended USD value (8 decimals). Pass 0 to skip
     *                   on-chain price verification.
     */
    function createNativePayment(
        bytes32 paymentId,
        address merchant,
        uint256 usdAmount,
        uint256 expiresAt
    ) external payable nonReentrant {
        require(payments[paymentId].status == PaymentStatus.None, "Payment exists");
        require(merchant != address(0), "Invalid merchant");
        require(msg.value > 0, "Invalid amount");
        require(expiresAt > block.timestamp, "Already expired");

        uint256 exchangeRate = 0;

        if (usdAmount > 0 && address(priceFeed) != address(0)) {
            exchangeRate = _verifyPrice(address(0), msg.value, usdAmount);
        }

        uint256 fee = (msg.value * platformFeeBps) / 10000;

        payments[paymentId] = Payment({
            paymentId: paymentId,
            merchant: merchant,
            payer: msg.sender,
            token: address(0),
            amount: msg.value,
            platformFee: fee,
            usdAmount: usdAmount,
            exchangeRate: exchangeRate,
            createdAt: block.timestamp,
            expiresAt: expiresAt,
            status: PaymentStatus.Pending
        });

        emit PaymentCreated(paymentId, merchant, msg.sender, address(0), msg.value, usdAmount);
    }

    // ─── Confirm ───

    function confirmPayment(bytes32 paymentId) external onlyOwner nonReentrant {
        Payment storage p = payments[paymentId];
        require(p.status == PaymentStatus.Pending, "Not pending");

        p.status = PaymentStatus.Confirmed;
        uint256 merchantAmount = p.amount - p.platformFee;

        if (p.token == address(0)) {
            (bool sent1, ) = payable(p.merchant).call{value: merchantAmount}("");
            require(sent1, "Merchant transfer failed");
            if (p.platformFee > 0) {
                (bool sent2, ) = payable(feeRecipient).call{value: p.platformFee}("");
                require(sent2, "Fee transfer failed");
            }
        } else {
            IERC20(p.token).safeTransfer(p.merchant, merchantAmount);
            if (p.platformFee > 0) {
                IERC20(p.token).safeTransfer(feeRecipient, p.platformFee);
            }
        }

        emit PaymentConfirmed(paymentId, p.merchant, merchantAmount, p.platformFee);
    }

    // ─── Refund ───

    function refundPayment(bytes32 paymentId) external onlyOwner nonReentrant {
        Payment storage p = payments[paymentId];
        require(p.status == PaymentStatus.Pending, "Not pending");

        p.status = PaymentStatus.Refunded;

        if (p.token == address(0)) {
            (bool sent, ) = payable(p.payer).call{value: p.amount}("");
            require(sent, "Refund failed");
        } else {
            IERC20(p.token).safeTransfer(p.payer, p.amount);
        }

        emit PaymentRefunded(paymentId, p.payer, p.amount);
    }

    // ─── Expire ───

    function expirePayment(bytes32 paymentId) external nonReentrant {
        Payment storage p = payments[paymentId];
        require(p.status == PaymentStatus.Pending, "Not pending");
        require(block.timestamp >= p.expiresAt, "Not yet expired");

        p.status = PaymentStatus.Expired;

        if (p.token == address(0)) {
            (bool sent, ) = payable(p.payer).call{value: p.amount}("");
            require(sent, "Refund failed");
        } else {
            IERC20(p.token).safeTransfer(p.payer, p.amount);
        }

        emit PaymentExpired(paymentId);
    }

    // ─── Views ───

    function getPayment(bytes32 paymentId) external view returns (Payment memory) {
        return payments[paymentId];
    }

    /**
     * @notice Gas-free price quote for the frontend / backend.
     * @param token     Token to pay with (address(0) for native).
     * @param usdAmount USD value in 8 decimals ($100 = 100_00000000).
     * @return tokenAmount  Exact amount the oracle recommends.
     * @return exchangeRate Current token/USD rate (8 decimals).
     * @return minAmount    Minimum acceptable (after slippage).
     * @return maxAmount    Maximum acceptable (after slippage).
     */
    function getQuote(
        address token,
        uint256 usdAmount
    )
        external
        view
        returns (
            uint256 tokenAmount,
            uint256 exchangeRate,
            uint256 minAmount,
            uint256 maxAmount
        )
    {
        require(address(priceFeed) != address(0), "Price feed not set");
        require(usdAmount > 0, "Invalid USD amount");

        (exchangeRate, ) = priceFeed.getPrice(token);
        tokenAmount = priceFeed.getAmountInToken(token, usdAmount);
        minAmount = (tokenAmount * (10000 - slippageBps)) / 10000;
        maxAmount = (tokenAmount * (10000 + slippageBps)) / 10000;
    }

    // ─── Admin ───

    function setSupportedToken(address token, bool supported) external onlyOwner {
        supportedTokens[token] = supported;
        emit TokenSupported(token, supported);
    }

    function setPlatformFee(uint256 newFeeBps) external onlyOwner {
        require(newFeeBps <= 1000, "Fee too high");
        emit PlatformFeeUpdated(platformFeeBps, newFeeBps);
        platformFeeBps = newFeeBps;
    }

    function setFeeRecipient(address newRecipient) external onlyOwner {
        require(newRecipient != address(0), "Invalid recipient");
        feeRecipient = newRecipient;
    }

    function setPriceFeed(address _priceFeed) external onlyOwner {
        emit PriceFeedUpdated(address(priceFeed), _priceFeed);
        priceFeed = IOpenPayPriceFeed(_priceFeed);
    }

    function setSlippage(uint256 newSlippageBps) external onlyOwner {
        require(newSlippageBps <= 500, "Slippage too high"); // max 5 %
        emit SlippageUpdated(slippageBps, newSlippageBps);
        slippageBps = newSlippageBps;
    }

    // ─── Internal ───

    /**
     * @dev Verify that `amount` of `token` is within slippage tolerance of
     *      the oracle-derived value for `usdAmount`.
     * @return exchangeRate  The current token/USD price from the oracle.
     */
    function _verifyPrice(
        address token,
        uint256 amount,
        uint256 usdAmount
    ) internal view returns (uint256 exchangeRate) {
        (exchangeRate, ) = priceFeed.getPrice(token);
        uint256 expectedAmount = priceFeed.getAmountInToken(token, usdAmount);

        uint256 minAmount = (expectedAmount * (10000 - slippageBps)) / 10000;
        uint256 maxAmount = (expectedAmount * (10000 + slippageBps)) / 10000;

        require(
            amount >= minAmount && amount <= maxAmount,
            "Price deviation too high"
        );
    }
}
