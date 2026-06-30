# Feature 06 — Assignments, Submissions & Gradebook

## 1. Assignment scope

Teacher:

- Create/edit/schedule assignment.
- Attach resources.
- View submissions.
- Review feedback/grade.

Student:

- View instructions/deadline.
- Draft text/upload files.
- Submit/resubmit if allowed.
- View published feedback/grade.

## 2. Assignment form

Fields:

- Title.
- Rich instructions.
- Class/target.
- Open/due/close times.
- Submission type.
- File limits.
- Resubmission policy.
- Max score decimal string.

Use explicit timezone and server validation.

## 3. Submission state

```text
NOT_STARTED
-> DRAFT
-> SUBMITTED / LATE
-> GRADED
-> RETURNED
-> RESUBMISSION (if allowed)
```

Frontend renders backend status, not derive terminal workflow independently.

## 4. Text/file submission

- Text form can autosave backend draft if endpoint exists.
- File upload uses direct upload intent.
- Submit action explicit and idempotent if backend requires.
- Submission receipt shows version/time.
- Do not overwrite old submission version locally.

## 5. Review UI

```text
SubmissionReviewPage
├── Student/attempt metadata
├── Submission content/files
├── Feedback editor
├── Score field
├── Rubric (P1/basic if available)
└── Save draft / Finalize controls
```

- `PATCH /submissions/{id}/grade-draft` lưu draft score/feedback với `expected_version`.
- `POST /submissions/{id}/finalize-grade` finalize và trả điểm với `Idempotency-Key`.
- Finalize/return là explicit. Conflict uses version check.

## 6. Gradebook model

Display fields:

- Grade item.
- Raw score.
- Max score.
- Override score.
- Final score.
- Publication status.
- Special state: missing, exempt, not graded.

All numeric values remain strings.

## 7. Gradebook grid behavior

- Server pagination by students/items nếu dataset lớn; dùng cursor/page từ backend.
- Sticky first column/header.
- Cell editor supports keyboard.
- Save per cell or controlled batch.
- Error indicator per cell plus summary.
- Unsaved edits tracked.
- Publish separate from save.

## 8. Grade override

Requires:

- New score string.
- Reason.
- Confirmation.
- Permission.

No optimistic update. On success invalidate gradebook and audit-related views if relevant.

## 9. Student grade view

- Published only.
- Explain special states.
- Feedback accessible.
- No hidden item leak from client cache/navigation.

## 10. Tests

- Submission upload/submit/resubmit.
- Deadline state rendering.
- Decimal string preserved.
- Missing vs zero.
- Grade override reason and conflict.
- Publish behavior.
- Keyboard grid smoke test.
