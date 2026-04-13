import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: 'H' },
  { path: '/dashboard/get-started', label: 'Get Started', icon: 'G' },
  { path: '/dashboard/devices', label: 'Devices', icon: 'D' },
  { path: '/dashboard/models', label: 'Models', icon: 'M' },
  { path: '/dashboard/deployments', label: 'Deployments', icon: 'R' },
  { path: '/dashboard/ab-tests', label: 'A/B Tests', icon: 'A' },
  { path: '/dashboard/policies', label: 'Policies', icon: 'P' },
  { path: '/dashboard/drift', label: 'Drift', icon: 'W' },
  { path: '/dashboard/compile', label: 'Compile', icon: 'C' },
{ path: '/dashboard/settings', label: 'Settings', icon: 'S' },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, organization, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="flex h-screen bg-gray-50">
      <aside className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <Link to="/dashboard" className="block">
            <h1 className="text-xl font-bold">FleetML</h1>
            <p className="text-xs text-gray-400 mt-1">
              {organization?.name || 'Edge MLOps Platform'}
            </p>
          </Link>
          {organization && (
            <span className={`inline-block mt-2 px-2 py-0.5 rounded text-xs font-medium ${
              organization.plan === 'pro' ? 'bg-blue-600 text-white' :
              organization.plan === 'starter' ? 'bg-green-600 text-white' :
              organization.plan === 'enterprise' ? 'bg-purple-600 text-white' :
              'bg-gray-700 text-gray-300'
            }`}>
              {organization.plan.charAt(0).toUpperCase() + organization.plan.slice(1)}
            </span>
          )}
        </div>
        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center px-3 py-2 rounded-lg text-sm ${
                location.pathname === item.path
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:bg-gray-800 hover:text-white'
              }`}
            >
              <span className="w-6 h-6 flex items-center justify-center mr-3 text-xs font-bold bg-gray-700 rounded">
                {item.icon}
              </span>
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="p-4 border-t border-gray-700">
          {user && (
            <div className="mb-3">
              <p className="text-sm text-white truncate">{user.email}</p>
              <p className="text-xs text-gray-400 capitalize">{user.role}</p>
            </div>
          )}
          <button
            onClick={handleLogout}
            className="w-full text-left text-sm text-gray-400 hover:text-white px-2 py-1 rounded hover:bg-gray-800"
          >
            Sign out
          </button>
          <p className="text-xs text-gray-600 mt-2">v0.1.0</p>
        </div>
      </aside>
      <main className="flex-1 overflow-auto">
        <div className="p-8">{children}</div>
      </main>
    </div>
  );
}
