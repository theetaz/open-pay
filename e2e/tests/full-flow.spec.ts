import { test, expect } from '@playwright/test'

const MERCHANT_URL = 'http://localhost:3000'
const ADMIN_URL = 'http://localhost:3001'
const API_URL = 'http://localhost:8080'

const delay = (ms: number) => new Promise((r) => setTimeout(r, ms))

// Unique email for this test run
const MERCHANT_EMAIL = `e2e-${Date.now()}@curryhouse.lk`
const MERCHANT_PASSWORD = 'TestPass1'
const MERCHANT_NAME = 'Kamal Perera'
const BUSINESS_NAME = 'Curry House Colombo'

test.describe.serial('Open Lanka Payment — Full E2E Flow', () => {
  // ==========================================
  // ACT 1: MERCHANT REGISTRATION
  // ==========================================
  test('1. Merchant Portal — Register new merchant', async ({ page }) => {
    await page.goto(`${MERCHANT_URL}/register`)
    await delay(1500)

    // Fill registration form
    await page.fill('#fullName', MERCHANT_NAME)
    await delay(500)
    await page.fill('#businessName', BUSINESS_NAME)
    await delay(500)
    await page.fill('#email', MERCHANT_EMAIL)
    await delay(500)
    await page.fill('#password', MERCHANT_PASSWORD)
    await delay(500)
    await page.fill('#confirmPassword', MERCHANT_PASSWORD)
    await delay(500)

    // Check terms
    await page.locator('#terms').click()
    await delay(500)

    // Submit
    await page.click('button[type="submit"]')
    await delay(3000)

    // Should redirect to activate (KYC) page
    await expect(page).toHaveURL(/activate/, { timeout: 10000 })
    await delay(2000)
  })

  test('2. Merchant Portal — Fill KYC Wizard Step 1 (Personal Details)', async ({ page }) => {
    // Login first
    await page.goto(`${MERCHANT_URL}/login`)
    await delay(1000)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    // Navigate to activate
    await page.goto(`${MERCHANT_URL}/activate`)
    await delay(2000)

    // Fill Step 1 - Personal Details
    await page.fill('input[name="firstName"]', 'Kamal')
    await delay(400)
    await page.fill('input[name="lastName"]', 'Perera')
    await delay(400)
    await page.fill('input[name="addressLine1"]', '123 Galle Road')
    await delay(400)
    await page.fill('input[name="city"]', 'Colombo')
    await delay(400)
    await page.fill('input[name="postalCode"]', '00300')
    await delay(400)
    await page.fill('input[name="idNumber"]', '199012345678')
    await delay(400)
    await page.fill('input[name="dateOfBirth"]', '1990-05-15')
    await delay(400)
    await page.fill('input[name="mobileNo"]', '+94771234567')
    await delay(1000)

    // Click Next
    await page.click('text=Next')
    await delay(2000)

    // Should be on Step 2 now
    await expect(page.locator('text=Business Details')).toBeVisible()
  })

  test('3. Merchant Portal — Login and view Dashboard', async ({ page }) => {
    await page.goto(`${MERCHANT_URL}/login`)
    await delay(1500)

    await page.fill('#email', MERCHANT_EMAIL)
    await delay(500)
    await page.fill('#password', MERCHANT_PASSWORD)
    await delay(500)
    await page.click('button[type="submit"]')
    await delay(3000)

    // Should be on dashboard
    await expect(page).toHaveURL(/\/$/, { timeout: 10000 })
    await delay(2000)

    // Check dashboard elements are visible
    await expect(page.locator('text=Total Revenue')).toBeVisible()
    await expect(page.locator('text=Quick Actions')).toBeVisible()
    await delay(2000)
  })

  test('4. Merchant Portal — Create a Payment', async ({ page }) => {
    // Login
    await page.goto(`${MERCHANT_URL}/login`)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    // Click New Payment button
    await page.click('text=New Payment')
    await delay(1500)

    // Fill payment form in dialog
    await page.fill('input[type="number"]', '25')
    await delay(500)

    // Click Create Payment
    await page.click('text=Create Payment')
    await delay(3000)

    // Should see success with payment number
    await expect(page.locator('text=Payment Created')).toBeVisible({ timeout: 10000 })
    await delay(2000)

    // Copy checkout link
    await page.click('text=Copy Checkout Link')
    await delay(1500)

    // Close dialog
    await page.click('text=Done')
    await delay(1000)
  })

  test('5. Merchant Portal — View Payments page', async ({ page }) => {
    // Login
    await page.goto(`${MERCHANT_URL}/login`)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    // Navigate to payments
    await page.goto(`${MERCHANT_URL}/payments`)
    await delay(2000)

    // Page should load
    await expect(page.locator('text=Payments')).toBeVisible()
    await delay(2000)
  })

  test('6. Merchant Portal — View Withdrawal page', async ({ page }) => {
    // Login
    await page.goto(`${MERCHANT_URL}/login`)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    await page.goto(`${MERCHANT_URL}/withdrawal`)
    await delay(2000)

    await expect(page.locator('text=Available Balance')).toBeVisible()
    await delay(2000)
  })

  test('7. Merchant Portal — View Subscriptions page', async ({ page }) => {
    // Login
    await page.goto(`${MERCHANT_URL}/login`)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    await page.goto(`${MERCHANT_URL}/subscriptions`)
    await delay(2000)

    await expect(page.locator('text=Subscriptions')).toBeVisible()
    await delay(2000)
  })

  test('8. Merchant Portal — View Settings page', async ({ page }) => {
    // Login
    await page.goto(`${MERCHANT_URL}/login`)
    await page.fill('#email', MERCHANT_EMAIL)
    await page.fill('#password', MERCHANT_PASSWORD)
    await page.click('button[type="submit"]')
    await delay(3000)

    await page.goto(`${MERCHANT_URL}/settings`)
    await delay(2000)

    // Check merchant info is displayed (not hardcoded dashes)
    await expect(page.locator(`text=${BUSINESS_NAME}`)).toBeVisible()
    await delay(2000)
  })

  // ==========================================
  // ACT 2: ADMIN DASHBOARD
  // ==========================================
  test('9. Admin Dashboard — Login', async ({ page }) => {
    await page.goto(`${ADMIN_URL}/login`)
    await delay(1500)

    // Should see admin login form
    await expect(page.locator('text=Admin Login')).toBeVisible()
    await delay(1000)

    // Fill credentials
    await page.fill('#email', 'admin@openlankapay.lk')
    await delay(500)
    await page.fill('#password', 'Admin@2024')
    await delay(500)

    // Submit
    await page.click('button[type="submit"]')
    await delay(3000)

    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/$/, { timeout: 10000 })
    await delay(2000)

    // Dashboard should show stats
    await expect(page.locator('text=Total Merchants')).toBeVisible()
    await expect(page.locator('text=Pending KYC')).toBeVisible()
    await delay(2000)
  })

  test('10. Admin Dashboard — View Merchants List', async ({ page }) => {
    // Login
    await page.goto(`${ADMIN_URL}/login`)
    await page.fill('#email', 'admin@openlankapay.lk')
    await page.fill('#password', 'Admin@2024')
    await page.click('button[type="submit"]')
    await delay(3000)

    // Navigate to merchants
    await page.goto(`${ADMIN_URL}/merchants`)
    await delay(2000)

    // Should see merchants table with our registered merchant
    await expect(page.locator('text=Merchants')).toBeVisible()
    await delay(1000)

    // Look for Curry House in the table
    await expect(page.locator(`text=${BUSINESS_NAME}`)).toBeVisible({ timeout: 10000 })
    await delay(2000)
  })

  test('11. Admin Dashboard — View Withdrawals', async ({ page }) => {
    // Login
    await page.goto(`${ADMIN_URL}/login`)
    await page.fill('#email', 'admin@openlankapay.lk')
    await page.fill('#password', 'Admin@2024')
    await page.click('button[type="submit"]')
    await delay(3000)

    await page.goto(`${ADMIN_URL}/withdrawals`)
    await delay(2000)

    await expect(page.locator('text=Withdrawal Approvals')).toBeVisible()
    await delay(2000)
  })

  test('12. Admin Dashboard — View Treasury', async ({ page }) => {
    // Login
    await page.goto(`${ADMIN_URL}/login`)
    await page.fill('#email', 'admin@openlankapay.lk')
    await page.fill('#password', 'Admin@2024')
    await page.click('button[type="submit"]')
    await delay(3000)

    await page.goto(`${ADMIN_URL}/treasury`)
    await delay(2000)

    // Should show real exchange rate
    await expect(page.locator('text=1 USDT =')).toBeVisible()
    await expect(page.locator('text=325')).toBeVisible()
    await delay(2000)
  })

  // ==========================================
  // ACT 3: CHECKOUT FLOW
  // ==========================================
  test('13. Checkout Page — View payment checkout', async ({ page }) => {
    // First create a payment via API
    const regRes = await fetch(`${API_URL}/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        businessName: 'Checkout Test',
        email: `checkout-${Date.now()}@test.com`,
        password: 'TestPass1',
        name: 'Test',
      }),
    })
    const regData = await regRes.json()
    const token = regData.data.accessToken

    const payRes = await fetch(`${API_URL}/v1/payments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ amount: '15.00', currency: 'USDT', provider: 'TEST' }),
    })
    const payData = await payRes.json()
    const paymentId = payData.data.id

    // Open checkout page
    await page.goto(`${MERCHANT_URL}/checkout/${paymentId}`)
    await delay(2000)

    // Should show payment details
    await expect(page.locator('text=15')).toBeVisible()
    await expect(page.locator('text=USDT')).toBeVisible()
    await expect(page.locator('text=Waiting for payment')).toBeVisible()
    await delay(3000)

    // Simulate payment via API
    const providerPayId = payData.data.providerPayId
    await fetch(`${API_URL}/test/simulate/${providerPayId}`, { method: 'POST' })
    await fetch(`${API_URL}/v1/payments/${paymentId}/callback`, { method: 'POST' })

    // Wait for polling to detect PAID status
    await delay(5000)

    // Should show payment successful
    await expect(page.locator('text=Payment Successful')).toBeVisible({ timeout: 15000 })
    await delay(3000)
  })
})
