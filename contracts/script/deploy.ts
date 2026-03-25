import { ethers } from "hardhat";

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deploying with:", deployer.address);
  console.log("Balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "BNB");

  // ─── 1. Mock Tokens ───
  console.log("\n--- Deploying Mock Tokens ---");

  const MockUSDT = await ethers.getContractFactory("MockUSDT");
  const usdt = await MockUSDT.deploy();
  await usdt.waitForDeployment();
  const usdtAddr = await usdt.getAddress();
  console.log("MockUSDT deployed:", usdtAddr);

  const MockUSDC = await ethers.getContractFactory("MockUSDC");
  const usdc = await MockUSDC.deploy();
  await usdc.waitForDeployment();
  const usdcAddr = await usdc.getAddress();
  console.log("MockUSDC deployed:", usdcAddr);

  // ─── 2. Mock Price Feeds ───
  console.log("\n--- Deploying Mock Price Feeds ---");

  const MockPriceFeed = await ethers.getContractFactory("MockPriceFeed");

  const bnbFeed = await MockPriceFeed.deploy(600_00000000, 8); // BNB = $600
  await bnbFeed.waitForDeployment();
  console.log("BNB/USD Feed deployed:", await bnbFeed.getAddress(), "(BNB = $600)");

  const usdtFeed = await MockPriceFeed.deploy(1_00000000, 8); // USDT = $1
  await usdtFeed.waitForDeployment();
  console.log("USDT/USD Feed deployed:", await usdtFeed.getAddress(), "(USDT = $1)");

  // USDC uses stablecoin shortcut (no feed needed)

  // ─── 3. Price Feed Oracle ───
  console.log("\n--- Deploying OpenPayPriceFeed ---");

  const PriceFeed = await ethers.getContractFactory("OpenPayPriceFeed");
  const priceFeed = await PriceFeed.deploy();
  await priceFeed.waitForDeployment();
  const priceFeedAddr = await priceFeed.getAddress();
  console.log("OpenPayPriceFeed deployed:", priceFeedAddr);

  // Configure feeds
  let tx = await priceFeed.setNativeFeed(await bnbFeed.getAddress());
  await tx.wait();
  console.log("  → BNB/USD feed configured");

  tx = await priceFeed.configureToken(usdtAddr, await usdtFeed.getAddress(), true);
  await tx.wait();
  console.log("  → USDT configured (with feed)");

  tx = await priceFeed.configureToken(usdcAddr, ethers.ZeroAddress, true);
  await tx.wait();
  console.log("  → USDC configured (stablecoin shortcut, $1.00)");

  // ─── 4. Escrow ───
  console.log("\n--- Deploying OpenPayEscrow ---");

  const Escrow = await ethers.getContractFactory("OpenPayEscrow");
  const escrow = await Escrow.deploy(200, deployer.address); // 2% fee
  await escrow.waitForDeployment();
  const escrowAddr = await escrow.getAddress();
  console.log("OpenPayEscrow deployed:", escrowAddr);

  // Wire price feed
  tx = await escrow.setPriceFeed(priceFeedAddr);
  await tx.wait();
  console.log("  → Price feed wired");

  // Enable tokens
  tx = await escrow.setSupportedToken(usdtAddr, true);
  await tx.wait();
  console.log("  → MockUSDT enabled");

  tx = await escrow.setSupportedToken(usdcAddr, true);
  await tx.wait();
  console.log("  → MockUSDC enabled");

  // ─── 5. Mint Test Tokens ───
  console.log("\n--- Minting Test Tokens ---");

  tx = await usdt.faucet(ethers.parseUnits("100000", 6));
  await tx.wait();
  console.log("  → 100,000 mUSDT minted to deployer");

  tx = await usdc.faucet(ethers.parseUnits("100000", 6));
  await tx.wait();
  console.log("  → 100,000 mUSDC minted to deployer");

  // ─── Summary ───
  const network = await ethers.provider.getNetwork();
  console.log("\n╔══════════════════════════════════════════════════════════╗");
  console.log("║              DEPLOYMENT SUMMARY                        ║");
  console.log("╠══════════════════════════════════════════════════════════╣");
  console.log(`║  Network:          ${network.name} (chain ${network.chainId})`);
  console.log(`║  Deployer:         ${deployer.address}`);
  console.log("║                                                        ║");
  console.log("║  CONTRACTS:                                            ║");
  console.log(`║  MockUSDT:         ${usdtAddr}`);
  console.log(`║  MockUSDC:         ${usdcAddr}`);
  console.log(`║  BNB/USD Feed:     ${await bnbFeed.getAddress()}`);
  console.log(`║  USDT/USD Feed:    ${await usdtFeed.getAddress()}`);
  console.log(`║  OpenPayPriceFeed: ${priceFeedAddr}`);
  console.log(`║  OpenPayEscrow:    ${escrowAddr}`);
  console.log("║                                                        ║");
  console.log("║  CONFIG:                                               ║");
  console.log("║  Platform Fee:     2% (200 bps)                        ║");
  console.log("║  Slippage:         2% (200 bps)                        ║");
  console.log("║  Staleness:        3600s (1 hour)                      ║");
  console.log("║  Supported:        mUSDT, mUSDC, BNB (native)         ║");
  console.log("╚══════════════════════════════════════════════════════════╝");

  console.log("\nRemaining balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "BNB");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
