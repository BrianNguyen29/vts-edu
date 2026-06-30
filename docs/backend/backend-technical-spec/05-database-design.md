# 05. Database Design

## 1. General conventions

- PostgreSQL 15+.
- Primary key: UUID v7 nếu thư viện và hệ sinh thái đã ổn định; nếu không dùng UUID v4. Không dùng ID tuần tự làm public identifier.
- Timestamp: `timestamptz`, lưu UTC.
- Money/score: `numeric(p,s)`.
- Status: `text` + `CHECK`, tránh PostgreSQL enum ở MVP để migration linh hoạt.
- Flexible payload: `jsonb` có schema validation ở application layer và check cơ bản khi hữu ích.
- Soft delete chỉ dùng nơi cần giữ lịch sử; ưu tiên `archived_at`/status thay vì `deleted_at` khắp nơi.
- Mọi tenant table có `organization_id NOT NULL`.

## 2. Core schemas

Có thể dùng một schema `public` ở MVP. Nếu muốn phân nhóm, dùng schema logic sau nhưng không bắt buộc:

```text
identity
academic
content
assessment
grading
system
```

Với dự án solo, một schema và tên bảng rõ ràng thường dễ vận hành hơn.

## 3. Table inventory

### Identity & tenancy

| Table | Mục đích | Key columns |
|---|---|---|
| `organizations` | Tenant | `id`, `name`, `slug`, `status`, `settings` |
| `users` | Identity toàn hệ thống | `id`, `email_normalized`, `status`, `auth_version` |
| `membership_login_names` | Tên đăng nhập theo org | `id`, `organization_id`, `username_normalized`, `user_id`, `status` |
| `user_profiles` | Hồ sơ | `user_id`, `display_name`, `student_code`, metadata hạn chế |
| `organization_memberships` | User thuộc organization | `organization_id`, `user_id`, `status` |
| `roles` | Role definition | `organization_id nullable`, `code`, `name` |
| `permissions` | Permission catalog | `code` |
| `role_permissions` | Role ↔ permission | `role_id`, `permission_id` |
| `membership_roles` | Membership ↔ role | `membership_id`, `role_id` |
| `refresh_sessions` | Refresh token/session | `id`, `user_id`, `membership_id`, `organization_id`, `auth_version`, `token_hash`, `family_id`, `device_metadata_json`, `expires_at`, `revoked_at` |
| `login_attempts` | Security telemetry | actor identifier, IP hash/prefix, outcome |

### Academic

| Table | Mục đích |
|---|---|
| `academic_terms` | Học kỳ/năm học |
| `subjects` | Môn học |
| `courses` | Course logical template |
| `class_sections` | Lớp cụ thể theo kỳ |
| `class_teachers` | Giáo viên/trợ giảng của lớp |
| `enrollments` | Học sinh trong lớp |
| `student_groups` | Nhóm trong lớp |
| `student_group_members` | Thành viên nhóm |

### Resources

| Table | Mục đích |
|---|---|
| `resource_folders` | Cây thư mục |
| `resources` | Metadata resource |
| `resource_versions` | Phiên bản nội dung/metadata |
| `files` | Object metadata |
| `resource_files` | Link resource ↔ file |
| `resource_targets` | Phân phối theo class/group/user |
| `resource_views` | Theo dõi xem ở mức cần thiết |

### Question bank

| Table | Mục đích |
|---|---|
| `question_banks` | Kho câu hỏi |
| `questions` | Logical question identity |
| `question_versions` | Nội dung/version immutable sau publish |
| `question_tags` | Tag dictionary |
| `question_tag_links` | N-N |
| `learning_outcomes` | Chuẩn đầu ra/chủ đề |
| `question_outcome_links` | Mapping |

`question_versions` key fields:

```text
id
organization_id
question_id
version_number
question_type
prompt_json
answer_key_json
scoring_config_json
explanation_json
status
created_by
created_at
published_at
```

### Assessment

| Table | Mục đích |
|---|---|
| `assessments` | Definition và settings |
| `assessment_sections` | Sections |
| `assessment_items` | Question refs trong draft |
| `assessment_item_rules` | Random selection rules |
| `assessment_publications` | Lần publish/version |
| `assessment_item_snapshots` | Immutable snapshot |
| `assessment_targets` | Class/group/user target |
| `assessment_accommodations` | Điều chỉnh theo user |

### Runtime attempts

| Table | Mục đích |
|---|---|
| `attempts` | Một lần làm bài |
| `attempt_items` | Items được chọn/thứ tự cho attempt |
| `attempt_answers` | Answer hiện hành |
| `attempt_events` | Timeline quan trọng |
| `idempotency_keys` | Chống duplicate write |

`attempt_answers` gợi ý:

```sql
CREATE TABLE attempt_answers (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    attempt_item_id uuid NOT NULL,
    answer_payload jsonb NOT NULL,
    revision bigint NOT NULL DEFAULT 1,
    answered_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (organization_id, attempt_id, attempt_item_id)
);
```

### Grading

| Table | Mục đích |
|---|---|
| `grading_runs` | Một lần chạy chấm |
| `grading_results` | Kết quả theo item |
| `manual_reviews` | Review tự luận |
| `grade_categories` | Nhóm điểm |
| `grade_items` | Bài có điểm |
| `grade_entries` | Điểm theo student |
| `grade_entry_history` | Lịch sử thay đổi |
| `grade_publications` | Công bố điểm |

### Assignment

| Table | Mục đích |
|---|---|
| `assignments` | Bài tập |
| `assignment_targets` | Đối tượng nhận |
| `submissions` | Submission logical |
| `submission_versions` | Mỗi lần nộp |
| `submission_files` | File nộp |
| `assignment_feedback` | Feedback |

