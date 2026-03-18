import { Routes, Route, Navigate } from 'react-router-dom'
import { LoginPage } from '#/pages/login'
import { DashboardLayout } from '#/layouts/dashboard'
import { DashboardIndex } from '#/pages/dashboard'
import { MerchantsPage } from '#/pages/merchants'
import { WithdrawalsPage } from '#/pages/withdrawals'
import { TreasuryPage } from '#/pages/treasury'
import { AuditLogsPage } from '#/pages/audit-logs'
import { SystemHealthPage } from '#/pages/system-health'
import { SettingsGeneralPage } from '#/pages/settings/general'
import { SettingsFeesPage } from '#/pages/settings/fees'
import { SettingsEmailPage } from '#/pages/settings/email'
import { SettingsLegalDocumentsPage } from '#/pages/settings/legal-documents'
import { SettingsEmailTemplatesPage } from '#/pages/settings/email-templates'
import { SettingsTeamPage } from '#/pages/settings/team'
import { SettingsRolesPage } from '#/pages/settings/roles'
import { ChangePasswordPage } from '#/pages/change-password'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/change-password" element={<ChangePasswordPage />} />
      <Route element={<DashboardLayout />}>
        <Route index element={<DashboardIndex />} />
        <Route path="merchants" element={<MerchantsPage />} />
        <Route path="withdrawals" element={<WithdrawalsPage />} />
        <Route path="treasury" element={<TreasuryPage />} />
        <Route path="audit-logs" element={<AuditLogsPage />} />
        <Route path="system-health" element={<SystemHealthPage />} />
        <Route path="settings">
          <Route index element={<Navigate to="general" replace />} />
          <Route path="general" element={<SettingsGeneralPage />} />
          <Route path="fees" element={<SettingsFeesPage />} />
          <Route path="email" element={<SettingsEmailPage />} />
          <Route path="legal-documents" element={<SettingsLegalDocumentsPage />} />
          <Route path="email-templates" element={<SettingsEmailTemplatesPage />} />
          <Route path="team" element={<SettingsTeamPage />} />
          <Route path="roles" element={<SettingsRolesPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
