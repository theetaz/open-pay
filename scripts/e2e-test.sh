#!/bin/bash
set -e
PASS=0; FAIL=0; GW="http://localhost:8080"
pass() { echo "  PASS: $1"; PASS=$((PASS+1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL+1)); }
check() { if [ $? -eq 0 ]; then pass "$1"; else fail "$1"; fi; }

echo "=========================================="
echo "  COMPREHENSIVE E2E TEST"
echo "=========================================="
echo ""

# --- ADMIN AUTH ---
echo "--- ADMIN AUTH ---"
echo "1. Admin Login"
AR=$(curl -sf $GW/v1/admin/auth/login -H "Content-Type: application/json" -d '{"email":"admin@openlankapay.lk","password":"Admin@2024"}')
AT=$(echo "$AR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
[ -n "$AT" ] && pass "got admin token" || fail "no admin token"

echo "2. Admin Me"
R=$(curl -sf $GW/v1/admin/auth/me -H "Authorization: Bearer $AT")
echo "$R" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; assert d['role']['name']=='SUPER_ADMIN'" 2>/dev/null && pass "SUPER_ADMIN" || fail "wrong role"

echo "3. Admin Wrong Password"
CODE=$(curl -s -o /dev/null -w "%{http_code}" $GW/v1/admin/auth/login -H "Content-Type: application/json" -d '{"email":"admin@openlankapay.lk","password":"wrong"}')
[ "$CODE" = "401" ] && pass "rejected 401" || fail "expected 401 got $CODE"

# --- MERCHANT REGISTRATION ---
echo ""
echo "--- MERCHANT REGISTRATION ---"
EM="curry-$RANDOM@test.com"
echo "4. Register Merchant ($EM)"
MR=$(curl -sf $GW/v1/auth/register -H "Content-Type: application/json" -d "{\"businessName\":\"Curry House\",\"email\":\"$EM\",\"password\":\"TestPass1\",\"name\":\"Kamal Perera\"}")
MT=$(echo "$MR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
MID=$(echo "$MR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['merchant']['id'])")
[ -n "$MT" ] && pass "merchant registered, ID: ${MID:0:8}" || fail "registration failed"

echo "5. Merchant Me"
R=$(curl -sf $GW/v1/auth/me -H "Authorization: Bearer $MT")
BNAME=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['merchant']['businessName'])")
[ "$BNAME" = "Curry House" ] && pass "$BNAME" || fail "expected Curry House got $BNAME"

echo "6. KYC Submit"
R=$(curl -sf -X PUT "$GW/v1/merchants/$MID" -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d '{"businessType":"Restaurant","bankName":"Commercial Bank","bankAccountNo":"123456789","bankAccountName":"Kamal Perera","city":"Colombo","submitKyc":true}')
KYC=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['kycStatus'])")
[ "$KYC" = "INSTANT_ACCESS" ] && pass "KYC: $KYC" || fail "expected INSTANT_ACCESS got $KYC"

# --- ADMIN VIEWS MERCHANTS ---
echo ""
echo "--- ADMIN VIEWS MERCHANTS ---"
echo "7. Admin Lists Merchants"
R=$(curl -sf "$GW/v1/merchants" -H "Authorization: Bearer $AT")
TOTAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['meta']['total'])")
[ "$TOTAL" -gt 0 ] && pass "total: $TOTAL merchants" || fail "no merchants found"

echo "8. Admin Sees Curry House in List"
echo "$R" | python3 -c "import sys,json; ms=json.load(sys.stdin)['data']; found=[m for m in ms if m['businessName']=='Curry House']; assert len(found)>0" 2>/dev/null && pass "Curry House found in list" || fail "Curry House not in list"

echo "9. Admin Approves Merchant"
R=$(curl -sf -X POST "$GW/v1/merchants/$MID/approve" -H "Authorization: Bearer $AT")
STATUS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
[ "$STATUS" = "approved" ] && pass "merchant approved" || fail "approve failed: $STATUS"

echo "10. Verify KYC Now APPROVED"
R=$(curl -sf "$GW/v1/merchants/$MID" -H "Authorization: Bearer $MT")
KYC=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['kycStatus'])")
[ "$KYC" = "APPROVED" ] && pass "KYC: $KYC" || fail "expected APPROVED got $KYC"

# --- PAYMENT FLOW ---
echo ""
echo "--- PAYMENT FLOW ---"
echo "11. Create Payment"
PR=$(curl -sf $GW/v1/payments -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d '{"amount":"25.50","currency":"USDT","provider":"TEST","merchantTradeNo":"TABLE-7"}')
PID=$(echo "$PR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
PROVID=$(echo "$PR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['providerPayId'])")
[ -n "$PID" ] && pass "payment ${PID:0:8}" || fail "create payment failed"

echo "12. Checkout (public)"
R=$(curl -sf "$GW/v1/payments/$PID/checkout")
PSTATUS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
[ "$PSTATUS" = "INITIATED" ] && pass "status: $PSTATUS" || fail "expected INITIATED got $PSTATUS"

echo "13. Simulate + Callback"
curl -sf -X POST "$GW/test/simulate/$PROVID" > /dev/null && curl -sf -X POST "$GW/v1/payments/$PID/callback" > /dev/null && pass "simulated + callback" || fail "simulate/callback"

echo "14. Verify PAID"
R=$(curl -sf "$GW/v1/payments/$PID/checkout")
PSTATUS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
[ "$PSTATUS" = "PAID" ] && pass "status: $PSTATUS" || fail "expected PAID got $PSTATUS"

echo "15. List Payments"
R=$(curl -sf "$GW/v1/payments" -H "Authorization: Bearer $MT")
PTOTAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['meta']['total'])")
[ "$PTOTAL" -gt 0 ] && pass "total: $PTOTAL" || fail "no payments"

