import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export default function SettingsPage() {
  const { data: health } = useQuery({
    queryKey: ['health'],
    queryFn: () => api.health(),
  });

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Settings</h2>

      <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Server Info</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Status:</span>{' '}
            <span className="font-medium">{health?.status || '-'}</span>
          </div>
          <div>
            <span className="text-gray-500">Version:</span>{' '}
            <span className="font-medium">{health?.version || '-'}</span>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">User Management</h3>
        <p className="text-sm text-gray-500">User management coming soon.</p>
      </div>
    </div>
  );
}
