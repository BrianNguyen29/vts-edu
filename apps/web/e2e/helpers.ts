import type { Page } from '@playwright/test';

const credentials: Record<string, { username: string; password: string; newPassword: string; home: string }> = {
  student: { username: 'hs001', password: 'Password123!', newPassword: 'Password123!', home: 'student' },
  teacher: { username: 'gv001', password: 'Password123!', newPassword: 'NewPassword123!', home: 'teacher' },
  admin: { username: 'admin001', password: 'Password123!', newPassword: 'AdminPass123!', home: 'admin' },
};

export async function loginAs(page: Page, role: 'student' | 'teacher' | 'admin') {
  const creds = credentials[role];

  await page.goto('/login');
  await page.getByTestId('organization-code-input').fill('school-a');
  await page.getByTestId('username-input').fill(creds.username);
  await page.getByTestId('password-input').fill(creds.password);
  await page.getByTestId('login-submit').click();

  // The router may briefly land on /app before redirecting restricted users to
  // /app/change-password, so wait for a concrete destination instead of /app.
  const destination = new RegExp(`/app/(change-password|${creds.home})`);
  try {
    await page.waitForURL(destination, { timeout: 6_000 });
  } catch {
    // If the old password no longer works (e.g. a previous test already changed
    // it), retry with the new password.
    if (creds.password !== creds.newPassword && page.url().includes('/login')) {
      await page.getByTestId('password-input').fill(creds.newPassword);
      await page.getByTestId('login-submit').click();
      await page.waitForURL(destination);
    } else {
      throw new Error(`Login failed for ${role}; current url: ${page.url()}`);
    }
  }

  if (page.url().includes('/app/change-password')) {
    await page.getByTestId('current-password-input').fill(creds.password);
    await page.getByTestId('new-password-input').fill(creds.newPassword);
    await page.getByTestId('confirm-password-input').fill(creds.newPassword);
    await page.getByTestId('change-password-submit').click();

    // After a successful change the app logs out and returns to /login.
    await page.waitForURL('/login', { timeout: 10_000 });

    await page.getByTestId('organization-code-input').fill('school-a');
    await page.getByTestId('username-input').fill(creds.username);
    await page.getByTestId('password-input').fill(creds.newPassword);
    await page.getByTestId('login-submit').click();
    await page.waitForURL(new RegExp(`/app/${creds.home}`));
  }
}
