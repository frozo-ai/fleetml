import { useState } from 'react';
import { useDriftReports } from '../hooks/useDrift';
import type { DriftReport } from '../api/client';

const severityColors: Record<string, string> = {
  none: 'bg-green-100 text-green-800',
  low: 'bg-yellow-100 text-yellow-800',
  medium: 'bg-orange-100 text-orange-800',
  high: 'bg-red-100 text-red-800',
};

function DriftReportRow({ report }: { report: DriftReport }) {
  return (
    <tr className="hover:bg-gray-50">
      <td className="px-4 py-3 text-sm font-mono text-gray-700">
        {report.feature_name}
      </td>
      <td className="px-4 py-3 text-sm font-mono text-gray-500 truncate max-w-32">
        {report.device_id}
      </td>
      <td className="px-4 py-3 text-sm font-mono text-gray-500 truncate max-w-32">
        {report.model_id}
      </td>
      <td className="px-4 py-3 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-16 bg-gray-200 rounded-full h-1.5">
            <div
              className={`h-1.5 rounded-full ${report.psi_score > 0.25 ? 'bg-red-500' : report.psi_score > 0.1 ? 'bg-yellow-500' : 'bg-green-500'}`}
              style={{ width: `${Math.min(report.psi_score * 200, 100)}%` }}
            />
          </div>
          <span className="text-xs font-mono">{report.psi_score.toFixed(3)}</span>
        </div>
      </td>
      <td className="px-4 py-3 text-sm font-mono text-gray-600">
        {report.ks_statistic.toFixed(3)}
      </td>
      <td className="px-4 py-3">
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${severityColors[report.severity] || severityColors.none}`}>
          {report.severity}
        </span>
      </td>
      <td className="px-4 py-3 text-xs text-gray-400">
        {new Date(report.created_at).toLocaleString()}
      </td>
    </tr>
  );
}

export default function DriftAlertsPage() {
  const [driftOnly, setDriftOnly] = useState(false);
  const { data, isLoading } = useDriftReports(driftOnly ? { drift_only: 'true' } : undefined);
  const reports = data?.reports || [];

  const driftCount = reports.filter((r) => r.drift_detected).length;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Drift Detection</h2>
          <p className="text-sm text-gray-500 mt-1">
            Monitor data distribution drift across deployed models
          </p>
        </div>
      </div>

      {/* Summary stats */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-sm text-gray-500">Total Reports</p>
          <p className="text-2xl font-bold text-gray-900">{reports.length}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-sm text-gray-500">Drift Detected</p>
          <p className="text-2xl font-bold text-red-600">{driftCount}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-sm text-gray-500">No Drift</p>
          <p className="text-2xl font-bold text-green-600">{reports.length - driftCount}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-sm text-gray-500">Drift Rate</p>
          <p className="text-2xl font-bold text-gray-900">
            {reports.length > 0 ? ((driftCount / reports.length) * 100).toFixed(0) : 0}%
          </p>
        </div>
      </div>

      {/* Filter toggle */}
      <div className="flex gap-2 mb-4">
        <button
          onClick={() => setDriftOnly(false)}
          className={`px-3 py-1 rounded-full text-sm ${!driftOnly ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
        >
          All Reports
        </button>
        <button
          onClick={() => setDriftOnly(true)}
          className={`px-3 py-1 rounded-full text-sm ${driftOnly ? 'bg-red-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
        >
          Drift Only ({driftCount})
        </button>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading drift reports...</div>
      ) : reports.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
          <p className="text-gray-500 mb-2">No drift reports yet.</p>
          <p className="text-sm text-gray-400">
            Set baselines and analyze distributions via the API to detect data drift.
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Feature</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Device</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Model</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">PSI Score</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">KS Stat</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {reports.map((report) => (
                <DriftReportRow key={report.id} report={report} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
