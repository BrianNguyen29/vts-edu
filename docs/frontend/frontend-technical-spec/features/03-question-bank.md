# Feature 03 — Question Bank

## 1. Scope

MVP question types:

- Single choice.
- Multiple choice.
- True/false.
- Short text.
- Numeric.
- Essay.

Capabilities:

- List/filter/search.
- Create logical question and version.
- Preview.
- Status workflow.
- Version history.

## 2. Page/widget composition

```text
QuestionBankPage
├── QuestionFilters
├── QuestionTable/CardList
├── BulkActionBar (later/minimal)
└── CreateQuestionButton

QuestionEditorPage
├── QuestionMetadataForm
├── QuestionContentEditor
├── AnswerConfiguration
├── ExplanationEditor
├── PreviewPanel
└── VersionStatusPanel
```

## 3. Form model

Use discriminated union by question type. Shared metadata:

- Subject/grade/topic.
- Difficulty.
- Tags.
- Estimated time.
- Score default suggestion if API allows.

Type-specific payload mapped to generated request.

## 4. Rich content

- TipTap lazy-loaded.
- KaTeX extension/rendering.
- Paste sanitization.
- Images inserted through file upload, not base64.
- Preview uses same renderer as student exam where possible.

## 5. Version rules

UI must distinguish:

- Logical question.
- Current/latest version.
- Published/used immutable version.

If version immutable:

- Edit action becomes “Tạo phiên bản mới”.
- Clone content into new draft.
- Never send update to immutable version endpoint.

## 6. List filtering

URL state:

```text
?query=&type=&status=&difficulty=&tag=&sort=-updated_at
```

Cursor pagination from API. Debounce text search 300–500ms.

## 7. Question preview registry

Renderer is shared with assessment preview/exam but modes differ:

```ts
type QuestionRenderMode = 'author-preview' | 'student-runtime' | 'result-review';
```

Author preview may show answer key/explanation; student runtime must not.

## 8. Validation

Examples:

- Single choice: at least 2 choices, exactly one correct.
- Multiple choice: at least 2 choices, at least one correct.
- Numeric: target and tolerance valid decimal strings.
- Essay: no automatic key required.

Backend validates final.

## 9. Conflict behavior

PATCH/update draft with version/If-Match. On conflict:

- Preserve local form values.
- Show latest server metadata.
- Offer reload/copy local content.
- Do not overwrite automatically.

## 10. Tests

- Schema per question type.
- API/form mapping round-trip.
- Add/remove/reorder choices.
- Immutable version action.
- Rich text sanitizer.
- Author vs student preview data leakage test.
- Keyboard choice editing.
