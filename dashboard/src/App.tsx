import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import Layout from './components/Layout'
import LandingPage from './pages/LandingPage'
import LoginPage from './pages/Login'
import SignupPage from './pages/Signup'
import DashboardPage from './pages/Dashboard'
import DevicesPage from './pages/Devices'
import DeviceDetailPage from './pages/DeviceDetail'
import ModelsPage from './pages/Models'
import DeploymentsPage from './pages/Deployments'
import ABTestsPage from './pages/ABTests'
import PoliciesPage from './pages/Policies'
import DriftAlertsPage from './pages/DriftAlerts'
import CompilePage from './pages/Compile'
import BillingPage from './pages/Billing'
import OnboardingPage from './pages/Onboarding'
import SettingsPage from './pages/Settings'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/dashboard" element={<ProtectedRoute><Layout><DashboardPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/get-started" element={<ProtectedRoute><Layout><OnboardingPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/devices" element={<ProtectedRoute><Layout><DevicesPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/devices/:deviceId" element={<ProtectedRoute><Layout><DeviceDetailPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/models" element={<ProtectedRoute><Layout><ModelsPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/deployments" element={<ProtectedRoute><Layout><DeploymentsPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/ab-tests" element={<ProtectedRoute><Layout><ABTestsPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/policies" element={<ProtectedRoute><Layout><PoliciesPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/drift" element={<ProtectedRoute><Layout><DriftAlertsPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/compile" element={<ProtectedRoute><Layout><CompilePage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/billing" element={<ProtectedRoute><Layout><BillingPage /></Layout></ProtectedRoute>} />
      <Route path="/dashboard/settings" element={<ProtectedRoute><Layout><SettingsPage /></Layout></ProtectedRoute>} />
    </Routes>
  )
}

export default App
