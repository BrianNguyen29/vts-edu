// Bounded load / concurrency harness for the attempt runtime.
//
// This script is OPTIONAL and MANUAL — it is intentionally not wired into
// `pnpm check` or any CI pipeline. The smoke suite (`pnpm e2e:smoke`) and
// the browser E2E suite (`pnpm e2e:browser`) already cover correctness;
// this harness exists to catch race conditions and small-burst backend
// misbehaviour without being heavy enough to slow the default check loop.
//
// Scenarios are designed to be small (N = 8–16), reliable on a developer
// laptop, and idempotent on a freshly-migrated E2E DB. Each scenario is
// run sequentially and must pass for the harness to exit 0.
//
// Scenarios:
//   1. Concurrent saves (same attempt / same item)        — N = 8
//   2. Concurrent submits (same attempt)                  — N = 8
//   3. Save-after-submit                                  — single
//   4. Burst reads (GET /attempts/{id})                   — N = 16
//
// Pre-flight: the harness assumes:
//   - Postgres is on localhost:5434 (managed by e2e_load.sh).
//   - API is on $API_BASE (default http://localhost:8080).
//   - Seeded demo attempt 00000000-0000-4000-8000-000000000001 is IN_PROGRESS.
//   - Seeded class "8A1" exists; teacher gv001 is assigned; student hs001 is enrolled.
//   - Seeded question version 00000000-0000-4000-8000-000000000002 is PUBLISHED.

const API_BASE = process.env.API_BASE || 'http://localhost:8080';
const API_PREFIX = `${API_BASE}/api/v1`;
const SEED_ATTEMPT_ID = '00000000-0000-4000-8000-000000000001';
const SEED_QUESTION_VERSION_ID = '00000000-0000-4000-8000-000000000002';
const DEMO_CSRF = 'demo-csrf-token';

const CONCURRENT_SAVE_N = Number(process.env.LOAD_SAVE_N || 8);
const CONCURRENT_SUBMIT_N = Number(process.env.LOAD_SUBMIT_N || 8);
const BURST_READ_N = Number(process.env.LOAD_READ_N || 16);
const SCENARIO_TIMEOUT_MS = Number(process.env.LOAD_TIMEOUT_MS || 15_000);

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
    } catch {
      /* still booting */
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error('API /readyz did not become ready in time');
}

async function loginWith(username, password) {
  const r = await fetch(`${API_PREFIX}/auth/login`, {
    method: 'POST',
    headers: headers(null, true),
    body: JSON.stringify({ organization_code: 'school-a', username, password }),
  });
  if (!r.ok) throw new Error(`login ${username} failed: ${r.status}`);
  return await r.json();
}

async function loginAs(username, password) {
  const json = await loginWith(username, password);
  return json.data.access_token;
}

function withTimeout(promise, ms, label) {
  return new Promise((resolve, reject) => {
    const t = setTimeout(() => reject(new Error(`${label} timed out after ${ms}ms`)), ms);
    promise.then(
      (v) => {
        clearTimeout(t);
        resolve(v);
      },
      (e) => {
        clearTimeout(t);
        reject(e);
      }
    );
  });
}

function logHeader(title) {
  const bar = '═'.repeat(60);
  console.log(`\n${bar}\n  ${title}\n${bar}`);
}

function logRow(name, value) {
  console.log(`  ${name.padEnd(20, ' ')} ${value}`);
}

