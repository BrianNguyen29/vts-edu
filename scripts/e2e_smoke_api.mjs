const API_BASE = process.env.API_BASE || 'http://localhost:8080';
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

async function duplicateSection(token, assessmentID, sectionID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/sections/${sectionID}/duplicate`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`duplicate section failed: ${r.status}`);
  const json = await r.json();
  console.log('  duplicated section:', json.data.title, json.data.id);
  return json.data;
}

async function duplicateItem(token, sectionID, itemID) {
  const r = await fetch(`${API_PREFIX}/assessment-sections/${sectionID}/items/${itemID}/duplicate`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`duplicate item failed: ${r.status}`);
  const json = await r.json();
  console.log('  duplicated item:', json.data.question_version_id, json.data.id);
  return json.data;
}

async function previewAssessment(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/preview`, {
    headers: headers(token),
  });
  if (!r.ok) throw new Error(`preview assessment failed: ${r.status}`);
  const json = await r.json();
  console.log('  preview sections:', json.data.sections.length);
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

async function updateSection(token, sectionID, payload) {
  const r = await fetch(`${API_PREFIX}/assessment-sections/${sectionID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(`update section failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated section:', json.data.title);
  return json.data;
}

async function updateItem(token, itemID, payload) {
  const r = await fetch(`${API_PREFIX}/assessment-items/${itemID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(`update item failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated item:', json.data.question_version_id, json.data.points);
  return json.data;
}

async function reorderSections(token, assessmentID, sectionIDs) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/sections/reorder`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ section_ids: sectionIDs }),
  });
  if (!r.ok) throw new Error(`reorder sections failed: ${r.status}`);
}

async function reorderItems(token, sectionID, itemIDs) {
  const r = await fetch(`${API_PREFIX}/assessment-sections/${sectionID}/items/reorder`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ item_ids: itemIDs }),
  });
  if (!r.ok) {
    const text = await r.text();
    throw new Error(`reorder items failed: ${r.status} ${text}`);
  }
}

async function deleteItem(token, itemID) {
  const r = await fetch(`${API_PREFIX}/assessment-items/${itemID}`, {
    method: 'DELETE',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`delete item failed: ${r.status}`);
}

async function deleteTarget(token, assessmentID, targetID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/targets/${targetID}`, {
    method: 'DELETE',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`delete target failed: ${r.status}`);
}

async function listQuestions(token, query = '') {
  const url = new URL(`${API_PREFIX}/questions`);
  if (query) url.searchParams.set('q', query);
  url.searchParams.set('limit', '10');
  const r = await fetch(url, { headers: headers(token) });
  if (!r.ok) throw new Error(`list questions failed: ${r.status}`);
  const json = await r.json();
  console.log('  questions:', json.data.length);
  return json.data;
}

async function listPublications(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/publications`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list publications failed: ${r.status}`);
  const json = await r.json();
  console.log('  publications:', json.data.length);
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

async function importUsers(token, csv, dryRun = false) {
  const r = await fetch(`${API_PREFIX}/users/imports`, {
    method: 'POST',
    headers: headers(token),
    body: JSON.stringify({ csv, dry_run: dryRun }),
  });
  const expectedStatus = dryRun ? 200 : 201;
  if (r.status !== expectedStatus) {
    throw new Error(`import users failed: ${r.status}`);
  }
  const json = await r.json();
  console.log('  import users:', json.data.created, 'created,', json.data.failed, 'failed, dry_run:', json.data.dry_run);
  return json.data;
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

async function updateAcademicTerm(token, termID, name, startDate, endDate) {
  const r = await fetch(`${API_PREFIX}/academic-terms/${termID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify({ name, start_date: startDate, end_date: endDate }),
  });
  if (!r.ok) throw new Error(`update academic term failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated term:', json.data.name);
  return json.data;
}

async function updateSubject(token, subjectID, code, name) {
  const r = await fetch(`${API_PREFIX}/subjects/${subjectID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify({ code, name }),
  });
  if (!r.ok) throw new Error(`update subject failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated subject:', json.data.code, json.data.name);
  return json.data;
}

async function updateCourse(token, courseID, subjectID, termID, code, name) {
  const r = await fetch(`${API_PREFIX}/courses/${courseID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify({ subject_id: subjectID, academic_term_id: termID, code, name }),
  });
  if (!r.ok) throw new Error(`update course failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated course:', json.data.code, json.data.name);
  return json.data;
}

async function updateClass(token, classID, courseID, name) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}`, {
    method: 'PATCH',
    headers: headers(token, true),
    body: JSON.stringify({ course_id: courseID, name }),
  });
  if (!r.ok) throw new Error(`update class failed: ${r.status}`);
  const json = await r.json();
  console.log('  updated class:', json.data.name);
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

async function bulkEnrollStudents(token, classID, userIDs, dryRun = false) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/enrollments/bulk`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ user_ids: userIDs, dry_run: dryRun }),
  });
  const expectedStatus = dryRun ? 200 : 201;
  if (r.status !== expectedStatus) {
    throw new Error(`bulk enroll students failed: ${r.status}`);
  }
  const json = await r.json();
  console.log('  bulk enroll:', json.data.enrolled, 'enrolled,', json.data.failed, 'failed, dry_run:', json.data.dry_run);
  return json.data;
}

