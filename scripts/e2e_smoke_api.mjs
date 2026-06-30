import { exec } from 'node:child_process';
import { promisify } from 'node:util';

const execAsync = promisify(exec);
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

async function me(token) {
  const r = await fetch(`${API_PREFIX}/me`, { headers: headers(token) });
  if (!r.ok) throw new Error(`/me failed: ${r.status}`);
  const json = await r.json();
  console.log('  actor:', json.data.id, json.data.roles, '| must_change_password:', json.data.must_change_password);
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
  return json.data;
}

async function assertStudentCannotListAssessments(token) {
  const r = await fetch(`${API_PREFIX}/assessments`, { headers: headers(token) });
  if (r.status !== 403) {
    throw new Error(`expected student /assessments to return 403, got ${r.status}`);
  }
  console.log('  student /assessments correctly rejected:', r.status);
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

async function assertAuditLogs(orgID, expectedActions) {
  const { stdout } = await execAsync(
    `docker exec vts-e2e-postgres psql -U postgres -v ON_ERROR_STOP=1 -t -c "` +
      `SELECT action FROM audit_logs WHERE organization_id = '${orgID}' ORDER BY created_at DESC` +
    `"`
  );
  const rows = stdout.split('\n').map((s) => s.trim()).filter(Boolean);
  console.log('  audit log actions:', rows);
  for (const action of expectedActions) {
    if (!rows.includes(action)) {
      throw new Error(`expected audit log action ${action} not found; got ${JSON.stringify(rows)}`);
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

  const users = await listUsers(adminAfter.data.access_token);
  if (users.length < 3) {
    throw new Error(`expected at least 3 seeded users, got ${users.length}`);
  }

  console.log('Checking user search/limit...');
  const searchedUsers = await listUsersSearchAndLimit(adminAfter.data.access_token, 'hs', 1);
  if (searchedUsers.length > 1) {
    throw new Error(`expected at most 1 user with limit=1, got ${searchedUsers.length}`);
  }

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
  await assertAuditLogs(org.id, ['organization.update', 'user.reset_password', 'user.update_roles', 'user.create']);

  await assertNonAdminCannotAccessAdmin(token, 'student');
  await assertNonAdminCannotAccessAdmin(teacherAfter.data.access_token, 'teacher');

  console.log('Smoke passed.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
