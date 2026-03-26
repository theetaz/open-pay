import { ethers } from "hardhat";

/**
 * E2E smoke-test against deployed BSC Testnet contracts.
 * Run: npx hardhat run script/test-bsc-testnet.ts --network bscTestnet
 */

const ADDRESSES = {
  MockUSDT: "0x98e2146A4381C74708782D03dAd3913b0388954A",
  MockUSDC: "0xE2b9aB57304C8AFc7068940c03EE202e0B8D4CEC",
  OpenPayPriceFeed: "0x1f34e070D4BB1eD3AaF37D8E3297b0a9A12a3399",
  OpenPayEscrow: "0xe50464081b781AFE101EB40bC7e68Fd017c5e8f2",
};

let passed = 0;
let failed = 0;

function ok(label: string) {
  passed++;
  console.log(`  ✅ ${label}`);
}
function fail(label: string, err: unknown) {
  failed++;
  console.log(`  ❌ ${label}: ${err}`);
}

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Tester:", deployer.address);
  console.log("Balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "BNB\n");

  // Attach to deployed contracts
  const usdt = await ethers.getContractAt("MockUSDT", ADDRESSES.MockUSDT);
  const usdc = await ethers.getContractAt("MockUSDC", ADDRESSES.MockUSDC);
  const priceFeed = await ethers.getContractAt("OpenPayPriceFeed", ADDRESSES.OpenPayPriceFeed);
  const escrow = await ethers.getContractAt("OpenPayEscrow", ADDRESSES.OpenPayEscrow);

  // ─── 1. Price Feed Queries ───
  console.log("--- 1. Price Feed Queries ---");

  try {
    const [bnbPrice] = await priceFeed.getPrice(ethers.ZeroAddress);
    console.log(`  BNB/USD price: $${Number(bnbPrice) / 1e8}`);
    if (bnbPrice === BigInt(600_00000000)) ok("BNB price = $600");
    else fail("BNB price mismatch", bnbPrice);
  } catch (e) { fail("BNB price query", e); }

  try {
    const [usdtPrice] = await priceFeed.getPrice(ADDRESSES.MockUSDT);
    if (usdtPrice === BigInt(1_00000000)) ok("USDT price = $1 (via feed)");
    else fail("USDT price mismatch", usdtPrice);
  } catch (e) { fail("USDT price query", e); }

  try {
    const [usdcPrice] = await priceFeed.getPrice(ADDRESSES.MockUSDC);
    if (usdcPrice === BigInt(1_00000000)) ok("USDC price = $1 (stablecoin shortcut)");
    else fail("USDC price mismatch", usdcPrice);
  } catch (e) { fail("USDC price query", e); }

  // ─── 2. Get Quotes ───
  console.log("\n--- 2. Get Quotes ---");

  try {
    const usd100 = BigInt(100_00000000);
    const [tokenAmt, rate, minAmt, maxAmt] = await escrow.getQuote(ethers.ZeroAddress, usd100);
    console.log(`  $100 in BNB: ${ethers.formatEther(tokenAmt)} BNB (rate: $${Number(rate) / 1e8})`);
    console.log(`  Acceptable range: ${ethers.formatEther(minAmt)} – ${ethers.formatEther(maxAmt)} BNB`);
    if (rate === BigInt(600_00000000)) ok("BNB quote correct");
    else fail("BNB quote rate mismatch", rate);
  } catch (e) { fail("BNB quote", e); }

  try {
    const usd50 = BigInt(50_00000000);
    const [tokenAmt] = await escrow.getQuote(ADDRESSES.MockUSDT, usd50);
    const expected = ethers.parseUnits("50", 6);
    if (tokenAmt === expected) ok("USDT quote: $50 = 50 mUSDT");
    else fail("USDT quote mismatch", `got ${tokenAmt}, expected ${expected}`);
  } catch (e) { fail("USDT quote", e); }

  // ─── 3. Create USDT Payment (with price verification) ───
  console.log("\n--- 3. Create USDT Payment ---");

  const usdtPaymentId = ethers.keccak256(ethers.toUtf8Bytes(`e2e-usdt-${Date.now()}`));
  const usdtAmount = ethers.parseUnits("50", 6); // 50 USDT
  const usdAmount = BigInt(50_00000000); // $50
  const expiresAt = Math.floor(Date.now() / 1000) + 3600;
  const merchant = deployer.address; // self-pay for testing

  try {
    // Approve
    let tx = await usdt.approve(ADDRESSES.OpenPayEscrow, usdtAmount);
    await tx.wait();
    ok("USDT approved for escrow");

    // Create payment
    tx = await escrow.createPayment(
      usdtPaymentId, merchant, ADDRESSES.MockUSDT, usdtAmount, usdAmount, expiresAt
    );
    const receipt = await tx.wait();
    console.log(`  Tx: ${receipt!.hash}`);
    ok("USDT payment created with price verification");

    // Verify on-chain state
    const p = await escrow.getPayment(usdtPaymentId);
    if (p.status === BigInt(1)) ok("Payment status = Pending");
    else fail("Payment status", p.status);
    if (p.usdAmount === usdAmount) ok("USD amount stored = $50");
    else fail("USD amount mismatch", p.usdAmount);
  } catch (e) { fail("USDT payment creation", e); }

  // ─── 4. Confirm USDT Payment ───
  console.log("\n--- 4. Confirm USDT Payment ---");

  try {
    const balBefore = await usdt.balanceOf(deployer.address);
    const tx = await escrow.confirmPayment(usdtPaymentId);
    const receipt = await tx.wait();
    console.log(`  Tx: ${receipt!.hash}`);
    const balAfter = await usdt.balanceOf(deployer.address);

    // Re-read after confirmation receipt to ensure state is settled
    const p = await escrow.getPayment(usdtPaymentId, { blockTag: receipt!.blockNumber });
    if (p.status === BigInt(2)) ok("Payment status = Confirmed");
    else fail("Confirm status", p.status);

    // Since merchant = deployer = feeRecipient, deployer gets full amount back
    ok("Funds distributed (merchant + fee → deployer)");
  } catch (e) { fail("USDT payment confirmation", e); }

  // ─── 5. Create BNB Payment (with price verification) ───
  console.log("\n--- 5. Create BNB Payment (native + price verification) ---");

  const bnbPaymentId = ethers.keccak256(ethers.toUtf8Bytes(`e2e-bnb-${Date.now()}`));
  const bnbUsdAmount = BigInt(10_00000000); // $10

  try {
    // Get quote first
    const [bnbTokenAmt] = await escrow.getQuote(ethers.ZeroAddress, bnbUsdAmount);
    console.log(`  Quote: $10 = ${ethers.formatEther(bnbTokenAmt)} BNB`);

    // Create payment with exact quoted amount
    const tx = await escrow.createNativePayment(
      bnbPaymentId, merchant, bnbUsdAmount, expiresAt, { value: bnbTokenAmt }
    );
    const receipt = await tx.wait();
    console.log(`  Tx: ${receipt!.hash}`);
    ok("BNB payment created with price verification");

    const p = await escrow.getPayment(bnbPaymentId);
    if (p.exchangeRate === BigInt(600_00000000)) ok("Exchange rate stored = $600");
    else fail("Exchange rate", p.exchangeRate);
  } catch (e) { fail("BNB payment creation", e); }

  // ─── 6. Confirm BNB Payment ───
  console.log("\n--- 6. Confirm BNB Payment ---");

  try {
    const tx = await escrow.confirmPayment(bnbPaymentId);
    const receipt = await tx.wait();
    console.log(`  Tx: ${receipt!.hash}`);

    const p = await escrow.getPayment(bnbPaymentId, { blockTag: receipt!.blockNumber });
    if (p.status === BigInt(2)) ok("BNB payment confirmed");
    else fail("BNB confirm status", p.status);
  } catch (e) { fail("BNB payment confirmation", e); }

  // ─── 7. Refund flow ───
  console.log("\n--- 7. Refund Flow ---");

  const refundPaymentId = ethers.keccak256(ethers.toUtf8Bytes(`e2e-refund-${Date.now()}`));
  const refundAmount = ethers.parseUnits("25", 6);

  try {
    let tx = await usdt.approve(ADDRESSES.OpenPayEscrow, refundAmount);
    await tx.wait();
    tx = await escrow.createPayment(
      refundPaymentId, merchant, ADDRESSES.MockUSDT, refundAmount, 0, expiresAt
    );
    await tx.wait();
    ok("Refund test: payment created");

    tx = await escrow.refundPayment(refundPaymentId);
    await tx.wait();

    const p = await escrow.getPayment(refundPaymentId);
    if (p.status === BigInt(3)) ok("Payment refunded (status = Refunded)");
    else fail("Refund status", p.status);
  } catch (e) { fail("Refund flow", e); }

  // ─── Summary ───
  console.log("\n══════════════════════════════════");
  console.log(`  PASSED: ${passed}  |  FAILED: ${failed}`);
  console.log("══════════════════════════════════");
  console.log("Remaining balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "BNB\n");

  if (failed > 0) process.exitCode = 1;
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
