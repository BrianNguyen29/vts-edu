import { useMutation, useQuery } from '@tanstack/react-query';
import {
  exportAssessmentAttemptsCSV,
  exportClassGradebookCSV,
  getAssessmentResults,
  getClassGradebook,
  listAssessmentAttempts,
  type AssessmentAttempt,
  type AssessmentResult,
  type ClassGradebookEntry,
} from './gradebook';
import { assessmentKeys, classKeys } from './query-keys';

export function useAssessmentAttempts(assessmentId: string | undefined) {
  return useQuery<AssessmentAttempt[], Error>({
    queryKey: assessmentId
      ? assessmentKeys.attempts(assessmentId)
      : assessmentKeys.all,
    queryFn: () => listAssessmentAttempts(assessmentId!),
    enabled: !!assessmentId,
  });
}

export function useAssessmentResults(assessmentId: string | undefined) {
  return useQuery<AssessmentResult, Error>({
    queryKey: assessmentId
      ? assessmentKeys.results(assessmentId)
      : assessmentKeys.all,
    queryFn: () => getAssessmentResults(assessmentId!),
    enabled: !!assessmentId,
  });
}

export function useClassGradebook(classId: string | undefined) {
  return useQuery<ClassGradebookEntry[], Error>({
    queryKey: classId ? classKeys.gradebook(classId) : classKeys.all,
    queryFn: () => getClassGradebook(classId!),
    enabled: !!classId,
  });
}

export function useExportAssessmentAttemptsCSV() {
  return useMutation<void, Error, string>({
    mutationFn: exportAssessmentAttemptsCSV,
  });
}

export function useExportClassGradebookCSV() {
  return useMutation<void, Error, string>({
    mutationFn: exportClassGradebookCSV,
  });
}
