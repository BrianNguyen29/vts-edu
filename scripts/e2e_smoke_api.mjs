const API_BASE = 'http://localhost:8080';
const API_PREFIX = `${API_BASE}/api/v1`;
const ATTEMPT_ID = '00000000-0000-4000-8000-000000000001';
const DEMO_CSRF = 'demo-csrf-token';

function headers(token, withCsrf = false) {
  const h = { 'Content-Type': 'application/json' };
  if (token) h['Authorization'] = `Bearer ${token}`;
  if (withCsrf) {
    h['X-CSRF-Token'] = DEMO_CSRF;
    h['Cookie'] = `vts_csrf=${DEMO_CSRF}`;
  }
  return h;
}

async function ready() {
  for (let i = 0; i < 30; i++) {
    try {
      const r = await fetch(`${API_BASE}/readyz`);
      if (r.ok) return;
    } catch {}
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error('API /readyz did not become ready in time');
}

async function loginWith(username, password) {
  const r = await fetch(`${API_PREFIX}/auth/login`, {
    method: 'POST',
    headers: headers(null, true),
    body: JSON.stringify({
      organization_code: 'school-a',
      username,
      password,
    }),
  });
  if (!r.ok) throw new Error(`login ${username} failed: ${r.status}`);
  return await r.json();
}

async function login() {
  const json = await loginWith('hs001', 'Password123!');
  return json.data.access_token;
}

async function assertLoginLockout(username) {
  for (let i = 0; i < 5; i++) {
    const r = await fetch(`${API_PREFIX}/auth/login`, {
      method: 'POST',
      headers: headers(null, true),
      body: JSON.stringify({
        organization_code: 'school-a',
        username,
        password: 'WrongPassword123!',
      }),
    });
    if (r.status !== 401) {
      throw new Error(`expected failed login to return 401, got ${r.status}`);
    }
  }
  const r = await fetch(`${API_PREFIX}/auth/login`, {
    method: 'POST',
    headers: headers(null, true),
    body: JSON.stringify({
      organization_code: 'school-a',
      username,
      password: 'Password123!',
    }),
  });
  if (r.status !== 429) {
    throw new Error(`expected locked account login to return 429, got ${r.status}`);
  }
  console.log('  login lockout returned 429 after 5 failures');
}

async function me(token) {
  const json = await meData(token);
  console.log('  actor:', json.id, json.roles, '| must_change_password:', json.must_change_password);
  return json;
}

async function meData(token) {
  const r = await fetch(`${API_PREFIX}/me`, { headers: headers(token) });
  if (!r.ok) throw new Error(`/me failed: ${r.status}`);
  const json = await r.json();
  return json.data;
}

async function assertRoleLogin(username, expectedRole) {
  const json = await loginWith(username, 'Password123!');
  const data = json.data;
  console.log(`  ${username} roles:`, data.roles, '| must_change_password:', data.must_change_password);
  if (!data.roles.includes(expectedRole)) {
    throw new Error(`expected ${username} to have role ${expectedRole}, got ${data.roles}`);
  }
  if (!data.must_change_password) {
    throw new Error(`expected ${username} to require password change on first login`);
  }
}

async function changePassword(token, currentPassword, newPassword) {
  const r = await fetch(`${API_PREFIX}/auth/change-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
  if (!r.ok) throw new Error(`change password failed: ${r.status}`);
  const json = await r.json();
  console.log('  change password success:', json.data.success);
}

async function assertChangePasswordRejected(token, currentPassword, newPassword) {
  const r = await fetch(`${API_PREFIX}/auth/change-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
  if (r.status !== 400) {
    throw new Error(`expected weak password change to return 400, got ${r.status}`);
  }
  console.log('  change-password weak rejected:', r.status);
}

async function assertChangePasswordReused(token, currentPassword, oldPassword) {
  const r = await fetch(`${API_PREFIX}/auth/change-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ current_password: currentPassword, new_password: oldPassword }),
  });
  if (r.status !== 400) {
    throw new Error(`expected reused password change to return 400, got ${r.status}`);
  }
  console.log('  change-password reused rejected:', r.status);
}

async function listAssessments(token) {
  const r = await fetch(`${API_PREFIX}/assessments`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list assessments failed: ${r.status}`);
  const json = await r.json();
  console.log('  assessments:', json.data.length);
  if (!Array.isArray(json.data) || json.data.length === 0) {
    throw new Error('expected non-empty assessment list for teacher');
  }
  return json.data;
}

async function listAssessmentsSearch(token, query, limit) {
  const url = new URL(`${API_PREFIX}/assessments`);
  url.searchParams.set('q', query);
  url.searchParams.set('limit', String(limit));
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`list assessments search failed: ${r.status}`);
  const json = await r.json();
  console.log('  assessments search:', json.data.length, '| page:', json.page);
  if (!Array.isArray(json.data)) {
    throw new Error('expected data array in paginated assessment response');
  }
  if (!json.page || json.page.limit !== limit) {
    throw new Error(`expected page.limit=${limit}, got ${JSON.stringify(json.page)}`);
  }
  if (json.page.has_more !== false) {
    throw new Error(`expected has_more=false for single assessment, got ${JSON.stringify(json.page)}`);
  }
  return json.data;
}

async function assertStudentCannotListAssessments(token) {
  const r = await fetch(`${API_PREFIX}/assessments`, { headers: headers(token) });
  if (r.status !== 403) {
    throw new Error(`expected student /assessments to return 403, got ${r.status}`);
  }
  console.log('  student /assessments correctly rejected:', r.status);
}

async function createAssessmentForClass(token, classID, title, durationMinutes) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/assessments`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ title, duration_minutes: durationMinutes, max_attempts: 1 }),
  });
  if (!r.ok) throw new Error(`create assessment failed: ${r.status}`);
  const json = await r.json();
  console.log('  created assessment:', json.data.title, json.data.id, 'status:', json.data.status);
  return json.data;
}

async function listAssessmentsByClass(token, classID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/assessments`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list assessments by class failed: ${r.status}`);
  const json = await r.json();
  console.log('  assessments by class:', json.data.length);
  return json.data;
}

async function getAssessment(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get assessment failed: ${r.status}`);
  const json = await r.json();
  return json.data;
}

