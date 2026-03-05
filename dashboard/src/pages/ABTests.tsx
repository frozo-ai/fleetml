import { useState } from 'react';
import { useABTests, useStopABTest } from '../hooks/useABTests';
import type { ABTest } from '../api/client';

const stateColors: Record<string, string> = {
  pending: 'bg-gray-100 text-gray-800',
  running: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  stopped: 'bg-red-100 text-red-800',
};

function ABTestCard({ test, onStop }: { test: ABTest; onStop: () => void }) {
  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-semibold text-gray-900">{test.name}</h3>
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${stateColors[test.state] || stateColors.pending}`}>
          {test.state}
        </span>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-3">
        <div className="text-sm">
          <p className="text-gray-500">Model A</p>
          <p className="font-mono text-xs text-gray-700 truncate">{test.model_a_id}</p>
          <div className="mt-1">
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div className="bg-blue-600 h-2 rounded-full" style={{ width: `${test.split_a}%` }} />
            </div>
            <span className="text-xs text-gray-500">{test.split_a}% traffic</span>
          </div>
        </div>
        <div className="text-sm">
          <p className="text-gray-500">Model B</p>
          <p className="font-mono text-xs text-gray-700 truncate">{test.model_b_id}</p>
          <div className="mt-1">
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div className="bg-orange-500 h-2 rounded-full" style={{ width: `${test.split_b}%` }} />
            </div>
            <span className="text-xs text-gray-500">{test.split_b}% traffic</span>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-gray-500 border-t pt-3">
        <div className="flex gap-4">
          <span>Metric: <strong>{test.metric || 'accuracy'}</strong></span>
          {test.duration && <span>Duration: {test.duration}</span>}
          {test.auto_promote && <span className="text-green-600">Auto-promote</span>}
        </div>
        <div className="flex items-center gap-2">
          {test.winner && (
            <span className="text-green-700 font-medium">Winner: {test.winner}</span>
          )}
          {test.state === 'running' && (
            <button
              onClick={onStop}
              className="px-2 py-1 text-xs bg-red-50 text-red-700 rounded hover:bg-red-100"
            >
              Stop
            </button>
          )}
        </div>
      </div>

      <div className="text-xs text-gray-400 mt-2">
        Created {new Date(test.created_at).toLocaleString()}
      </div>
    </div>
  );
}

export default function ABTestsPage() {
  const [stateFilter, setStateFilter] = useState<string>('');
  const { data, isLoading } = useABTests(stateFilter || undefined);
  const stopMutation = useStopABTest();
  const tests = data?.ab_tests || [];

  const states = ['', 'running', 'pending', 'completed', 'stopped'];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">A/B Tests</h2>
      </div>

      <div className="flex gap-2 mb-4">
        {states.map((s) => (
          <button
            key={s}
            onClick={() => setStateFilter(s)}
            className={`px-3 py-1 rounded-full text-sm ${
              stateFilter === s
                ? 'bg-gray-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading A/B tests...</div>
      ) : tests.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500 mb-2">No A/B tests found.</p>
          <p className="text-sm text-gray-400">
            Create A/B tests via CLI: <code className="bg-gray-100 px-1 rounded">fleetml ab-test create</code>
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {tests.map((test) => (
            <ABTestCard
              key={test.id}
              test={test}
              onStop={() => stopMutation.mutate({ id: test.id })}
            />
          ))}
        </div>
      )}
    </div>
  );
}
