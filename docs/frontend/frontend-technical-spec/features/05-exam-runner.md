# Feature 05 — Exam Runner

This feature follows the full specification in `../10-exam-runtime-frontend.md`.

## 1. Slice map

```text
features/attempts/
├── start-attempt/
├── load-runtime/
├── save-answer/
├── flag-question/
├── submit-attempt/
└── reconcile-attempt/

widgets/exam-navigation/
widgets/exam-header/
entities/attempt/
shared/db/exam-draft-repository.ts
```

## 2. Public API

```ts
export {
  useAttemptRuntime,
  useAnswerController,
  useSubmitAttempt,
  ExamSaveStatus,
  ExamTimer,
} from '@/features/attempts';
```

## 3. Required services

```ts
interface ExamRuntimeServices {
  api: ApiClient;
  drafts: ExamDraftRepository;
  clock: ServerOffsetClock;
  connectivity: ConnectivityService;
  logger: FrontendLogger;
}
```

## 4. UI components

- `ExamHeader`.
- `ExamTimer`.
- `ExamSaveStatus`.
- `QuestionRenderer`.
- `QuestionNavigator`.
- `UnansweredSummary`.
- `SubmitAttemptDialog`.
- `OfflineBanner`.
- `AttemptConflictScreen`.

## 5. Prohibited simplifications

- No token in IndexedDB.
- No answer save only on next/submit.
- No client-only timer authority.
- No `navigator.sendBeacon` as primary save.
- No auto-delete pending drafts on generic error.
- No service-worker-only sync.
- No final score computation.

## 6. Test requirement

Any PR changing this feature must include at least one of:

- Pure sync state machine test.
- Browser IndexedDB integration test.
- Playwright network/reload scenario.

And must run existing critical exam suite.
