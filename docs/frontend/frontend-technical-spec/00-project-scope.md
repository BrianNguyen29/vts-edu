# 00. Project Scope & Frontend Invariants

## 1. Product context

Nền tảng phục vụ hai actor chính trong MVP:

- **Giáo viên:** quản lý lớp, học sinh, học liệu, ngân hàng câu hỏi, đề kiểm tra, bài tập và bảng điểm.
- **Học sinh:** truy cập bằng account được cấp, xem học liệu, làm bài tập, thi trực tuyến và xem kết quả đã công bố.

Actor nền tảng:

- **Organization administrator:** quản lý tài khoản, học kỳ, lớp, import và audit.

Actor sau MVP:

- Teaching assistant.
- Parent/guardian.
- System operator.

## 2. Ràng buộc dự án

| Ràng buộc | Hệ quả frontend |
|---|---|
| Dự án solo | Giảm dependency, tránh abstraction không tạo giá trị trực tiếp |
| Tối ưu chi phí | SPA static, một codebase, không Node production server |
| Backend Go | Contract qua OpenAPI, không chia sẻ runtime model trực tiếp |
| pnpm workspace | Generated API package và frontend nằm cùng monorepo |
| Pilot 100–300 người thi đồng thời | UI phải giảm request thừa và có autosave điều tiết |
| Học sinh có thể dùng thiết bị yếu | Bundle, animation và chart phải có budget |
| Dữ liệu trẻ em và điểm | Không lưu PII/token bừa bãi; telemetry phải tối thiểu |
| Mạng có thể chập chờn | Answer pending phải tồn tại qua reload bằng IndexedDB |

## 3. MVP frontend scope

| Module | MVP | Ghi chú |
|---|:---:|---|
| Authentication & session bootstrap | Có | Login, refresh, logout, change password |
| Role/permission-aware navigation | Có | Admin, teacher, student |
| Dashboard cơ bản | Có | Việc cần làm, thông báo, trạng thái gần đây |
| Class & enrollment views | Có | Danh sách, chi tiết, học sinh |
| Resources & file access | Có | Upload intent, signed URL, preview cơ bản |
| Question bank | Có | CRUD logical question + immutable versions |
| Assessment builder | Có | Fixed questions, sections, shuffle, schedule |
| Exam runner | Có | Start, autosave, resume, submit, expiry |
| Manual/auto grading UI | Có | Review tự luận và kết quả |
| Assignment/submission | Có | Text/file submission |
| Gradebook | Có, cơ bản | View/edit/publish/export |
| Notifications | Có, in-app | Email không phải concern trực tiếp của UI |
| PWA installability | Có thể sau core | Không chặn pilot |
| AI assistant | Không | Phase sau |
| Gamification | Không | Phase sau |
| Parent portal | Không | Phase sau |

## 4. Out of scope giai đoạn đầu

- Server-side rendering và SEO application pages.
- Native Android/iOS.
- Livestream hoặc video conference tự xây.
- Webcam proctoring.
- Realtime chat.
- Full offline LMS.
- Redux, Zustand hoặc event bus toàn cục nếu chưa có nhu cầu đo được.
- Micro-frontend.
- Module federation.
- Storybook bắt buộc; chỉ thêm khi design system đủ lớn.
- AI tạo/chấm nội dung chính thức.
- Dashboard animation phức tạp giống ảnh tham khảo.

## 5. Non-negotiable frontend invariants

### 5.1. API contract

- Request/response type protected API phải sinh từ OpenAPI.
- Không copy-paste interface API vào feature.
- Không dùng `any` cho response hoặc error payload.
- Generated code không được sửa bằng tay.

### 5.2. Authentication

- Access JWT chỉ tồn tại trong memory.
- Refresh token chỉ nằm trong cookie `HttpOnly`; JavaScript không đọc được.
- Reload trang phải bootstrap bằng refresh endpoint.
- Logout phải xóa memory token, query cache, exam state của user và phát tín hiệu sang tab khác.
- Không lưu token trong `localStorage`, `sessionStorage` hoặc IndexedDB.

### 5.3. Authorization

- Route/menu/button có thể ẩn hoặc disable theo permission để cải thiện UX.
- Frontend không được coi kiểm tra permission là security boundary.
- 403 từ backend luôn được xử lý như nguồn sự thật.
- Không suy ra quyền chỉ từ role name nếu backend trả permission list.

### 5.4. Attempt integrity

- Server time là nguồn quyết định expiry.
- Answer chỉ được coi là đã lưu khi server trả 2xx và revision mới.
- Answer chưa được xác nhận phải được giữ trong IndexedDB.
- Submit phải dùng idempotency key và không tạo hai lần nộp.
- Page reload hoặc browser crash không được làm mất answer pending.
- Client không tự final score.

### 5.5. Score integrity

- Điểm nhận từ API được giữ dạng decimal string.
- Không dùng `parseFloat` để cộng, nhân hoặc làm tròn điểm nghiệp vụ.
- Frontend chỉ format điểm; backend chịu trách nhiệm tính final score.
- Học sinh chỉ thấy grade đã publish theo response backend.

### 5.6. Accessibility

- Luồng lõi phải đạt WCAG 2.2 AA ở mức thiết kế và kiểm thử.
- Mọi thao tác chính dùng được bằng bàn phím.
- Focus không bị mất sau navigation, modal hoặc validation failure.
- Không truyền đạt trạng thái chỉ bằng màu.
- Exam timer phải có thông báo tiếp cận được nhưng không spam screen reader.

### 5.7. Performance

- Page route phải lazy-load.
- Editor, chart, PDF preview và exam runner không nằm trong initial bundle chung nếu không cần.
- Không polling dày nếu không có requirement.
- Không render toàn bộ bảng hàng nghìn dòng; pagination hoặc virtualization theo dữ liệu.

### 5.8. Security & privacy

- Rich text render phải sanitize defense-in-depth.
- Không chèn raw HTML từ API bằng `dangerouslySetInnerHTML` nếu chưa sanitize.
- Không log answer content, grade content, access token hoặc PII.
- URL không chứa token, password hoặc dữ liệu nhạy cảm.
- Signed URL không được lưu lâu dài trong persistent state.

## 6. Giao diện tham khảo: nguyên tắc chuyển đổi

Ảnh tham khảo là dashboard học sinh có sidebar, hero, KPI, chart, lộ trình, AI panel và gamification. MVP chỉ giữ:

- App shell với sidebar responsive.
- Global search placeholder hoặc quick navigation.
- “Việc cần làm hôm nay”.
- Bài tập và bài kiểm tra sắp đến hạn.
- Điểm mới công bố.
- Thông báo gần đây.
- Tiến độ hoàn thành đơn giản.

Tạm hoãn:

- Hero animation lớn.
- AI panel cố định.
- Radar năng lực.
- XP, huy hiệu, streak.
- Bảng xếp hạng lớp.

## 7. Success criteria cho pilot

| Chỉ tiêu | Mục tiêu |
|---|---:|
| Initial JS gzip cho app shell | ≤ 250 KB, không gồm route lazy chunks |
| Route interaction p95 trên thiết bị trung bình | < 200 ms sau khi data đã có |
| Largest Contentful Paint dashboard | < 2,5 giây trong điều kiện pilot |
| Autosave acknowledgement p95 | Phù hợp backend target < 800 ms |
| Answer đã được server xác nhận bị mất | 0 |
| E2E happy path lõi | Pass trên Chromium, WebKit, Firefox |
| Critical accessibility issue | 0 trên luồng lõi |
| Token trong persistent browser storage | 0 |
