import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import LandingPage from './pages/LandingPage'
import DashboardPage from './pages/Dashboard'
import DevicesPage from './pages/Devices'
import DeviceDetailPage from './pages/DeviceDetail'
import ModelsPage from './pages/Models'
import DeploymentsPage from './pages/Deployments'
import ABTestsPage from './pages/ABTests'
import PoliciesPage from './pages/Policies'
import DriftAlertsPage from './pages/DriftAlerts'
import CompilePage from './pages/Compile'
import SettingsPage from './pages/Settings'

function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/dashboard" element={<Layout><DashboardPage /></Layout>} />
      <Route path="/dashboard/devices" element={<Layout><DevicesPage /></Layout>} />
      <Route path="/dashboard/devices/:deviceId" element={<Layout><DeviceDetailPage /></Layout>} />
      <Route path="/dashboard/models" element={<Layout><ModelsPage /></Layout>} />
      <Route path="/dashboard/deployments" element={<Layout><DeploymentsPage /></Layout>} />
      <Route path="/dashboard/ab-tests" element={<Layout><ABTestsPage /></Layout>} />
      <Route path="/dashboard/policies" element={<Layout><PoliciesPage /></Layout>} />
      <Route path="/dashboard/drift" element={<Layout><DriftAlertsPage /></Layout>} />
      <Route path="/dashboard/compile" element={<Layout><CompilePage /></Layout>} />
      <Route path="/dashboard/settings" element={<Layout><SettingsPage /></Layout>} />
    </Routes>
  )
}

export default App
