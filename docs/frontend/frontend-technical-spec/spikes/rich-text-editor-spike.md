# Rich text / TipTap + KaTeX question editor — feasibility spike (2026-07-02)

> Spike-only. Lives on branch `spike/rich-text-editor`. **No merge to `main` yet**
> — go/no-go decision is pending the spike evidence collected below.

## TL;DR — GO with caveats

- **Verdict: GO**, with a **~184 kB gz deferred-cost ceiling** for the rich text mode
  (editor chunk 106.57 kB gz + KaTeX chunk 77.57 kB gz) loaded **only** when a
  teacher explicitly opts into "Rich text" mode and inserts a math formula.
- All `+5 kB gz` (initial chunk) and `+80 kB gz` (question-banks route chunk)
  budgets are met without splitting on every visit — the heavy deps live in
  their own lazy chunks.
- XSS surface is fully sanitized (DOMPurify + custom validator) and 20 unit
  tests cover the attack vectors.
- Backward compatibility: existing `{text: '...'}` envelopes still render
  through the same renderer; no data migration needed.

The spike is **production-shaped** in code structure (sanitization as a separate
layer, two-envelope detector, dynamic KaTeX) so the cost of upgrading to a
real "rich prompt" feature is mostly schema work, not a re-architecture.

## 1. Goal & scope

The current question bank editor accepts only a plain `<textarea>`. This spike
asks:

1. Can TipTap + KaTeX be added to the question bank editor **without** blowing
   the initial-chunk or question-banks route bundle budgets?
2. Can the resulting prompt envelopes still round-trip through the existing
   backend (`prompt: { text: '...' }`) without a DB migration?
3. Is the sanitization story strong enough to ship a teacher-controlled rich
   prompt that is later rendered to students in exam / review screens?

**In scope** (this spike):

- TipTap (StarterKit + Underline + Link) + KaTeX (math) in the question bank
  editor prompt field.
- Two-envelope detector (`{text}` legacy, `{doc}` rich).
- DOMPurify allowlist sanitization + custom validator for the rich doc shape.
- Backward-compat rendering for the legacy envelope.

**Out of scope** (deferred to follow-up slices):

- Rich text in any other field (choices, rubric, accepted-answers).
- Rich prompt rendering in the exam runner, attempt review, or grading detail
  pages (still uses raw `text`).
- DB migration to a typed `prompt_doc` column; we JSON-stringify the envelope
  into the existing `text` field for the spike.
- Media / image upload, collaborative editing, prompt search indexing.
- Full WCAG audit for the new editor (we only verified the toolbar uses
  `role="toolbar"` + `aria-pressed` on toggle buttons).

## 2. Dependencies added

All under `apps/web` — `dependencies` (production runtime):

```json
"@tiptap/core": "^2.10.3"
"@tiptap/extension-link": "^2.10.3"
"@tiptap/extension-underline": "^2.10.3"
"@tiptap/pm": "^2.10.3"
"@tiptap/react": "^2.10.3"
"@tiptap/starter-kit": "^2.10.3"
"dompurify": "^3.2.3"
"katex": "^0.16.11"
```

`devDependencies` (types only):

```json
"@types/dompurify": "^3.2.0"
"@types/katex": "^0.16.7"
```

`pnpm audit --prod` reports zero known vulnerabilities. The full-tree audit
shows pre-existing high-severity Vercel/undici issues in dev tooling, not
introduced by this spike.

`pnpm install --frozen-lockfile` runs cleanly (lockfile change is expected
because dependency metadata changed).

## 3. Bundle impact (vs. main baseline)

All sizes from `pnpm web:build` on the spike branch.

| Chunk                    | Pre-spike                | Post-spike                                             | Δ                                       | Budget              | Verdict     |
| ------------------------ | ------------------------ | ------------------------------------------------------ | --------------------------------------- | ------------------- | ----------- |
| `index-*.js` (initial)     | 360.77 kB / 114.49 kB gz | 360.81 kB / 114.53 kB gz                               | **+0.04 kB / +0.04 kB gz**              | ≤ +5 kB gz          | ✅ within    |
| `question-banks-page-*.js` | 6.61 kB / 2.23 kB gz     | 41.24 kB / 15.04 kB gz                                 | **+34.63 kB / +12.81 kB gz**            | ≤ +80 kB gz         | ✅ within    |
| `rich-prompt-editor-*.js`  | (n/a)                    | 335.27 kB / 106.57 kB gz *(lazy)*                      | n/a                                     | (deferred cost)     | ✅ opt-in    |
| `katex-*.js`               | (n/a)                    | 261.33 kB / 77.57 kB gz *(lazy)*                       | n/a                                     | (deferred cost)     | ✅ opt-in    |

