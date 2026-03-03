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
