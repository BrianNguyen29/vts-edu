import { test, expect } from '@playwright/test';

test.describe('error pages', () => {
  test('shows 403 state with request id and copy button', async ({ page }) => {
    await page.goto('/error/403?requestId=test-req-123');
    await expect(page.getByTestId('error-page')).toBeVisible();
    await expect(page.getByTestId('error-message')).toContainText(
      'Không có quyền truy cập.'
    );
    await expect(page.getByTestId('error-request-id')).toContainText(
      'test-req-123'
    );
    await expect(page.getByTestId('error-copy-button')).toBeVisible();
  });

  test('shows 500 error state', async ({ page }) => {
    await page.goto('/error/500');
    await expect(page.getByTestId('error-page')).toBeVisible();
    await expect(page.getByTestId('error-message')).toContainText(
      'Máy chủ gặp sự cố'
    );
  });

  test('shows 404 not found for unknown routes', async ({ page }) => {
    await page.goto('/this-route-does-not-exist');
    await expect(page.getByTestId('not-found-page')).toBeVisible();
    await expect(page.getByText('404')).toBeVisible();
  });
});
