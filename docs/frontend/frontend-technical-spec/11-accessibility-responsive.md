# 11. Accessibility & Responsive Design

## 1. Accessibility target

Mục tiêu là **WCAG 2.2 AA** cho các luồng cốt lõi:

- Login và đổi mật khẩu.
- Điều hướng workspace.
- Xem lớp/học liệu.
- Tạo câu hỏi và đề kiểm tra.
- Làm bài thi.
- Nộp bài tập.
- Xem/chỉnh bảng điểm.

Automated tests không thay thế manual keyboard/screen-reader review.

## 2. Semantic structure

Mỗi page phải có:

```text
header/banner (nếu cần)
nav
main
  h1 duy nhất
  section với heading theo thứ tự
aside chỉ khi nội dung bổ trợ
```

- Không dùng `div` clickable thay button/link.
- Link dùng cho navigation; button dùng cho action.
- Bảng dữ liệu dùng table semantic khi đúng bản chất.

## 3. Focus management

- Focus visible rõ trên mọi control.
- Sau route navigation, focus vào page heading hoặc main.
- Dialog open focus phần tử đầu tiên phù hợp; close trả trigger.
- Validation failure focus error summary/field đầu tiên.
- Toast không giành focus.
- Không dùng `outline: none` nếu không thay bằng focus ring tương đương.

## 4. Keyboard

- Sidebar, menu, dropdown, tab và dialog dùng keyboard theo pattern chuẩn.
- Data table action không chỉ xuất hiện khi hover.
- Gradebook hỗ trợ arrow/tab navigation theo thiết kế, nhưng không phá hành vi browser cơ bản.
- Exam shortcut chỉ thêm nếu không xung đột input và có hướng dẫn.

## 5. Color and contrast

- Text/interactive contrast theo AA.
- Status `success/warning/danger` có text/icon, không chỉ màu.
- Chart có legend/labels và table/text alternative.
- Disabled state vẫn đọc được nhưng phân biệt rõ.

## 6. Forms

- Label visible cho field nghiệp vụ.
- Placeholder không thay label.
- Error message liên kết bằng `aria-describedby`.
- Required state có semantics.
- Help text ngắn, không ẩn toàn bộ trong tooltip.
- File dropzone có button input fallback.

## 7. Live regions

Dùng hạn chế:

| Event | Live region |
|---|---|
| Form validation summary | assertive một lần |
| Exam save status | polite, debounce |
| Timer warning | polite/announced theo mốc |
| Upload complete | polite |
| Background notification badge | Không announce liên tục |

Không announce mỗi giây timer.

## 8. Rich text and math

- Heading/list/table semantics được giữ.
- Image cần alt text hoặc decorative marker.
- KaTeX output phải có MathML/accessibility representation nếu khả dụng.
- Link text có nghĩa, không chỉ “bấm vào đây”.
- Teacher preview phải phản ánh student accessible rendering.

## 9. Responsive breakpoints

Không xem breakpoint là thiết bị cụ thể; dùng theo layout cần thiết.

| Range | Layout |
|---|---|
| `< 640px` | Single column, bottom/compact navigation, full-width dialog |
| `640–1023px` | 1–2 columns, sidebar drawer |
| `1024–1439px` | Collapsed sidebar, main content rộng |
| `≥ 1440px` | Full sidebar, optional contextual panel |

## 10. App shell behavior

### Desktop

- Sidebar 240–272px.
- Header sticky nếu không che focus/anchor.
- Main max-width theo page type; gradebook/editor có full bleed.

### Mobile

- Sidebar thành sheet.
- Quick actions không quá 4–5 mục.
- Breadcrumb rút gọn.
- Table chuyển card/list hoặc horizontal scroll có hướng dẫn.

## 11. Exam responsive

- Mobile vẫn hiển thị timer và save status.
- Question navigator thành drawer/bottom sheet.
- Submit không bị che bởi browser UI.
- Choice touch target lớn.
- Essay editor không dùng fixed height quá nhỏ.
- Không yêu cầu landscape trừ dạng câu đặc biệt chưa thuộc MVP.