The deferred-cost chunks are only loaded when:

1. A teacher toggles the form to "Rich text (spike)" mode → `rich-prompt-editor-*.js`
   loads.
2. A math placeholder is in the rendered prompt and the renderer mounts →
   `katex-*.js` loads.

If neither condition triggers, the question-banks page is exactly **+12.81 kB gz**
heavier than baseline (renderer + sanitize + lazy-loader).

**Important**: the *first* spike iteration that did not split these chunks
landed `question-banks-page-*.js` at **+196.22 kB gz** — over the budget by
~116 kB gz. The split we ship in this report is the one that meets the budget.

### Math vs. budgets

- Initial chunk: 0.8 % of the +5 kB gz envelope.
- Route chunk: 16 % of the +80 kB gz envelope.
- Opt-in cost: 184 kB gz total (editor + KaTeX) — only paid by teachers who
  insert a math formula.

## 4. XSS / sanitization

Threat model: a teacher (authenticated, trusted role) authors a prompt that
later renders to students in the exam runner / review pages. The attacker
controls the prompt text. Goal: prevent script execution, event handlers,
CSS exfiltration, javascript: URLs, raw `<math>` / `<svg>` from the rich
prompt.

Defense layers:

1. **Validator** (`isValidRichDoc`): walks the doc, refuses unknown node or
   mark types, clamps depth to 12 and total node count to 2 000, allows
   `link` marks only with `http(s)://` or `mailto:` href.
2. **Builder** (`renderNode`): every text node passes through
   `escapeHtml`; marks and link hrefs are encoded; math nodes emit a sanitized
   `<code data-math="...">` placeholder, never raw LaTeX.
3. **DOMPurify** (`renderRichPromptToHtml`): defense-in-depth pass with a
   locked-down allowlist of tags/attrs, no `style`, no event-handler attrs,
   and a hard `FORBID_TAGS: ['script','iframe','object','embed','form',
   'input','svg','math']`.
4. **KaTeX mounting** (`RichPromptRenderer`): KaTeX renders into a pre-existing
   `<code>` element via `katex.render(src, node, { throwOnError: false })`,
   never via `dangerouslySetInnerHTML` of unsanitized input.

### Tests

20 new vitest cases under `apps/web/src/shared/rich-prompt/__tests__/sanitize.test.ts`:

- `detectPrompt` (5): plain / rich / null / primitive / wrong-shape rejection.
- `isValidRichDoc` (7): minimal valid doc, non-doc root, unknown node type,
  unknown mark, text without `text`, javascript: link rejection,
  http/https/mailto link acceptance, depth clamp.
- `renderRichPromptToHtml` (8): script-tag escape in plain envelopes, script
  + onerror strip from rich docs, math placeholder, event handler strip,
  javascript: href strip in link marks, text-node HTML escape, full structural
  block (heading, lists, blockquote, code, hr, br).

Final test suite: **77 / 77 passing** (was 57 / 57).

The DOMPurify normalization caveat was hit during testing: `<hr/>` and
`<br/>` are emitted as `<hr>` and `<br>` after the sanitization pass. Tests
were adjusted to match the actual normalized output (the tag set is still
correctly bounded).

## 5. Backward compatibility

Two envelope shapes are accepted by the renderer and the form's preview:

1. **Legacy** `{ text: 'hello' }` — used by every question that exists today.
2. **Rich** `{ doc: { type: 'doc', content: [...] } }` — produced by TipTap.

The spike does **not** persist a typed `doc` column. Instead, the rich editor
serializes the doc to JSON and stores it in the existing `text` field. The
consumer side is unchanged; only the form's "rich" mode ever produces a
non-plain envelope, and the renderer auto-detects the envelope on read.

This means:

- Existing question bank data renders exactly as before.
- A question authored in "rich" mode will have a JSON-stringified `text` field
  on the next list/refetch, which will round-trip through the rich renderer.
- A follow-up slice can swap the persistence layer to a typed `prompt_doc`
  JSONB column with zero renderer changes (the envelope detector already
  accepts both shapes).

## 6. What ships in this branch

```
apps/web/package.json                                       (deps + devDeps)
apps/web/src/shared/rich-prompt/sanitize.ts                 (validator + builder + DOMPurify)
apps/web/src/shared/rich-prompt/rich-prompt-renderer.tsx    (envelope detector + KaTeX post-mount)
apps/web/src/shared/rich-prompt/rich-prompt-editor.tsx      (TipTap editor + math insertion)
apps/web/src/shared/rich-prompt/rich-prompt-editor.css      (toolbar + content styles)
apps/web/src/shared/rich-prompt/__tests__/sanitize.test.ts (20 cases)
apps/web/src/pages/question-banks/question-banks-page.tsx   (mode toggle, lazy editor, preview)
docs/frontend/frontend-technical-spec/spikes/rich-text-editor-spike.md (this report)
pnpm-lock.yaml                                              (expected — deps changed)
```