async function updateAssessment(token, assessmentID, payload) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(`update assessment failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated assessment:', json.data.title);
  return json.data;
}

async function createSection(token, assessmentID, title, position) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/sections`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ title, position }),
  });
  if (!r.ok) throw new Error(`create section failed: ${r.status}`);
  const json = await r.json();
  console.log('  created section:', json.data.title, json.data.id);
  return json.data;
}

async function createItem(token, sectionID, questionVersionID, position, points = '1.00') {
  const r = await fetch(`${API_PREFIX}/assessment-sections/${sectionID}/items`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ question_version_id: questionVersionID, position, points }),
  });
  if (!r.ok) throw new Error(`create item failed: ${r.status}`);
  const json = await r.json();
  console.log('  created item:', json.data.question_version_id, json.data.id);
  return json.data;
}

async function createTarget(token, assessmentID, classSectionID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/targets`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ class_section_id: classSectionID }),
  });
  if (!r.ok) throw new Error(`create target failed: ${r.status}`);
  const json = await r.json();
  console.log('  created target:', json.data.class_section_id);
  return json.data;
}

async function validateAssessment(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/validate`, {
    method: 'POST',
    headers: headers(token),
  });
  if (!r.ok) throw new Error(`validate assessment failed: ${r.status}`);
  const json = await r.json();
  console.log('  validation:', json.data.valid, json.data.errors?.length ? `errors: ${json.data.errors.length}` : '');
  return json.data;
}

async function publishAssessment(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/publish`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`publish assessment failed: ${r.status}`);
  const json = await r.json();
  console.log('  published assessment:', json.data.status, 'revision:', json.data.revision);
  return json.data;
}

async function listUsers(token) {
  const r = await fetch(`${API_PREFIX}/users`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list users failed: ${r.status}`);
  const json = await r.json();
  console.log('  users:', json.data.length);
  return json.data;
}

async function listUsersSearchAndLimit(token, query, limit) {
  const url = new URL(`${API_PREFIX}/users`);
  url.searchParams.set('q', query);
  url.searchParams.set('limit', String(limit));
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`list users search failed: ${r.status}`);
  const json = await r.json();
  console.log('  users search:', json.data.length, '| page:', json.page);
  if (!Array.isArray(json.data)) {
    throw new Error('expected data array in paginated user response');
  }
  if (!json.page || json.page.limit !== limit) {
    throw new Error(`expected page.limit=${limit}, got ${JSON.stringify(json.page)}`);
  }
  return json.data;
}

