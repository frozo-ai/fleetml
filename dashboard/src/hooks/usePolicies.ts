import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, CreatePolicyRequest } from '../api/client';

export function usePolicies(type?: string) {
  return useQuery({
    queryKey: ['policies', type],
    queryFn: () => api.listPolicies(type),
  });
}

export function usePolicy(id: string) {
  return useQuery({
    queryKey: ['policy', id],
    queryFn: () => api.getPolicy(id),
    enabled: !!id,
  });
}

export function useCreatePolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePolicyRequest) => api.createPolicy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] });
    },
  });
}

export function useDeletePolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deletePolicy(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] });
    },
  });
}
