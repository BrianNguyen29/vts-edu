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
