import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthLayout } from '#/layouts/auth'
import { DashboardLayout } from '#/layouts/dashboard'
import { LoginPage } from '#/pages/login'
import { RegisterPage } from '#/pages/register'
import { DashboardIndex } from '#/pages/dashboard'
import { ActivatePage } from '#/pages/activate'
import { PaymentsPage } from '#/pages/payments'
import { PaymentLinksPage } from '#/pages/payment-links'
import { SubscriptionsPage } from '#/pages/subscriptions'
import { WithdrawalPage } from '#/pages/withdrawal'
import { BranchesPage } from '#/pages/branches'
import { UsersPage } from '#/pages/users'
import { SettingsPage } from '#/pages/settings'
import { ProfilePage } from '#/pages/profile'
import { SecurityPage } from '#/pages/security'
import { AuditLogPage } from '#/pages/audit-log'
import { IntegrationsPage } from '#/pages/integrations'
import { ExamplePage } from '#/pages/example'
import { PaymentLinkCheckout } from '#/pages/pay-slug'
import { CheckoutPage } from '#/pages/checkout'
import { SandboxPayPage } from '#/pages/sandbox-pay'
import { VerifyDirectorPage } from '#/pages/verify-director'

export function App() {
  return (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
      </Route>
      <Route element={<DashboardLayout />}>
        <Route index element={<DashboardIndex />} />
        <Route path="activate" element={<ActivatePage />} />
        <Route path="payments" element={<PaymentsPage />} />
        <Route path="payment-links" element={<PaymentLinksPage />} />
        <Route path="subscriptions" element={<SubscriptionsPage />} />
        <Route path="withdrawal" element={<WithdrawalPage />} />
        <Route path="branches" element={<BranchesPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="security" element={<SecurityPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="integrations" element={<IntegrationsPage />} />
        <Route path="audit-log" element={<AuditLogPage />} />
        <Route path="example" element={<ExamplePage />} />
      </Route>
      <Route path="/pay/:slug" element={<PaymentLinkCheckout />} />
      <Route path="/checkout/:paymentId" element={<CheckoutPage />} />
      <Route path="/sandbox/pay/:providerPayId" element={<SandboxPayPage />} />
      <Route path="/verify/director/:token" element={<VerifyDirectorPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
