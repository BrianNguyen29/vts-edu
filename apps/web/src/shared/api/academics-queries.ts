import { useQuery } from '@tanstack/react-query';
import {
  listClasses,
  listCourses,
  listSubjects,
  listTerms,
  type ClassSectionList,
  type CourseList,
  type SubjectList,
  type TermList,
} from './academics';
import { academicKeys } from './query-keys';

export function useTerms() {
  return useQuery<TermList, Error>({
    queryKey: academicKeys.terms(),
    queryFn: listTerms,
  });
}

export function useSubjects() {
  return useQuery<SubjectList, Error>({
    queryKey: academicKeys.subjects(),
    queryFn: listSubjects,
  });
}

export function useCourses() {
  return useQuery<CourseList, Error>({
    queryKey: academicKeys.courses(),
    queryFn: listCourses,
  });
}

export function useClasses() {
  return useQuery<ClassSectionList, Error>({
    queryKey: academicKeys.classes(),
    queryFn: listClasses,
  });
}
