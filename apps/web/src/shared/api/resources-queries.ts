import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  archiveResource,
  createResource,
  listResources,
  publishResource,
  uploadResourceFile,
  type CreateResourceRequest,
  type Resource,
  type ResourceFile,
  type ResourceList,
} from './resources';
import { resourceKeys } from './query-keys';

export function useResourcesQuery() {
  return useQuery<ResourceList>({
    queryKey: resourceKeys.list(),
    queryFn: listResources,
  });
}

export function useCreateResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<Resource, Error, CreateResourceRequest>({
    mutationFn: createResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.list() });
    },
  });
}

export function usePublishResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<Resource, Error, string>({
    mutationFn: publishResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.list() });
    },
  });
}

export function useArchiveResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: archiveResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.list() });
    },
  });
}

export function useUploadResourceFileMutation(resourceId: string) {
  const queryClient = useQueryClient();
  return useMutation<ResourceFile, Error, File>({
    mutationFn: (file) => uploadResourceFile(resourceId, file),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.list() });
    },
  });
}
