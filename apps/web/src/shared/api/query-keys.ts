export const attemptKeys = {
  all: ['attempts'] as const,
  assigned: () => [...attemptKeys.all, 'assigned'] as const,
  history: () => [...attemptKeys.all, 'history'] as const,
  detail: (attemptId: string) => [...attemptKeys.all, 'detail', attemptId] as const,
  result: (attemptId: string) => [...attemptKeys.all, 'result', attemptId] as const,
};

export const assessmentKeys = {
  all: ['assessments'] as const,
  list: (opts: { q?: string; limit?: number } = {}) =>
    [...assessmentKeys.all, 'list', opts.q ?? '', opts.limit ?? 10] as const,
  infinite: (q: string) =>
    [...assessmentKeys.all, 'infinite', q] as const,
  detail: (id: string) => [...assessmentKeys.all, 'detail', id] as const,
  preview: (id: string) => [...assessmentKeys.all, 'preview', id] as const,
  publications: (id: string) =>
    [...assessmentKeys.all, 'publications', id] as const,
  attempts: (id: string) =>
    [...assessmentKeys.all, 'attempts', id] as const,
  results: (id: string) => [...assessmentKeys.all, 'results', id] as const,
  questions: (opts: { q?: string; bank_id?: string } = {}) =>
    [...assessmentKeys.all, 'questions', opts.q ?? '', opts.bank_id ?? ''] as const,
};

export const classKeys = {
  all: ['classes'] as const,
  list: () => classKeys.all,
  detail: (id: string) => [...classKeys.all, 'detail', id] as const,
  teachers: (id: string) => [...classKeys.all, 'teachers', id] as const,
  enrollments: (id: string) => [...classKeys.all, 'enrollments', id] as const,
  gradebook: (id: string) => [...classKeys.all, 'gradebook', id] as const,
};

export const adminKeys = {
  all: ['admin'] as const,
  users: (opts: { q?: string; limit?: number } = {}) =>
    [...adminKeys.all, 'users', opts.q ?? '', opts.limit ?? 10] as const,
  org: () => [...adminKeys.all, 'organization'] as const,
  audit: (opts: { action?: string; limit?: number } = {}) =>
    [...adminKeys.all, 'audit', opts.action ?? '', opts.limit ?? 10] as const,
};

export const academicKeys = {
  all: ['academics'] as const,
  terms: () => [...academicKeys.all, 'terms'] as const,
  subjects: () => [...academicKeys.all, 'subjects'] as const,
  courses: () => [...academicKeys.all, 'courses'] as const,
  classes: () => [...academicKeys.all, 'classes'] as const,
};

export const resourceKeys = {
  all: ['resources'] as const,
  list: (filter?: { contextType?: string; contextID?: string }) =>
    [...resourceKeys.all, 'list', filter ?? {}] as const,
  files: (resourceId: string) =>
    [...resourceKeys.all, 'files', resourceId] as const,
};

export const gradingKeys = {
  all: ['grading'] as const,
  reviewQueue: (assessmentId: string) =>
    [...gradingKeys.all, 'review-queue', assessmentId] as const,
  attemptReview: (attemptId: string) =>
    [...gradingKeys.all, 'attempt-review', attemptId] as const,
};

export const notificationKeys = {
  all: ['notifications'] as const,
  list: (opts: { limit?: number; before?: string } = {}) =>
    [...notificationKeys.all, 'list', opts.limit ?? '', opts.before ?? ''] as const,
  unreadCount: () => [...notificationKeys.all, 'unread-count'] as const,
};