### System

| Table | Mục đích |
|---|---|
| `notifications` | Nội dung notification |
| `notification_recipients` | Trạng thái theo user |
| `audit_logs` | Append-only audit |
| River tables | Queue framework |

## 4. Critical path DDL (authoritative samples)

MVP cần DDL đầy đủ cho critical path trước khi viết migration. Dưới đây là các mẫu cơ bản với tenant key, FK, unique/check, delete policy.

### Identity & tenancy

```sql
CREATE TABLE organizations (
    id uuid PRIMARY KEY,
    code text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','SUSPENDED')),
    settings jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (code)
);

CREATE TABLE users (
    id uuid PRIMARY KEY,
    email_normalized text,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('INVITED','ACTIVE','SUSPENDED','ARCHIVED')),
    auth_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE organization_memberships (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations(id),
    user_id uuid NOT NULL REFERENCES users(id),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','SUSPENDED','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, user_id)
);

CREATE TABLE membership_login_names (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations(id),
    username_normalized text NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id),
    status text NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, username_normalized),
    FOREIGN KEY (organization_id, user_id) REFERENCES organization_memberships(organization_id, user_id)
);

CREATE TABLE refresh_sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    membership_id uuid NOT NULL REFERENCES organization_memberships(id),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    token_hash text NOT NULL,
    family_id uuid NOT NULL,
    auth_version bigint NOT NULL,
    device_metadata_json jsonb NOT NULL DEFAULT '{}',
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);
```

Refresh phải xác nhận `organization_id`/`membership_id` vẫn active.

## 5. Index strategy

### Tenant + lookup indexes

```sql
CREATE INDEX idx_class_sections_org_term
ON class_sections (organization_id, academic_term_id);

CREATE UNIQUE INDEX uq_enrollments_active
ON enrollments (organization_id, class_section_id, student_user_id)
WHERE status = 'ACTIVE';
```

### Attempt indexes

```sql
CREATE INDEX idx_attempts_student_assessment
ON attempts (organization_id, student_user_id, assessment_id, created_at DESC);

CREATE INDEX idx_attempts_in_progress_expiry
ON attempts (expires_at)
WHERE status = 'IN_PROGRESS';
```

### Grade indexes

```sql
CREATE UNIQUE INDEX uq_grade_entry_item_student
ON grade_entries (organization_id, grade_item_id, student_user_id);
```

### Search

MVP có thể tạo generated `tsvector` hoặc expression GIN index cho:

- Resource title/content text.
- Question prompt plain-text projection.
- User display name/student code.

Không index mọi JSONB path mặc định.

## 6. Foreign key policy

- FK bắt buộc cho quan hệ dữ liệu lõi.
- `ON DELETE CASCADE` chỉ với child thuần kỹ thuật không có lịch sử độc lập.
- User/class/assessment có dữ liệu lịch sử: dùng restrict/archive.
- Audit log không FK cứng tới actor/resource nếu việc xóa/ẩn danh có thể xảy ra; lưu actor/resource ID snapshot.

## 7. Multi-tenancy

### Application-layer scoping

Mọi query tenant có dạng:

```sql
SELECT ...
FROM assessments
WHERE organization_id = $1
  AND id = $2;
```

### RLS — optional defense in depth

RLS có thể bật sau khi repository pattern ổn định. Nếu bật:

```sql
SET LOCAL app.organization_id = '...';
```

Policy dùng `current_setting`. Cần test connection pool không làm rò context; dùng `SET LOCAL` trong transaction.

Không xem RLS là thay thế authorization application layer.

## 8. Concurrency controls

| Use case | Cơ chế |
|---|---|
| Save answer | `revision` optimistic check |
| Submit attempt | `SELECT ... FOR UPDATE` trên attempt |
| Publish assessment | Lock assessment row + status check |
| Grade override | Lock grade entry + version/history |
| Refresh token rotation | Lock session/family row hoặc atomic update condition |
| Start attempt count | Transaction + unique/locking strategy |

## 9. Idempotency table

```text
idempotency_keys
  organization_id
  actor_user_id
  scope
  key_hash
  request_hash
  response_status
  response_body_json
  resource_id
  expires_at
  created_at
```

Unique:

```text
(organization_id, actor_user_id, scope, key_hash)
```

Không lưu raw idempotency key.

## 10. Audit log design

```text
audit_logs
  id
  organization_id
  actor_user_id nullable
  actor_type
  action
  resource_type
  resource_id nullable
  request_id
  ip_prefix_or_hash
  user_agent_hash
  before_json redacted nullable
  after_json redacted nullable
  metadata_json
  created_at
```

Không lưu:

- Password.
- Token.
- Full answer essay mặc định.
- Sensitive personal data không cần thiết.

## 11. Migration rules

1. Migration chỉ append, không sửa migration đã chạy shared environment.
2. DDL nguy hiểm phải có rollout plan.
3. Add column theo pattern expand → backfill → enforce → contract.
4. Index lớn dùng `CREATE INDEX CONCURRENTLY` khi production, với migration `NO TRANSACTION` phù hợp.
5. Mỗi migration có `Up` và `Down` khi down an toàn; destructive down có thể ghi rõ không hỗ trợ.
6. CI chạy migration từ database trống và upgrade từ snapshot gần nhất.

## 12. Backup

- Automated daily full backup/snapshot.
- WAL/PITR nếu managed provider hỗ trợ và dữ liệu production thật.
- Object storage versioning hoặc lifecycle phù hợp.
- Restore drill định kỳ, không chỉ kiểm tra “backup succeeded”.
