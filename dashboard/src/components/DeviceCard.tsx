import { Device } from '../api/client';

const statusColors: Record<string, string> = {
  healthy: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  offline: 'bg-red-100 text-red-800',
  registered: 'bg-blue-100 text-blue-800',
};

export default function DeviceCard({ device }: { device: Device }) {
  return (
    <div className="bg-white rounded-lg shadow-sm border p-4 hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-medium text-gray-900">{device.device_id}</h3>
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[device.status] || 'bg-gray-100'}`}>
          {device.status}
        </span>
      </div>
      {device.name && (
        <p className="text-sm text-gray-500 mb-2">{device.name}</p>
      )}
      <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
        <div>Arch: {device.arch}</div>
        <div>GPU: {device.gpu_type || 'none'}</div>
        <div>Runtime: {device.runtime}</div>
        <div>RAM: {device.ram_mb} MB</div>
      </div>
      {device.cpu_percent !== undefined && (
        <div className="mt-3 pt-3 border-t">
          <div className="flex justify-between text-xs text-gray-500">
            <span>CPU: {device.cpu_percent?.toFixed(1)}%</span>
            <span>RAM: {device.ram_mb_used} MB</span>
            <span>Disk: {device.disk_percent?.toFixed(1)}%</span>
          </div>
        </div>
      )}
    </div>
  );
}
