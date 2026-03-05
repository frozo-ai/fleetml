import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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

export function useDeviceLogs(id: string, params?: { limit?: number; level?: string }) {
  return useQuery({
    queryKey: ['device-logs', id, params],
    queryFn: () => api.getDeviceLogs(id, params),
    enabled: !!id,
  });
}

export function useUpdateDevice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { labels?: Record<string, string>; fleet_id?: string } }) =>
      api.updateDevice(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.invalidateQueries({ queryKey: ['device'] });
    },
  });
}

export function useDeleteDevice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteDevice(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
}
