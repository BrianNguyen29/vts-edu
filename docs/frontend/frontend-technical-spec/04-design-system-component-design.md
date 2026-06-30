# 04. Design System & Component Design

## 1. Mục tiêu

Thiết kế hệ thống component phải:

- Cho phép dựng dashboard gần tinh thần ảnh tham khảo nhưng thực dụng hơn.
- Tái sử dụng giữa student, teacher và admin workspace.
- Có accessibility mặc định.
- Không khóa vào package UI black-box.
- Cho AI agents một vocabulary thống nhất.

## 2. Design token layers

```text
primitive tokens
  -> semantic tokens
      -> component tokens
          -> feature-specific composition
```

### Primitive tokens

- Color scale.
- Spacing scale.
- Radius.
- Shadow.
- Typography scale.
- Motion duration/easing.

### Semantic tokens

```css
:root {
  --background: ...;
  --foreground: ...;
  --surface: ...;
  --surface-muted: ...;
  --primary: ...;
  --primary-foreground: ...;
  --success: ...;
  --warning: ...;
  --danger: ...;
  --border: ...;
  --focus-ring: ...;
}
```

Không dùng màu như `purple-600` trực tiếp cho trạng thái nghiệp vụ; dùng semantic token.

## 3. Component categories

| Category | Location | Ví dụ |
|---|---|---|
| UI primitives | `shared/ui` | Button, Input, Dialog, Tabs |
| Generic composites | `shared/ui` | DataTable, EmptyState, FileDropzone |
| Domain components | `entities/*/ui` | QuestionCard, GradeBadge |
| User-action components | `features/*/ui` | PublishButton, LoginForm |
| Large page widgets | `widgets/*` | GradebookGrid, UpcomingWorkPanel |
| Page composition | `pages/*` | TeacherDashboardPage |

## 4. shadcn/ui policy

- Chọn Radix-backed primitives cho Dialog, Popover, Dropdown, Tabs và Select.
- Copy source vào repository.
- Chuẩn hóa variant bằng `class-variance-authority`.
- Không sửa accessibility primitives bằng cách loại bỏ role, label hoặc focus trap.
- Component shared không chứa copy tiếng Việt; text truyền qua props hoặc i18n.

## 5. Component API principles

### Ưu tiên composition

```tsx
<Card>
  <CardHeader>
    <CardTitle>...</CardTitle>
  </CardHeader>
  <CardContent>...</CardContent>
</Card>
```

Thay vì component có quá nhiều boolean props:

```tsx
// Không nên
<Card compact blue shadowless teacher />
```

### Variant rõ nghĩa

```ts
type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
type ButtonSize = 'sm' | 'md' | 'lg' | 'icon';
```

### Controlled/uncontrolled

- Form components dùng controlled adapter với React Hook Form.
- Dialog generic có thể controlled hoặc uncontrolled.
- Business workflow dialog nên controlled bởi feature.

## 6. Component state contract

Mọi component data-driven phải có chiến lược cho:

- Loading.
- Empty.
- Error.
- Partial/stale.
- Read-only/disabled.
- Permission denied.

Không dùng một spinner giữa màn hình cho mọi tình huống.

## 7. Forms

Component chuẩn:

```text
FormField
  |- Label
  |- Control
  |- Description (optional)
  |- ErrorMessage
```

Yêu cầu:

- `id`, `htmlFor`, `aria-describedby`, `aria-invalid` đúng.
- Error summary ở đầu form khi submit fail.
- Focus field lỗi đầu tiên.
- Required state có text/semantics, không chỉ dấu sao màu đỏ.

## 8. Tables và grids

### DataTable dùng cho

- User list.
- Question list.
- Assignment submissions.
- Audit logs.

### GradebookGrid riêng

Gradebook không nên ép vào DataTable generic vì có:

- Sticky student column.
- Horizontal scroll.
- Cell edit.
- Keyboard navigation.
- Publish state.
- Decimal formatting.
- Large data windowing sau này.

## 9. Charts

MVP chỉ dùng chart khi trả lời câu hỏi cụ thể:

- Tiến độ hoàn thành theo tuần.
- Phân bố điểm.
- Tỷ lệ nộp bài.

Quy tắc:

- Có text/table alternative.
- Tooltip dùng keyboard nếu thư viện hỗ trợ; nếu không, dữ liệu quan trọng phải có dạng text.
- Không dùng animation dài.
- Color palette có contrast và không chỉ dựa vào hue.
- Lazy-load chart library ở route dashboard/analytics.

## 10. Dashboard theo ảnh tham khảo

### Desktop ≥ 1440px

```text
[Sidebar 256] [Main flexible] [Optional contextual panel 320]
```

### Laptop 1024–1439px

```text
[Collapsed sidebar 72] [Main flexible]
Contextual panel -> drawer
```

### Tablet/mobile

- Sidebar -> sheet/drawer.
- Cards -> 1–2 columns.
- Header actions rút gọn.
- AI panel không tồn tại ở MVP.

### Ưu tiên nội dung

1. Việc sắp đến hạn.
2. Continue/Resume.
3. Điểm và feedback mới.
4. Thông báo.
5. Thống kê đơn giản.

## 11. Icons và imagery

- Lucide cho functional icons.
- Icon luôn có accessible label hoặc `aria-hidden` nếu text đã mô tả.
- Illustration/robot chỉ là decorative và phải lazy-load.
- Không phụ thuộc illustration để truyền tải status.

## 12. Motion

- Duration 120–220ms cho UI transition.
- Tôn trọng `prefers-reduced-motion`.
- Không animation liên tục trên exam page.
- Không thay đổi layout gây mất focus.

## 13. Responsive component rules

- Component tự co giãn theo container, không biết toàn bộ viewport nếu không cần.
- Dùng CSS grid/flex và container queries khi có giá trị.
- Không hardcode chiều cao card theo ảnh mockup.
- Touch target tối thiểu 44×44 CSS px cho hành động chính trên mobile.

## 14. Example domain component

```tsx
interface AssignmentStatusCardProps {
  title: string;
  dueAt: string;
  status: 'not_started' | 'draft' | 'submitted' | 'late' | 'graded';
  href: string;
  score?: string | null;
}
```

Component chỉ trình bày. Việc quyết định học sinh có được submit lại hay không thuộc feature/service response.

## 15. Component review checklist

- Tên component phản ánh trách nhiệm.
- Props không lộ API response nguyên khối nếu chỉ cần vài field.
- Không fetch data trong shared UI.
- Không hardcode permission.
- Không hardcode route string.
- Có keyboard/focus behavior.
- Có test cho interaction quan trọng.
- Không tạo abstraction chỉ dùng một lần nếu không giảm phức tạp.
