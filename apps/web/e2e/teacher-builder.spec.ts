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

    // Add a multiple_choice question to the section. The picker now exposes
    // short_answer/essay options after the non-MCQ foundation; this spec only
    // covers the MCQ publish path, so we filter by the `[TN]` label prefix.
    await page.getByTestId('add-question-button').first().click();
    const pickerSelect = page.getByTestId('picker-question-select');
    await expect.poll(async () => await pickerSelect.locator('option').count()).toBeGreaterThan(1);
    const mcqValue = await pickerSelect.evaluate((el) => {
      const select = el as HTMLSelectElement;
      const opt = Array.from(select.options).find((o) => /^\[TN\]/.test(o.textContent ?? ''));
      return opt?.value ?? '';
    });
    if (!mcqValue) {
      throw new Error('No multiple_choice question available in the picker.');
    }
    await pickerSelect.selectOption(mcqValue);
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