async function bulkAssignTeachers(token, classID, items, dryRun = false) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/teachers/bulk`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({ items, dry_run: dryRun }),
  });
  const expectedStatus = dryRun ? 200 : 201;
  if (r.status !== expectedStatus) {
    throw new Error(`bulk assign teachers failed: ${r.status}`);
  }
  const json = await r.json();
  console.log('  bulk assign teachers:', json.data.assigned, 'assigned,', json.data.failed, 'failed, dry_run:', json.data.dry_run);
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

async function exportAuditLogsCSV(token) {
  const r = await fetch(`${API_PREFIX}/audit-logs/export`, { headers: headers(token) });
  if (!r.ok) throw new Error(`export audit logs failed: ${r.status}`);
  const text = await r.text();
  console.log('  audit logs csv rows:', text.split('\n').length);
  return text;
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

async function listAssignedAssessments(token) {
  const r = await fetch(`${API_PREFIX}/me/assessments`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list assigned assessments failed: ${r.status}`);
  const json = await r.json();
  console.log('  assigned assessments:', json.data.length);
  return json.data;
}

async function listMyAttempts(token) {
  const r = await fetch(`${API_PREFIX}/me/attempts`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list my attempts failed: ${r.status}`);
  const json = await r.json();
  console.log('  my attempts:', json.data.length);
  return json.data;
}

async function listAssessmentAttempts(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/attempts`, { headers: headers(token) });
  if (!r.ok) throw new Error(`list assessment attempts failed: ${r.status}`);
  const json = await r.json();
  console.log('  assessment attempts:', json.data.length);
  return json.data;
}

async function getAssessmentResults(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/results`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get assessment results failed: ${r.status}`);
  const json = await r.json();
  console.log('  assessment results:', json.data);
  return json.data;
}

async function exportAssessmentAttemptsCSV(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/attempts/export`, { headers: headers(token) });
  if (!r.ok) throw new Error(`export assessment attempts failed: ${r.status}`);
  const text = await r.text();
  console.log('  assessment attempts csv rows:', text.split('\n').length);
  return text;
}

async function getClassGradebook(token, classID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/gradebook`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get class gradebook failed: ${r.status}`);
  const json = await r.json();
  console.log('  class gradebook entries:', json.data.length);
  return json.data;
}

async function exportClassGradebookCSV(token, classID) {
  const r = await fetch(`${API_PREFIX}/classes/${classID}/gradebook/export`, { headers: headers(token) });
  if (!r.ok) throw new Error(`export class gradebook failed: ${r.status}`);
  const text = await r.text();
  console.log('  class gradebook csv rows:', text.split('\n').length);
  return text;
}

async function assertStudentCannotAccessGradebook(token, assessmentID, classID) {
  const r1 = await fetch(`${API_PREFIX}/assessments/${assessmentID}/attempts`, { headers: headers(token) });
  if (r1.status !== 403) throw new Error(`expected student /assessments/{id}/attempts 403, got ${r1.status}`);
  const r2 = await fetch(`${API_PREFIX}/assessments/${assessmentID}/results`, { headers: headers(token) });
  if (r2.status !== 403) throw new Error(`expected student /assessments/{id}/results 403, got ${r2.status}`);
  const r3 = await fetch(`${API_PREFIX}/classes/${classID}/gradebook`, { headers: headers(token) });
  if (r3.status !== 403) throw new Error(`expected student /classes/{id}/gradebook 403, got ${r3.status}`);
  console.log('  student gradebook endpoints correctly rejected:', r1.status);
}

async function saveAnswerForAttempt(token, attemptId, itemId, selectedOption, payloadOverride) {
  const payload = payloadOverride ?? { selected_option: selectedOption };
  const r = await fetch(`${API_PREFIX}/attempts/${attemptId}/answers/${itemId}`, {
    method: 'PUT',
    headers: headers(token, true),
    body: JSON.stringify({ answer_payload: payload }),
  });
  if (!r.ok) throw new Error(`save answer failed: ${r.status}`);
  const json = await r.json();
  console.log('  answer revision:', json.data.revision);
}

async function submitAttemptById(token, attemptId) {
  const r = await fetch(`${API_PREFIX}/attempts/${attemptId}/submit`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`submit failed: ${r.status}`);
  const json = await r.json();
  console.log('  submit status:', json.data.status);
  console.log('  score:', json.data.score, '/', json.data.max_score, '| grading:', json.data.grading_status);
  if (!['GRADED', 'PENDING_REVIEW', 'NOT_GRADED'].includes(json.data.grading_status)) {
    throw new Error(`unexpected grading_status: ${json.data.grading_status}`);
  }
  return json.data;
}

async function getAttemptResult(token, attemptId) {
  const r = await fetch(`${API_PREFIX}/attempts/${attemptId}/result`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get attempt result failed: ${r.status}`);
  const json = await r.json();
  console.log('  result status:', json.data.status, '| items:', json.data.items.length);
  return json.data;
}

