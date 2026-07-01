import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import {
  createAssessment,
  getAssessment,
  listAssessments,
  listQuestions,
  previewAssessment,
  type AssessmentDetail,
  type AssessmentPreview,
  type CreateAssessmentRequest,
  type PagedAssessments,
  type PagedQuestions,
} from './assessments';
import { assessmentKeys } from './query-keys';

export function useAssessments(opts: { q?: string; limit?: number } = {}) {
  return useQuery<PagedAssessments, Error>({
    queryKey: assessmentKeys.list(opts),
    queryFn: () => listAssessments(opts),
  });
}

export function useInfiniteAssessments(searchQuery: string) {
  return useInfiniteQuery<PagedAssessments, Error>({
    queryKey: assessmentKeys.infinite(searchQuery),
    queryFn: ({ pageParam }) =>
      listAssessments({
        q: searchQuery || undefined,
        limit: 10,
        cursor: pageParam as string | undefined,
      }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.page?.has_more ? lastPage.page?.next_cursor : undefined,
  });
}

export function useAssessment(id: string | undefined) {
  return useQuery<AssessmentDetail, Error>({
    queryKey: id ? assessmentKeys.detail(id) : assessmentKeys.all,
    queryFn: () => getAssessment(id!),
    enabled: !!id,
  });
}

export function useAssessmentPreview(id: string | undefined) {
  return useQuery<AssessmentPreview, Error>({
    queryKey: id ? assessmentKeys.preview(id) : assessmentKeys.all,
    queryFn: () => previewAssessment(id!),
    enabled: !!id,
  });
}

export function useCreateAssessment() {
  const queryClient = useQueryClient();

  return useMutation<AssessmentDetail, Error, CreateAssessmentRequest & { classSectionId: string }>({
    mutationFn: ({ classSectionId, ...req }) =>
      createAssessment(classSectionId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: assessmentKeys.all });
    },
  });
}

export function useQuestions(opts: { q?: string; bank_id?: string } = {}) {
  return useQuery<PagedQuestions, Error>({
    queryKey: assessmentKeys.questions(opts),
    queryFn: () => listQuestions(opts),
  });
}