async function fetchUsersPage(token, opts = {}) {
  const url = new URL(`${API_PREFIX}/users`);
  if (opts.q) url.searchParams.set('q', opts.q);
  if (opts.limit) url.searchParams.set('limit', String(opts.limit));
  if (opts.cursor) url.searchParams.set('cursor', opts.cursor);
  if (opts.count) url.searchParams.set('count', 'true');
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`fetch users page failed: ${r.status}`);
  return await r.json();
}

async function assertUsersCursorPagination(token, limit) {
  const first = await fetchUsersPage(token, { limit });
  console.log('  users cursor page 1:', first.data.length, '| page:', first.page);
  if (!Array.isArray(first.data) || first.data.length !== limit) {
    throw new Error(`expected ${limit} users on first cursor page, got ${first.data?.length}`);
  }
  if (!first.page || !first.page.has_more || !first.page.next_cursor) {
    throw new Error(`expected first users page to have has_more and next_cursor, got ${JSON.stringify(first.page)}`);
  }

  const second = await fetchUsersPage(token, { limit, cursor: first.page.next_cursor });
  console.log('  users cursor page 2:', second.data.length, '| page:', second.page);
  if (!Array.isArray(second.data) || second.data.length === 0) {
    throw new Error('expected non-empty second users cursor page');
  }
  if (!second.page || second.page.limit !== limit) {
    throw new Error(`expected second users page metadata, got ${JSON.stringify(second.page)}`);
  }
}

async function assertUsersCount(token) {
  const json = await fetchUsersPage(token, { limit: 1, count: true });
  console.log('  users count page:', json.data.length, '| page:', json.page);
  if (typeof json.page?.total_count !== 'number' || json.page.total_count < 3) {
    throw new Error(`expected total_count >= 3, got ${JSON.stringify(json.page)}`);
  }
}

async function createUser(token, loginName, displayName, roles, temporaryPassword) {
  const r = await fetch(`${API_PREFIX}/users`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ login_name: loginName, display_name: displayName, roles, temporary_password: temporaryPassword }),
  });
  if (!r.ok) throw new Error(`create user failed: ${r.status}`);
  const json = await r.json();
  console.log('  created user:', json.data.login_name, json.data.id);
  return json.data;
}

async function assertCreateUserRejected(token, loginName, displayName, roles, temporaryPassword) {
  const r = await fetch(`${API_PREFIX}/users`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ login_name: loginName, display_name: displayName, roles, temporary_password: temporaryPassword }),
  });
  if (r.status !== 400) {
    throw new Error(`expected weak password create user to return 400, got ${r.status}`);
  }
  console.log('  create-user weak rejected:', r.status);
}

async function updateUserRoles(token, userID, roles) {
  const r = await fetch(`${API_PREFIX}/users/${userID}/roles`, {
    method: 'PUT',
    headers: headers(token, true),
    body: JSON.stringify({ roles }),
  });
  if (!r.ok) throw new Error(`update roles failed: ${r.status}`);
  console.log('  updated roles for user:', userID);
}

async function resetUserPassword(token, userID, temporaryPassword) {
  const r = await fetch(`${API_PREFIX}/users/${userID}/reset-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ temporary_password: temporaryPassword }),
  });
  if (!r.ok) throw new Error(`reset password failed: ${r.status}`);
  console.log('  reset password for user:', userID);
}

async function assertResetPasswordRejected(token, userID, temporaryPassword) {
  const r = await fetch(`${API_PREFIX}/users/${userID}/reset-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ temporary_password: temporaryPassword }),
  });
  if (r.status !== 400) {
    throw new Error(`expected weak password reset to return 400, got ${r.status}`);
  }
  console.log('  reset-password weak rejected:', r.status);
}

async function assertResetPasswordReused(token, userID, temporaryPassword) {
  const r = await fetch(`${API_PREFIX}/users/${userID}/reset-password`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ temporary_password: temporaryPassword }),
  });
  if (r.status !== 400) {
    throw new Error(`expected reused password reset to return 400, got ${r.status}`);
  }
  console.log('  reset-password reused rejected:', r.status);
}