async function startAttempt(token, assessmentID) {
  const r = await fetch(`${API_PREFIX}/assessments/${assessmentID}/attempts`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`start attempt failed: ${r.status}`);
  const json = await r.json();
  console.log('  started attempt:', json.data.id, 'status:', json.data.status, 'items:', json.data.items.length);
  return json.data;
}

async function listResourcesRaw(token) {
  const r = await fetch(`${API_PREFIX}/resources`, { headers: headers(token) });
  if (!r.ok) {
    const body = await r.text();
    throw new Error(`list resources failed: ${r.status} body=${body}`);
  }
  return await r.json();
}

async function listResources(token) {
  return (await listResourcesRaw(token)).data ?? [];
}

async function createResource(token, title, description, contextID) {
  const r = await fetch(`${API_PREFIX}/resources`, {
    method: 'POST',
    headers: headers(token, true),
    body: JSON.stringify({
      title,
      description,
      context_type: 'organization',
      context_id: contextID,
    }),
  });
  if (!r.ok) throw new Error(`create resource failed: ${r.status}`);
  return (await r.json()).data;
}

async function publishResource(token, resourceID) {
  const r = await fetch(`${API_PREFIX}/resources/${resourceID}/publish`, {
    method: 'POST',
    headers: headers(token, true),
  });
  if (!r.ok) throw new Error(`publish resource failed: ${r.status}`);
  return (await r.json()).data;
}

async function uploadResourceFile(token, resourceID, filename, contentType, payload) {
  const form = new FormData();
  form.append('file', new Blob([payload], { type: contentType }), filename);
  const h = headers(token, true);
  delete h['Content-Type'];
  const r = await fetch(`${API_PREFIX}/resources/${resourceID}/files`, {
    method: 'POST',
    headers: h,
    body: form,
  });
  if (!r.ok) {
    const body = await r.text();
    throw new Error(`upload resource file failed: ${r.status} body=${body}`);
  }
  return (await r.json()).data;
}

async function downloadResourceFile(token, resourceID) {
  const r = await fetch(`${API_PREFIX}/resources/${resourceID}/download`, {
    headers: headers(token),
  });
  if (!r.ok) throw new Error(`download resource file failed: ${r.status}`);
  return r.text();
}

async function assertStudentCannotDownloadDraft(token, resourceID) {
  const r = await fetch(`${API_PREFIX}/resources/${resourceID}/download`, {
    headers: headers(token),
  });
  if (r.status !== 403) {
    throw new Error(`expected 403 for draft download by student, got ${r.status}`);
  }
}

async function assertResourcesFlow(teacherToken, studentToken, orgID) {
  console.log('Checking resources/files MVP...');
  const title = `Tài liệu smoke ${Date.now()}`;
  const resource = await createResource(teacherToken, title, 'smoke description', orgID);
  if (resource.status !== 'DRAFT') {
    throw new Error(`expected new resource in DRAFT, got ${resource.status}`);
  }

  const payload = `hello resources ${Date.now()}`;
  await uploadResourceFile(teacherToken, resource.id, 'hello.txt', 'text/plain', payload);

  await assertStudentCannotDownloadDraft(studentToken, resource.id);

  const published = await publishResource(teacherToken, resource.id);
  if (published.status !== 'PUBLISHED') {
    throw new Error(`expected PUBLISHED, got ${published.status}`);
  }

  const teacherListRaw = await listResourcesRaw(teacherToken);
  const teacherList = teacherListRaw.data ?? [];
  if (!teacherList.find((r) => r && r.data && r.data.id === resource.id)) {
    throw new Error(`teacher should see the new resource; got ${JSON.stringify(teacherList)}`);
  }

  const studentList = await listResources(studentToken);
  const studentResource = studentList.find((r) => r && r.data && r.data.id === resource.id);
  if (!studentResource) {
    throw new Error('student should see published resource');
  }
  if (studentResource.data.status !== 'PUBLISHED') {
    throw new Error(`student should only see PUBLISHED, got ${studentResource.data.status}`);
  }

  const downloaded = await downloadResourceFile(studentToken, resource.id);
  if (downloaded !== payload) {
    throw new Error(`downloaded payload mismatch: got ${JSON.stringify(downloaded)}`);
  }
  console.log('  resources: create, upload, publish, student list, download — ok');
}

