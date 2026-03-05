import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useDevices } from '../hooks/useDevices';

export default function DevicesPage() {
  const [statusFilter, setStatusFilter] = useState('');
  const { data, isLoading } = useDevices(statusFilter ? { status: statusFilter } : undefined);

  const devices = data?.devices || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Devices</h2>
        <div className="flex gap-2">
          {['', 'healthy', 'warning', 'offline'].map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={`px-3 py-1 rounded text-sm ${
                statusFilter === s
                  ? 'bg-gray-900 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {s || 'All'}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading devices...</div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Device ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Hardware</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Runtime</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">CPU</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Last Seen</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {devices.map((device) => (
                <tr key={device.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link to={`/dashboard/devices/${device.device_id}`} className="text-blue-600 hover:underline font-medium">
                      {device.device_id}
                    </Link>
                    {device.name && <div className="text-xs text-gray-500">{device.name}</div>}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={device.status} />
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-700">
                    {device.arch} / {device.gpu_type || 'no GPU'}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-700">{device.runtime}</td>
                  <td className="px-4 py-3 text-sm text-gray-700">
                    {device.cpu_percent !== undefined ? `${device.cpu_percent.toFixed(1)}%` : '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {device.last_heartbeat
                      ? new Date(device.last_heartbeat).toLocaleString()
                      : 'Never'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {devices.length === 0 && (
            <div className="text-center py-8 text-gray-500">No devices found.</div>
          )}
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    healthy: 'bg-green-100 text-green-800',
    warning: 'bg-yellow-100 text-yellow-800',
    offline: 'bg-red-100 text-red-800',
    registered: 'bg-blue-100 text-blue-800',
  };

  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${colors[status] || 'bg-gray-100 text-gray-800'}`}>
      {status}
    </span>
  );
}
