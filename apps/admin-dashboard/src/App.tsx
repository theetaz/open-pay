import { Routes, Route, Navigate } from 'react-router-dom'
import { LoginPage } from '#/pages/login'
import { DashboardLayout } from '#/layouts/dashboard'
import { DashboardIndex } from '#/pages/dashboard'
import { MerchantsPage } from '#/pages/merchants'
import { WithdrawalsPage } from '#/pages/withdrawals'
import { TreasuryPage } from '#/pages/treasury'
import { AuditLogsPage } from '#/pages/audit-logs'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<DashboardLayout />}>
        <Route index element={<DashboardIndex />} />
        <Route path="merchants" element={<MerchantsPage />} />
        <Route path="withdrawals" element={<WithdrawalsPage />} />
        <Route path="treasury" element={<TreasuryPage />} />
        <Route path="audit-logs" element={<AuditLogsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