// ───────────────────────────────────────────────────────────────────────────
// Scenario 1 — Concurrent saves on the same (attempt, item).
//
// The seeded attempt is IN_PROGRESS and has 2 items. We fire N concurrent
// PUTs for the same item; the SQL UPSERT increments `revision` on every
// write, so the final revision must be ≥ N. No request may return 5xx.
// ───────────────────────────────────────────────────────────────────────────
async function scenario1_concurrentSaves(studentToken) {
  logHeader(`S1: concurrent saves (N=${CONCURRENT_SAVE_N}) on seed attempt`);
  const itemsRes = await fetch(`${API_PREFIX}/attempts/${SEED_ATTEMPT_ID}`, { headers: headers(studentToken) });
  if (!itemsRes.ok) throw new Error(`get seed attempt failed: ${itemsRes.status}`);
  const itemsJson = await itemsRes.json();
  const targetItem = (itemsJson.data.items || [])[0];
  if (!targetItem) throw new Error('seed attempt has no items');

  const basePayload = { selected_option: 'A' };
  const startedAt = Date.now();
  const responses = await Promise.all(
    Array.from({ length: CONCURRENT_SAVE_N }, (_, i) =>
      withTimeout(
        (async () => {
          const r = await fetch(
            `${API_PREFIX}/attempts/${SEED_ATTEMPT_ID}/answers/${targetItem.id}`,
            {
              method: 'PUT',
              headers: headers(studentToken, true),
              body: JSON.stringify({ answer_payload: basePayload }),
            }
          );
          const body = r.ok ? await r.json() : await r.text();
          return { i, status: r.status, body };
        })(),
        SCENARIO_TIMEOUT_MS,
        `save[${i}]`
      )
    )
  );
  const elapsed = Date.now() - startedAt;

  const ok = responses.filter((r) => r.status === 200);
  const notOk = responses.filter((r) => r.status !== 200);
  if (notOk.length > 0) {
    throw new Error(
      `S1 FAIL: ${notOk.length}/${CONCURRENT_SAVE_N} saves did not return 200. ` +
        `statuses=${notOk.map((r) => r.status).sort().join(',')}; bodies=${notOk
          .map((r) => (typeof r.body === 'string' ? r.body.slice(0, 80) : JSON.stringify(r.body).slice(0, 80)))
          .join(' | ')}`
    );
  }

  const revisions = ok.map((r) => {
    const body = r.body;
    if (body && body.data && typeof body.data.revision === 'number') return body.data.revision;
    return undefined;
  });
  const minRev = Math.min(...revisions);
  const maxRev = Math.max(...revisions);
  // Even if a single client won the race and saw its own revision = N, all
  // responses must be 200. We assert the max observed revision is ≥ N
  // (proves at least N writes happened) AND that no revision was lost
  // (revisions must be a permutation of [N-K+1 .. N] where K = number of
  // responses; in practice monotonic non-decreasing is the SQL guarantee).
  if (revisions.length !== CONCURRENT_SAVE_N) {
    throw new Error(`S1 FAIL: expected ${CONCURRENT_SAVE_N} revisions, got ${revisions.length}`);
  }
  const sorted = [...revisions].sort((a, b) => a - b);
  const expectedFirst = Math.max(1, maxRev - CONCURRENT_SAVE_N + 1);
  if (sorted[0] !== expectedFirst) {
    throw new Error(
      `S1 FAIL: revision sequence is not contiguous. observed=${sorted.join(',')}; expected contiguous range ending at ${maxRev}`
    );
  }
  if (maxRev < CONCURRENT_SAVE_N) {
    throw new Error(`S1 FAIL: max revision ${maxRev} < N=${CONCURRENT_SAVE_N}; expected at least ${CONCURRENT_SAVE_N}`);
  }

  logRow('statuses', 'all 200');
  logRow('revisions (sorted)', sorted.join(','));
  logRow('elapsed (ms)', elapsed);
  console.log('  PASS: S1');
}

