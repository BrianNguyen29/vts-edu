import { test, expect, type Page } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { loginAs } from './helpers';

/** Shape of a single axe violation returned by AxeBuilder.analyze(). */
type AxeResults = Awaited<ReturnType<InstanceType<typeof AxeBuilder>['analyze']>>;
type AxeViolation = AxeResults['violations'][number];
type AxeNode = AxeViolation['nodes'][number];

/**
 * Bounded WCAG/axe accessibility gate for stable routes.
 *
 * This spec complements (not replaces) the structural a11y checks in
 * `a11y.spec.ts` by running a real axe-core scan on the same set of
 * stable surfaces. It is intentionally bounded:
 *
 *   - Uses `test.describe.serial` so the whole suite shares one browser
 *     context and one login session per role, matching the existing
 *     `critical-flow.spec.ts` pattern.
 *   - Skips deep flows that need expensive setup (assessment builder,
 *     exam runner, grading detail) and surfaces where the demo data
 *     shape is flaky (attempt review depends on the seeded attempt).
 *   - Uses WCAG 2.0 A+AA + 2.1 AA + best-practice rule sets so the
 *     scan matches the level the spec aims at.
 *   - Disables `color-contrast` and `target-size` rules: the design
 *     system palette and click-target sizes are tracked separately
 *     and are not stable enough to gate on in a unit test.
 *
 * Run only this spec with:
 *
 *   pnpm web:e2e -- axe.spec.ts
 *
 * The whole gate (DB + API + Vite + spec) is wired as
 * `pnpm e2e:a11y`. It is intentionally NOT part of `pnpm check`
 * because it needs a running database and API, just like the
 * existing `pnpm e2e:browser` / `pnpm e2e:load` flows.
 */

const RULES_TO_DISABLE: string[] = [
  // The design system palette is audited manually; axe will flag the
  // muted placeholders / placeholder dots that the team has already
  // reviewed. Keeping this rule off avoids a noisy diff every time
  // someone tweaks a token.
  'color-contrast',
  // Many icon buttons (close, back, status badge) are intentionally
  // smaller than 24px; a target-size rule would block benign UI.
  'target-size',
];

type AxeScan = {
  /** Display name of the surface (used in test titles + log output). */
  surface: string;
  /** Login role required for the route (omit for public routes). */
  role?: 'student' | 'teacher' | 'admin';
  /** Path to navigate to inside the (already-routed) SPA. */
  path: string;
  /**
   * Extra wait predicate that resolves once the page is "ready" for
   * scanning. Defaults to waiting for the page's main heading. Routes
   * that lazy-load a panel (e.g. admin) should pass a more specific
   * selector so axe does not scan a partial render.
   */
  ready?: (page: Page) => Promise<unknown>;
};

const PUBLIC_SURFACES: AxeScan[] = [
  { surface: 'login', path: '/login', ready: (page) => page.getByRole('heading', { level: 1 }).waitFor() },
  { surface: 'error-403', path: '/error/403?requestId=axe-gate', ready: (page) => page.getByTestId('error-page').waitFor() },
  { surface: 'not-found', path: '/this-does-not-exist', ready: (page) => page.getByTestId('not-found-page').waitFor() },
];

const STUDENT_SURFACES: AxeScan[] = [
  {
    surface: 'student-dashboard',
    path: '/app/student',
    ready: (page) => page.getByRole('heading', { name: /Trang làm việc/ }).waitFor(),
  },
];

const TEACHER_SURFACES: AxeScan[] = [
  {
    surface: 'teacher-dashboard',
    path: '/app/teacher',
    ready: (page) => page.getByRole('heading', { name: /Trang giáo viên/ }).waitFor(),
  },
  {
    surface: 'teacher-resources',
    path: '/app/resources',
    ready: (page) => page.getByRole('heading', { name: /Tài liệu/ }).waitFor(),
  },
  {
    surface: 'change-password',
    path: '/app/change-password',
    ready: (page) => page.getByRole('heading', { name: /Đổi mật khẩu/ }).waitFor(),
  },
];

const ADMIN_SURFACES: AxeScan[] = [
  {
    surface: 'admin-dashboard',
    path: '/app/admin',
    ready: (page) => page.getByRole('tablist', { name: /Quản lý quản trị/ }).waitFor(),
  },
];

function summarizeViolations(violations: AxeViolation[]): string {
  if (violations.length === 0) return 'no axe violations';
  return violations
    .map((v) => {
      const nodes = v.nodes
        .slice(0, 3)
        .map((n: AxeNode) => formatNodeTarget(n))
        .join('\n');
      const more = v.nodes.length > 3 ? `\n      … and ${v.nodes.length - 3} more` : '';
      return `    [${v.impact ?? 'unknown'}] ${v.id} — ${v.description}\n      help: ${v.helpUrl}\n${nodes}${more}`;
    })
    .join('\n');
}

function formatNodeTarget(node: AxeNode): string {
  const target = Array.isArray(node.target)
    ? node.target.map((t) => String(t)).join(' ')
    : String(node.target);
  const summary = node.failureSummary ? node.failureSummary.split('\n').join(' | ') : 'no summary';
  return `      • ${target} (${summary})`;
}

async function runAxe(page: Page, surface: string) {
  // Wait for the route to be in a stable, interactive state before
  // scanning. axe mutates the DOM (highlights) and we want the scan
  // to reflect the user-visible page, not a partial render.
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'best-practice'])
    .disableRules(RULES_TO_DISABLE)
    .analyze();
  // axe will sometimes report `incomplete` results that are not
  // actionable from a unit test (e.g. need a real color picker to
  // disambiguate). We log them but only fail on `violations`.
  if (results.incomplete.length > 0) {
    test.info().annotations.push({
      type: `axe-incomplete:${surface}`,
      description: `${results.incomplete.length} incomplete result(s): ${results.incomplete
        .map((i) => i.id)
        .join(', ')}`,
    });
  }
  expect(
    results.violations,
    `axe gate failed for surface "${surface}":\n${summarizeViolations(results.violations)}`
  ).toEqual([]);
}

test.describe.serial('@axe accessibility gate', () => {
  test('public surfaces (login, error, 404)', async ({ page }) => {
    for (const scan of PUBLIC_SURFACES) {
      await page.goto(scan.path);
      if (scan.ready) {
        await scan.ready(page);
      }
      await runAxe(page, scan.surface);
    }
  });

  test('student surfaces', async ({ page }) => {
    await loginAs(page, 'student');
    for (const scan of STUDENT_SURFACES) {
      await page.goto(scan.path);
      if (scan.ready) {
        await scan.ready(page);
      }
      await runAxe(page, scan.surface);
    }
  });

  test('teacher surfaces (dashboard, resources, change-password)', async ({ page }) => {
    await loginAs(page, 'teacher');
    for (const scan of TEACHER_SURFACES) {
      await page.goto(scan.path);
      if (scan.ready) {
        await scan.ready(page);
      }
      await runAxe(page, scan.surface);
    }
  });

  test('admin surfaces', async ({ page }) => {
    await loginAs(page, 'admin');
    for (const scan of ADMIN_SURFACES) {
      await page.goto(scan.path);
      if (scan.ready) {
        await scan.ready(page);
      }
      await runAxe(page, scan.surface);
    }
  });
});
