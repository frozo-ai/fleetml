import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, CreateABTestRequest } from '../api/client';

export function useABTests(state?: string) {
  return useQuery({
    queryKey: ['ab-tests', state],
    queryFn: () => api.listABTests(state),
  });
}

export function useABTest(id: string) {
  return useQuery({
    queryKey: ['ab-test', id],
    queryFn: () => api.getABTest(id),
    enabled: !!id,
  });
}

export function useCreateABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateABTestRequest) => api.createABTest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ab-tests'] });
    },
  });
}

export function useStopABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, winner }: { id: string; winner?: string }) => api.stopABTest(id, winner),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ab-tests'] });
    },
  });
}
