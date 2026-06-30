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

async function login() {
  const r = await fetch(`${API_PREFIX}/auth/login`, {
    method: 'POST',
    headers: headers(null, true),
    body: JSON.stringify({
      organization_code: 'school-a',
      username: 'hs001',
      password: 'Password123!',
    }),
  });
  if (!r.ok) throw new Error(`login failed: ${r.status}`);
  const json = await r.json();
  return json.data.access_token;
}

async function me(token) {
  const r = await fetch(`${API_PREFIX}/me`, { headers: headers(token) });
  if (!r.ok) throw new Error(`/me failed: ${r.status}`);
  const json = await r.json();
  console.log('  actor:', json.data.id, json.data.roles);
}

async function getAttemptItems(token) {
  const r = await fetch(`${API_PREFIX}/attempts/${ATTEMPT_ID}`, { headers: headers(token) });
  if (!r.ok) throw new Error(`get attempt failed: ${r.status}`);
  const json = await r.json();
  const items = json.data.items || [];
  console.log('  attempt status:', json.data.status, '| items:', items.length);
  if (items.length === 0) throw new Error('attempt has no items');
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

  console.log('Smoke passed.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
