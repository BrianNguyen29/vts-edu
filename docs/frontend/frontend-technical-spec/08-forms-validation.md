# 08. Forms & Validation

## 1. Stack

- React Hook Form quản lý field, dirty, touched, submit state.
- Zod 4 định nghĩa client schema.
- `@hookform/resolvers/zod` liên kết schema.
- Backend vẫn thực hiện structural và business validation cuối cùng.

## 2. Validation layers

```text
HTML semantics
  -> Zod client structural validation
  -> feature precondition checks
  -> backend structural/business validation
  -> Problem Details field error mapping
```

Client validation không được giả lập mọi business rule nếu rule phụ thuộc database hoặc thời gian server.

## 3. Form module structure

```text
features/questions/create-question/
├── model/
│   ├── schema.ts
│   ├── defaults.ts
│   ├── mapper.ts
│   └── use-create-question-form.ts
├── ui/
│   ├── create-question-form.tsx
│   └── answer-options-field.tsx
└── index.ts
```

## 4. Schema example

```ts
const questionSchema = z.discriminatedUnion('type', [
  z.object({
    type: z.literal('SINGLE_CHOICE'),
    title: z.string().trim().min(1).max(500),
    choices: z.array(choiceSchema).min(2),
    correctChoiceId: z.string().uuid(),
  }),
  z.object({
    type: z.literal('ESSAY'),
    title: z.string().trim().min(1).max(500),
    rubric: z.string().max(10_000).optional(),
  }),
]);
```

Discriminated union phù hợp câu hỏi nhiều loại; không dùng một object hàng chục optional fields.

## 5. API mapping

Form model không nhất thiết giống API model 1:1.

```text
API response
 -> mapToFormDefaults()
 -> form model
 -> mapFormToRequest()
 -> generated request type
```

Mapper phải có test.

## 6. Default values và async data

```ts
const form = useForm<FormValues>({
  resolver: zodResolver(schema),
  defaultValues: EMPTY_DEFAULTS,
});

useEffect(() => {
  if (query.data && !form.formState.isDirty) {
    form.reset(mapToFormDefaults(query.data));
  }
}, [query.data]);
```

Không reset khi user đã sửa trừ khi họ xác nhận.

## 7. Server field error mapping

Backend field path như `body.choices.1.text` được map sang `choices.1.text`.

```ts
for (const [path, messages] of Object.entries(error.fieldErrors ?? {})) {
  form.setError(normalizeServerFieldPath(path), {
    type: 'server',
    message: messages[0],
  });
}
```

Nếu field không tồn tại trong form, hiển thị general error summary.

## 8. Submit lifecycle

```text
user submit
-> client validate
-> focus first invalid field nếu fail
-> disable duplicate submit
-> mutation
   -> success: reset dirty state / navigate / toast
   -> 422: field mapping
   -> 409/412: conflict dialog
   -> other: form-level alert with request ID
```

Button disabled chỉ trong request đang active; không khóa toàn form nếu người dùng cần đọc/sửa sau lỗi.

## 9. Unsaved changes

Form editor có dirty state:

- Route blocker khi rời page.
- Browser `beforeunload` chỉ khi dirty.
- Sau save thành công, `reset(savedValues)` để clear dirty.
- Không block navigation nếu chỉ query cache stale.

## 10. Dynamic fields

Dùng `useFieldArray` cho:

- Choice list.
- Assessment sections.
- Rubric criteria.

Mỗi item có stable client ID. Không dùng array index làm React key khi reorder.

## 11. Autosave forms

Không autosave mọi form. Có thể dùng cho:

- Assessment builder draft.
- Long assignment/question editor.

Autosave policy:

- Debounce 1–2 giây sau change.
- Chỉ save khi schema subset valid.
- Hiển thị `Saving / Saved / Error`.
- Dùng version/If-Match để phát hiện conflict.
- Local draft là enhancement, không thay backend draft.

Exam answer autosave là subsystem riêng, không dùng generic form autosave.

## 12. File fields

File input model chỉ giữ:

- Local File object tạm.
- Upload state.
- Server file ID sau confirm.

Không serialize File vào form JSON.

Validation:

- Extension chỉ là UX hint.
- MIME/size client check để feedback sớm.
- Backend/storage pipeline kiểm tra cuối cùng.

## 13. Date/time fields

- UI nhập theo timezone người dùng/tổ chức.
- Convert thành UTC RFC3339 trước request.
- Hiển thị timezone rõ ở lịch thi.
- Không dùng `new Date('YYYY-MM-DD')` không timezone cho schedule nghiệp vụ.

## 14. Decimal fields

- Điểm nhập dạng string.
- Regex/decimal parser riêng xác thực.
- Không dùng `<input type="number">` nếu browser locale làm mất kiểm soát decimal separator; có thể dùng `inputMode="decimal"` và text input.
- Backend trả lỗi nếu vượt thang điểm.

## 15. Accessibility

- Mỗi form có `aria-describedby` liên kết help/error.
- Error summary dùng heading và links/focus đến field.
- Không chỉ toast cho validation error.
- Dynamic choice add/remove thông báo qua polite live region khi cần.
- Submit error không làm mất input.

## 16. Form testing

Test tối thiểu:

- Required/format validation.
- Dynamic fields.
- Mapping API defaults.
- Mapping server field errors.
- Duplicate submit prevention.
- Conflict handling.
- Keyboard/focus behavior.
