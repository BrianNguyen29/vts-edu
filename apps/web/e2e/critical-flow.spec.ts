import { test, expect, type BrowserContext, type Page } from '@playwright/test';
import { loginAs } from './helpers';

const ASSESSMENT_TITLE = 'E2E Kiểm tra Playwright';

async function createAndPublishAssessment(page: Page) {
  await loginAs(page, 'teacher');
  await page.getByTestId('create-assessment-button').click();
  await page.getByTestId('create-assessment-title').fill(ASSESSMENT_TITLE);
  await page.getByTestId('create-assessment-class').selectOption({ index: 1 });
  await page.getByTestId('create-assessment-submit').click();

  await expect(page).toHaveURL(/\/app\/teacher\/assessments\//);

  await page.getByTestId('section-title-input').fill('Phần trắc nghiệm');
  await page.getByTestId('add-section-button').click();
  await expect(page.getByTestId('builder-section')).toHaveCount(1);

  await page.getByTestId('add-question-button').first().click();
  await page.getByTestId('picker-question-select').selectOption({ index: 1 });
  await page.getByTestId('picker-add-button').click();

  await page.getByTestId('target-class-select').selectOption({ index: 1 });
  await page.getByTestId('add-target-button').click();

  await page.getByTestId('validate-button').click();
  await expect(page.getByTestId('builder-success')).toContainText('hợp lệ');

  page.on('dialog', (dialog) => dialog.accept());
  await page.getByTestId('publish-button').click();
  await expect(page.getByTestId('builder-success')).toContainText('xuất bản');
}

test.describe.serial('critical end-to-end flow', () => {
  test('teacher publishes an assessment, student takes it, teacher views gradebook', async ({ browser }) => {
    // 1. Teacher publishes the assessment.
    const teacherContext = await browser.newContext();
    const teacherPage = await teacherContext.newPage();
    await createAndPublishAssessment(teacherPage);
    await teacherContext.close();

    // 2. Student starts and submits the assessment.
    const studentContext = await browser.newContext();
    const studentPage = await studentContext.newPage();
    await loginAs(studentPage, 'student');

    const assessmentCard = studentPage.locator('.assessment-list-item', {
      hasText: ASSESSMENT_TITLE,
    });
    await expect(assessmentCard).toBeVisible();
    await assessmentCard.getByTestId('start-assessment-button').click();

    await expect(studentPage).toHaveURL(/\/exam\/attempts\//);

    const questions = studentPage.getByTestId('exam-question');
    await expect(questions.first()).toBeVisible({ timeout: 10_000 });
    const count = await questions.count();
    expect(count).toBeGreaterThan(0);
    for (let i = 0; i < count; i++) {
      await questions.nth(i).locator('input[type="radio"]').first().check();
    }

    studentPage.on('dialog', (dialog) => dialog.accept());
    await studentPage.getByTestId('submit-exam-button').click();

    await expect(studentPage.getByText(/Điểm:/)).toBeVisible();
    await studentContext.close();

    // 3. Teacher views the gradebook and can export CSV.
    const teacherContext2 = await browser.newContext();
    const teacherPage2 = await teacherContext2.newPage();
    await loginAs(teacherPage2, 'teacher');
    await teacherPage2.goto('/app/teacher/gradebook?tab=assessment');

    await teacherPage2.getByTestId('gradebook-assessment-select').selectOption({ label: ASSESSMENT_TITLE });
    await expect(teacherPage2.getByTestId('gradebook-table')).toBeVisible();
    await expect(teacherPage2.getByTestId('export-assessment-csv')).toBeEnabled();
    await teacherContext2.close();
  });

  test('admin bulk imports users with dry-run and confirm', async ({ browser }) => {
    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();
    await loginAs(adminPage, 'admin');

    await adminPage.getByTestId('users-tab').click();
    await adminPage.getByTestId('import-csv-button').click();

    const csv = [
      'login_name,display_name,email,temporary_password,roles',
      'e2e-student,E2E Student,e2e-student@example.com,TempPass123!,student',
      'e2e-invalid,,invalid@example.com,TempPass123!,student',
    ].join('\n');

    await adminPage.getByTestId('import-csv-textarea').fill(csv);
    await adminPage.getByTestId('dry-run-import-button').click();
    await expect(adminPage.getByTestId('import-preview')).toBeVisible();

    await adminPage.getByTestId('confirm-import-button').click();
    await expect(adminPage.getByText(/Đã nhập/)).toBeVisible();

    await adminContext.close();
  });
});