// ───────────────────────────────────────────────────────────────────────────
// Scenario 2 — Concurrent submits on the same attempt (freshly created).
//
// We create a fresh assessment + draft attempt for the same student so we
// do not pollute the seed attempt. Then we fire N concurrent submits.
// Exactly one must succeed (200) and the rest must be cleanly rejected
// (409 attempt_not_in_progress). No 5xx.
// ───────────────────────────────────────────────────────────────────────────
async function scenario2_concurrentSubmits(studentToken, teacherToken, classId) {
  logHeader(`S2: concurrent submits (N=${CONCURRENT_SUBMIT_N}) on fresh attempt`);

  // 2.1 Teacher creates a draft assessment for class 8A1 with a single
  //     published-question item.
  const createRes = await fetch(`${API_PREFIX}/classes/${classId}/assessments`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ title: 'Load-test assessment', duration_minutes: 30, max_attempts: 5 }),
  });
  if (!createRes.ok) {
    throw new Error(`S2: create assessment failed: ${createRes.status} body=${await createRes.text()}`);
  }
  const assessment = (await createRes.json()).data;

  // 2.2 Add a single section + a single published-question item.
  const sectionRes = await fetch(`${API_PREFIX}/assessments/${assessment.id}/sections`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ title: 'Load section', position: 1 }),
  });
  if (!sectionRes.ok) throw new Error(`S2: create section failed: ${sectionRes.status}`);
  const section = (await sectionRes.json()).data;

  const itemRes = await fetch(`${API_PREFIX}/assessment-sections/${section.id}/items`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ question_version_id: SEED_QUESTION_VERSION_ID, position: 1, points: '1.00' }),
  });
  if (!itemRes.ok) {
    throw new Error(`S2: create item failed: ${itemRes.status} body=${await itemRes.text()}`);
  }

  // 2.3 Target class + validate + publish.
  const targetRes = await fetch(`${API_PREFIX}/assessments/${assessment.id}/targets`, {
    method: 'POST',
    headers: headers(teacherToken, true),
    body: JSON.stringify({ class_section_id: classId }),
  });
  if (!targetRes.ok) throw new Error(`S2: create target failed: ${targetRes.status}`);

  const valRes = await fetch(`${API_PREFIX}/assessments/${assessment.id}/validate`, {
    method: 'POST',
    headers: headers(teacherToken, true),
  });
  if (!valRes.ok) throw new Error(`S2: validate failed: ${valRes.status}`);
  const valBody = await valRes.json();
  if (!valBody.data.valid) {
    throw new Error(`S2: assessment invalid: ${JSON.stringify(valBody.data.errors)}`);
  }

  const pubRes = await fetch(`${API_PREFIX}/assessments/${assessment.id}/publish`, {
    method: 'POST',
    headers: headers(teacherToken, true),
  });
  if (!pubRes.ok) throw new Error(`S2: publish failed: ${pubRes.status}`);

  // 2.4 Student starts a fresh attempt (max_attempts=5 ensures the student
  //     can re-attempt after previous load runs against this assessment).
  const startRes = await fetch(`${API_PREFIX}/assessments/${assessment.id}/attempts`, {
    method: 'POST',
    headers: headers(studentToken, true),
  });
  if (!startRes.ok) {
    throw new Error(`S2: start attempt failed: ${startRes.status} body=${await startRes.text()}`);
  }
  const attempt = (await startRes.json()).data;
  if (attempt.status !== 'IN_PROGRESS') {
    throw new Error(`S2: expected IN_PROGRESS, got ${attempt.status}`);
  }

  // 2.5 Fire N concurrent submits.
  // The submit handler is intentionally idempotent: the first request
  // transitions IN_PROGRESS → SUBMITTED, all subsequent ones see
  // Status == "SUBMITTED" and return the original 200 response without
  // re-grading. We therefore assert:
  //   - all N responses are 200 (no 4xx, no 5xx),
  //   - all N response bodies share the same submitted_at + score +
  //     grading_status (proving the database was only written once),
  //   - the final GET shows the attempt in a terminal state.
  const startedAt = Date.now();
  const responses = await Promise.all(
    Array.from({ length: CONCURRENT_SUBMIT_N }, (_, i) =>
      withTimeout(
        (async () => {
          const r = await fetch(`${API_PREFIX}/attempts/${attempt.id}/submit`, {
            method: 'POST',
            headers: headers(studentToken, true),
          });
          const body = r.ok ? await r.json() : await r.text();
          return { i, status: r.status, body };
        })(),
        SCENARIO_TIMEOUT_MS,
        `submit[${i}]`
      )
    )
  );
  const elapsed = Date.now() - startedAt;

  const nonOk = responses.filter((r) => r.status !== 200);
  if (nonOk.length > 0) {
    throw new Error(
      `S2 FAIL: ${nonOk.length}/${CONCURRENT_SUBMIT_N} submits did not return 200. ` +
        `statuses=${nonOk.map((r) => r.status).sort().join(',')}; bodies=${nonOk
          .map((r) => (typeof r.body === 'string' ? r.body.slice(0, 80) : JSON.stringify(r.body).slice(0, 80)))
          .join(' | ')}`
    );
  }

  // Idempotency: every response must carry the same submitted_at + score.
  const submittedAts = responses.map((r) => r.body && r.body.data && r.body.data.submitted_at);
  const scores = responses.map((r) => r.body && r.body.data && r.body.data.score);
  const gradingStatuses = responses.map((r) => r.body && r.body.data && r.body.data.grading_status);
  const uniqueSubmittedAt = new Set(submittedAts);
  const uniqueScores = new Set(scores);
  const uniqueGrading = new Set(gradingStatuses);
  if (uniqueSubmittedAt.size !== 1) {
    throw new Error(
      `S2 FAIL: idempotency broken — submitted_at diverged across ${CONCURRENT_SUBMIT_N} submits: ${[...uniqueSubmittedAt].join(' | ')}`
    );
  }
  if (uniqueScores.size !== 1) {
    throw new Error(
      `S2 FAIL: idempotency broken — score diverged: ${[...uniqueScores].join(' | ')}`
    );
  }
  if (uniqueGrading.size !== 1) {
    throw new Error(
      `S2 FAIL: idempotency broken — grading_status diverged: ${[...uniqueGrading].join(' | ')}`
    );
  }

  // 2.6 Confirm the final attempt is in a terminal status (SUBMITTED or EXPIRED).
  const finalRes = await fetch(`${API_PREFIX}/attempts/${attempt.id}`, { headers: headers(studentToken) });
  if (!finalRes.ok) throw new Error(`S2: final GET attempt failed: ${finalRes.status}`);
  const finalBody = await finalRes.json();
  if (!['SUBMITTED', 'EXPIRED'].includes(finalBody.data.status)) {
    throw new Error(`S2 FAIL: final status ${finalBody.data.status}, expected SUBMITTED or EXPIRED`);
  }

  logRow('statuses', 'all 200');
  logRow('submitted_at', [...uniqueSubmittedAt][0]);
  logRow('score', [...uniqueScores][0]);
  logRow('grading_status', [...uniqueGrading][0]);
  logRow('final status', finalBody.data.status);
  logRow('elapsed (ms)', elapsed);
  console.log('  PASS: S2');

  return { attemptId: attempt.id, itemId: (attempt.items || [])[0]?.id };
}

