import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useModels } from '../hooks/useModels';
import { api } from '../api/client';
import type { Model, CompileResult } from '../api/client';

const runtimes = [
  { value: 'tensorrt', label: 'TensorRT', desc: 'NVIDIA GPUs (Jetson, T4, A100)' },
  { value: 'openvino', label: 'OpenVINO', desc: 'Intel CPUs/VPUs (NUC, Movidius)' },
  { value: 'tflite', label: 'TFLite', desc: 'ARM CPUs (Raspberry Pi, mobile)' },
  { value: 'mock', label: 'Mock', desc: 'Testing only (copies ONNX as-is)' },
];

function CompileCard({ model }: { model: Model }) {
  const [selectedRuntime, setSelectedRuntime] = useState('');
  const [result, setResult] = useState<CompileResult | null>(null);
  const queryClient = useQueryClient();

  const compileMutation = useMutation({
    mutationFn: () => api.compileModel(model.id, selectedRuntime),
    onSuccess: (data) => {
      setResult(data);
      queryClient.invalidateQueries({ queryKey: ['models'] });
    },
  });

  const isOnnx = model.format === 'onnx';

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <div className="flex items-center justify-between mb-3">
        <div>
          <h3 className="font-semibold text-gray-900">{model.name}</h3>
          <p className="text-xs text-gray-500">
            v{model.version} - {model.format.toUpperCase()} - {formatBytes(model.artifact_size)}
          </p>
        </div>
        {!isOnnx && (
          <span className="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded">
            Non-ONNX
          </span>
        )}
      </div>

      {isOnnx ? (
        <div>
          <div className="grid grid-cols-2 gap-2 mb-3">
            {runtimes.map((rt) => (
              <button
                key={rt.value}
                onClick={() => setSelectedRuntime(rt.value)}
                className={`p-2 rounded border text-left text-sm ${
                  selectedRuntime === rt.value
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <p className="font-medium">{rt.label}</p>
                <p className="text-xs text-gray-500">{rt.desc}</p>
              </button>
            ))}
          </div>

          <button
            onClick={() => compileMutation.mutate()}
            disabled={!selectedRuntime || compileMutation.isPending}
            className="w-full py-2 bg-gray-900 text-white rounded text-sm hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {compileMutation.isPending ? 'Compiling...' : `Compile to ${selectedRuntime || '...'}`}
          </button>

          {compileMutation.isError && (
            <p className="mt-2 text-sm text-red-600">
              {(compileMutation.error as Error).message}
            </p>
          )}

          {result && (
            <div className="mt-3 p-3 bg-green-50 rounded border border-green-200">
              <p className="text-sm font-medium text-green-800">Compilation complete</p>
              <div className="text-xs text-green-700 mt-1 space-y-1">
                <p>Runtime: {result.runtime}</p>
                <p>Size: {formatBytes(result.file_size)}</p>
                <p>Time: {result.compile_time_seconds.toFixed(1)}s</p>
                <p className="font-mono truncate">Checksum: {result.checksum}</p>
              </div>
            </div>
          )}
        </div>
      ) : (
        <p className="text-sm text-gray-500">
          Only ONNX models can be compiled. Convert this model to ONNX first.
        </p>
      )}
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export default function CompilePage() {
  const { data, isLoading } = useModels();
  const models = data?.models || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Model Compilation</h2>
          <p className="text-sm text-gray-500 mt-1">
            Compile ONNX models to chip-specific runtimes for optimized edge deployment
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading models...</div>
      ) : models.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500">No models in the registry. Upload a model first.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {models.map((model) => (
            <CompileCard key={model.id} model={model} />
          ))}
        </div>
      )}
    </div>
  );
}
