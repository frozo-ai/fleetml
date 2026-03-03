import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useDevices(params?: { status?: string; fleet_id?: string }) {
  return useQuery({
    queryKey: ['devices', params],
    queryFn: () => api.listDevices(params),
  });
}

export function useDevice(id: string) {
  return useQuery({
    queryKey: ['device', id],
    queryFn: () => api.getDevice(id),
    enabled: !!id,
  });
}