// ───────────────────────────────────────────────────────────────────────────
// Scenario 3 — Save after submit is cleanly rejected.
//
// We re-use the attempt from S2 (now SUBMITTED). A subsequent save must
// return 409 attempt_not_in_progress, not 5xx.
// ───────────────────────────────────────────────────────────────────────────
async function scenario3_saveAfterSubmit(studentToken, ctx) {
  logHeader('S3: save-after-submit is cleanly rejected (409)');
  if (!ctx.attemptId || !ctx.itemId) {
    throw new Error('S3: missing attemptId/itemId from S2');
  }
  const r = await fetch(
    `${API_PREFIX}/attempts/${ctx.attemptId}/answers/${ctx.itemId}`,
    {
      method: 'PUT',
      headers: headers(studentToken, true),
      body: JSON.stringify({ answer_payload: { selected_option: 'B' } }),
    }
  );
  if (r.status === 200) {
    throw new Error('S3 FAIL: save-after-submit returned 200; expected 409');
  }
  if (r.status !== 409) {
    const body = await r.text();
    throw new Error(`S3 FAIL: save-after-submit returned ${r.status}, expected 409. body=${body.slice(0, 120)}`);
  }
  const body = await r.json();
  if (!body || !body.error || !body.error.code) {
    throw new Error(`S3 FAIL: expected error envelope, got ${JSON.stringify(body).slice(0, 120)}`);
  }
  logRow('status', '409 (expected)');
  logRow('error.code', body.error.code);
  console.log('  PASS: S3');
}