# --- SETTLEMENT & WITHDRAWAL ---
echo ""
echo "--- SETTLEMENT & WITHDRAWAL ---"
echo "16. Credit Balance"
curl -sf -X POST "$GW/v1/settlements/credit" -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d "{\"merchantId\":\"$MID\",\"netUsdt\":\"24.99\",\"feesUsdt\":\"0.51\"}" > /dev/null && pass "credited" || fail "credit"

echo "17. Check Balance"
R=$(curl -sf "$GW/v1/settlements/balance" -H "Authorization: Bearer $MT")
BAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['availableUsdt'])")
[ "$BAL" = "24.99" ] && pass "balance: $BAL USDT" || fail "expected 24.99 got $BAL"

echo "18. Request Withdrawal"
WR=$(curl -sf $GW/v1/withdrawals -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d '{"amountUsdt":"10","exchangeRate":"325","bankName":"Commercial Bank","bankAccountNo":"123456789","bankAccountName":"Kamal"}')
WID=$(echo "$WR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
WSTATUS=$(echo "$WR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
[ "$WSTATUS" = "REQUESTED" ] && pass "withdrawal $WSTATUS" || fail "expected REQUESTED got $WSTATUS"

echo "19. Admin Approves Withdrawal"
curl -sf -X POST "$GW/v1/withdrawals/$WID/approve" -H "Authorization: Bearer $AT" > /dev/null && pass "approved" || fail "approve"

echo "20. Admin Completes Withdrawal"
curl -sf -X POST "$GW/v1/withdrawals/$WID/complete" -H "Content-Type: application/json" -H "Authorization: Bearer $AT" -d '{"bankReference":"TXN-20260316-001"}' > /dev/null && pass "completed" || fail "complete"

# --- SUBSCRIPTIONS ---
echo ""
echo "--- SUBSCRIPTIONS ---"
echo "21. Create Plan"
SP=$(curl -sf $GW/v1/subscription-plans -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d '{"name":"Premium Monthly","description":"Full access","amount":"10","currency":"USDT","intervalType":"MONTHLY","intervalCount":1,"trialDays":7}')
SPID=$(echo "$SP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
[ -n "$SPID" ] && pass "plan ${SPID:0:8}" || fail "create plan"

echo "22. List Plans"
R=$(curl -sf "$GW/v1/subscription-plans" -H "Authorization: Bearer $MT")
PCOUNT=$(echo "$R" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))")
[ "$PCOUNT" -gt 0 ] && pass "plans: $PCOUNT" || fail "no plans"

echo "23. Subscribe Customer"
R=$(curl -sf -X POST "$GW/v1/subscription-plans/$SPID/subscribe" -H "Content-Type: application/json" -d '{"email":"customer@example.com","wallet":"0xABC123"}')
SSTATUS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
[ "$SSTATUS" = "TRIAL" ] && pass "subscription: $SSTATUS" || fail "expected TRIAL got $SSTATUS"

# --- BRANCHES ---
echo ""
echo "--- BRANCHES ---"
echo "24. Create Branch"
R=$(curl -sf $GW/v1/branches -H "Content-Type: application/json" -H "Authorization: Bearer $MT" -d '{"name":"Colombo Branch","city":"Colombo","address":"123 Galle Road"}')
BID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
[ -n "$BID" ] && pass "branch ${BID:0:8}" || fail "create branch"

echo "25. List Branches"
R=$(curl -sf "$GW/v1/branches" -H "Authorization: Bearer $MT")
BCOUNT=$(echo "$R" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))")
[ "$BCOUNT" -gt 0 ] && pass "branches: $BCOUNT" || fail "no branches"

# --- EXCHANGE RATE ---
echo ""
echo "--- EXCHANGE RATE ---"
echo "26. Get Active Rate"
R=$(curl -sf "$GW/v1/exchange-rates/active")
RATE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['rate'])")
[ -n "$RATE" ] && pass "1 USDT = $RATE LKR" || fail "no rate"

# --- SECURITY ---
echo ""
echo "--- SECURITY ---"
echo "27. Merchant Token Cannot Access Admin Me"
CODE=$(curl -s -o /dev/null -w "%{http_code}" $GW/v1/admin/auth/me -H "Authorization: Bearer $MT")
[ "$CODE" = "403" ] && pass "forbidden (403)" || fail "expected 403 got $CODE"

echo "28. No Token Gets 401"
CODE=$(curl -s -o /dev/null -w "%{http_code}" $GW/v1/payments)
[ "$CODE" = "401" ] && pass "unauthorized (401)" || fail "expected 401 got $CODE"

echo "29. Health Check"
curl -sf $GW/healthz > /dev/null && pass "ok" || fail "healthz"

echo ""
echo "=========================================="
echo "  RESULTS: $PASS passed, $FAIL failed out of 29"
echo "=========================================="
[ $FAIL -eq 0 ] && exit 0 || exit 1
