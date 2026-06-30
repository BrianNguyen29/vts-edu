# 01. Tech Stack & Libraries

## 1. Tóm tắt lựa chọn

| Nhóm | Công nghệ | MVP | Lý do |
|---|---|:---:|---|
| Runtime tooling | Node.js 24 LTS | Có | LTS hiện hành, tương thích pnpm 11, Vite 8 và React Router 8 |
| Package manager | pnpm 11.x | Có | Workspace tích hợp, tiết kiệm disk, security defaults tốt |
| UI library | React 19.2 | Có | Component model trưởng thành, ecosystem lớn |
| Language | TypeScript strict | Có | Giảm lỗi contract, refactor an toàn |
| Build tool | Vite 8 | Có | Dev/build nhanh, SPA static, code splitting |
| Router | React Router 8 | Có | Nested layouts, lazy routes, error boundaries, data APIs |
| Server state | TanStack Query 5 | Có | Cache, request lifecycle, mutation và invalidation |
| API client | `openapi-fetch` | Có | Typed native Fetch client từ OpenAPI, runtime nhỏ |
| Query adapter | `openapi-react-query` | Có | Liên kết OpenAPI với TanStack Query, giảm boilerplate |
| Forms | React Hook Form | Có | Hiệu quả với form lớn, ít re-render |
| Validation | Zod 4 | Có | Type-safe schema, phù hợp form và runtime validation |
| Styling | Tailwind CSS 4.3 | Có | Tốc độ triển khai cao, token hóa qua CSS variables |
| Component source | shadcn/ui, Radix primitives | Có | Accessible primitives, copy source và tự sở hữu code |
| Icons | Lucide React | Có | Nhất quán, tree-shakeable |
| Charts | Recharts 3 | Hạn chế | Đủ line/bar/radar cơ bản, tích hợp shadcn chart |
| Rich text | TipTap OSS | Có ở editor | ProseMirror-based, extension model rõ |
| Math rendering | KaTeX | Có | Render công thức nhanh, không cần server |
| Content sanitization | DOMPurify | Có | Defense-in-depth khi render HTML |
| Date/time | `date-fns` | Có | Function-based, tree-shakeable |
| Decimal display | String + formatter riêng | Có | Không tính điểm bằng float phía client |
| i18n | i18next + react-i18next | Có, vi trước | Namespace, lazy loading và TypeScript support |
| Persistent local DB | `idb` | Có cho exam | Wrapper nhỏ cho IndexedDB |
| Toast | Sonner hoặc component tương đương | Có | Feedback không chặn flow |
| Unit/component tests | Vitest 4 + Testing Library | Có | Cùng pipeline Vite, test theo hành vi |
| API mocking | MSW | Có | Mock ở network boundary cho browser và Node tests |
| E2E | Playwright | Có | Chromium/WebKit/Firefox, fixtures và tracing |
| Lint | ESLint 9 flat config | Có | Rules cho TS, React Hooks, a11y và tests |
| Formatting | Prettier | Có | Format ổn định cho người và AI agents |
| Global state library | Không dùng | Không | Chưa có use case biện minh Redux/Zustand |
| HTTP library | Không dùng Axios | Không | Native Fetch + OpenAPI đủ nhu cầu và ít dependency |
| PWA plugin | `vite-plugin-pwa` | Phase sau | Chỉ thêm khi core exam ổn định |

## 2. Version policy

### 2.1. Pinning

- `packageManager` trong root `package.json` pin pnpm exact major/minor phù hợp.
- `engines.node` yêu cầu Node `>=24 <25` trong giai đoạn đầu.
- Dependency runtime dùng caret trong cùng major khi dự án ổn định; lockfile là nguồn chính cho CI.
- Vite, React Router, Tailwind và TypeScript upgrade trong PR riêng.
- Không tự động merge major upgrade.
- CI chạy `pnpm install --frozen-lockfile`.

### 2.2. TypeScript