### Editor toolbar

Bold / Italic / Underline / Code / Bullet list / Ordered list / H1 / H2 /
Blockquote / Link / Undo / Redo + a dedicated LaTeX input ("Chèn công thức
LaTeX"). Each toolbar button is `aria-pressed` aware and `data-testid`-tagged
for the eventual E2E suite (`rich-prompt-bold`, `rich-prompt-math-input`,
`rich-prompt-math-insert`, etc.).

### Custom math node

A 30-line `MathInline` TipTap node (`apps/web/src/shared/rich-prompt/rich-prompt-editor.tsx`)
is the only place that emits the `mathInline` node type. It serializes to
`<code data-math="..." class="rich-math">` so the renderer post-mount can
hand off to KaTeX without any `dangerouslySetInnerHTML` of unsanitized
content.

## 7. E2E / test commands & results

| Command                | Result   |
| ---------------------- | -------- |
| `pnpm web:typecheck`   | clean    |
| `pnpm web:build`       | 24 chunks built; 2 new lazy chunks for the editor + KaTeX; budgets met |
| `pnpm web:test`        | 9 test files, **77/77 passing** (was 8/57) |
| `pnpm e2e:smoke`       | "Smoke passed" (full API flow including question bank creation) |
| `pnpm e2e:browser`     | **23/23** chromium (no e2e coverage added for the spike — the form is a new control surface, not in any existing test path) |
| `pnpm check`           | green (web typecheck + build, go test + vet + gofmt all clean) |
| `pnpm audit --prod`    | 0 known vulnerabilities |

### KaTeX / WebKit manual test

The WebKit project in `apps/web/playwright.config.ts` is gated behind
`PLAYWRIGHT_BROWSERS=1` and currently **cannot launch on the WSL2 host** —
the system is missing `libgtk-4.so.1`, `libgraphene-1.0.so.0`, `libxslt.so.1`,
`libevent-2.1.so.7`, `libopus.so.0`, `libgstallocators-1.0.so.0`, and a
handful of other WebKit-only system libraries. `playwright install --with-deps`
requires `sudo` which is not available in this environment. The
`scripts/e2e_browser_all.sh` probe (slice-14) already detects and gracefully
falls back when those deps are missing.

For the spike, KaTeX behavior on WebKit/Safari is therefore **unverified on
this host**. KaTeX is widely used in production on Safari, but the spike does
not provide an automated proof. A follow-up slice should add a WebKit
matrix run on a CI runner with the right system libs before any production
merge.

## 8. Open follow-ups (post GO)

1. **Persistence layer.** Replace the JSON-in-`text` hack with a typed
   `prompt_doc JSONB` column on `question_versions` and `attempt_items`.
2. **Renderer coverage.** Wire `<RichPromptRenderer>` into the exam runner,
   attempt review, and grading detail pages. Currently the spike is
   question-bank-editor-only.
3. **WCAG audit.** Add a full axe-core scan of the toolbar + LaTeX input.
   The toolbar has correct roles / pressed-state semantics, but contrast,
   focus order, and keyboard escape from the link prompt (`window.prompt`)
   need a real check.
4. **WebKit CI.** Get a CI runner with `libgtk-4` / `libgraphene` / `libxslt`
   etc. and rerun the Playwright matrix including a `rich-prompt.spec.ts`.
5. **TipTap Pro extensions.** We deliberately kept the surface to
   StarterKit + Underline + Link. Tables, code-block with syntax highlight,
   image, and callout would each add chunk weight and need their own budget
   review.
6. **CSP / Trusted Types.** If the platform ever ships a strict CSP, the
   sanitization path should also be reviewed against `require-trusted-types-for`.

## 9. Final go/no-go

**GO** — the spike shows that:

- Rich prompt editing can fit inside the question-banks route chunk
  (≤ +80 kB gz) **only** by splitting TipTap and KaTeX into their own lazy
  chunks. The 184 kB gz deferred cost is acceptable because the feature is
  opt-in (only teachers in the rich mode pay the cost).
- Sanitization is layered (validator + builder + DOMPurify + safe KaTeX
  mount), with 20 unit tests covering the practical XSS vectors and full
  backward compatibility for `{text: '...'}` envelopes.
- No backend change is required for the spike; a follow-up slice can lift
  the doc persistence into a real column when the team is ready to commit
  to a renderer migration across exam / review / grading screens.

**Hold** before production merge:

- KaTeX-on-Safari unverified on this host.
- A WCAG 2.1 AA pass on the new editor.

Both are scoped, follow-up work — they do not block the spike's verdict.
