import { test, expect } from '@playwright/test';
import { loginAs } from './helpers';

test.describe('teacher assessment builder', () => {
  test('creates, validates and publishes an assessment', async ({ page }) => {
    await loginAs(page, 'teacher');

    // Open create-assessment form.
    await page.getByTestId('create-assessment-button').click();
    await page.getByTestId('create-assessment-title').fill('E2E Kiểm tra Playwright');
    await page.getByTestId('create-assessment-class').selectOption({ index: 1 });
    await page.getByTestId('create-assessment-submit').click();

    // Should navigate to builder.
    await expect(page).toHaveURL(/\/app\/teacher\/assessments\//);

    // Add a section.
    await page.getByTestId('section-title-input').fill('Phần trắc nghiệm');
    await page.getByTestId('add-section-button').click();
    await expect(page.getByTestId('builder-section')).toHaveCount(1);

    // Add the first available question to the section.
    await page.getByTestId('add-question-button').first().click();
    const pickerSelect = page.getByTestId('picker-question-select');
    await expect.poll(async () => await pickerSelect.locator('option').count()).toBeGreaterThan(1);
    await pickerSelect.selectOption({ index: 1 });
    await page.getByTestId('picker-add-button').click();

    // Assign the same class as target.
    const targetSelect = page.getByTestId('target-class-select');
    await expect.poll(async () => await targetSelect.locator('option').count()).toBeGreaterThan(1);
    await targetSelect.selectOption({ index: 1 });
    await page.getByTestId('add-target-button').click();

    // Validate.
    await page.getByTestId('validate-button').click();
    await expect(page.getByTestId('builder-success')).toContainText('hợp lệ');

    // Publish.
    page.on('dialog', (dialog) => dialog.accept());
    await page.getByTestId('publish-button').click();
    await expect(page.getByTestId('builder-success')).toContainText('xuất bản');
    await expect(page.getByTestId('publication-table')).toBeVisible();
  });
});