Dùng `strict: true` và các tùy chọn:

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitOverride": true,
    "useUnknownInCatchVariables": true,
    "noFallthroughCasesInSwitch": true,
    "verbatimModuleSyntax": true
  }
}
```

Không dùng:

- `skipLibCheck: false` như một mục tiêu bắt buộc nếu dependency gây cản trở; có thể giữ `true` nhưng không được che lỗi code ứng dụng.
- `any` để vượt type errors.
- Type assertion `as` liên tục thay cho validation.

## 3. React + Vite thay vì Next.js

### Lý do

- LMS chủ yếu nằm sau đăng nhập, SEO không phải yêu cầu chính.
- Frontend có thể build thành static assets và được Go hoặc CDN phục vụ.
- Không cần vận hành Node server production.
- Triển khai và rollback đơn giản.
- Route-level code splitting đủ cho dashboard và editor.

### Hệ quả

- Meta tags public marketing không phải trọng tâm.
- Runtime configuration cần endpoint hoặc file cấu hình do Go cung cấp.
- Authentication bootstrap diễn ra phía client.
- Không dùng server components hoặc server actions.

## 4. React Router 8

Chọn Declarative/Data Router bằng `createBrowserRouter` để có:

- Nested layouts.
- Lazy route modules.
- `errorElement` theo boundary.
- Loader cho session/permission gate và route preconditions nhẹ.
- Navigation blocker cho form chưa lưu.
- `handle` metadata cho title, breadcrumb và permission.

Không dùng loader để thay toàn bộ TanStack Query. Server state vẫn do Query quản lý; loader chỉ bootstrap hoặc `ensureQueryData` ở route cần thiết.

## 5. TanStack Query

TanStack Query quản lý:

- Query cache.
- Dedup request.
- Mutation lifecycle.
- Retry policy.
- Invalidation.
- Background refetch có kiểm soát.

Không dùng Query cache cho:

- Access token.
- Form draft chưa submit.
- Exam answer pending chưa được server xác nhận.
- UI modal state.

## 6. API client: openapi-fetch

Luồng sinh contract:

```text
Go Huma routes/types
      -> openapi.json
      -> openapi-typescript
      -> packages/api-client/src/generated/schema.d.ts
      -> openapi-fetch client
      -> openapi-react-query hooks/options
```

Lợi ích:

- URL, query params, body và response được type-check.
- Không cần tự viết generic `ApiResponse<T>`.
- Giảm drift giữa Go và TypeScript.
- Có middleware để gắn bearer token và xử lý auth.

## 7. Không dùng Axios

Native Fetch đáp ứng:

- Abort qua `AbortSignal`.
- `credentials: include` cho refresh cookie.
- Streaming/download cơ bản.
- Tích hợp service worker và browser platform.

Riêng upload cần progress có thể dùng wrapper `XMLHttpRequest` nhỏ. Không thêm Axios chỉ vì một use case.

## 8. State management

Không chọn Redux/Zustand ở MVP vì:

- Server state đã do TanStack Query quản lý.
- Form state do React Hook Form quản lý.
- Filter/sort/page thuộc URL.
- Auth chỉ cần memory store nhỏ.
- Theme/locale/sidebar dùng context hoặc persistent preference hook.
- Exam pending queue dùng IndexedDB repository riêng.

Điều kiện thêm Zustand sau này:

- Có state client phức tạp dùng chung qua nhiều route.
- State có nhiều transition cần devtools.
- Context gây re-render đo được.
- Có ADR mô tả use case cụ thể.

## 9. Styling và component source

### Tailwind CSS

- Theme qua CSS custom properties.
- Không rải màu hex trong feature component.
- Không tạo utility class tự phát nếu token có thể tái sử dụng.
- Class composition dùng `cn()` với `clsx` và `tailwind-merge`.

### shadcn/ui

shadcn/ui được dùng như nguồn component code, không phải black-box dependency. Quy tắc:

- Component được copy vào `shared/ui`.
- Mọi chỉnh sửa thuộc repository.
- Không import trực tiếp Radix primitives rải rác ngoài `shared/ui` nếu đã có wrapper.
- A11y behavior của primitive không được phá khi styling.

## 10. Rich text và công thức

### Canonical format

- Editor dùng TipTap JSON làm form state.
- Backend quyết định format canonical cuối cùng theo API contract.
- Không lưu base64 image trong editor content.
- Image/file reference dùng file ID và signed URL tạm thời.

### Rendering

- KaTeX render công thức.
- HTML fallback phải sanitize bằng DOMPurify.
- Link scheme chỉ cho `http`, `https`, `mailto` nếu policy cho phép.

## 11. Testing stack

| Test | Công cụ | Mục tiêu |
|---|---|---|
| Pure logic | Vitest | Permission, timer, queue, formatters |
| Component | Testing Library | Hành vi người dùng, keyboard, validation |
| API integration | MSW | Query/mutation/error states |
| E2E | Playwright | Luồng login, thi, chấm, gradebook |
| Accessibility | axe integration + manual | Lỗi tự động và focus flow |
| Visual smoke | Playwright screenshot hạn chế | App shell và exam layout |

## 12. Dependency policy

Trước khi thêm package, phải trả lời:

1. Platform API hoặc dependency hiện tại có giải quyết được không?
2. Package có maintained và ESM-compatible không?
3. Bundle cost là bao nhiêu?
4. Có cần chạy install script không?
5. Có gây lock-in dữ liệu không?
6. Có thể thay bằng module dưới 100 dòng không?
7. Test và migration path là gì?

Dependency mới có ảnh hưởng kiến trúc phải có ADR.
