# Frontend Technical Specification — LMS & Online Assessment Platform

**Phiên bản:** 1.0  
**Ngày chốt:** 2026-06-29  
**Đối tượng triển khai:** Dự án solo, tối ưu chi phí giai đoạn đầu, có khả năng mở rộng theo tải thực tế  
**Frontend:** React SPA/PWA-ready, TypeScript, Vite, pnpm  
**Backend contract:** Go modular monolith, OpenAPI 3.1, JWT access token + rotating opaque refresh token  
**MVP demo hosting:** Vercel Hobby + Render Free + Supabase Free (see backend ADR-0005)

## 1. Mục đích

Thư mục này là nguồn đặc tả kỹ thuật chuẩn cho frontend của nền tảng LMS và kiểm tra trực tuyến. Tài liệu được viết để:

- Một lập trình viên solo có thể triển khai tuần tự mà không tạo thừa abstraction hoặc dependency.
- AI coding agents có đủ context để tạo page, route, component, query hook, form, test và tài liệu đúng kiến trúc.
- Frontend bám chặt API contract được sinh từ OpenAPI, không tự định nghĩa DTO trùng lặp.
- Luồng thi trực tuyến có cơ chế autosave, resume và submit đáng tin cậy.
- Authentication, authorization UX, accessibility và responsive layout có quy tắc thống nhất.
- Giao diện tham khảo được chuyển thành một hệ thống dashboard thực dụng, không ưu tiên hiệu ứng hơn nghiệp vụ.

## 2. Kiến trúc chốt

```text
Browser
  |
  | Static assets from Vercel Hobby + HTTPS REST/JSON cross-origin
  v
React 19 SPA / PWA-ready (installable ở Phase 2)
  |- React Router 8
  |- TanStack Query 5
  |- openapi-fetch + openapi-react-query
  |- React Hook Form + Zod
  |- Tailwind CSS 4 + shadcn/ui
  |- IndexedDB for exam durability
  |
  +--> Go API on Render Free /api/v1
  +--> Supabase Storage signed upload/download URLs
```

MVP demo hosting: Vercel Hobby (SPA) → Render Free (Go API + River in-process) → Supabase Free (Postgres + Storage).

### Quyết định chính

1. **SPA bằng React + Vite**, không chạy Node server ở production.
2. **Feature-based architecture lấy cảm hứng từ Feature-Sliced Design**, nhưng giản lược cho dự án solo.
3. **TanStack Query là nguồn quản lý server state**; không copy dữ liệu API vào Redux/Zustand.
4. **Không dùng Redux/Zustand trong MVP**; auth dùng memory store nhỏ, form dùng React Hook Form, URL dùng search params.
5. **Không dùng Axios**; dùng `openapi-fetch` trên native Fetch để đồng bộ type với OpenAPI.
6. **Access token chỉ lưu trong memory**; refresh token nằm trong cookie `HttpOnly` do backend quản lý.
7. **IndexedDB là nguồn durable phía client cho answer chưa được server xác nhận** trong bài thi.
8. **Server time quyết định thời gian thi**; timer trình duyệt chỉ để hiển thị.
9. **Không gọi API trực tiếp từ page/component trình bày**; đi qua feature query/mutation hooks.
10. **Mọi route và action nhạy cảm đều có permission metadata**, nhưng backend vẫn là nơi quyết định authorization cuối cùng.

## 3. Quan hệ với Backend Technical Specification

Frontend phải đọc và tuân theo các tài liệu backend sau:

- `docs/backend-technical-spec/06-api-conventions.md`
- `docs/backend-technical-spec/api/*.md`
- `docs/backend-technical-spec/08-core-functions.md`
- `docs/backend-technical-spec/09-security.md`
- `docs/backend-technical-spec/AGENTS.md`

Các invariant liên hệ trực tiếp:

- Question version và published assessment snapshot là immutable.
- Submit attempt là idempotent.
- Answer save dùng revision để chống ghi đè cũ.
- Điểm được biểu diễn bằng decimal string, không tính bằng binary floating point.
- Chỉ grade đã publish mới hiển thị cho học sinh.
- API error dùng `application/problem+json`.

## 4. Thứ tự đọc khuyến nghị