async function getCurrentOrg(token) {
  const r = await fetch(`${API_PREFIX}/organizations/current`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get current org failed: ${r.status}`);
  const json = await r.json();
  console.log('  current org:', json.data.code, json.data.name);
  return json.data;
}

async function updateCurrentOrg(token, name) {
  const r = await fetch(`${API_PREFIX}/organizations/current`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify({ name }),
  });
  if (!r.ok) throw new Error(`update current org failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated org name:', json.data.name);
}

async function assertNonAdminCannotAccessAdmin(token, label) {
  const r = await fetch(`${API_PREFIX}/users`, { headers: headers(token) });
  if (r.status !== 403) {
    throw new Error(`expected ${label} /users to return 403, got ${r.status}`);
  }
  console.log(`  ${label} /users correctly rejected:`, r.status);
}

async function listAcademicTerms(token) {
  const r = await fetch(`${API_PREFIX}/academic-terms`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list academic terms failed: ${r.status}`);
  const json = await r.json();
  console.log('  academic terms:', json.data.length);
  return json.data;
}

async function createAcademicTerm(token, name, startDate, endDate) {
  const r = await fetch(`${API_PREFIX}/academic-terms`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ name, start_date: startDate, end_date: endDate }),
  });
  if (!r.ok) throw new Error(`create academic term failed: ${r.status}`);
  const json = await r.json();
  console.log('  created term:', json.data.name, json.data.id);
  return json.data;
}

async function listSubjects(token) {
  const r = await fetch(`${API_PREFIX}/subjects`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list subjects failed: ${r.status}`);
  const json = await r.json();
  console.log('  subjects:', json.data.length);
  return json.data;
}

async function createSubject(token, code, name) {
  const r = await fetch(`${API_PREFIX}/subjects`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ code, name }),
  });
  if (!r.ok) throw new Error(`create subject failed: ${r.status}`);
  const json = await r.json();
  console.log('  created subject:', json.data.code, json.data.id);
  return json.data;
}

async function listCourses(token) {
  const r = await fetch(`${API_PREFIX}/courses`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list courses failed: ${r.status}`);
  const json = await r.json();
  console.log('  courses:', json.data.length);
  return json.data;
}

async function createCourse(token, subjectID, termID, code, name) {
  const r = await fetch(`${API_PREFIX}/courses`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ subject_id: subjectID, academic_term_id: termID, code, name }),
  });
  if (!r.ok) throw new Error(`create course failed: ${r.status}`);
  const json = await r.json();
  console.log('  created course:', json.data.code, json.data.id);
  return json.data;
}

async function listClasses(token) {
  const r = await fetch(`${API_PREFIX}/classes`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list classes failed: ${r.status}`);
  const json = await r.json();
  console.log('  classes:', json.data.length);
  return json.data;
}

async function listMyTeachingClasses(token) {
  const r = await fetch(`${API_PREFIX}/me/teaching/classes`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list my teaching classes failed: ${r.status}`);
  const json = await r.json();
  console.log('  my teaching classes:', json.data.length);
  return json.data;
}

async function createClass(token, courseID, name) {
  const r = await fetch(`${API_PREFIX}/classes`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ course_id: courseID, name }),
  });
  if (!r.ok) throw new Error(`create class failed: ${r.status}`);
  const json = await r.json();
  console.log('  created class:', json.data.name, json.data.id);
  return json.data;
}

async function listClassTeachers(token, classID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/teachers`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list class teachers failed: ${r.status}`);
  const json = await r.json();
  console.log('  class teachers:', json.data.length);
  return json.data;
}

async function addClassTeacher(token, classID, userID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/teachers`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ user_id: userID }),
  });
  if (!r.ok) throw new Error(`add class teacher failed: ${r.status}`);
  const json = await r.json();
  console.log('  added teacher:', json.data.user_id);
  return json.data;
}

async function listEnrollments(token, classID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/enrollments`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list enrollments failed: ${r.status}`);
  const json = await r.json();
  console.log('  enrollments:', json.data.length);
  return json.data;
}