## 12. Reduced motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    scroll-behavior: auto !important;
    transition-duration: 0.01ms !important;
  }
}
```

Không dùng motion để báo trạng thái duy nhất.

## 13. Zoom and text resize

- Layout dùng được ở 200% zoom.
- Không khóa viewport zoom.
- Không dùng fixed pixel height cho text container.
- Horizontal scroll chỉ ở vùng bảng/editor cần thiết, không toàn page.

## 14. Accessibility testing

### Automated

- `eslint-plugin-jsx-a11y`.
- axe trong component/E2E tests.
- Playwright scan các route lõi.

### Manual

- Keyboard-only.
- Screen reader smoke test trên ít nhất NVDA/VoiceOver theo môi trường có sẵn.
- 200% zoom.
- High contrast/reduced motion.
- Mobile screen reader cho exam nếu pilot có nhu cầu.

## 15. A11y Definition of Done

- Không có axe critical/serious issue được chấp nhận không lý do.
- Keyboard hoàn thành được workflow.
- Focus order logic.
- Error/status được announced đúng.
- Text alternative cho chart và non-text content.
- Route title/lang được cập nhật.

## 16. Recent improvements (slice-8-accessibility-audit, 2026-07-01)

Đợt kiểm tra này tập trung vào các luồng cốt lõi (login, đổi mật khẩu, dashboards, builder, exam, gradebook, resources, error) với những thay đổi sau:

- **Skip link + landmark**: thêm skip link "Bỏ qua đến nội dung chính" trong `AuthLayout`, `AppShellLayout`, `ExamLayout`; cả ba đều có `<main id="main-content" tabIndex={-1}>` để skip link đặt focus đúng vị trí.
- **Focus ring**: chuyển từ `:focus` (hiển thị cả với mouse) sang `:focus-visible` (chỉ bàn phím). Áp dụng cho `button`, `a`, `select`, `[role="tab"]`, `[role="button"]`, `input`, `textarea`. Mouse click không còn để lại outline xanh.
- **Trạng thái lưu/loading**: các vùng loading có `role="status" aria-live="polite"`; banner lỗi có `role="alert" aria-live="assertive"`; submit button có `aria-busy`.
- **Form error association**: form login/change-password dùng `aria-describedby` trỏ tới banner lỗi; hint mật khẩu (`PasswordPolicyHints`) nhận `id` và được liên kết với input.
- **Skip link / focus restoration cho dialog**: preview dialog của assessment builder chuyển focus vào nút Đóng khi mở và trả focus về phần tử kích hoạt khi đóng.
- **Tab semantics**: admin tabs và gradebook tabs dùng `role="tablist"` + `role="tab"` + `role="tabpanel"` với `aria-selected`, `aria-controls`, `tabIndex` roving.
- **Table caption**: tất cả bảng dữ liệu (users, audit logs, gradebook assessment, gradebook class, resources, publication history, roster, terms, subjects, courses, classes, bulk import preview) đều có `<caption>` (thường là visually-hidden vì heading đã đặt tên bảng).
- **Nhãn cho search input**: các ô `<input type="search">` trên teacher dashboard, admin user list, audit actor filter, user picker trong academic management đều có label (visually-hidden nếu chỉ dùng placeholder).
- **Trạng thái đọc được**: status badge có `aria-label` mô tả (ví dụ "Trạng thái SUBMITTED"); decorative "·" được `aria-hidden`.
- **Route title**: hook `useDocumentTitle` đặt tab title cho tất cả trang; restore title cũ khi unmount.
- **Reduced motion**: thêm `@media (prefers-reduced-motion: reduce)` để tắt transition/animation.
- **Heading hierarchy**: chỉnh lại một số heading (vd. card title trong dashboard từ h2 xuống h3) để tránh skip level.
- **A11y smoke e2e**: thêm `apps/web/e2e/a11y.spec.ts` (10 case) kiểm tra skip link, landmark, h1, role của error/alert, tab semantics, table caption, form description, accessible name.

### Limitations / known gaps

- **Chưa tích hợp axe-core**: dùng Playwright locators thay vì `@axe-core/playwright` để giữ dependency footprint tối thiểu. Smoke spec kiểm tra các tiêu chí axe thường gặp cho luồng cốt lõi nhưng không thay thế axe scan đầy đủ.
- **Chưa có manual screen reader / keyboard review**: chưa chạy NVDA/VoiceOver thực tế, chưa test với 200% zoom trên production build, chưa test high-contrast mode.
- **Tiêu điểm cho mobile**: chưa test screen reader trên thiết bị di động, chưa test landscape/portrait exam UI.
- **Dialog focus trap**: preview dialog hiện tại chuyển focus đến close button nhưng chưa có focus trap đầy đủ (Tab có thể thoát ra ngoài panel).
- **PWA / offline status**: chưa có a11y cho thông báo background sync / push notification.
- **Charts / non-text content**: chưa có chart/canvas nào, nên chưa áp dụng text alternative.
- **Dark mode / high contrast theme**: chưa có theme, chưa test contrast ratio tự động.