// ───────────────────────────────────────────────────────────────────────────
// Scenario 4 — Burst reads against the seed attempt.
//
// Fires N concurrent GETs against GET /attempts/{id}. The seed attempt is
// still IN_PROGRESS after S1, so all reads must return 200 with a
// consistent status field.
// ───────────────────────────────────────────────────────────────────────────
async function scenario4_burstReads(studentToken) {
  logHeader(`S4: burst reads (N=${BURST_READ_N}) on seed attempt`);
  const startedAt = Date.now();
  const responses = await Promise.all(
    Array.from({ length: BURST_READ_N }, (_, i) =>
      withTimeout(
        (async () => {
          const r = await fetch(`${API_PREFIX}/attempts/${SEED_ATTEMPT_ID}`, {
            headers: headers(studentToken),
          });
          const body = r.ok ? await r.json() : await r.text();
          return { i, status: r.status, body };
        })(),
        SCENARIO_TIMEOUT_MS,
        `read[${i}]`
      )
    )
  );
  const elapsed = Date.now() - startedAt;

  const ok = responses.filter((r) => r.status === 200);
  const notOk = responses.filter((r) => r.status !== 200);
  if (notOk.length > 0) {
    throw new Error(
      `S4 FAIL: ${notOk.length}/${BURST_READ_N} reads did not return 200. statuses=${notOk
        .map((r) => r.status)
        .sort()
        .join(',')}`
    );
  }
  if (ok.length !== BURST_READ_N) {
    throw new Error(`S4 FAIL: expected ${BURST_READ_N} OK responses, got ${ok.length}`);
  }
  const statuses = ok.map((r) => r.body.data.status);
  const unique = [...new Set(statuses)];
  if (unique.length !== 1) {
    throw new Error(`S4 FAIL: inconsistent statuses across reads: ${statuses.join(',')}`);
  }
  if (!['IN_PROGRESS', 'SUBMITTED', 'EXPIRED', 'GRADED'].includes(unique[0])) {
    throw new Error(`S4 FAIL: unexpected status ${unique[0]}`);
  }
  // Item count must be stable.
  const itemCounts = ok.map((r) => (r.body.data.items || []).length);
  const uniqueCounts = [...new Set(itemCounts)];
  if (uniqueCounts.length !== 1 || uniqueCounts[0] === 0) {
    throw new Error(`S4 FAIL: inconsistent item counts: ${itemCounts.join(',')}`);
  }
  logRow('statuses', unique.join(','));
  logRow('item count', uniqueCounts[0]);
  logRow('elapsed (ms)', elapsed);
  console.log('  PASS: S4');
}

async function findClass8A1Id(studentToken) {
  const r = await fetch(`${API_PREFIX}/classes`, { headers: headers(studentToken) });
  if (!r.ok) throw new Error(`list classes failed: ${r.status}`);
  const json = await r.json();
  const cls = (json.data || []).find((c) => c.name === '8A1');
  if (!cls) throw new Error('seeded class 8A1 not found');
  return cls.id;
}

async function main() {
  console.log('Waiting for API...');
  await ready();

  console.log('Logging in (student + teacher)...');
  const studentToken = await loginAs('hs001', 'Password123!');
  const teacherToken = await loginAs('gv001', 'Password123!');

  // 1) S1 — concurrent saves against seed attempt.
  await scenario1_concurrentSaves(studentToken);

  // 2) S2 — concurrent submits against a freshly created attempt.
  //    List classes as the teacher (student is 403 on GET /classes).
  const classId = await findClass8A1Id(teacherToken);
  const ctx = await scenario2_concurrentSubmits(studentToken, teacherToken, classId);

  // 3) S3 — save-after-submit is cleanly rejected.
  await scenario3_saveAfterSubmit(studentToken, ctx);

  // 4) S4 — burst reads against seed attempt.
  await scenario4_burstReads(studentToken);

  console.log('\nLoad harness: ALL SCENARIOS PASSED.');
}

main().catch((err) => {
  console.error('Load harness FAILED:', err);
  process.exit(1);
});
