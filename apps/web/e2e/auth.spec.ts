import { test, expect } from '@playwright/test';
import { loginAs } from './helpers';

test.describe('auth & role redirect', () => {
  test('student logs in and lands on student dashboard', async ({ page }) => {
    await loginAs(page, 'student');
    await expect(page).toHaveURL(/\/app\/student/);
    await expect(page.getByRole('heading', { name: /Trang làm việc/ })).toBeVisible();
    await expect(page.getByTestId('assigned-assessments-section')).toBeVisible();
  });

  test('teacher logs in, changes forced password, and lands on teacher dashboard', async ({ page }) => {
    await loginAs(page, 'teacher');
    await expect(page).toHaveURL(/\/app\/teacher/);
    await expect(page.getByRole('heading', { name: /Trang giáo viên/ })).toBeVisible();
  });

  test('admin logs in, changes forced password, and lands on admin dashboard', async ({ page }) => {
    await loginAs(page, 'admin');
    await expect(page).toHaveURL(/\/app\/admin/);
    await expect(page.getByRole('heading', { name: /Trang quản trị/ })).toBeVisible();
  });

  test('invalid credentials show login error', async ({ page }) => {
    await page.goto('/login');
    await page.getByTestId('organization-code-input').fill('school-a');
    await page.getByTestId('username-input').fill('hs001');
    await page.getByTestId('password-input').fill('WrongPassword123!');
    await page.getByTestId('login-submit').click();
    await expect(page.getByTestId('login-error')).toBeVisible();
  });
});
