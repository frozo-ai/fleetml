import { useModels } from '../hooks/useModels';
import ModelCard from '../components/ModelCard';

export default function ModelsPage() {
  const { data, isLoading } = useModels();
  const models = data?.models || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Models</h2>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading models...</div>
      ) : models.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500 mb-2">No models uploaded yet.</p>
          <p className="text-sm text-gray-400">Use the CLI to upload your first model:</p>
          <code className="block mt-2 bg-gray-100 rounded p-2 text-sm">
            fleetml deploy model.onnx --name my-model --version 1.0
          </code>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {models.map((model) => (
            <ModelCard key={model.id} model={model} />
          ))}
        </div>
      )}
    </div>
  );
}
