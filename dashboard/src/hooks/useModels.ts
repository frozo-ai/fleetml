import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useModels() {
  return useQuery({
    queryKey: ['models'],
    queryFn: () => api.listModels(),
  });
}

export function useModel(id: string) {
  return useQuery({
    queryKey: ['model', id],
    queryFn: () => api.getModel(id),
    enabled: !!id,
  });
}
