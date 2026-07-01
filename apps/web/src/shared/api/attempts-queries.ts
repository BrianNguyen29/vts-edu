import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import {
  getAttempt,
  getAttemptResult,
  listAssignedAssessments,
  listAttemptHistory,
  saveAnswer,
  startAttempt,
  submitAttempt,
  type AnswerSaved,
  type AttemptResult,
  type AttemptSnapshot,
  type AssignedAssessment,
  type PagedList,
  type StudentAttempt,
} from './attempts';
import { attemptKeys } from './query-keys';

export function useAssignedAssessments() {
  return useQuery<AssignedAssessment[], Error>({
    queryKey: attemptKeys.assigned(),
    queryFn: listAssignedAssessments,
  });
}

export function useAttemptHistory() {
  return useInfiniteQuery<PagedList<StudentAttempt>, Error>({
    queryKey: attemptKeys.history(),
    queryFn: ({ pageParam }) =>
      listAttemptHistory({ limit: 10, cursor: pageParam as string | undefined }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.page?.next_cursor ?? undefined,
  });
}

export function useAttempt(attemptId: string | undefined) {
  return useQuery<AttemptSnapshot, Error>({
    queryKey: attemptId ? attemptKeys.detail(attemptId) : attemptKeys.all,
    queryFn: () => getAttempt(attemptId!),
    enabled: !!attemptId,
  });
}

export function useAttemptResult(attemptId: string | undefined) {
  return useQuery<AttemptResult, Error>({
    queryKey: attemptId ? attemptKeys.result(attemptId) : attemptKeys.all,
    queryFn: () => getAttemptResult(attemptId!),
    enabled: !!attemptId,
  });
}

export function useStartAttempt() {
  const queryClient = useQueryClient();

  return useMutation<AttemptSnapshot, Error, string>({
    mutationFn: startAttempt,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: attemptKeys.history() });
    },
  });
}

export function useSaveAnswer() {
  return useMutation<
    AnswerSaved,
    Error,
    { attemptId: string; itemId: string; answerPayload: unknown }
  >({
    mutationFn: ({ attemptId, itemId, answerPayload }) =>
      saveAnswer(attemptId, itemId, answerPayload),
  });
}

export function useSubmitAttempt() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: submitAttempt,
    onSuccess: (_data, attemptId) => {
      queryClient.invalidateQueries({
        queryKey: attemptKeys.detail(attemptId),
      });
      queryClient.invalidateQueries({ queryKey: attemptKeys.history() });
      queryClient.invalidateQueries({
        queryKey: attemptKeys.result(attemptId),
      });
    },
  });
}
