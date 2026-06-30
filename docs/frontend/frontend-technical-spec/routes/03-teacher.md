# Routes 03 — Teacher Workspace

Base: `/app/teacher`

## 1. Route table

| Route | Page | Permission | Main capability |
|---|---|---|---|
| `/app/teacher` | Dashboard | `teacher:workspace` | Pending grading, upcoming tests, classes |
| `/app/teacher/classes` | Classes | `class:view` | List/filter/open |
| `/app/teacher/classes/:classId` | Class Workspace | `class:view` | Overview/students/content/results |
| `/app/teacher/classes/:classId/students` | Students | `enrollment:view` | View enrollment/progress |
| `/app/teacher/classes/:classId/resources` | Resources | `resource:view` | Create/upload/publish |
| `/app/teacher/question-banks` | Question Banks | `question:view` | Bank/filter |
| `/app/teacher/question-banks/:bankId` | Bank Detail | `question:view` | Questions/tags |
| `/app/teacher/questions/new` | New Question | `question:create` | Create version |
| `/app/teacher/questions/:questionId` | Question Detail | `question:view` | Preview/version history |
| `/app/teacher/questions/:questionId/new-version` | New Version | `question:update` | Clone/edit/publish |
| `/app/teacher/assessments` | Assessments | `assessment:view` | List/status; backend aggregate qua `/me/teaching/assessments` |
| `/app/teacher/assessments/new` | New Assessment | `assessment:create` | Create draft |
| `/app/teacher/assessments/:assessmentId/edit` | Builder | `assessment:update` | Sections/items/settings |
| `/app/teacher/assessments/:assessmentId/preview` | Preview | `assessment:view` | Student-like preview |
| `/app/teacher/assessments/:assessmentId/results` | Results | `attempt:grade` | Attempts/statistics |
| `/app/teacher/attempts/:attemptId/review` | Manual Review | `attempt:grade` | Essay/manual grading |
| `/app/teacher/assignments` | Assignments | `assignment:view` | List/create; backend aggregate qua `/me/teaching/assignments` |
| `/app/teacher/assignments/new` | New Assignment | `assignment:create` | Create/schedule |
| `/app/teacher/assignments/:assignmentId` | Assignment Detail | `assignment:view` | Submissions/settings |
| `/app/teacher/submissions/:submissionId/review` | Submission Review | `submission:grade` | Feedback/grade |
| `/app/teacher/gradebook/:classId` | Gradebook | `grade:view` | Edit/publish/export |

## 2. Dashboard

Priority order:

1. Bài chờ chấm.
2. Học sinh chưa nộp/bài quá hạn.
3. Assessment sắp mở.
4. Quick create actions.
5. Class activity summary.

No decorative hero taking most viewport in working dashboard.

## 3. Class workspace

Nested areas:

- Overview.
- Students.
- Resources.
- Assignments.
- Assessments.
- Gradebook.

Class ID is always from route, but API authorization verifies teacher scope.

## 4. Question editor route

Page sections:

```text
Question metadata
Question content editor
Answer configuration
Explanation
Preview
Version/status panel
```

Create new version rather than editing immutable published version.

## 5. Assessment builder route

Full-width editor layout inside AppShell or specialized builder layout.

Components:

- Outline/sections panel.
- Item canvas.
- Question picker drawer.
- Settings panel.
- Validation summary.
- Save status.
- Preview/publish actions.

Route blocker if dirty/local unsaved changes.

## 6. Results and manual review

- Result list uses server pagination/filter.
- Manual review locks or version checks according API.
- Score input string decimal.
- Feedback draft not lost on navigation.
- Finalize action explicit.

## 7. Gradebook

- Full-bleed horizontal workspace.
- Sticky names and headers.
- Keyboard cell navigation where practical.
- Edit state per cell/batch.
- Conflict/error visible at cell and page summary.
- Publish is explicit, not automatic after edit.

## 8. Permission-aware actions

Teacher may have view without publish/update. Buttons rendered based on exact permissions, not merely teacher role.
