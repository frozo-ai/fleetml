import { Model } from '../api/client';

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export default function ModelCard({ model }: { model: Model }) {
  return (
    <div className="bg-white rounded-lg shadow-sm border p-4 hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-medium text-gray-900">{model.name}</h3>
        <span className="text-sm text-gray-500">v{model.version}</span>
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs text-gray-600 mb-2">
        <div>Format: {model.format}</div>
        <div>Size: {formatBytes(model.artifact_size)}</div>
      </div>
      {model.description && (
        <p className="text-xs text-gray-500 mb-2">{model.description}</p>
      )}
      {model.tags && model.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {model.tags.map((tag) => (
            <span key={tag} className="px-2 py-0.5 bg-gray-100 text-gray-600 rounded text-xs">
              {tag}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
