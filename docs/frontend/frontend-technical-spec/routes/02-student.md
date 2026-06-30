# Routes 02 — Student Workspace

Base: `/app/student`

## 1. Route table

| Route | Page | Permission | Primary data | Key actions |
|---|---|---|---|---|
| `/app/student` | Dashboard | `student:workspace` | upcoming work, recent grades, notices | Continue/open |
| `/app/student/classes` | My Classes | `class:own:view` | enrolled classes | Open class |
| `/app/student/classes/:classId` | Class Overview | `class:own:view` | class, teacher, modules | Navigate tabs |
| `/app/student/classes/:classId/resources` | Resources | `resource:assigned:view` | folders/resources | View/download |
| `/app/student/assignments` | Assignments | `assignment:own:view` | assigned list | Filter/open |
| `/app/student/assignments/:assignmentId` | Assignment Detail | `assignment:own:view` | assignment, submission | Draft/submit |
| `/app/student/assessments` | Assessments | `assessment:assigned:view` | availability/list | View/start |
| `/app/student/assessments/:assessmentId` | Assessment Instructions | `assessment:assigned:view` | rules/availability | Start attempt |
| `/exam/attempts/:attemptId` | Exam Runner | `attempt:own:continue` | runtime snapshot | Answer/submit |
| `/app/student/results/:attemptId` | Attempt Result | `attempt:own:view` | published result | Review allowed data |
| `/app/student/grades` | Grades | `grade:own:view` | published grade items | Filter/details |

## 2. Dashboard composition

```text
PageHeader
├── ContinueLearningCard
├── UpcomingAssignments
├── UpcomingAssessments
├── RecentlyPublishedGrades
└── RecentNotifications
```

MVP excludes AI, XP, leaderboard and radar chart.

## 3. Class overview

Tabs:

- Overview.
- Resources.
- Assignments.
- Assessments.
- Grades if class-specific.

Tab can be nested route or search param. Prefer nested route when deep-link matters.

## 4. Assignment detail

States:

| Submission status | UI |
|---|---|
| Not started | Start form |
| Draft | Restore draft |
| Submitted | Read-only receipt |
| Late accepted | Label late, submit if allowed |
| Graded | Feedback/score published |
| Resubmission requested | New version form |

File upload uses direct signed URL flow. Text draft may persist locally only if long-form and clearly scoped; server draft preferred.

## 5. Assessment instructions

Must show:

- Open/close time in timezone.
- Duration.
- Attempts remaining.
- Navigation/shuffle rules.
- Result/answer release policy.
- Connection/device advice.
- Capability check.

Start action uses idempotency key.

## 6. Exam result

Render only fields backend returns. Possible states:

- Submitted, grading pending.
- Manual review pending.
- Score published.
- Detailed answers hidden until release.
- Expired/terminated.

Do not infer correct answers from snapshot or client state.

## 7. Grades page

- Decimal score display from strings.
- Published only.
- Category/period filters in URL.
- Missing/not graded/absent/exempt statuses distinct from zero.
- No rank in MVP.

## 8. Mobile priority

- Dashboard single column.
- Upcoming deadline cards first.
- Class tabs horizontally scrollable or select.
- Exam route uses separate layout.