| Thứ tự | Tài liệu | Mục đích |
|---:|---|---|
| 1 | [00-project-scope.md](00-project-scope.md) | Phạm vi và invariant frontend |
| 2 | [01-tech-stack.md](01-tech-stack.md) | Công nghệ, dependency và version policy |
| 3 | [02-frontend-architecture.md](02-frontend-architecture.md) | Kiến trúc layer và dependency rules |
| 4 | [03-project-structure.md](03-project-structure.md) | Folder tree và naming conventions |
| 5 | [04-design-system-component-design.md](04-design-system-component-design.md) | Design tokens và component strategy |
| 6 | [05-routing-layouts.md](05-routing-layouts.md) | Router, protected routes và layouts |
| 7 | [routes/](routes/) | Route catalog theo vai trò |
| 8 | [06-state-management-dataflow.md](06-state-management-dataflow.md) | State taxonomy và dataflow |
| 9 | [07-api-client-auth.md](07-api-client-auth.md) | API client, token, refresh, errors |
| 10 | [08-forms-validation.md](08-forms-validation.md) | Form conventions và validation |
| 11 | [09-core-utilities.md](09-core-utilities.md) | Utilities toàn cục |
| 12 | [10-exam-runtime-frontend.md](10-exam-runtime-frontend.md) | Luồng thi, autosave và IndexedDB |
| 13 | [features/](features/) | Đặc tả theo feature |
| 14 | [11-accessibility-responsive.md](11-accessibility-responsive.md) | WCAG và responsive |
| 15 | [12-testing-strategy.md](12-testing-strategy.md) | Unit/component/E2E |
| 16 | [13-performance-pwa.md](13-performance-pwa.md) | Performance budget và PWA |
| 17 | [14-security.md](14-security.md) | Frontend security controls |
| 18 | [15-observability-deployment.md](15-observability-deployment.md) | Telemetry và deployment |
| 19 | [16-ai-agent-guide.md](16-ai-agent-guide.md) | Quy tắc cho AI agents |
| 20 | [17-implementation-roadmap.md](17-implementation-roadmap.md) | Lộ trình triển khai solo |

## 5. Route catalogs

- [Public & Authentication Routes](routes/01-public-auth.md)
- [Student Routes](routes/02-student.md)
- [Teacher Routes](routes/03-teacher.md)
- [Administrator Routes](routes/04-admin.md)

## 6. Feature specifications

- [Authentication](features/01-auth.md)
- [Classes & Resources](features/02-classes-resources.md)
- [Question Bank](features/03-question-bank.md)
- [Assessment Builder](features/04-assessment-builder.md)
- [Exam Runner](features/05-exam-runner.md)
- [Assignments & Gradebook](features/06-assignments-gradebook.md)
- [Notifications & Dashboard](features/07-notifications-dashboard.md)

## 7. Architectural Decision Records

- [ADR-0001: React SPA with Vite](adr/0001-react-vite-spa.md)
- [ADR-0002: Simplified Feature-Sliced Architecture](adr/0002-simplified-feature-sliced.md)
- [ADR-0003: TanStack Query for Server State](adr/0003-tanstack-query-server-state.md)
- [ADR-0004: In-memory Access Token](adr/0004-auth-token-storage.md)
- [ADR-0005: Native Fetch + OpenAPI Client](adr/0005-fetch-openapi-client.md)
- [ADR-0006: IndexedDB for Exam Draft Durability](adr/0006-indexeddb-exam-drafts.md)

## 8. Definition of Done cấp hệ thống

Một feature frontend chỉ được coi là hoàn thành khi:

- Route và permission metadata đã được khai báo.
- API type được lấy từ generated OpenAPI package; không tự tạo DTO trùng.
- Có loading, empty, error và permission-denied states.
- Form có client validation, server error mapping và accessible error summary.
- Query/mutation có cache policy và invalidation rõ.
- Có unit hoặc component test cho logic quan trọng.
- Có Playwright test cho happy path nếu feature thuộc luồng lõi.
- Keyboard navigation và focus flow đã được kiểm tra.
- Mobile width 360px và desktop width 1440px không vỡ layout.
- Không lưu token hoặc dữ liệu nhạy cảm trong `localStorage`.
- Không log request body chứa câu trả lời, điểm hoặc PII.
- Không sửa file generated bằng tay.
- Tài liệu feature và route liên quan được cập nhật.

## 9. Tệp dành cho AI agents

AI agent phải đọc [AGENTS.md](AGENTS.md) trước khi thay đổi code. Template nhận task nằm trong [templates/agent-task-template.md](templates/agent-task-template.md).
