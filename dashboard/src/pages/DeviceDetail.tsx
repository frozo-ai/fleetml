import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useDevice, useDeviceLogs, useDeleteDevice } from '../hooks/useDevices';
import type { Device, DeviceLog } from '../api/client';

const statusColors: Record<string, string> = {
  healthy: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  offline: 'bg-red-100 text-red-800',
  registered: 'bg-blue-100 text-blue-800',
  decommissioned: 'bg-gray-100 text-gray-800',
};

const logLevelColors: Record<string, string> = {
  error: 'text-red-600',
  warn: 'text-yellow-600',
  info: 'text-blue-600',
  debug: 'text-gray-400',
};

function MetricGauge({ label, value, max, unit, color }: {
  label: string; value: number | undefined; max: number; unit: string; color: string;
}) {
  const pct = value !== undefined ? Math.min((value / max) * 100, 100) : 0;
  const displayValue = value !== undefined ? value.toFixed(1) : '-';

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{displayValue}<span className="text-sm text-gray-400 ml-1">{unit}</span></p>
      <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
        <div className={`h-2 rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

function DeviceInfo({ device }: { device: Device }) {
  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <h3 className="font-semibold text-gray-900 mb-3">Hardware Info</h3>
      <div className="grid grid-cols-2 gap-y-2 text-sm">
        <div className="text-gray-500">Architecture</div>
        <div className="text-gray-900">{device.arch}</div>
        <div className="text-gray-500">GPU</div>
        <div className="text-gray-900">{device.gpu_type || 'None'}</div>
        <div className="text-gray-500">Runtime</div>
        <div className="text-gray-900">{device.runtime}</div>
        <div className="text-gray-500">Hardware Model</div>
        <div className="text-gray-900">{device.hardware_model || '-'}</div>
        <div className="text-gray-500">RAM</div>
        <div className="text-gray-900">{device.ram_mb} MB</div>
        <div className="text-gray-500">Disk</div>
        <div className="text-gray-900">{device.disk_gb} GB</div>
        <div className="text-gray-500">Last Heartbeat</div>
        <div className="text-gray-900">
          {device.last_heartbeat ? new Date(device.last_heartbeat).toLocaleString() : 'Never'}
        </div>
      </div>

      {device.labels && Object.keys(device.labels).length > 0 && (
        <div className="mt-4 pt-3 border-t">
          <p className="text-xs text-gray-500 mb-2">Labels</p>
          <div className="flex flex-wrap gap-1">
            {Object.entries(device.labels).map(([k, v]) => (
              <span key={k} className="px-2 py-0.5 bg-gray-100 text-gray-700 rounded text-xs">
                {k}={v}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function LogsPanel({ deviceId }: { deviceId: string }) {
  const [level, setLevel] = useState('');
  const { data, isLoading } = useDeviceLogs(deviceId, { limit: 50, level: level || undefined });
  const logs = data?.logs || [];
  const levels = ['', 'error', 'warn', 'info', 'debug'];

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-semibold text-gray-900">Logs</h3>
        <div className="flex gap-1">
          {levels.map((l) => (
            <button
              key={l}
              onClick={() => setLevel(l)}
              className={`px-2 py-0.5 rounded text-xs ${
                level === l ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {l || 'All'}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <p className="text-gray-500 text-sm">Loading logs...</p>
      ) : logs.length === 0 ? (
        <p className="text-gray-400 text-sm">No logs available.</p>
      ) : (
        <div className="space-y-1 max-h-80 overflow-auto font-mono text-xs">
          {logs.map((log: DeviceLog, i: number) => (
            <div key={i} className="flex gap-2 hover:bg-gray-50 px-1 py-0.5 rounded">
              <span className="text-gray-400 whitespace-nowrap">
                {new Date(log.timestamp).toLocaleTimeString()}
              </span>
              <span className={`uppercase w-10 ${logLevelColors[log.level] || 'text-gray-500'}`}>
                {log.level}
              </span>
              <span className="text-gray-700">{log.message}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function ControlPanel({ device }: { device: Device }) {
  const navigate = useNavigate();
  const deleteMutation = useDeleteDevice();

  const handleDecommission = () => {
    if (confirm(`Decommission device "${device.device_id}"? This will remove it from any fleet.`)) {
      deleteMutation.mutate(device.device_id, {
        onSuccess: () => navigate('/dashboard/devices'),
      });
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <h3 className="font-semibold text-gray-900 mb-3">Actions</h3>
      <div className="space-y-2">
        <button
          onClick={handleDecommission}
          className="w-full py-2 px-3 text-sm text-red-700 bg-red-50 rounded hover:bg-red-100 border border-red-200"
        >
          Decommission Device
        </button>
      </div>
    </div>
  );
}

export default function DeviceDetailPage() {
  const { deviceId } = useParams<{ deviceId: string }>();
  const { data: device, isLoading } = useDevice(deviceId || '');

  if (isLoading) {
    return <div className="text-gray-500">Loading device...</div>;
  }

  if (!device) {
    return (
      <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
        <p className="text-gray-500">Device not found.</p>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">{device.device_id}</h2>
          {device.name && <p className="text-sm text-gray-500">{device.name}</p>}
        </div>
        <span className={`px-3 py-1 rounded-full text-sm font-medium ${statusColors[device.status] || statusColors.registered}`}>
          {device.status}
        </span>
      </div>

      {/* Metrics gauges */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <MetricGauge label="CPU Usage" value={device.cpu_percent} max={100} unit="%" color="bg-blue-500" />
        <MetricGauge label="GPU Usage" value={device.gpu_percent} max={100} unit="%" color="bg-purple-500" />
        <MetricGauge label="RAM Used" value={device.ram_mb_used} max={device.ram_mb} unit="MB" color="bg-green-500" />
        <MetricGauge label="Disk Usage" value={device.disk_percent} max={100} unit="%" color="bg-orange-500" />
      </div>

      {/* Temperature if available */}
      {device.temperature_c !== undefined && (
        <div className="mb-6 bg-white rounded-lg shadow-sm border p-4 flex items-center gap-4">
          <span className="text-sm text-gray-500">Temperature:</span>
          <span className={`text-lg font-bold ${
            device.temperature_c > 80 ? 'text-red-600' : device.temperature_c > 60 ? 'text-yellow-600' : 'text-green-600'
          }`}>
            {device.temperature_c.toFixed(1)}°C
          </span>
        </div>
      )}

      {/* Two column layout: info + controls / logs */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="space-y-4">
          <DeviceInfo device={device} />
          <ControlPanel device={device} />
        </div>
        <div className="lg:col-span-2">
          <LogsPanel deviceId={device.device_id} />
        </div>
      </div>
    </div>
  );
}
