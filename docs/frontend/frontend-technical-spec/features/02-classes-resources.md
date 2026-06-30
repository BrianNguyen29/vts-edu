# Feature 02 — Classes & Resources

## 1. Scope

- Class list/detail.
- Enrollment/student list for authorized roles.
- Resource folders/items.
- Direct upload and confirm.
- Resource publication/access.
- Signed download and preview.

## 2. Class dataflow

```text
route filters
-> class list query
-> cards/table
-> open detail
-> nested resource/student/assignment queries
```

Class summary and full detail use separate query options to avoid overfetch.

## 3. Components

```text
entities/class/ui/
├── class-card.tsx
├── class-status-badge.tsx
└── class-summary.tsx

widgets/class-overview/
├── class-header.tsx
├── class-stats.tsx
└── class-tabs.tsx
```

## 4. Resource model

Types MVP:

- File.
- Link (`url`, `open_in_new`).
- Rich content/page (canonical format do backend quyết định, xem `backend/api/04-question-bank.md`).
- Folder.

Resource list provides:

- Name/type/size.
- Publication status/time.
- Last updated.
- Actions by permission.

Backend cần bổ sung folder CRUD: `POST /resources/folders`, `PATCH`, `POST /resources/folders/{id}/move`, `POST /resources/folders/{id}/items/reorder`, `DELETE`.

## 5. Upload state machine

```text
SELECTED
-> REQUESTING_INTENT
-> UPLOADING
-> CONFIRMING
-> PROCESSING
-> READY
```

Failure states retain enough context to retry safely.

### Direct upload

1. Validate size/MIME hint.
2. Request upload intent.
3. Upload to signed URL without bearer token.
4. Confirm file ID to backend.
5. Query processing status.

## 6. Upload UI

- Progress per file.
- Cancel.
- Retry.
- Do not navigate away silently during active upload.
- Large file size and network warning.
- Screen reader status updates throttled.

## 7. Resource preview

- Image: safe blob/object URL or signed URL.
- PDF: lazy viewer or browser native fallback.
- Office files: download in MVP unless converted preview exists.
- HTML/SVG upload: no direct inline rendering.
- Link: open with safe `rel="noopener noreferrer"` where new tab.

## 8. Cache policy

- Class list invalidate after create/archive/enrollment action.
- Resource folder list invalidate after upload/rename/publish.
- Signed URL response should have short `staleTime` or no persistent cache.
- Avoid storing binary in Query cache.

## 9. Tests

- Student cannot see teacher action controls.
- Backend 403 overrides stale UI permission.
- Upload cancel/retry/confirm.
- Signed URL expired -> reacquire.
- Empty folder and processing file states.
- Mobile list behavior.
