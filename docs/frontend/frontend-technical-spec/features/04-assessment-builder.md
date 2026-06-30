# Feature 04 — Assessment Builder

## 1. Scope

- Create assessment draft.
- Edit metadata/settings.
- Sections.
- Fixed question items.
- Reorder.
- Shuffle settings.
- Schedule/assignment to class.
- Validate/preview/publish.

Random pool rules can be P1 if backend supports later.

## 2. Builder layout

```text
AssessmentBuilder
├── TopBar
│   ├── title/status
│   ├── save state
│   ├── preview
│   └── publish
├── OutlinePanel
├── BuilderCanvas
│   ├── SectionEditor
│   └── AssessmentItemEditor
└── SettingsPanel/Drawer
```

On laptop, settings panel becomes drawer to preserve canvas width.

## 3. State ownership

- Server draft: TanStack Query.
- Current form/editor: RHF + local reducer.
- Dirty status: form state.
- Reorder: local optimistic representation, then explicit save.
- Conflict version: server ETag/version.

Do not put entire builder in global store initially.

## 4. Save strategy

Options:

- Explicit save for structural changes.
- Debounced autosave for title/settings if stable.

Recommended MVP:

- Explicit save + clear save indicator.
- Small debounced metadata save only after core stable.

Every save uses current version/If-Match.

## 5. Question picker

- Server-side search/filter.
- Preview question.
- Select one/many.
- Avoid loading full bank.
- Added item references specific question version per API.

## 6. Reordering

- Keyboard-accessible move up/down actions always available.
- Drag-and-drop optional enhancement.
- Stable item IDs.
- Update order as one explicit request `POST /assessment-sections/{section_id}/items/reorder` hoặc `POST /assessment-sections/reorder` với expected version.
- Failure restores prior order and displays error.

## 7. Missing mutations (P0-10)

Backend cần bổ sung các mutation để UI hoàn tất:

- PATCH `/assessment-items/{item_id}` sửa item/points.
- DELETE `/assessment-sections/{section_id}` xóa section.
- POST `/assessment-sections/{section_id}/items/reorder` batch reorder items.
- POST `/assessment-sections/reorder` batch reorder sections.
- PATCH `/assessment-random-rules/{rule_id}` sửa random rule.
- DELETE `/assessment-random-rules/{rule_id}` xóa random rule.
- DELETE `/assessments/{assessment_id}/targets/{target_id}` gỡ target.
- DELETE `/assessments/{assessment_id}/accommodations/{accommodation_id}` gỡ accommodation.

Mọi structural mutation cần optimistic concurrency (version/If-Match).

## 7. Validation summary

Before publish, show categorized issues:

- Missing title/schedule.
- Empty section.
- Invalid points.
- Unsupported/archived question version.
- Invalid duration.
- No assigned class if required.

Each issue links/focuses relevant section.

Backend publish response remains source of truth and may return additional 422 errors.

## 8. Preview

- Student-like rendering.
- No live attempt creation.
- Shows shuffle as representative only, not guarantee final order.
- Accessibility review.

## 9. Publish

```text
click publish
-> local validation
-> fetch/ensure latest draft version
-> consequence confirm
-> POST publish with idempotency key
-> pending non-dismissible state
-> success: set published snapshot/status, lock immutable fields
-> 409/412: conflict handling
-> 422: validation summary
```

Do not optimistic publish.

## 10. Tests

- Section/item add/remove/reorder.
- Dirty navigation.
- Conflict preservation.
- Publish idempotency key stable on retry.
- Validation summary focus links.
- Permission view-only mode.
- Responsive settings drawer.
