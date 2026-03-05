import { useState } from 'react';
import { usePolicies, useDeletePolicy } from '../hooks/usePolicies';
import type { Policy } from '../api/client';

const typeColors: Record<string, string> = {
  deployment: 'bg-blue-100 text-blue-800',
  scaling: 'bg-purple-100 text-purple-800',
  alerting: 'bg-yellow-100 text-yellow-800',
  compliance: 'bg-green-100 text-green-800',
};

function PolicyCard({ policy, onDelete }: { policy: Policy; onDelete: () => void }) {
  const [expanded, setExpanded] = useState(false);
  const ruleCount = Object.keys(policy.rules || {}).length;

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <h3 className="font-semibold text-gray-900">{policy.name}</h3>
          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${typeColors[policy.policy_type] || 'bg-gray-100 text-gray-800'}`}>
            {policy.policy_type}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${policy.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
          <span className="text-xs text-gray-500">{policy.enabled ? 'Active' : 'Disabled'}</span>
        </div>
      </div>

      {policy.description && (
        <p className="text-sm text-gray-600 mb-2">{policy.description}</p>
      )}

      <div className="flex items-center gap-4 text-xs text-gray-500 mb-2">
        <span>Priority: {policy.priority}</span>
        <span>{ruleCount} rule{ruleCount !== 1 ? 's' : ''}</span>
        {policy.target_fleet_id && <span>Fleet-scoped</span>}
      </div>

      {expanded && (
        <div className="mt-3 p-3 bg-gray-50 rounded text-xs font-mono overflow-auto max-h-40">
          <pre>{JSON.stringify(policy.rules, null, 2)}</pre>
        </div>
      )}

      <div className="flex items-center justify-between mt-3 pt-3 border-t">
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-xs text-blue-600 hover:text-blue-800"
        >
          {expanded ? 'Hide rules' : 'Show rules'}
        </button>
        <div className="flex gap-2">
          <button
            onClick={onDelete}
            className="text-xs text-red-500 hover:text-red-700"
          >
            Delete
          </button>
        </div>
      </div>

      <div className="text-xs text-gray-400 mt-2">
        Updated {new Date(policy.updated_at).toLocaleString()}
      </div>
    </div>
  );
}

export default function PoliciesPage() {
  const [typeFilter, setTypeFilter] = useState<string>('');
  const { data, isLoading } = usePolicies(typeFilter || undefined);
  const deleteMutation = useDeletePolicy();
  const policies = data?.policies || [];

  const types = ['', 'deployment', 'scaling', 'alerting', 'compliance'];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Policies</h2>
      </div>

      <div className="flex gap-2 mb-4">
        {types.map((t) => (
          <button
            key={t}
            onClick={() => setTypeFilter(t)}
            className={`px-3 py-1 rounded-full text-sm ${
              typeFilter === t
                ? 'bg-gray-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {t || 'All'}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading policies...</div>
      ) : policies.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500 mb-2">No policies configured.</p>
          <p className="text-sm text-gray-400">
            Policies control deployment behavior, scaling, and compliance rules.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {policies.map((policy) => (
            <PolicyCard
              key={policy.id}
              policy={policy}
              onDelete={() => {
                if (confirm(`Delete policy "${policy.name}"?`)) {
                  deleteMutation.mutate(policy.id);
                }
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}