async function enrollStudent(token, classID, userID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/enrollments`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ user_id: userID }),
  });
  if (!r.ok) throw new Error(`enroll student failed: ${r.status}`);
  const json = await r.json();
  console.log('  enrolled student:', json.data.user_id);
  return json.data;
}

async function assertStudentCannotAccessAcademics(token) {
  const r = await fetch(`${API_PREFIX}/classes`, { headers: headers(token) });
  if (r.status !== 403) {
    throw new Error(`expected student /classes to return 403, got ${r.status}`);
  }
  console.log('  student /classes correctly rejected:', r.status);
}

async function listAuditLogs(token, filters = {}) {
  const url = new URL(`${API_PREFIX}/audit-logs`);
  url.searchParams.set('limit', '50');
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== '') {
      url.searchParams.set(key, String(value));
    }
  }
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`list audit logs failed: ${r.status}`);
  const json = await r.json();
  console.log('  audit logs:', json.data.length, '| page:', json.page);
  return json.data;
}

async function fetchAuditLogsPage(token, opts = {}) {
  const url = new URL(`${API_PREFIX}/audit-logs`);
  if (opts.limit) url.searchParams.set('limit', String(opts.limit));
  if (opts.cursor) url.searchParams.set('cursor', opts.cursor);
  if (opts.action) url.searchParams.set('action', opts.action);
  if (opts.count) url.searchParams.set('count', 'true');
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`fetch audit logs page failed: ${r.status}`);
  return await r.json();
}

async function assertAuditLogsCursorPagination(token, limit) {
  const first = await fetchAuditLogsPage(token, { limit });
  console.log('  audit logs cursor page 1:', first.data.length, '| page:', first.page);
  if (!Array.isArray(first.data) || first.data.length !== limit) {
    throw new Error(`expected ${limit} audit logs on first cursor page, got ${first.data?.length}`);
  }
  if (!first.page || !first.page.has_more || !first.page.next_cursor) {
    throw new Error(`expected first audit logs page to have has_more and next_cursor, got ${JSON.stringify(first.page)}`);
  }

  const second = await fetchAuditLogsPage(token, { limit, cursor: first.page.next_cursor });
  console.log('  audit logs cursor page 2:', second.data.length, '| page:', second.page);
  if (!Array.isArray(second.data) || second.data.length === 0) {
    throw new Error('expected non-empty second audit logs cursor page');
  }
  if (!second.page || second.page.limit !== limit) {
    throw new Error(`expected second audit logs page metadata, got ${JSON.stringify(second.page)}`);
  }
}

async function assertAuditLogsCount(token) {
  const json = await fetchAuditLogsPage(token, { limit: 1, count: true });
  console.log('  audit logs count page:', json.data.length, '| page:', json.page);
  if (typeof json.page?.total_count !== 'number' || json.page.total_count < 4) {
    throw new Error(`expected total_count >= 4, got ${JSON.stringify(json.page)}`);
  }
}

async function assertAuditLogs(token, expectedActions) {
  const rows = await listAuditLogs(token);
  const actions = rows.map((row) => row.action);
  console.log('  audit log actions:', actions);
  for (const action of expectedActions) {
    if (!actions.includes(action)) {
      throw new Error(`expected audit log action ${action} not found; got ${JSON.stringify(actions)}`);
    }
  }
}

async function getAttemptItems(token) {
  const r = await fetch(`${API_PREFIX}/attempts/${ATTEMPT_ID}`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get attempt failed: ${r.status}`);
  const json = await r.json();
  const items = json.data.items || [];
  console.log('  attempt status:', json.data.status, '| items:', items.length);
  if (items.length === 0) throw new Error('attempt has no items');
  for (const item of items) {
    if (!item.prompt || !item.prompt.text) {
      throw new Error(`item ${item.id} missing prompt snapshot`);
    }
    if (!Array.isArray(item.choices) || item.choices.length === 0) {
      throw new Error(`item ${item.id} missing choices snapshot`);
    }
  }
  return items;
}

async function saveAnswer(token, itemId, selectedOption) {
  const r = await fetch(`${API_PREFIX}/attempts/${ATTEMPT_ID}/answers/${itemId}`, {
    method: 'PUT',
    headers: headers(token, true),
    body: JSON.stringify({ answer_payload: { selected_option: selectedOption } }),
  });
  if (!r.ok) throw new Error(`save answer failed: ${r.status}`);
  const json = await r.json();
  console.log('  answer revision:', json.data.revision);
}

async function submit(token) {
  const r = await fetch(`${API_PREFIX}/attempts/${ATTEMPT_ID}/submit`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`submit failed: ${r.status}`);
  const json = await r.json();
  console.log('  submit status:', json.data.status);
  console.log('  score:', json.data.score, '/', json.data.max_score, '| grading:', json.data.grading_status);
  if (json.data.grading_status !== 'GRADED') throw new Error(`unexpected grading_status: ${json.data.grading_status}`);
  return json.data;
}

