import { ApiResponseError, createApiError } from './attempts';
import { getOpenAPIClient } from './openapi-client';
import type { components } from './openapi-schema';

export type Term = components['schemas']['Term']['data'];
export type TermList = components['schemas']['TermList'];
export type CreateTermRequest = components['schemas']['CreateTermRequest'];
export type UpdateTermRequest = components['schemas']['UpdateTermRequest'];

export type Subject = components['schemas']['Subject']['data'];
export type SubjectList = components['schemas']['SubjectList'];
export type CreateSubjectRequest = components['schemas']['CreateSubjectRequest'];
export type UpdateSubjectRequest = components['schemas']['UpdateSubjectRequest'];

export type Course = components['schemas']['Course']['data'];
export type CourseList = components['schemas']['CourseList'];
export type CreateCourseRequest = components['schemas']['CreateCourseRequest'];
export type UpdateCourseRequest = components['schemas']['UpdateCourseRequest'];

export type ClassSection = components['schemas']['ClassSection']['data'];
export type ClassSectionList = components['schemas']['ClassSectionList'];
export type CreateClassRequest = components['schemas']['CreateClassRequest'];
export type UpdateClassRequest = components['schemas']['UpdateClassRequest'];

export type ClassTeacher = components['schemas']['ClassTeacher']['data'];
export type ClassTeacherList = components['schemas']['ClassTeacherList'];
export type AddClassTeacherRequest = components['schemas']['AddClassTeacherRequest'];

export type Enrollment = components['schemas']['Enrollment']['data'];
export type EnrollmentList = components['schemas']['EnrollmentList'];
export type EnrollStudentRequest = components['schemas']['EnrollStudentRequest'];

export async function listTerms(): Promise<TermList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/academic-terms');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function createTerm(body: CreateTermRequest): Promise<Term> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST('/academic-terms', {
    body,
  });
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function updateTerm(
  termId: string,
  body: UpdateTermRequest
): Promise<Term> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.PATCH(
    '/academic-terms/{term_id}',
    { params: { path: { term_id: termId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function archiveTerm(termId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE(
    '/academic-terms/{term_id}',
    { params: { path: { term_id: termId } } }
  );
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function listSubjects(): Promise<SubjectList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/subjects');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function createSubject(body: CreateSubjectRequest): Promise<Subject> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST('/subjects', { body });
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function updateSubject(
  subjectId: string,
  body: UpdateSubjectRequest
): Promise<Subject> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.PATCH(
    '/subjects/{subject_id}',
    { params: { path: { subject_id: subjectId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function archiveSubject(subjectId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE('/subjects/{subject_id}', {
    params: { path: { subject_id: subjectId } },
  });
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function listCourses(): Promise<CourseList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/courses');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function createCourse(body: CreateCourseRequest): Promise<Course> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST('/courses', { body });
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function updateCourse(
  courseId: string,
  body: UpdateCourseRequest
): Promise<Course> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.PATCH(
    '/courses/{course_id}',
    { params: { path: { course_id: courseId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function archiveCourse(courseId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE('/courses/{course_id}', {
    params: { path: { course_id: courseId } },
  });
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function listClasses(): Promise<ClassSectionList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/classes');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function createClass(body: CreateClassRequest): Promise<ClassSection> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST('/classes', { body });
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function updateClass(
  classId: string,
  body: UpdateClassRequest
): Promise<ClassSection> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.PATCH(
    '/classes/{class_id}',
    { params: { path: { class_id: classId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function archiveClass(classId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE('/classes/{class_id}', {
    params: { path: { class_id: classId } },
  });
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function listClassTeachers(
  classId: string
): Promise<ClassTeacherList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET(
    '/classes/{class_id}/teachers',
    { params: { path: { class_id: classId } } }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function addClassTeacher(
  classId: string,
  body: AddClassTeacherRequest
): Promise<ClassTeacher> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST(
    '/classes/{class_id}/teachers',
    { params: { path: { class_id: classId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function removeClassTeacher(
  classId: string,
  userId: string
): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE(
    '/classes/{class_id}/teachers/{user_id}',
    { params: { path: { class_id: classId, user_id: userId } } }
  );
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function listEnrollments(
  classId: string
): Promise<EnrollmentList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET(
    '/classes/{class_id}/enrollments',
    { params: { path: { class_id: classId } } }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function enrollStudent(
  classId: string,
  body: EnrollStudentRequest
): Promise<Enrollment> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST(
    '/classes/{class_id}/enrollments',
    { params: { path: { class_id: classId } }, body }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function unenrollStudent(
  classId: string,
  userId: string
): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE(
    '/classes/{class_id}/enrollments/{user_id}',
    { params: { path: { class_id: classId, user_id: userId } } }
  );
  if (error) {
    throw createApiError(response.status, error);
  }
}

export { ApiResponseError };
