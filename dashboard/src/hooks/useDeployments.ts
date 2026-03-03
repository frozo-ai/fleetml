import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, CreateDeploymentRequest } from '../api/client';

export function useDeployments() {
  return useQuery({
    queryKey: ['deployments'],
    queryFn: () => api.listDeployments(),
  });
}

export function useDeployment(id: string) {
  return useQuery({
    queryKey: ['deployment', id],
    queryFn: () => api.getDeployment(id),
    enabled: !!id,
  });
}

export function useCreateDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateDeploymentRequest) => api.createDeployment(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deployments'] });
    },
  });
}
