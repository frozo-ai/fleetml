import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import LandingPage from './pages/LandingPage'
import DashboardPage from './pages/Dashboard'
import DevicesPage from './pages/Devices'
import ModelsPage from './pages/Models'
import DeploymentsPage from './pages/Deployments'
import SettingsPage from './pages/Settings'

function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/dashboard" element={<Layout><DashboardPage /></Layout>} />
      <Route path="/dashboard/devices" element={<Layout><DevicesPage /></Layout>} />
      <Route path="/dashboard/models" element={<Layout><ModelsPage /></Layout>} />
      <Route path="/dashboard/deployments" element={<Layout><DeploymentsPage /></Layout>} />
      <Route path="/dashboard/settings" element={<Layout><SettingsPage /></Layout>} />
    </Routes>
  )
}

export default App
