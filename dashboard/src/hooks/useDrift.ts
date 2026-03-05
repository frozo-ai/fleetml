import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useDriftReports(params?: { model_id?: string; drift_only?: string }) {
  return useQuery({
    queryKey: ['drift-reports', params],
    queryFn: () => api.listDriftReports(params),
  });
}
