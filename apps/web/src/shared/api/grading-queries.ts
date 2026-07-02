import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import {
  getAttemptForReview,
  gradeAttemptItem,
  listReviewQueue,
  type AttemptGradingContext,
  type GradeItemRequest,
  type GradeItemResponse,
  type ReviewQueueEntry,
} from './grading';
import { attemptKeys, gradingKeys } from './query-keys';

export function useReviewQueue(assessmentId: string | undefined) {
  return useQuery<ReviewQueueEntry[], Error>({
    queryKey: assessmentId
      ? gradingKeys.reviewQueue(assessmentId)
      : gradingKeys.all,
    queryFn: () => listReviewQueue(assessmentId!),
    enabled: !!assessmentId,
  });
}

export function useAttemptForReview(attemptId: string | undefined) {
  return useQuery<AttemptGradingContext, Error>({
    queryKey: attemptId
      ? gradingKeys.attemptReview(attemptId)
      : gradingKeys.all,
    queryFn: () => getAttemptForReview(attemptId!),
    enabled: !!attemptId,
  });
}

export function useGradeAttemptItem(assessmentId?: string) {
  const queryClient = useQueryClient();
  return useMutation<
    GradeItemResponse,
    Error,
    { attemptId: string; itemId: string; payload: GradeItemRequest }
  >({
    mutationFn: ({ attemptId, itemId, payload }) =>
      gradeAttemptItem(attemptId, itemId, payload),
    onSuccess: (_data, vars) => {
      queryClient.invalidateQueries({
        queryKey: gradingKeys.attemptReview(vars.attemptId),
      });
      queryClient.invalidateQueries({
        queryKey: attemptKeys.result(vars.attemptId),
      });
      if (assessmentId) {
        queryClient.invalidateQueries({
          queryKey: gradingKeys.reviewQueue(assessmentId),
        });
        queryClient.invalidateQueries({
          queryKey: attemptKeys.all,
        });
      }
    },
  });
}