async function main() {
  console.log('Waiting for API...');
  await ready();

  console.log('Logging in...');
  const token = await login();

  console.log('Fetching /me...');
  await me(token);

  console.log('Fetching attempt snapshot...');
  const items = await getAttemptItems(token);

  console.log('Saving answers...');
  // Item 1 correct option is A, item 2 correct option is B.
  await saveAnswer(token, items[0].id, 'A');
  if (items.length > 1) {
    await saveAnswer(token, items[1].id, 'C');
  }

  console.log('Submitting attempt...');
  const result = await submit(token);
  if (result.score !== '1.00') throw new Error(`expected score 1.00, got ${result.score}`);
  if (result.max_score !== '2.00') throw new Error(`expected max_score 2.00, got ${result.max_score}`);

  console.log('Checking seeded role logins...');
  await assertRoleLogin('gv001', 'teacher');
  await assertRoleLogin('admin001', 'admin');

  console.log('Checking forced password change flow...');
  const teacher = await loginWith('gv001', 'Password123!');
  await assertChangePasswordRejected(teacher.data.access_token, 'Password123!', 'password');
  await changePassword(teacher.data.access_token, 'Password123!', 'NewPassword123!');
  const teacherAfter = await loginWith('gv001', 'NewPassword123!');
  if (teacherAfter.data.must_change_password) {
    throw new Error('teacher should not require password change after changing password');
  }
  if (!teacherAfter.data.roles.includes('teacher')) {
    throw new Error(`teacher re-login roles mismatch: ${teacherAfter.data.roles}`);
  }
  await assertChangePasswordReused(teacherAfter.data.access_token, 'NewPassword123!', 'Password123!');

  console.log('Checking teacher assessment list...');
  await listAssessments(teacherAfter.data.access_token);

  console.log('Checking teacher assessment search/limit...');
  const searchedAssessments = await listAssessmentsSearch(teacherAfter.data.access_token, 'Demo', 1);
  if (searchedAssessments.length > 1) {
    throw new Error(`expected at most 1 assessment with limit=1, got ${searchedAssessments.length}`);
  }

  console.log('Checking student cannot list assessments...');
  await assertStudentCannotListAssessments(token);

  console.log('Checking admin user management flow...');
  const admin = await loginWith('admin001', 'Password123!');
  await assertChangePasswordRejected(admin.data.access_token, 'Password123!', '12345678');
  await changePassword(admin.data.access_token, 'Password123!', 'AdminPass123!');
  const adminAfter = await loginWith('admin001', 'AdminPass123!');
  if (!adminAfter.data.roles.includes('admin')) {
    throw new Error(`admin re-login roles mismatch: ${adminAfter.data.roles}`);
  }
  await assertChangePasswordReused(adminAfter.data.access_token, 'AdminPass123!', 'Password123!');

  const users = await listUsers(adminAfter.data.access_token);
  if (users.length < 3) {
    throw new Error(`expected at least 3 seeded users, got ${users.length}`);
  }

  console.log('Checking user search/limit...');
  const searchedUsers = await listUsersSearchAndLimit(adminAfter.data.access_token, 'hs', 1);
  if (searchedUsers.length > 1) {
    throw new Error(`expected at most 1 user with limit=1, got ${searchedUsers.length}`);
  }

  console.log('Checking users cursor pagination/count...');
  await assertUsersCursorPagination(adminAfter.data.access_token, 1);
  await assertUsersCount(adminAfter.data.access_token);

  await assertCreateUserRejected(adminAfter.data.access_token, 'weakuser', 'Weak User', ['student'], 'password');

  const newUser = await createUser(adminAfter.data.access_token, 'testuser', 'Test User', ['student'], 'TempPass123!');
  if (!newUser.must_change_password) {
    throw new Error('newly created user must require password change');
  }

  const newUserLogin = await loginWith('testuser', 'TempPass123!');
  if (!newUserLogin.data.must_change_password) {
    throw new Error('login for new user must report must_change_password=true');
  }

  await updateUserRoles(adminAfter.data.access_token, newUser.id, ['student', 'teacher']);
  await assertResetPasswordRejected(adminAfter.data.access_token, newUser.id, 'password123');
  await assertResetPasswordReused(adminAfter.data.access_token, newUser.id, 'TempPass123!');
  await resetUserPassword(adminAfter.data.access_token, newUser.id, 'ResetPass123!');
  const resetLogin = await loginWith('testuser', 'ResetPass123!');
  if (!resetLogin.data.must_change_password) {
    throw new Error('login after reset password must report must_change_password=true');
  }
  if (!resetLogin.data.roles.includes('teacher')) {
    throw new Error(`roles after update mismatch: ${resetLogin.data.roles}`);
  }

  const org = await getCurrentOrg(adminAfter.data.access_token);
  await updateCurrentOrg(adminAfter.data.access_token, 'Trường THPT Demo A Updated');

  console.log('Checking audit logs...');
  await assertAuditLogs(adminAfter.data.access_token, ['organization.update', 'user.reset_password', 'user.update_roles', 'user.create']);

  console.log('Checking audit log action filter...');
  const filtered = await listAuditLogs(adminAfter.data.access_token, { action: 'user.create' });
  if (filtered.length === 0 || filtered.some((row) => row.action !== 'user.create')) {
    throw new Error(`expected only user.create audit rows, got ${JSON.stringify(filtered.map((r) => r.action))}`);
  }

  console.log('Checking audit logs cursor pagination/count...');
  await assertAuditLogsCursorPagination(adminAfter.data.access_token, 1);
  await assertAuditLogsCount(adminAfter.data.access_token);

  await assertNonAdminCannotAccessAdmin(token, 'student');
  await assertNonAdminCannotAccessAdmin(teacherAfter.data.access_token, 'teacher');

  console.log('Checking academics foundation...');
  const studentActor = await meData(token);

  const seededTerms = await listAcademicTerms(adminAfter.data.access_token);
  if (seededTerms.length === 0) throw new Error('expected at least one seeded academic term');

  const seededSubjects = await listSubjects(adminAfter.data.access_token);
  if (seededSubjects.length === 0) throw new Error('expected at least one seeded subject');

  const seededCourses = await listCourses(adminAfter.data.access_token);
  if (seededCourses.length === 0) throw new Error('expected at least one seeded course');

  const seededClasses = await listClasses(adminAfter.data.access_token);
  if (seededClasses.length === 0) throw new Error('expected at least one seeded class');
  const seeded8A1 = seededClasses.find((c) => c.name === '8A1');
  if (!seeded8A1) throw new Error('expected seeded class 8A1');
  if (seeded8A1.student_count !== 1) throw new Error(`expected 1 enrollment in 8A1, got ${seeded8A1.student_count}`);
  if (seeded8A1.teacher_count !== 1) throw new Error(`expected 1 teacher in 8A1, got ${seeded8A1.teacher_count}`);

  await assertStudentCannotAccessAcademics(token);

  const teacherClasses = await listClasses(teacherAfter.data.access_token);
  if (!teacherClasses.some((c) => c.name === '8A1')) {
    throw new Error('teacher should see assigned class 8A1');
  }

  console.log('Checking teacher my-teaching-classes endpoint...');
  const myTeachingClasses = await listMyTeachingClasses(teacherAfter.data.access_token);
  const my8A1 = myTeachingClasses.find((c) => c.name === '8A1');
  if (!my8A1) throw new Error('teacher /me/teaching/classes should include 8A1');
  if (my8A1.student_count !== 1) throw new Error(`expected 1 student in my teaching class 8A1, got ${my8A1.student_count}`);
  if (my8A1.teacher_count !== 1) throw new Error(`expected 1 teacher in my teaching class 8A1, got ${my8A1.teacher_count}`);

  const studentTeachingResp = await fetch(`${API_PREFIX}/me/teaching/classes`, { headers: headers(token) });
  if (studentTeachingResp.status !== 403) throw new Error(`student /me/teaching/classes should be 403, got ${studentTeachingResp.status}`);

  const adminTeachingResp = await fetch(`${API_PREFIX}/me/teaching/classes`, { headers: headers(adminAfter.data.access_token) });
  if (adminTeachingResp.status !== 403) throw new Error(`admin /me/teaching/classes should be 403, got ${adminTeachingResp.status}`);

  const seededTeachers = await listClassTeachers(teacherAfter.data.access_token, seeded8A1.id);
  const teacherUserId = teacherAfter.data.user?.id;
  if (!seededTeachers.some((t) => t.user_id === teacherUserId)) {
    throw new Error('teacher should be listed for class 8A1');
  }

  const seededEnrollments = await listEnrollments(teacherAfter.data.access_token, seeded8A1.id);
  if (!seededEnrollments.some((e) => e.user_id === studentActor.id)) {
    throw new Error('student should be enrolled in class 8A1');
  }

  const newTerm = await createAcademicTerm(adminAfter.data.access_token, 'Học kỳ 1 2026-2027', '2026-09-01', '2027-01-31');
  const newSubject = await createSubject(adminAfter.data.access_token, 'PHY', 'Vật lý');
  const newCourse = await createCourse(adminAfter.data.access_token, newSubject.id, newTerm.id, 'PHY9-HK1', 'Vật lý 9 - HK1');
  const newClass = await createClass(adminAfter.data.access_token, newCourse.id, '9A1');

  await addClassTeacher(adminAfter.data.access_token, newClass.id, teacherUserId);
  await enrollStudent(adminAfter.data.access_token, newClass.id, studentActor.id);

  const teacherClassesAfter = await listClasses(teacherAfter.data.access_token);
  if (!teacherClassesAfter.some((c) => c.name === '9A1')) {
    throw new Error('teacher should see newly assigned class 9A1');
  }

  const newEnrollments = await listEnrollments(teacherAfter.data.access_token, newClass.id);
  if (!newEnrollments.some((e) => e.user_id === studentActor.id)) {
    throw new Error('student should be enrolled in newly created class 9A1');
  }

  console.log('Checking assessment builder...');
  const FIXED_QUESTION_VERSION_ID = '00000000-0000-4000-8000-000000000002';
  const draftAssessment = await createAssessmentForClass(teacherAfter.data.access_token, seeded8A1.id, 'Bài kiểm tra 8A1', 30);
  if (draftAssessment.status !== 'DRAFT') {
    throw new Error(`expected DRAFT status, got ${draftAssessment.status}`);
  }

  const updatedAssessment = await updateAssessment(teacherAfter.data.access_token, draftAssessment.id, { instructions: 'Làm cẩn thận' });
  if (updatedAssessment.instructions !== 'Làm cẩn thận') {
    throw new Error('assessment instructions not updated');
  }

  const builderSection = await createSection(teacherAfter.data.access_token, draftAssessment.id, 'Phần I', 1);
  const builderItem = await createItem(teacherAfter.data.access_token, builderSection.id, FIXED_QUESTION_VERSION_ID, 1, '1.00');
  if (builderItem.question_version_id !== FIXED_QUESTION_VERSION_ID) {
    throw new Error('item question_version_id mismatch');
  }

  await createTarget(teacherAfter.data.access_token, draftAssessment.id, seeded8A1.id);

  const validationBeforePublish = await validateAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (!validationBeforePublish.valid) {
    throw new Error(`expected valid assessment, got errors: ${JSON.stringify(validationBeforePublish.errors)}`);
  }

  const publishedAssessment = await publishAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (publishedAssessment.status !== 'OPEN' && publishedAssessment.status !== 'PUBLISHED') {
    throw new Error(`expected OPEN or PUBLISHED status, got ${publishedAssessment.status}`);
  }
  if (publishedAssessment.revision < 1) {
    throw new Error('expected published revision >= 1');
  }

  const classAssessments = await listAssessmentsByClass(teacherAfter.data.access_token, seeded8A1.id);
  if (!classAssessments.some((a) => a.id === draftAssessment.id)) {
    throw new Error('published assessment should appear in class assessment list');
  }

  const assessmentDetail = await getAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (assessmentDetail.sections.length !== 1) {
    throw new Error(`expected 1 section in assessment detail, got ${assessmentDetail.sections.length}`);
  }
  if (assessmentDetail.targets.length !== 1) {
    throw new Error(`expected 1 target in assessment detail, got ${assessmentDetail.targets.length}`);
  }
  const sectionWithItems = assessmentDetail.sections.find((s) => s.items && s.items.length > 0);
  if (!sectionWithItems) {
    throw new Error('expected at least one section with nested items in assessment detail');
  }
  if (sectionWithItems.items[0].question_version_id !== FIXED_QUESTION_VERSION_ID) {
    throw new Error(`expected item question_version_id ${FIXED_QUESTION_VERSION_ID}, got ${sectionWithItems.items[0].question_version_id}`);
  }
  console.log('  assessment detail section has items:', sectionWithItems.items.length);

  console.log('Checking login lockout...');
  await assertLoginLockout('hs001');

  console.log('Smoke passed.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
