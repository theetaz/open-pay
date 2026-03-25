import { expect } from "chai";
import { ethers } from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-toolbox/network-helpers";

describe("OpenPayEscrow", function () {
  // ─── Shared fixture ───
  async function deployFixture() {
    const [owner, merchant, payer, feeRecipient] = await ethers.getSigners();

    // Mock tokens
    const MockUSDT = await ethers.getContractFactory("MockUSDT");
    const usdt = await MockUSDT.deploy();

    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await MockUSDC.deploy();

    // Mock Chainlink feeds (8 decimals)
    const MockPriceFeed = await ethers.getContractFactory("MockPriceFeed");
    const bnbFeed = await MockPriceFeed.deploy(600_00000000, 8);   // BNB = $600
    const usdtFeed = await MockPriceFeed.deploy(1_00000000, 8);    // USDT = $1
    const usdcFeed = await MockPriceFeed.deploy(1_00000000, 8);    // USDC = $1

    // Price Feed oracle
    const PriceFeed = await ethers.getContractFactory("OpenPayPriceFeed");
    const priceFeed = await PriceFeed.deploy();

    // Configure price feeds
    await priceFeed.setNativeFeed(await bnbFeed.getAddress());
    await priceFeed.configureToken(await usdt.getAddress(), await usdtFeed.getAddress(), true);
    await priceFeed.configureToken(await usdc.getAddress(), address0(), true); // stablecoin shortcut

    // Escrow with 2% fee
    const Escrow = await ethers.getContractFactory("OpenPayEscrow");
    const escrow = await Escrow.deploy(200, feeRecipient.address);

    // Wire price feed & enable tokens
    await escrow.setPriceFeed(await priceFeed.getAddress());
    await escrow.setSupportedToken(await usdt.getAddress(), true);
    await escrow.setSupportedToken(await usdc.getAddress(), true);

    // Mint tokens to payer
    await usdt.mint(payer.address, ethers.parseUnits("10000", 6));
    await usdc.mint(payer.address, ethers.parseUnits("10000", 6));

    return {
      escrow, usdt, usdc, priceFeed,
      bnbFeed, usdtFeed, usdcFeed,
      owner, merchant, payer, feeRecipient,
    };
  }

  function address0() {
    return ethers.ZeroAddress;
  }

  // ═══════════════════════════════════════════════════
  //  CORE: ERC-20 Payments
  // ═══════════════════════════════════════════════════

  describe("ERC20 Payments", function () {
    it("creates a payment (no price verification)", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("payment-001"));
      const amount = ethers.parseUnits("100", 6);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);

      await expect(
        escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt)
      ).to.emit(escrow, "PaymentCreated");

      const p = await escrow.getPayment(paymentId);
      expect(p.status).to.equal(1); // Pending
      expect(p.amount).to.equal(amount);
      expect(p.merchant).to.equal(merchant.address);
      expect(p.usdAmount).to.equal(0);
    });

    it("confirms a payment and distributes funds", async function () {
      const { escrow, usdt, merchant, payer, feeRecipient } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("payment-002"));
      const amount = ethers.parseUnits("1000", 6);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      const merchantBefore = await usdt.balanceOf(merchant.address);
      const feeBefore = await usdt.balanceOf(feeRecipient.address);

      await expect(escrow.confirmPayment(paymentId)).to.emit(escrow, "PaymentConfirmed");

      const merchantAfter = await usdt.balanceOf(merchant.address);
      const feeAfter = await usdt.balanceOf(feeRecipient.address);

      // 2% fee = 20 USDT, merchant gets 980
      expect(merchantAfter - merchantBefore).to.equal(ethers.parseUnits("980", 6));
      expect(feeAfter - feeBefore).to.equal(ethers.parseUnits("20", 6));

      const p = await escrow.getPayment(paymentId);
      expect(p.status).to.equal(2); // Confirmed
    });

    it("refunds a payment", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("payment-003"));
      const amount = ethers.parseUnits("500", 6);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      const payerBefore = await usdt.balanceOf(payer.address);
      await expect(escrow.refundPayment(paymentId)).to.emit(escrow, "PaymentRefunded");
      const payerAfter = await usdt.balanceOf(payer.address);
      expect(payerAfter - payerBefore).to.equal(amount);
    });

    it("prevents duplicate payment IDs", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("payment-dup"));
      const amount = ethers.parseUnits("100", 6);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount * 2n);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      await expect(
        escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt)
      ).to.be.revertedWith("Payment exists");
    });

    it("only owner can confirm", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("payment-auth"));
      const amount = ethers.parseUnits("100", 6);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      await expect(escrow.connect(payer).confirmPayment(paymentId)).to.be.reverted;
    });
  });

  // ═══════════════════════════════════════════════════
  //  CORE: Native Currency (BNB) Payments
  // ═══════════════════════════════════════════════════

  describe("Native Currency Payments", function () {
    it("creates and confirms a native payment", async function () {
      const { escrow, merchant, payer, feeRecipient } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("native-001"));
      const amount = ethers.parseEther("1.0");
      const expiresAt = (await time.latest()) + 3600;

      await escrow.connect(payer).createNativePayment(paymentId, merchant.address, 0, expiresAt, { value: amount });

      const p = await escrow.getPayment(paymentId);
      expect(p.status).to.equal(1);
      expect(p.amount).to.equal(amount);

      const merchantBefore = await ethers.provider.getBalance(merchant.address);
      await escrow.confirmPayment(paymentId);
      const merchantAfter = await ethers.provider.getBalance(merchant.address);

      // 2% fee = 0.02 ETH, merchant gets 0.98
      expect(merchantAfter - merchantBefore).to.equal(ethers.parseEther("0.98"));
    });
  });

  // ═══════════════════════════════════════════════════
  //  EXPIRATION
  // ═══════════════════════════════════════════════════

  describe("Expiration", function () {
    it("expires a payment after deadline", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("expire-001"));
      const amount = ethers.parseUnits("200", 6);
      const expiresAt = (await time.latest()) + 60;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      // Fast-forward past expiration
      await time.increase(120);

      const payerBefore = await usdt.balanceOf(payer.address);
      await expect(escrow.expirePayment(paymentId)).to.emit(escrow, "PaymentExpired");
      const payerAfter = await usdt.balanceOf(payer.address);
      expect(payerAfter - payerBefore).to.equal(amount);
    });

    it("cannot expire before deadline", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("expire-002"));
      const amount = ethers.parseUnits("100", 6);
      const expiresAt = (await time.latest()) + 86400;

      await usdt.connect(payer).approve(await escrow.getAddress(), amount);
      await escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), amount, 0, expiresAt);

      await expect(escrow.expirePayment(paymentId)).to.be.revertedWith("Not yet expired");
    });
  });

  // ═══════════════════════════════════════════════════
  //  ADMIN
  // ═══════════════════════════════════════════════════

  describe("Admin", function () {
    it("updates platform fee", async function () {
      const { escrow } = await loadFixture(deployFixture);
      await escrow.setPlatformFee(150);
      expect(await escrow.platformFeeBps()).to.equal(150);
    });

    it("rejects fee above 10%", async function () {
      const { escrow } = await loadFixture(deployFixture);
      await expect(escrow.setPlatformFee(1001)).to.be.revertedWith("Fee too high");
    });

    it("manages supported tokens", async function () {
      const { escrow, usdt } = await loadFixture(deployFixture);
      const addr = await usdt.getAddress();
      expect(await escrow.supportedTokens(addr)).to.be.true;

      await escrow.setSupportedToken(addr, false);
      expect(await escrow.supportedTokens(addr)).to.be.false;
    });

    it("sets slippage tolerance", async function () {
      const { escrow } = await loadFixture(deployFixture);
      await escrow.setSlippage(100); // 1%
      expect(await escrow.slippageBps()).to.equal(100);
    });

    it("rejects slippage above 5%", async function () {
      const { escrow } = await loadFixture(deployFixture);
      await expect(escrow.setSlippage(501)).to.be.revertedWith("Slippage too high");
    });
  });

  // ═══════════════════════════════════════════════════
  //  PRICE VERIFICATION & QUOTES
  // ═══════════════════════════════════════════════════

  describe("Price Verification", function () {
    it("getQuote returns correct BNB amount for $100", async function () {
      const { escrow } = await loadFixture(deployFixture);

      // BNB = $600 → $100 = 0.1666... BNB
      const usdAmount = 100_00000000n; // $100
      const [tokenAmount, exchangeRate, minAmount, maxAmount] = await escrow.getQuote(address0(), usdAmount);

      // Expected: 100e8 * 1e18 / 600e8 = 1e26 / 6e10 = 166666666666666666 (~0.1667 BNB)
      const expected = ethers.parseEther("100") / 600n;
      expect(tokenAmount).to.equal(expected);
      expect(exchangeRate).to.equal(600_00000000n);

      // Slippage 2%: min = expected * 98%, max = expected * 102%
      expect(minAmount).to.equal((expected * 9800n) / 10000n);
      expect(maxAmount).to.equal((expected * 10200n) / 10000n);
    });

    it("getQuote returns correct USDT amount (stablecoin via feed)", async function () {
      const { escrow, usdt } = await loadFixture(deployFixture);

      const usdAmount = 50_00000000n; // $50
      const [tokenAmount] = await escrow.getQuote(await usdt.getAddress(), usdAmount);

      // USDT = $1, 6 decimals → 50e8 * 1e6 / 1e8 = 50e6
      expect(tokenAmount).to.equal(ethers.parseUnits("50", 6));
    });

    it("getQuote returns correct USDC amount (stablecoin shortcut)", async function () {
      const { escrow, usdc } = await loadFixture(deployFixture);

      const usdAmount = 75_00000000n; // $75
      const [tokenAmount] = await escrow.getQuote(await usdc.getAddress(), usdAmount);

      expect(tokenAmount).to.equal(ethers.parseUnits("75", 6));
    });

    it("creates BNB payment with price verification", async function () {
      const { escrow, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("price-bnb-001"));

      const usdAmount = 100_00000000n; // $100
      const [tokenAmount] = await escrow.getQuote(address0(), usdAmount);
      const expiresAt = (await time.latest()) + 3600;

      await expect(
        escrow.connect(payer).createNativePayment(paymentId, merchant.address, usdAmount, expiresAt, { value: tokenAmount })
      ).to.emit(escrow, "PaymentCreated");

      const p = await escrow.getPayment(paymentId);
      expect(p.usdAmount).to.equal(usdAmount);
      expect(p.exchangeRate).to.equal(600_00000000n);
      expect(p.status).to.equal(1);
    });

    it("creates USDT payment with price verification", async function () {
      const { escrow, usdt, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("price-usdt-001"));

      const usdAmount = 200_00000000n; // $200
      const [tokenAmount] = await escrow.getQuote(await usdt.getAddress(), usdAmount);
      const expiresAt = (await time.latest()) + 3600;

      await usdt.connect(payer).approve(await escrow.getAddress(), tokenAmount);

      await expect(
        escrow.connect(payer).createPayment(paymentId, merchant.address, await usdt.getAddress(), tokenAmount, usdAmount, expiresAt)
      ).to.emit(escrow, "PaymentCreated");

      const p = await escrow.getPayment(paymentId);
      expect(p.usdAmount).to.equal(usdAmount);
      expect(p.amount).to.equal(ethers.parseUnits("200", 6));
    });

    it("rejects payment when price deviation exceeds slippage", async function () {
      const { escrow, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("price-reject-001"));

      const usdAmount = 100_00000000n; // $100
      const expiresAt = (await time.latest()) + 3600;

      // Send way too little BNB ($100 worth should be ~0.167 BNB, send 0.1)
      const tooLittle = ethers.parseEther("0.1");

      await expect(
        escrow.connect(payer).createNativePayment(paymentId, merchant.address, usdAmount, expiresAt, { value: tooLittle })
      ).to.be.revertedWith("Price deviation too high");
    });

    it("rejects payment when overpaying beyond slippage", async function () {
      const { escrow, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("price-reject-002"));

      const usdAmount = 100_00000000n; // $100
      const expiresAt = (await time.latest()) + 3600;

      // Send way too much BNB ($100 worth ≈ 0.167 BNB, send 0.5)
      const tooMuch = ethers.parseEther("0.5");

      await expect(
        escrow.connect(payer).createNativePayment(paymentId, merchant.address, usdAmount, expiresAt, { value: tooMuch })
      ).to.be.revertedWith("Price deviation too high");
    });

    it("accepts payment within slippage tolerance", async function () {
      const { escrow, merchant, payer } = await loadFixture(deployFixture);
      const paymentId = ethers.keccak256(ethers.toUtf8Bytes("price-slip-001"));

      const usdAmount = 100_00000000n;
      const [tokenAmount, , minAmount, maxAmount] = await escrow.getQuote(address0(), usdAmount);
      const expiresAt = (await time.latest()) + 3600;

      // Pay slightly above exact (within 2% tolerance)
      const slightlyOver = tokenAmount + (tokenAmount * 1n) / 100n; // +1%
      expect(slightlyOver).to.be.lte(maxAmount);

      await expect(
        escrow.connect(payer).createNativePayment(paymentId, merchant.address, usdAmount, expiresAt, { value: slightlyOver })
      ).to.emit(escrow, "PaymentCreated");
    });
  });

  // ═══════════════════════════════════════════════════
  //  PRICE FEED (Oracle)
  // ═══════════════════════════════════════════════════

  describe("PriceFeed Oracle", function () {
    it("rejects stale prices", async function () {
      const { priceFeed, bnbFeed } = await loadFixture(deployFixture);

      // Set updatedAt to 2 hours ago
      const twoHoursAgo = (await time.latest()) - 7200;
      await bnbFeed.setUpdatedAt(twoHoursAgo);

      await expect(
        priceFeed.getPrice(address0())
      ).to.be.revertedWith("Stale price");
    });

    it("updates price dynamically", async function () {
      const { escrow, bnbFeed } = await loadFixture(deployFixture);

      // BNB goes to $300 (halved)
      await bnbFeed.setPrice(300_00000000);

      const usdAmount = 100_00000000n; // $100
      const [tokenAmount] = await escrow.getQuote(address0(), usdAmount);

      // $100 at $300/BNB = 0.333... BNB
      const expected = ethers.parseEther("100") / 300n;
      expect(tokenAmount).to.equal(expected);
    });

    it("stablecoin shortcut returns $1 without oracle call", async function () {
      const { priceFeed, usdc } = await loadFixture(deployFixture);

      // USDC was configured with feed=address(0), isStablecoin=true
      const [price, decimals] = await priceFeed.getPrice(await usdc.getAddress());
      expect(price).to.equal(1_00000000n);
      expect(decimals).to.equal(8);
    });

    it("rejects unconfigured token", async function () {
      const { priceFeed } = await loadFixture(deployFixture);
      const randomToken = "0x0000000000000000000000000000000000000001";

      await expect(priceFeed.getPrice(randomToken)).to.be.revertedWith("Token not configured");
    });

    it("calculates USD value of token amount", async function () {
      const { priceFeed } = await loadFixture(deployFixture);

      // 0.5 BNB at $600 = $300
      const usd = await priceFeed.getAmountInUsd(address0(), ethers.parseEther("0.5"));
      expect(usd).to.equal(300_00000000n);
    });
  });
});
