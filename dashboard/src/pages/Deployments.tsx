import { useDeployments } from '../hooks/useDeployments';
import DeploymentProgress from '../components/DeploymentProgress';

export default function DeploymentsPage() {
  const { data, isLoading } = useDeployments();
  const deployments = data?.deployments || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Deployments</h2>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading deployments...</div>
      ) : deployments.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500">No deployments yet.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {deployments.map((d) => (
            <DeploymentProgress key={d.id} deployment={d} />
          ))}
        </div>
      )}
    </div>
  );
}
