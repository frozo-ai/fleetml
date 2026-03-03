import { useDevices } from '../hooks/useDevices';
import { useDeployments } from '../hooks/useDeployments';
import DeploymentProgress from '../components/DeploymentProgress';

export default function DashboardPage() {
  const { data: devicesData } = useDevices();
  const { data: deploymentsData } = useDeployments();

  const devices = devicesData?.devices || [];
  const deployments = deploymentsData?.deployments || [];

  const healthy = devices.filter((d) => d.status === 'healthy').length;
  const warning = devices.filter((d) => d.status === 'warning').length;
  const offline = devices.filter((d) => d.status === 'offline').length;

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Fleet Overview</h2>

      <div className="grid grid-cols-4 gap-4 mb-8">
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <div className="text-3xl font-bold text-gray-900">{devices.length}</div>
          <div className="text-sm text-gray-500">Total Devices</div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <div className="text-3xl font-bold text-green-600">{healthy}</div>
          <div className="text-sm text-gray-500">Healthy</div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <div className="text-3xl font-bold text-yellow-600">{warning}</div>
          <div className="text-sm text-gray-500">Warning</div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <div className="text-3xl font-bold text-red-600">{offline}</div>
          <div className="text-sm text-gray-500">Offline</div>
        </div>
      </div>

      <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Deployments</h3>
      <div className="space-y-3">
        {deployments.length === 0 ? (
          <p className="text-gray-500 text-sm">No deployments yet.</p>
        ) : (
          deployments.slice(0, 5).map((d) => (
            <DeploymentProgress key={d.id} deployment={d} />
          ))
        )}
      </div>
    </div>
  );
}