async function assertNonMcqFlow(teacherToken, studentToken, classId, studentUserId) {
  console.log('Checking non-MCQ question types...');

  // 1. Create a question bank
  const bankRes = await fetch(`${API_PREFIX}/question-banks`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ title: `Bộ câu hỏi smoke ${Date.now()}` }),
  });
  if (bankRes.status !== 201) {
    throw new Error(`expected 201 on create question bank, got ${bankRes.status}`);
  }
  const bankBody = await bankRes.json();
  const bankId = bankBody.data.id;

  // 2. Create a short_answer question
  const saRes = await fetch(`${API_PREFIX}/question-banks/${bankId}/questions`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({
      question_type: 'short_answer',
      prompt: { text: '3 + 4 bằng mấy?' },
      answer_key: { accepted_answers: ['7', 'bảy'] },
      max_score: '1.00',
    }),
  });
  if (saRes.status !== 201) {
    const errBody = await saRes.text();
    throw new Error(`expected 201 on create short_answer, got ${saRes.status}: ${errBody}`);
  }
  const saBody = await saRes.json();
  if (saBody.data.version?.question_type !== 'short_answer') {
    throw new Error(`expected short_answer version, got ${JSON.stringify(saBody.data.version)}`);
  }

  // 3. Create an essay question
  const essayRes = await fetch(`${API_PREFIX}/question-banks/${bankId}/questions`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({
      question_type: 'essay',
      prompt: { text: 'Trình bày cách giải phương trình bậc nhất.' },
      max_score: '2.00',
    }),
  });
  if (essayRes.status !== 201) {
    throw new Error(`expected 201 on create essay, got ${essayRes.status}`);
  }

  // 4. Verify list questions in bank
  const listRes = await fetch(`${API_PREFIX}/question-banks/${bankId}/questions?limit=10`, {
    headers: headers(teacherToken),
  });
  if (listRes.status !== 200) {
    throw new Error(`expected 200 on list questions, got ${listRes.status}`);
  }
  const listBody = await listRes.json();
  if (listBody.data.length !== 2) {
    throw new Error(`expected 2 questions in bank, got ${listBody.data.length}`);
  }

  // 5. Verify picker exposes question_type
  const pickerRes = await fetch(`${API_PREFIX}/questions?limit=10`, { headers: headers(teacherToken) });
  if (pickerRes.status !== 200) {
    const errBody = await pickerRes.text();
    throw new Error(`expected 200 on picker, got ${pickerRes.status}: ${errBody}`);
  }
  const pickerBody = await pickerRes.json();
  const allPicked = pickerBody.data;
  const hasShortAnswer = allPicked.some((q) => q.question_type === 'short_answer');
  const hasEssay = allPicked.some((q) => q.question_type === 'essay');
  if (!hasShortAnswer || !hasEssay) {
    throw new Error(`expected picker to expose short_answer and essay question types, got ${allPicked.map((q) => q.question_type).join(',')}`);
  }

  // 6. Validation: MCQ missing choices should reject
  const badMCQ = await fetch(`${API_PREFIX}/question-banks/${bankId}/questions`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({
      question_type: 'multiple_choice',
      prompt: { text: 'Câu trắc nghiệm thiếu choices' },
      answer_key: { correct_option: 'A' },
    }),
  });
  if (badMCQ.status !== 400) {
    throw new Error(`expected 400 on MCQ missing choices, got ${badMCQ.status}`);
  }

  // 7. Validation: short_answer missing accepted_answers should reject
  const badSA = await fetch(`${API_PREFIX}/question-banks/${bankId}/questions`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({
      question_type: 'short_answer',
      prompt: { text: 'Câu trả lời ngắn thiếu đáp án' },
      answer_key: {},
    }),
  });
  if (badSA.status !== 400) {
    throw new Error(`expected 400 on short_answer missing accepted_answers, got ${badSA.status}`);
  }

  // 8. Build a small assessment with the new types and submit
  const draft = await createAssessmentForClass(teacherToken, classId, 'Đề thi non-MCQ smoke', 30);
  if (!draft?.id) {
    throw new Error('expected assessment id from createAssessmentForClass');
  }
  const section = await createSection(teacherToken, draft.id, 'Phần A', 1);
  if (!section?.id) {
    throw new Error('expected section id');
  }
  const saQ = allPicked.find((q) => q.question_type === 'short_answer');
  const esQ = allPicked.find((q) => q.question_type === 'essay');
  if (!saQ || !esQ) {
    throw new Error('expected both types in picker');
  }
  await createItem(teacherToken, section.id, saQ.question_version_id, 1, '1.00');
  await createItem(teacherToken, section.id, esQ.question_version_id, 2, '2.00');
  const targetRes = await fetch(`${API_PREFIX}/assessments/${draft.id}/targets`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ class_section_id: classId }),
  });
  if (targetRes.status !== 201) {
    throw new Error(`expected 201 on target create, got ${targetRes.status}`);
  }
  const publish = await publishAssessment(teacherToken, draft.id);
  if (!['OPEN', 'PUBLISHED', 'SCHEDULED'].includes(publish.status)) {
    throw new Error(`expected published assessment, got ${publish.status}`);
  }

  // 9. Student starts attempt and submits
  const started = await startAttempt(studentToken, draft.id);
  if (started.status !== 'IN_PROGRESS') {
    throw new Error(`expected IN_PROGRESS, got ${started.status}`);
  }
  for (const item of started.items) {
    let payload;
    if (item.question_type === 'short_answer') {
      payload = { text: '7' };
    } else if (item.question_type === 'essay') {
      payload = { text: 'Trừ hai vế cho cùng một số để cô lập ẩn.' };
    } else {
      payload = { selected_option: 'A' };
    }
    await saveAnswerForAttempt(studentToken, started.id, item.id, 'A', payload);
  }
  const submitted = await submitAttemptById(studentToken, started.id);
  if (submitted.grading_status !== 'PENDING_REVIEW') {
    throw new Error(`expected PENDING_REVIEW (essay triggers it), got ${submitted.grading_status}`);
  }
  if (submitted.max_score !== '3.00') {
    throw new Error(`expected max_score 3.00, got ${submitted.max_score}`);
  }

  // 10. Verify result review shows pending and per-item type
  const result = await getAttemptResult(studentToken, started.id);
  if (result.grading_status !== 'PENDING_REVIEW') {
    throw new Error(`expected PENDING_REVIEW on result, got ${result.grading_status}`);
  }
  const hasPendingItem = result.items.some((it) => it.grading_status === 'PENDING_REVIEW');
  if (!hasPendingItem) {
    throw new Error('expected at least one PENDING_REVIEW item in result');
  }
  const essayItem = result.items.find((it) => it.question_type === 'essay');
  if (!essayItem || essayItem.is_correct !== undefined) {
    throw new Error('expected essay item with no is_correct');
  }

  console.log('  non-MCQ: bank, MCQ/SA/essay create+validate, picker, mixed-attempt PENDING_REVIEW — ok');
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

  console.log('Checking audit logs CSV export...');
  await exportAuditLogsCSV(adminAfter.data.access_token);

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
  const updatedTerm = await updateAcademicTerm(adminAfter.data.access_token, newTerm.id, 'Học kỳ 1 2026-2027 Updated', '2026-09-01', '2027-01-31');
  if (updatedTerm.name !== 'Học kỳ 1 2026-2027 Updated') {
    throw new Error('term name not updated');
  }

  const newSubject = await createSubject(adminAfter.data.access_token, 'PHY', 'Vật lý');
  const updatedSubject = await updateSubject(adminAfter.data.access_token, newSubject.id, 'PHY-UPD', 'Vật lý Updated');
  if (updatedSubject.code !== 'PHY-UPD' || updatedSubject.name !== 'Vật lý Updated') {
    throw new Error('subject not updated');
  }

  const newCourse = await createCourse(adminAfter.data.access_token, newSubject.id, newTerm.id, 'PHY9-HK1', 'Vật lý 9 - HK1');
  const updatedCourse = await updateCourse(adminAfter.data.access_token, newCourse.id, newSubject.id, newTerm.id, 'PHY9-HK1-UPD', 'Vật lý 9 - HK1 Updated');
  if (updatedCourse.code !== 'PHY9-HK1-UPD' || updatedCourse.name !== 'Vật lý 9 - HK1 Updated') {
    throw new Error('course not updated');
  }

  const newClass = await createClass(adminAfter.data.access_token, newCourse.id, '9A1');
  const updatedClass = await updateClass(adminAfter.data.access_token, newClass.id, updatedCourse.id, '9A1');
  if (updatedClass.name !== '9A1') {
    throw new Error('class name unexpectedly changed');
  }

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

  console.log('Checking bulk operations...');
  const importDryRun = await importUsers(
    adminAfter.data.access_token,
    'login_name,display_name,email,temporary_password,roles\n' +
      'bulk-student,Bulk Student,bulk-student@example.com,TempPass123!,student\n' +
      'bulk-teacher,Bulk Teacher,bulk-teacher@example.com,TempPass123!,teacher\n' +
      'bulk-invalid,,invalid@example.com,TempPass123!,student\n',
    true,
  );
  if (importDryRun.total !== 3 || importDryRun.failed !== 1) {
    throw new Error(`expected dry-run 3 total / 1 failed, got ${importDryRun.total} / ${importDryRun.failed}`);
  }

  const importConfirm = await importUsers(
    adminAfter.data.access_token,
    'login_name,display_name,email,temporary_password,roles\n' +
      'bulk-student,Bulk Student,bulk-student@example.com,TempPass123!,student\n' +
      'bulk-teacher,Bulk Teacher,bulk-teacher@example.com,TempPass123!,teacher\n',
    false,
  );
  if (importConfirm.created !== 2) {
    throw new Error(`expected 2 imported users, got ${importConfirm.created}`);
  }
  const importedStudent = importConfirm.rows.find((row) => row.login_name === 'bulk-student');
  const importedTeacher = importConfirm.rows.find((row) => row.login_name === 'bulk-teacher');
  if (!importedStudent?.user_id || !importedTeacher?.user_id) {
    throw new Error('missing imported user ids');
  }

  const bulkClass = await createClass(adminAfter.data.access_token, newCourse.id, '9A2');

  const assignDryRun = await bulkAssignTeachers(
    adminAfter.data.access_token,
    bulkClass.id,
    [{ user_id: importedTeacher.user_id, role: 'teacher' }],
    true,
  );
  if (assignDryRun.rows[0]?.status !== 'valid') {
    throw new Error(`expected teacher dry-run valid, got ${assignDryRun.rows[0]?.status}`);
  }

  const assignConfirm = await bulkAssignTeachers(
    adminAfter.data.access_token,
    bulkClass.id,
    [{ user_id: importedTeacher.user_id, role: 'teacher' }],
    false,
  );
  if (assignConfirm.assigned !== 1) {
    throw new Error(`expected 1 teacher assigned, got ${assignConfirm.assigned}`);
  }

  const enrollDryRun = await bulkEnrollStudents(
    adminAfter.data.access_token,
    bulkClass.id,
    [importedStudent.user_id],
    true,
  );
  if (enrollDryRun.rows[0]?.status !== 'valid') {
    throw new Error(`expected student dry-run valid, got ${enrollDryRun.rows[0]?.status}`);
  }

  const enrollConfirm = await bulkEnrollStudents(
    adminAfter.data.access_token,
    bulkClass.id,
    [importedStudent.user_id],
    false,
  );
  if (enrollConfirm.enrolled !== 1) {
    throw new Error(`expected 1 student enrolled, got ${enrollConfirm.enrolled}`);
  }

  const bulkTeachers = await listClassTeachers(adminAfter.data.access_token, bulkClass.id);
  if (!bulkTeachers.some((t) => t.user_id === importedTeacher.user_id)) {
    throw new Error('bulk teacher not assigned');
  }
  const bulkEnrollments = await listEnrollments(adminAfter.data.access_token, bulkClass.id);
  if (!bulkEnrollments.some((e) => e.user_id === importedStudent.user_id)) {
    throw new Error('bulk student not enrolled');
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

  const builderTarget = await createTarget(teacherAfter.data.access_token, draftAssessment.id, seeded8A1.id);

  // Builder upgrades: edit, reorder, delete/re-add, question picker, publications.
  await updateSection(teacherAfter.data.access_token, builderSection.id, { title: 'Phần I - updated' });
  await updateItem(teacherAfter.data.access_token, builderItem.id, { points: '2.00' });
  await reorderSections(teacherAfter.data.access_token, draftAssessment.id, [builderSection.id]);
  await reorderItems(teacherAfter.data.access_token, builderSection.id, [builderItem.id]);

  await deleteItem(teacherAfter.data.access_token, builderItem.id);
  const invalidAfterDelete = await validateAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (invalidAfterDelete.valid) {
    throw new Error('expected validation to fail after deleting item');
  }
  const readdedItem = await createItem(teacherAfter.data.access_token, builderSection.id, FIXED_QUESTION_VERSION_ID, 2, '1.50');

  await deleteTarget(teacherAfter.data.access_token, draftAssessment.id, builderTarget.id);
  const invalidAfterTargetDelete = await validateAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (invalidAfterTargetDelete.valid) {
    throw new Error('expected validation to fail after deleting target');
  }
  await createTarget(teacherAfter.data.access_token, draftAssessment.id, newClass.id);

  // Builder polish: duplicate section/item and preview.
  const duplicatedSection = await duplicateSection(teacherAfter.data.access_token, draftAssessment.id, builderSection.id);
  if (!duplicatedSection.title.includes('(copy)')) {
    throw new Error(`expected duplicated section title to contain (copy), got ${duplicatedSection.title}`);
  }
  if (duplicatedSection.items.length !== 1) {
    throw new Error(`expected 1 item in duplicated section, got ${duplicatedSection.items.length}`);
  }

  const duplicatedItem = await duplicateItem(teacherAfter.data.access_token, builderSection.id, readdedItem.id);
  if (duplicatedItem.question_version_id !== readdedItem.question_version_id) {
    throw new Error('duplicated item question_version_id mismatch');
  }

  await reorderSections(teacherAfter.data.access_token, draftAssessment.id, [builderSection.id, duplicatedSection.id]);

  const preview = await previewAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (preview.sections.length < 2) {
    throw new Error(`expected at least 2 sections in preview, got ${preview.sections.length}`);
  }
  const previewItems = preview.sections.flatMap((s) => s.items || []);
  if (previewItems.length < 3) {
    throw new Error(`expected at least 3 items in preview, got ${previewItems.length}`);
  }
  const previewItem = previewItems[0];
  if (!previewItem.prompt || Object.keys(previewItem.prompt).length === 0) {
    throw new Error('expected preview item to include prompt');
  }
  if (!previewItem.choices || previewItem.choices.length === 0) {
    throw new Error('expected preview item to include choices');
  }
  if ('answer_key' in previewItem) {
    throw new Error('preview item should not expose answer_key');
  }

  const questions = await listQuestions(teacherAfter.data.access_token, 'Giá trị');
  if (questions.length === 0) {
    throw new Error('expected at least one question in picker');
  }

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

  const publications = await listPublications(teacherAfter.data.access_token, draftAssessment.id);
  if (publications.length !== 1) {
    throw new Error(`expected 1 publication, got ${publications.length}`);
  }

  const classAssessments = await listAssessmentsByClass(teacherAfter.data.access_token, newClass.id);
  if (!classAssessments.some((a) => a.id === draftAssessment.id)) {
    throw new Error('published assessment should appear in class assessment list');
  }

  const assessmentDetail = await getAssessment(teacherAfter.data.access_token, draftAssessment.id);
  if (assessmentDetail.sections.length !== 2) {
    throw new Error(`expected 2 sections in assessment detail, got ${assessmentDetail.sections.length}`);
  }
  if (assessmentDetail.targets.length !== 1) {
    throw new Error(`expected 1 target in assessment detail, got ${assessmentDetail.targets.length}`);
  }
  const sectionWithItems = assessmentDetail.sections.find((s) => s.items && s.items.length > 0);
  if (!sectionWithItems) {
    throw new Error('expected at least one section with nested items in assessment detail');
  }
  const detailItem = sectionWithItems.items.find((it) => it.id === readdedItem.id);
  if (!detailItem) {
    throw new Error('re-added item not found in assessment detail');
  }
  if (detailItem.question_version_id !== FIXED_QUESTION_VERSION_ID) {
    throw new Error(`expected item question_version_id ${FIXED_QUESTION_VERSION_ID}, got ${detailItem.question_version_id}`);
  }
  console.log('  assessment detail section has items:', sectionWithItems.items.length);

  console.log('Checking student attempt generation...');
  const assigned = await listAssignedAssessments(token);
  const assignedAssessment = assigned.find((a) => a.id === draftAssessment.id);
  if (!assignedAssessment) {
    throw new Error('student should see assigned published assessment');
  }
  if (assignedAssessment.availability !== 'open') {
    throw new Error(`expected assigned assessment availability open, got ${assignedAssessment.availability}`);
  }
  if (typeof assignedAssessment.attempts_used !== 'number') {
    throw new Error('expected attempts_used to be a number');
  }
  if (assignedAssessment.publication_id !== publications[0].id) {
    throw new Error('assigned assessment publication_id mismatch');
  }

  const generatedAttempt = await startAttempt(token, draftAssessment.id);
  if (generatedAttempt.status !== 'IN_PROGRESS') {
    throw new Error(`expected IN_PROGRESS generated attempt, got ${generatedAttempt.status}`);
  }
  if (generatedAttempt.items.length === 0) {
    throw new Error('generated attempt has no items');
  }
  if (!generatedAttempt.server_time) {
    throw new Error('expected server_time in generated attempt snapshot');
  }
  if (!generatedAttempt.expires_at) {
    throw new Error('expected expires_at in generated attempt snapshot');
  }
  for (const item of generatedAttempt.items) {
    if (!item.prompt || !item.prompt.text) {
      throw new Error(`generated item ${item.id} missing prompt snapshot`);
    }
    if (!Array.isArray(item.choices) || item.choices.length === 0) {
      throw new Error(`generated item ${item.id} missing choices snapshot`);
    }
  }

  const resumedAttempt = await startAttempt(token, draftAssessment.id);
  if (resumedAttempt.id !== generatedAttempt.id) {
    throw new Error('expected resume of existing in-progress attempt');
  }

  console.log('Checking student save/submit/result on generated attempt...');
  for (const item of generatedAttempt.items) {
    await saveAnswerForAttempt(token, generatedAttempt.id, item.id, 'B');
  }
  const generatedResult = await submitAttemptById(token, generatedAttempt.id);
  if (generatedResult.score !== generatedResult.max_score) {
    throw new Error(`expected full score, got ${generatedResult.score}/${generatedResult.max_score}`);
  }
  const generatedReview = await getAttemptResult(token, generatedAttempt.id);
  if (generatedReview.status !== 'SUBMITTED') {
    throw new Error(`expected SUBMITTED result, got ${generatedReview.status}`);
  }
  if (!Array.isArray(generatedReview.items) || generatedReview.items.length === 0) {
    throw new Error('expected result items');
  }
  for (const item of generatedReview.items) {
    if (!item.is_correct) {
      throw new Error(`expected result item ${item.id} to be correct`);
    }
    if (!item.correct_answer || !item.correct_answer.correct_option) {
      throw new Error(`expected correct_answer on result item ${item.id}`);
    }
  }

  const history = await listMyAttempts(token);
  if (!history.some((a) => a.id === generatedAttempt.id)) {
    throw new Error('attempt history should include generated attempt');
  }

  console.log('Checking teacher gradebook...');
  const attempts = await listAssessmentAttempts(teacherAfter.data.access_token, draftAssessment.id);
  if (attempts.length === 0) {
    throw new Error('expected at least one assessment attempt for teacher gradebook');
  }
  const generatedAttemptRow = attempts.find((a) => a.id === generatedAttempt.id);
  if (!generatedAttemptRow) {
    throw new Error('gradebook attempts should include generated attempt');
  }
  if (generatedAttemptRow.student_user_id !== studentActor.id) {
    throw new Error('gradebook attempt student mismatch');
  }

  const results = await getAssessmentResults(teacherAfter.data.access_token, draftAssessment.id);
  if (results.total_attempts < 1) {
    throw new Error(`expected total_attempts >= 1, got ${results.total_attempts}`);
  }
  if (results.submitted_count < 1) {
    throw new Error(`expected submitted_count >= 1, got ${results.submitted_count}`);
  }

  const csvAttempts = await exportAssessmentAttemptsCSV(teacherAfter.data.access_token, draftAssessment.id);
  if (!csvAttempts.includes('attempt_id')) {
    throw new Error('expected CSV header for assessment attempts');
  }

  const classGradebook = await getClassGradebook(teacherAfter.data.access_token, newClass.id);
  if (classGradebook.length === 0) {
    throw new Error('expected non-empty class gradebook');
  }
  const studentGradebookEntry = classGradebook.find((e) => e.student_user_id === studentActor.id && e.assessment_id === draftAssessment.id);
  if (!studentGradebookEntry) {
    throw new Error('class gradebook should include student/assessment entry');
  }
  if (!studentGradebookEntry.attempt_id) {
    throw new Error('class gradebook entry should reference attempt');
  }

  const csvGradebook = await exportClassGradebookCSV(teacherAfter.data.access_token, newClass.id);
  if (!csvGradebook.includes('student_user_id')) {
    throw new Error('expected CSV header for class gradebook');
  }

  await assertStudentCannotAccessGradebook(token, draftAssessment.id, newClass.id);

  await assertResourcesFlow(
    teacherAfter.data.access_token,
    token,
    studentActor.organization_id,
  );

  await assertNonMcqFlow(
    teacherAfter.data.access_token,
    token,
    newClass.id,
    studentActor.id,
  );

  console.log('Checking login lockout...');
  await assertLoginLockout('hs001');

  console.log('Smoke passed.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
