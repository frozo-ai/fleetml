import { Deployment } from '../api/client';

const stateColors: Record<string, string> = {
  pending: 'bg-gray-100 text-gray-800',
  rolling_out: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  cancelled: 'bg-gray-100 text-gray-800',
  rolled_back: 'bg-yellow-100 text-yellow-800',
};

export default function DeploymentProgress({ deployment }: { deployment: Deployment }) {
  const progress = deployment.total_devices > 0
    ? Math.round((deployment.completed_devices / deployment.total_devices) * 100)
    : 0;

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <div className="flex items-center justify-between mb-3">
        <div>
          <span className="text-sm font-medium text-gray-900">
            {deployment.id.substring(0, 8)}
          </span>
          <span className={`ml-2 px-2 py-0.5 rounded-full text-xs font-medium ${stateColors[deployment.state] || 'bg-gray-100'}`}>
            {deployment.state}
          </span>
        </div>
        <span className="text-xs text-gray-500">
          {new Date(deployment.created_at).toLocaleString()}
        </span>
      </div>

      <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
        <div
          className="bg-blue-600 h-2 rounded-full transition-all duration-300"
          style={{ width: `${progress}%` }}
        />
      </div>

      <div className="flex justify-between text-xs text-gray-500">
        <span>{deployment.completed_devices}/{deployment.total_devices} completed</span>
        {deployment.failed_devices > 0 && (
          <span className="text-red-500">{deployment.failed_devices} failed</span>
        )}
        {deployment.queued_devices > 0 && (
          <span>{deployment.queued_devices} queued</span>
        )}
      </div>
    </div>
  );
}
