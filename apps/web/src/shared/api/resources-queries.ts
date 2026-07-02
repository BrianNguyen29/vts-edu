import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  archiveResource,
  createResource,
  fetchResourceDownload,
  listResourceFiles,
  listResources,
  publishResource,
  uploadResourceFile,
  type CreateResourceRequest,
  type Resource,
  type ResourceFile,
  type ResourceList,
  type ResourceListFilter,
} from './resources';
import { resourceKeys } from './query-keys';

export function useResourcesQuery(filter: ResourceListFilter = {}) {
  return useQuery<ResourceList>({
    queryKey: resourceKeys.list(filter),
    queryFn: () => listResources(filter),
  });
}

export function useResourceFilesQuery(resourceId: string | null) {
  return useQuery<ResourceFile[]>({
    queryKey: resourceKeys.files(resourceId ?? ''),
    queryFn: () => listResourceFiles(resourceId ?? ''),
    enabled: Boolean(resourceId),
  });
}

export function useCreateResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<Resource, Error, CreateResourceRequest>({
    mutationFn: createResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.all });
    },
  });
}

export function usePublishResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<Resource, Error, string>({
    mutationFn: publishResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.all });
    },
  });
}

export function useArchiveResourceMutation() {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: archiveResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.all });
    },
  });
}

export function useUploadResourceFileMutation(resourceId: string) {
  const queryClient = useQueryClient();
  return useMutation<ResourceFile, Error, File>({
    mutationFn: (file) => uploadResourceFile(resourceId, file),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: resourceKeys.all });
    },
  });
}

export function useDownloadResourceMutation() {
  return useMutation<
    void,
    Error,
    { resourceId: string; filename: string; fileId?: string }
  >({
    mutationFn: ({ resourceId, filename, fileId }) =>
      fetchResourceDownload(resourceId, filename, { fileId }),
  });
}
