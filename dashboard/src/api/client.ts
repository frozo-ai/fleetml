const API_BASE = import.meta.env.VITE_API_URL || '';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('fleetml_token');

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export const api = {
  // Health
  health: () => request<{ status: string; version: string }>('/api/v1/health'),

  // Auth
  login: (email: string, password: string) =>
    request<{ token: string; user: { id: string; email: string; role: string } }>(
      '/api/v1/auth/login',
      { method: 'POST', body: JSON.stringify({ email, password }) }
    ),

  // Devices
  listDevices: (params?: { status?: string; fleet_id?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return request<{ devices: Device[]; total: number }>(`/api/v1/devices?${query}`);
  },
  getDevice: (id: string) => request<Device>(`/api/v1/devices/${id}`),
  getDeviceLogs: (id: string, params?: { limit?: number; level?: string }) => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.level) query.set('level', params.level);
    return request<{ logs: DeviceLog[] }>(`/api/v1/devices/${id}/logs?${query}`);
  },
  updateDevice: (id: string, data: { labels?: Record<string, string>; fleet_id?: string }) =>
    request<Device>(`/api/v1/devices/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteDevice: (id: string) =>
    request<void>(`/api/v1/devices/${id}`, { method: 'DELETE' }),

  // Models
  listModels: () => request<{ models: Model[]; total: number }>('/api/v1/models'),
  getModel: (id: string) => request<Model>(`/api/v1/models/${id}`),

  // Deployments
  listDeployments: () => request<{ deployments: Deployment[]; total: number }>('/api/v1/deployments'),
  getDeployment: (id: string) => request<Deployment>(`/api/v1/deployments/${id}`),
  createDeployment: (data: CreateDeploymentRequest) =>
    request<Deployment>('/api/v1/deployments', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Fleets
  listFleets: () => request<{ fleets: Fleet[] }>('/api/v1/fleets'),
  getFleetStats: (id: string) => request<FleetStats>(`/api/v1/fleets/${id}/stats`),

  // A/B Tests
  listABTests: (state?: string) => {
    const query = state ? `?state=${state}` : '';
    return request<{ ab_tests: ABTest[]; total: number }>(`/api/v1/ab-tests${query}`);
  },
  getABTest: (id: string) => request<ABTest>(`/api/v1/ab-tests/${id}`),
  createABTest: (data: CreateABTestRequest) =>
    request<ABTest>('/api/v1/ab-tests', { method: 'POST', body: JSON.stringify(data) }),
  stopABTest: (id: string, winner?: string) =>
    request<ABTest>(`/api/v1/ab-tests/${id}/stop`, {
      method: 'POST',
      body: JSON.stringify({ winner: winner || '' }),
    }),

  // Policies
  listPolicies: (type?: string) => {
    const query = type ? `?type=${type}` : '';
    return request<{ policies: Policy[]; total: number }>(`/api/v1/policies${query}`);
  },
  getPolicy: (id: string) => request<Policy>(`/api/v1/policies/${id}`),
  createPolicy: (data: CreatePolicyRequest) =>
    request<Policy>('/api/v1/policies', { method: 'POST', body: JSON.stringify(data) }),
  updatePolicy: (id: string, data: Partial<CreatePolicyRequest>) =>
    request<Policy>(`/api/v1/policies/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deletePolicy: (id: string) =>
    request<void>(`/api/v1/policies/${id}`, { method: 'DELETE' }),

  // Drift Detection
  listDriftReports: (params?: { model_id?: string; drift_only?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return request<{ reports: DriftReport[] }>(`/api/v1/drift/reports?${query}`);
  },

  // Compile
  compileModel: (modelId: string, targetRuntime: string, options?: Record<string, unknown>) =>
    request<CompileResult>(`/api/v1/models/${modelId}/compile`, {
      method: 'POST',
      body: JSON.stringify({ target_runtime: targetRuntime, options: options || {} }),
    }),
};

export interface Device {
  id: string;
  device_id: string;
  name: string;
  status: string;
  arch: string;
  gpu_type: string;
  runtime: string;
  ram_mb: number;
  disk_gb: number;
  hardware_model: string;
  labels: Record<string, string>;
  last_heartbeat: string;
  cpu_percent?: number;
  gpu_percent?: number;
  ram_mb_used?: number;
  disk_percent?: number;
  temperature_c?: number;
}

export interface Model {
  id: string;
  name: string;
  version: string;
  format: string;
  artifact_size: number;
  checksum: string;
  description: string;
  tags: string[];
  created_at: string;
}

export interface Deployment {
  id: string;
  model_id: string;
  state: string;
  target_type: string;
  total_devices: number;
  completed_devices: number;
  failed_devices: number;
  queued_devices: number;
  deployment_policy: string;
  created_at: string;
}

export interface Fleet {
  id: string;
  name: string;
  description: string;
  labels: Record<string, string>;
}

export interface CreateDeploymentRequest {
  model_name: string;
  model_version: string;
  target_type: string;
  target_id: string;
  policy: string;
}

export interface ABTest {
  id: string;
  name: string;
  model_a_id: string;
  model_b_id: string;
  split_a: number;
  split_b: number;
  target_fleet_id?: string;
  metric: string;
  duration?: string;
  auto_promote: boolean;
  state: string;
  winner?: string;
  model_a_metrics?: Record<string, unknown>;
  model_b_metrics?: Record<string, unknown>;
  started_at?: string;
  stopped_at?: string;
  created_at: string;
}

export interface CreateABTestRequest {
  name: string;
  model_a_id: string;
  model_b_id: string;
  split_a: number;
  split_b: number;
  target_fleet_id?: string;
  metric: string;
  duration?: string;
  auto_promote: boolean;
}

export interface Policy {
  id: string;
  name: string;
  description: string;
  policy_type: string;
  rules: Record<string, unknown>;
  enabled: boolean;
  priority: number;
  target_fleet_id?: string;
  target_labels?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface CreatePolicyRequest {
  name: string;
  description: string;
  policy_type: string;
  rules: Record<string, unknown>;
  enabled: boolean;
  priority: number;
  target_fleet_id?: string;
  target_labels?: Record<string, string>;
}

export interface DriftReport {
  id: string;
  device_id: string;
  model_id: string;
  feature_name: string;
  psi_score: number;
  ks_statistic: number;
  ks_p_value: number;
  drift_detected: boolean;
  severity: string;
  created_at: string;
}

export interface CompileResult {
  runtime: string;
  artifact_url: string;
  checksum: string;
  file_size: number;
  compile_time_seconds: number;
  metadata?: Record<string, unknown>;
}

export interface FleetStats {
  total_devices: number;
  online_devices: number;
  offline_devices: number;
  warning_devices: number;
  runtime_counts: Record<string, number>;
  arch_counts: Record<string, number>;
}

export interface DeviceLog {
  timestamp: string;
  level: string;
  message: string;
  source: string;
}
