import { useEffect, useMemo, useRef, useState } from 'react';
import { ApiResponseError } from '@/shared/api/attempts';
import {
  addClassTeacher,
  archiveClass,
  archiveCourse,
  archiveSubject,
  archiveTerm,
  bulkAssignTeachers,
  bulkEnrollStudents,
  createClass,
  createCourse,
  createSubject,
  createTerm,
  enrollStudent,
  listClasses,
  listClassTeachers,
  listCourses,
  listEnrollments,
  listSubjects,
  listTerms,
  removeClassTeacher,
  unenrollStudent,
  updateClass,
  updateCourse,
  updateSubject,
  updateTerm,
  type AddClassTeacherRequest,
  type BulkAssignTeacherItem,
  type BulkAssignTeachersResult,
  type BulkEnrollmentResult,
  type ClassSection,
  type ClassTeacher,
  type Course,
  type CreateClassRequest,
  type CreateCourseRequest,
  type CreateSubjectRequest,
  type CreateTermRequest,
  type Enrollment,
  type Subject,
  type Term,
} from '@/shared/api/academics';
import { listUsers, type User } from '@/shared/api/admin';

type Section = 'terms' | 'subjects' | 'courses' | 'classes';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền thực hiện thao tác này.';
      case 404:
        return 'Không tìm thấy dữ liệu.';
      case 409:
        return 'Dữ liệu bị trùng lặp hoặc đã tồn tại.';
      default:
        return err.body.error.message || 'Yêu cầu thất bại.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function AcademicManagementPanel() {
  const [section, setSection] = useState<Section>('terms');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [terms, setTerms] = useState<Term[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [courses, setCourses] = useState<Course[]>([]);
  const [classes, setClasses] = useState<ClassSection[]>([]);

  // create forms
  const [termForm, setTermForm] = useState<CreateTermRequest>({
    name: '',
    start_date: '',
    end_date: '',
  });
  const [subjectForm, setSubjectForm] = useState<CreateSubjectRequest>({
    code: '',
    name: '',
    description: '',
  });
  const [courseForm, setCourseForm] = useState<CreateCourseRequest>({
    subject_id: '',
    academic_term_id: '',
    code: '',
    name: '',
  });
  const [classForm, setClassForm] = useState<CreateClassRequest>({
    course_id: '',
    name: '',
  });

  // inline editing
  const [editingTermId, setEditingTermId] = useState<string | null>(null);
  const [editingTermDraft, setEditingTermDraft] = useState<CreateTermRequest>({
    name: '',
    start_date: '',
    end_date: '',
  });
  const [editingSubjectId, setEditingSubjectId] = useState<string | null>(null);
  const [editingSubjectDraft, setEditingSubjectDraft] =
    useState<CreateSubjectRequest>({ code: '', name: '', description: '' });
  const [editingCourseId, setEditingCourseId] = useState<string | null>(null);
  const [editingCourseDraft, setEditingCourseDraft] =
    useState<CreateCourseRequest>({
      subject_id: '',
      academic_term_id: '',
      code: '',
      name: '',
    });
  const [editingClassId, setEditingClassId] = useState<string | null>(null);
  const [editingClassDraft, setEditingClassDraft] =
    useState<CreateClassRequest>({ course_id: '', name: '' });

  // class detail
  const [managingClassId, setManagingClassId] = useState<string | null>(null);
  const [classTeachers, setClassTeachers] = useState<ClassTeacher[]>([]);
  const [classStudents, setClassStudents] = useState<Enrollment[]>([]);
  const [detailLoading, setDetailLoading] = useState(false);

  // bulk operations
  const [bulkStudentIds, setBulkStudentIds] = useState<string[]>([]);
  const [bulkTeacherItems, setBulkTeacherItems] = useState<
    BulkAssignTeacherItem[]
  >([]);
  const [bulkPreview, setBulkPreview] = useState<
    BulkEnrollmentResult | BulkAssignTeachersResult | null
  >(null);
  const [bulkMode, setBulkMode] = useState<'students' | 'teachers' | null>(
    null
  );
  const [bulkLoading, setBulkLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const [t, s, c, cl] = await Promise.all([
          listTerms(),
          listSubjects(),
          listCourses(),
          listClasses(),
        ]);
        if (cancelled) return;
        setTerms(t.data);
        setSubjects(s.data);
        setCourses(c.data);
        setClasses(cl.data);
      } catch (err) {
        if (cancelled) return;
        setError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!managingClassId) {
      setClassTeachers([]);
      setClassStudents([]);
      resetBulkState();
      return;
    }

    resetBulkState();
    let cancelled = false;

    async function loadDetail() {
      setDetailLoading(true);
      try {
        const [teachers, students] = await Promise.all([
          listClassTeachers(managingClassId!),
          listEnrollments(managingClassId!),
        ]);
        if (cancelled) return;
        setClassTeachers(teachers.data);
        setClassStudents(students.data);
      } catch (err) {
        if (cancelled) return;
        setError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setDetailLoading(false);
      }
    }

    void loadDetail();

    return () => {
      cancelled = true;
    };
  }, [managingClassId]);

  function clearMessages() {
    setError(null);
    setSuccess(null);
  }

  async function handleCreateTerm(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();
    try {
      const created = await createTerm(termForm);
      setTerms((prev) => [...prev, created]);
      setTermForm({ name: '', start_date: '', end_date: '' });
      setSuccess('Đã tạo học kỳ.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUpdateTerm(termId: string) {
    clearMessages();
    try {
      const updated = await updateTerm(termId, editingTermDraft);
      setTerms((prev) =>
        prev.map((t) => (t.id === termId ? updated : t))
      );
      setEditingTermId(null);
      setSuccess('Đã cập nhật học kỳ.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleArchiveTerm(termId: string) {
    if (!window.confirm('Bạn có chắc muốn lưu trữ học kỳ này?')) return;
    clearMessages();
    try {
      await archiveTerm(termId);
      setTerms((prev) => prev.filter((t) => t.id !== termId));
      setSuccess('Đã lưu trữ học kỳ.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleCreateSubject(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();
    try {
      const created = await createSubject(subjectForm);
      setSubjects((prev) => [...prev, created]);
      setSubjectForm({ code: '', name: '', description: '' });
      setSuccess('Đã tạo môn học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUpdateSubject(subjectId: string) {
    clearMessages();
    try {
      const updated = await updateSubject(subjectId, editingSubjectDraft);
      setSubjects((prev) =>
        prev.map((s) => (s.id === subjectId ? updated : s))
      );
      setEditingSubjectId(null);
      setSuccess('Đã cập nhật môn học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleArchiveSubject(subjectId: string) {
    if (!window.confirm('Bạn có chắc muốn lưu trữ môn học này?')) return;
    clearMessages();
    try {
      await archiveSubject(subjectId);
      setSubjects((prev) => prev.filter((s) => s.id !== subjectId));
      setSuccess('Đã lưu trữ môn học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleCreateCourse(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();
    try {
      const created = await createCourse(courseForm);
      setCourses((prev) => [...prev, created]);
      setCourseForm({
        subject_id: '',
        academic_term_id: '',
        code: '',
        name: '',
      });
      setSuccess('Đã tạo khóa học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUpdateCourse(courseId: string) {
    clearMessages();
    try {
      const updated = await updateCourse(courseId, editingCourseDraft);
      setCourses((prev) =>
        prev.map((c) => (c.id === courseId ? updated : c))
      );
      setEditingCourseId(null);
      setSuccess('Đã cập nhật khóa học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleArchiveCourse(courseId: string) {
    if (!window.confirm('Bạn có chắc muốn lưu trữ khóa học này?')) return;
    clearMessages();
    try {
      await archiveCourse(courseId);
      setCourses((prev) => prev.filter((c) => c.id !== courseId));
      setSuccess('Đã lưu trữ khóa học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleCreateClass(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();
    try {
      const created = await createClass(classForm);
      setClasses((prev) => [...prev, created]);
      setClassForm({ course_id: '', name: '' });
      setSuccess('Đã tạo lớp học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUpdateClass(classId: string) {
    clearMessages();
    try {
      const updated = await updateClass(classId, editingClassDraft);
      setClasses((prev) =>
        prev.map((cl) => (cl.id === classId ? updated : cl))
      );
      setEditingClassId(null);
      setSuccess('Đã cập nhật lớp học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleArchiveClass(classId: string) {
    if (!window.confirm('Bạn có chắc muốn lưu trữ lớp học này?')) return;
    clearMessages();
    try {
      await archiveClass(classId);
      setClasses((prev) => prev.filter((cl) => cl.id !== classId));
      if (managingClassId === classId) setManagingClassId(null);
      setSuccess('Đã lưu trữ lớp học.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleAddTeacher(userId: string, role: AddClassTeacherRequest['role']) {
    if (!managingClassId) return;
    clearMessages();
    try {
      const added = await addClassTeacher(managingClassId, { user_id: userId, role });
      setClassTeachers((prev) => [...prev, added]);
      setClasses((prev) =>
        prev.map((cl) =>
          cl.id === managingClassId
            ? { ...cl, teacher_count: cl.teacher_count + 1 }
            : cl
        )
      );
      setSuccess('Đã thêm giáo viên.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleRemoveTeacher(userId: string) {
    if (!managingClassId) return;
    if (!window.confirm('Bạn có chắc muốn gỡ giáo viên này?')) return;
    clearMessages();
    try {
      await removeClassTeacher(managingClassId, userId);
      setClassTeachers((prev) => prev.filter((t) => t.user_id !== userId));
      setClasses((prev) =>
        prev.map((cl) =>
          cl.id === managingClassId
            ? { ...cl, teacher_count: Math.max(0, cl.teacher_count - 1) }
            : cl
        )
      );
      setSuccess('Đã gỡ giáo viên.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleEnrollStudent(userId: string) {
    if (!managingClassId) return;
    clearMessages();
    try {
      const added = await enrollStudent(managingClassId, { user_id: userId });
      setClassStudents((prev) => [...prev, added]);
      setClasses((prev) =>
        prev.map((cl) =>
          cl.id === managingClassId
            ? { ...cl, student_count: cl.student_count + 1 }
            : cl
        )
      );
      setSuccess('Đã thêm học sinh.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUnenrollStudent(userId: string) {
    if (!managingClassId) return;
    if (!window.confirm('Bạn có chắc muốn rút học sinh này khỏi lớp?')) return;
    clearMessages();
    try {
      await unenrollStudent(managingClassId, userId);
      setClassStudents((prev) => prev.filter((s) => s.user_id !== userId));
      setClasses((prev) =>
        prev.map((cl) =>
          cl.id === managingClassId
            ? { ...cl, student_count: Math.max(0, cl.student_count - 1) }
            : cl
        )
      );
      setSuccess('Đã rút học sinh.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  function resetBulkState() {
    setBulkStudentIds([]);
    setBulkTeacherItems([]);
    setBulkPreview(null);
    setBulkMode(null);
  }

  function addBulkStudent(userId: string) {
    setBulkStudentIds((prev) =>
      prev.includes(userId) ? prev : [...prev, userId]
    );
    setBulkPreview(null);
  }

  function removeBulkStudent(userId: string) {
    setBulkStudentIds((prev) => prev.filter((id) => id !== userId));
    setBulkPreview(null);
  }

  function addBulkTeacher(userId: string, role: BulkAssignTeacherItem['role']) {
    setBulkTeacherItems((prev) => {
      const next = prev.filter((item) => item.user_id !== userId);
      return [...next, { user_id: userId, role }];
    });
    setBulkPreview(null);
  }

  function removeBulkTeacher(userId: string) {
    setBulkTeacherItems((prev) => prev.filter((item) => item.user_id !== userId));
    setBulkPreview(null);
  }

  async function handleBulkEnroll(dryRun: boolean) {
    if (!managingClassId || bulkStudentIds.length === 0) return;
    clearMessages();
    setBulkLoading(true);
    try {
      const result = await bulkEnrollStudents(managingClassId, {
        user_ids: bulkStudentIds,
        dry_run: dryRun,
      });
      setBulkPreview(result);
      if (!dryRun) {
        setSuccess(`Đã ghi danh ${result.enrolled}/${result.total} học sinh.`);
        const [teachers, students] = await Promise.all([
          listClassTeachers(managingClassId),
          listEnrollments(managingClassId),
        ]);
        setClassTeachers(teachers.data);
        setClassStudents(students.data);
        setClasses((prev) =>
          prev.map((cl) =>
            cl.id === managingClassId
              ? { ...cl, student_count: students.data.length }
              : cl
          )
        );
        resetBulkState();
      }
    } catch (err) {
      setError(formatFriendlyError(err));
      setBulkPreview(null);
    } finally {
      setBulkLoading(false);
    }
  }

  async function handleBulkAssignTeachers(dryRun: boolean) {
    if (!managingClassId || bulkTeacherItems.length === 0) return;
    clearMessages();
    setBulkLoading(true);
    try {
      const result = await bulkAssignTeachers(managingClassId, {
        items: bulkTeacherItems,
        dry_run: dryRun,
      });
      setBulkPreview(result);
      if (!dryRun) {
        setSuccess(
          `Đã phân công ${result.assigned}/${result.total} giáo viên.`
        );
        const [teachers, students] = await Promise.all([
          listClassTeachers(managingClassId),
          listEnrollments(managingClassId),
        ]);
        setClassTeachers(teachers.data);
        setClassStudents(students.data);
        setClasses((prev) =>
          prev.map((cl) =>
            cl.id === managingClassId
              ? { ...cl, teacher_count: teachers.data.length }
              : cl
          )
        );
        resetBulkState();
      }
    } catch (err) {
      setError(formatFriendlyError(err));
      setBulkPreview(null);
    } finally {
      setBulkLoading(false);
    }
  }

  const activeTerms = useMemo(
    () => terms.filter((t) => t.status === 'ACTIVE'),
    [terms]
  );
  const activeSubjects = useMemo(
    () => subjects.filter((s) => s.status === 'ACTIVE'),
    [subjects]
  );
  const activeCourses = useMemo(
    () => courses.filter((c) => c.status === 'ACTIVE'),
    [courses]
  );
  const activeClasses = useMemo(
    () => classes.filter((cl) => cl.status === 'ACTIVE'),
    [classes]
  );

  return (
    <section className="admin-section">
      <h2>Quản lý học vụ</h2>

      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}
      {success && (
        <div className="success-banner" role="status">
          {success}
        </div>
      )}

      <nav className="admin-subtabs" aria-label="Quản lý học vụ">
        {[
          { key: 'terms', label: 'Học kỳ' },
          { key: 'subjects', label: 'Môn học' },
          { key: 'courses', label: 'Khóa học' },
          { key: 'classes', label: 'Lớp học' },
        ].map((tab) => (
          <button
            key={tab.key}
            type="button"
            className={section === tab.key ? 'active' : ''}
            onClick={() => setSection(tab.key as Section)}
            aria-current={section === tab.key ? 'true' : undefined}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {loading && section !== 'classes' && (
        <p className="dashboard-status">Đang tải dữ liệu…</p>
      )}

      {section === 'terms' && (
        <TermManager
          terms={activeTerms}
          form={termForm}
          onFormChange={setTermForm}
          onCreate={handleCreateTerm}
          editingId={editingTermId}
          editingDraft={editingTermDraft}
          onEdit={(term) => {
            setEditingTermId(term.id);
            setEditingTermDraft({
              name: term.name,
              start_date: term.start_date,
              end_date: term.end_date,
            });
          }}
          onEditChange={setEditingTermDraft}
          onSave={() => editingTermId && handleUpdateTerm(editingTermId)}
          onCancelEdit={() => setEditingTermId(null)}
          onArchive={handleArchiveTerm}
        />
      )}

      {section === 'subjects' && (
        <SubjectManager
          subjects={activeSubjects}
          form={subjectForm}
          onFormChange={setSubjectForm}
          onCreate={handleCreateSubject}
          editingId={editingSubjectId}
          editingDraft={editingSubjectDraft}
          onEdit={(subject) => {
            setEditingSubjectId(subject.id);
            setEditingSubjectDraft({
              code: subject.code,
              name: subject.name,
              description: subject.description ?? '',
            });
          }}
          onEditChange={setEditingSubjectDraft}
          onSave={() =>
            editingSubjectId && handleUpdateSubject(editingSubjectId)
          }
          onCancelEdit={() => setEditingSubjectId(null)}
          onArchive={handleArchiveSubject}
        />
      )}

      {section === 'courses' && (
        <CourseManager
          courses={activeCourses}
          subjects={activeSubjects}
          terms={activeTerms}
          form={courseForm}
          onFormChange={setCourseForm}
          onCreate={handleCreateCourse}
          editingId={editingCourseId}
          editingDraft={editingCourseDraft}
          onEdit={(course) => {
            setEditingCourseId(course.id);
            setEditingCourseDraft({
              subject_id: course.subject_id,
              academic_term_id: course.academic_term_id,
              code: course.code,
              name: course.name,
            });
          }}
          onEditChange={setEditingCourseDraft}
          onSave={() =>
            editingCourseId && handleUpdateCourse(editingCourseId)
          }
          onCancelEdit={() => setEditingCourseId(null)}
          onArchive={handleArchiveCourse}
        />
      )}

      {section === 'classes' && (
        <ClassManager
          classes={activeClasses}
          courses={activeCourses}
          form={classForm}
          onFormChange={setClassForm}
          onCreate={handleCreateClass}
          editingId={editingClassId}
          editingDraft={editingClassDraft}
          onEdit={(cl) => {
            setEditingClassId(cl.id);
            setEditingClassDraft({
              course_id: cl.course_id,
              name: cl.name,
            });
          }}
          onEditChange={setEditingClassDraft}
          onSave={() =>
            editingClassId && handleUpdateClass(editingClassId)
          }
          onCancelEdit={() => setEditingClassId(null)}
          onArchive={handleArchiveClass}
          managingClassId={managingClassId}
          onManage={setManagingClassId}
          teachers={classTeachers}
          students={classStudents}
          detailLoading={detailLoading}
          onAddTeacher={handleAddTeacher}
          onRemoveTeacher={handleRemoveTeacher}
          onEnrollStudent={handleEnrollStudent}
          onUnenrollStudent={handleUnenrollStudent}
          bulkStudentIds={bulkStudentIds}
          bulkTeacherItems={bulkTeacherItems}
          bulkPreview={bulkPreview}
          bulkMode={bulkMode}
          bulkLoading={bulkLoading}
          onAddBulkStudent={addBulkStudent}
          onRemoveBulkStudent={removeBulkStudent}
          onAddBulkTeacher={addBulkTeacher}
          onRemoveBulkTeacher={removeBulkTeacher}
          onSetBulkMode={setBulkMode}
          onBulkEnroll={handleBulkEnroll}
          onBulkAssignTeachers={handleBulkAssignTeachers}
        />
      )}
    </section>
  );
}

function TermManager({
  terms,
  form,
  onFormChange,
  onCreate,
  editingId,
  editingDraft,
  onEdit,
  onEditChange,
  onSave,
  onCancelEdit,
  onArchive,
}: {
  terms: Term[];
  form: CreateTermRequest;
  onFormChange: (form: CreateTermRequest) => void;
  onCreate: (e: React.FormEvent) => void;
  editingId: string | null;
  editingDraft: CreateTermRequest;
  onEdit: (term: Term) => void;
  onEditChange: (form: CreateTermRequest) => void;
  onSave: () => void;
  onCancelEdit: () => void;
  onArchive: (id: string) => void;
}) {
  return (
    <div className="academic-manager">
      <form onSubmit={onCreate} className="inline-form academic-create-form">
        <input
          type="text"
          placeholder="Tên học kỳ"
          value={form.name}
          onChange={(e) => onFormChange({ ...form, name: e.target.value })}
          required
        />
        <input
          type="date"
          value={form.start_date}
          onChange={(e) =>
            onFormChange({ ...form, start_date: e.target.value })
          }
          required
        />
        <input
          type="date"
          value={form.end_date}
          onChange={(e) => onFormChange({ ...form, end_date: e.target.value })}
          required
        />
        <button type="submit" className="primary">
          Thêm học kỳ
        </button>
      </form>

      {terms.length === 0 ? (
        <p className="dashboard-status">Chưa có học kỳ nào.</p>
      ) : (
        <table className="academic-table">
          <thead>
            <tr>
              <th>Tên</th>
              <th>Bắt đầu</th>
              <th>Kết thúc</th>
              <th>Thao tác</th>
            </tr>
          </thead>
          <tbody>
            {terms.map((term) =>
              editingId === term.id ? (
                <tr key={term.id}>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.name}
                      onChange={(e) =>
                        onEditChange({ ...editingDraft, name: e.target.value })
                      }
                    />
                  </td>
                  <td>
                    <input
                      type="date"
                      value={editingDraft.start_date}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          start_date: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td>
                    <input
                      type="date"
                      value={editingDraft.end_date}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          end_date: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td className="row-actions">
                    <button type="button" onClick={onSave}>
                      Lưu
                    </button>
                    <button type="button" onClick={onCancelEdit}>
                      Hủy
                    </button>
                  </td>
                </tr>
              ) : (
                <tr key={term.id}>
                  <td>{term.name}</td>
                  <td>{term.start_date}</td>
                  <td>{term.end_date}</td>
                  <td className="row-actions">
                    <button type="button" onClick={() => onEdit(term)}>
                      Sửa
                    </button>
                    <button type="button" onClick={() => onArchive(term.id)}>
                      Lưu trữ
                    </button>
                  </td>
                </tr>
              )
            )}
          </tbody>
        </table>
      )}
    </div>
  );
}

function SubjectManager({
  subjects,
  form,
  onFormChange,
  onCreate,
  editingId,
  editingDraft,
  onEdit,
  onEditChange,
  onSave,
  onCancelEdit,
  onArchive,
}: {
  subjects: Subject[];
  form: CreateSubjectRequest;
  onFormChange: (form: CreateSubjectRequest) => void;
  onCreate: (e: React.FormEvent) => void;
  editingId: string | null;
  editingDraft: CreateSubjectRequest;
  onEdit: (subject: Subject) => void;
  onEditChange: (form: CreateSubjectRequest) => void;
  onSave: () => void;
  onCancelEdit: () => void;
  onArchive: (id: string) => void;
}) {
  return (
    <div className="academic-manager">
      <form onSubmit={onCreate} className="inline-form academic-create-form">
        <input
          type="text"
          placeholder="Mã môn học"
          value={form.code}
          onChange={(e) => onFormChange({ ...form, code: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Tên môn học"
          value={form.name}
          onChange={(e) => onFormChange({ ...form, name: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Mô tả"
          value={form.description ?? ''}
          onChange={(e) =>
            onFormChange({ ...form, description: e.target.value })
          }
        />
        <button type="submit" className="primary">
          Thêm môn học
        </button>
      </form>

      {subjects.length === 0 ? (
        <p className="dashboard-status">Chưa có môn học nào.</p>
      ) : (
        <table className="academic-table">
          <thead>
            <tr>
              <th>Mã</th>
              <th>Tên</th>
              <th>Mô tả</th>
              <th>Thao tác</th>
            </tr>
          </thead>
          <tbody>
            {subjects.map((subject) =>
              editingId === subject.id ? (
                <tr key={subject.id}>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.code}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          code: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.name}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          name: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.description ?? ''}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          description: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td className="row-actions">
                    <button type="button" onClick={onSave}>
                      Lưu
                    </button>
                    <button type="button" onClick={onCancelEdit}>
                      Hủy
                    </button>
                  </td>
                </tr>
              ) : (
                <tr key={subject.id}>
                  <td>{subject.code}</td>
                  <td>{subject.name}</td>
                  <td>{subject.description || '—'}</td>
                  <td className="row-actions">
                    <button type="button" onClick={() => onEdit(subject)}>
                      Sửa
                    </button>
                    <button
                      type="button"
                      onClick={() => onArchive(subject.id)}
                    >
                      Lưu trữ
                    </button>
                  </td>
                </tr>
              )
            )}
          </tbody>
        </table>
      )}
    </div>
  );
}

function CourseManager({
  courses,
  subjects,
  terms,
  form,
  onFormChange,
  onCreate,
  editingId,
  editingDraft,
  onEdit,
  onEditChange,
  onSave,
  onCancelEdit,
  onArchive,
}: {
  courses: Course[];
  subjects: Subject[];
  terms: Term[];
  form: CreateCourseRequest;
  onFormChange: (form: CreateCourseRequest) => void;
  onCreate: (e: React.FormEvent) => void;
  editingId: string | null;
  editingDraft: CreateCourseRequest;
  onEdit: (course: Course) => void;
  onEditChange: (form: CreateCourseRequest) => void;
  onSave: () => void;
  onCancelEdit: () => void;
  onArchive: (id: string) => void;
}) {
  return (
    <div className="academic-manager">
      <form onSubmit={onCreate} className="inline-form academic-create-form">
        <select
          value={form.subject_id}
          onChange={(e) =>
            onFormChange({ ...form, subject_id: e.target.value })
          }
          required
        >
          <option value="">Chọn môn học…</option>
          {subjects.map((s) => (
            <option key={s.id} value={s.id}>
              {s.code} — {s.name}
            </option>
          ))}
        </select>
        <select
          value={form.academic_term_id}
          onChange={(e) =>
            onFormChange({ ...form, academic_term_id: e.target.value })
          }
          required
        >
          <option value="">Chọn học kỳ…</option>
          {terms.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
        <input
          type="text"
          placeholder="Mã khóa học"
          value={form.code}
          onChange={(e) => onFormChange({ ...form, code: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Tên khóa học"
          value={form.name}
          onChange={(e) => onFormChange({ ...form, name: e.target.value })}
          required
        />
        <button type="submit" className="primary">
          Thêm khóa học
        </button>
      </form>

      {courses.length === 0 ? (
        <p className="dashboard-status">Chưa có khóa học nào.</p>
      ) : (
        <table className="academic-table">
          <thead>
            <tr>
              <th>Mã</th>
              <th>Tên</th>
              <th>Môn học</th>
              <th>Học kỳ</th>
              <th>Thao tác</th>
            </tr>
          </thead>
          <tbody>
            {courses.map((course) => {
              const subject = subjects.find((s) => s.id === course.subject_id);
              const term = terms.find((t) => t.id === course.academic_term_id);
              return editingId === course.id ? (
                <tr key={course.id}>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.code}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          code: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.name}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          name: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td>
                    <select
                      value={editingDraft.subject_id}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          subject_id: e.target.value,
                        })
                      }
                    >
                      {subjects.map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.code} — {s.name}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td>
                    <select
                      value={editingDraft.academic_term_id}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          academic_term_id: e.target.value,
                        })
                      }
                    >
                      {terms.map((t) => (
                        <option key={t.id} value={t.id}>
                          {t.name}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="row-actions">
                    <button type="button" onClick={onSave}>
                      Lưu
                    </button>
                    <button type="button" onClick={onCancelEdit}>
                      Hủy
                    </button>
                  </td>
                </tr>
              ) : (
                <tr key={course.id}>
                  <td>{course.code}</td>
                  <td>{course.name}</td>
                  <td>{subject ? subject.name : course.subject_id}</td>
                  <td>{term ? term.name : course.academic_term_id}</td>
                  <td className="row-actions">
                    <button type="button" onClick={() => onEdit(course)}>
                      Sửa
                    </button>
                    <button
                      type="button"
                      onClick={() => onArchive(course.id)}
                    >
                      Lưu trữ
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}

function ClassManager({
  classes,
  courses,
  form,
  onFormChange,
  onCreate,
  editingId,
  editingDraft,
  onEdit,
  onEditChange,
  onSave,
  onCancelEdit,
  onArchive,
  managingClassId,
  onManage,
  teachers,
  students,
  detailLoading,
  onAddTeacher,
  onRemoveTeacher,
  onEnrollStudent,
  onUnenrollStudent,
  bulkStudentIds,
  bulkTeacherItems,
  bulkPreview,
  bulkMode,
  bulkLoading,
  onAddBulkStudent,
  onRemoveBulkStudent,
  onAddBulkTeacher,
  onRemoveBulkTeacher,
  onSetBulkMode,
  onBulkEnroll,
  onBulkAssignTeachers,
}: {
  classes: ClassSection[];
  courses: Course[];
  form: CreateClassRequest;
  onFormChange: (form: CreateClassRequest) => void;
  onCreate: (e: React.FormEvent) => void;
  editingId: string | null;
  editingDraft: CreateClassRequest;
  onEdit: (cl: ClassSection) => void;
  onEditChange: (form: CreateClassRequest) => void;
  onSave: () => void;
  onCancelEdit: () => void;
  onArchive: (id: string) => void;
  managingClassId: string | null;
  onManage: (id: string | null) => void;
  teachers: ClassTeacher[];
  students: Enrollment[];
  detailLoading: boolean;
  onAddTeacher: (userId: string, role: AddClassTeacherRequest['role']) => void;
  onRemoveTeacher: (userId: string) => void;
  onEnrollStudent: (userId: string) => void;
  onUnenrollStudent: (userId: string) => void;
  bulkStudentIds: string[];
  bulkTeacherItems: BulkAssignTeacherItem[];
  bulkPreview: BulkEnrollmentResult | BulkAssignTeachersResult | null;
  bulkMode: 'students' | 'teachers' | null;
  bulkLoading: boolean;
  onAddBulkStudent: (userId: string) => void;
  onRemoveBulkStudent: (userId: string) => void;
  onAddBulkTeacher: (userId: string, role: BulkAssignTeacherItem['role']) => void;
  onRemoveBulkTeacher: (userId: string) => void;
  onSetBulkMode: (mode: 'students' | 'teachers' | null) => void;
  onBulkEnroll: (dryRun: boolean) => void;
  onBulkAssignTeachers: (dryRun: boolean) => void;
}) {
  const managingClass = classes.find((cl) => cl.id === managingClassId);
  const teacherRoleRef = useRef<HTMLSelectElement>(null);

  return (
    <div className="academic-manager">
      <form onSubmit={onCreate} className="inline-form academic-create-form">
        <select
          value={form.course_id}
          onChange={(e) =>
            onFormChange({ ...form, course_id: e.target.value })
          }
          required
        >
          <option value="">Chọn khóa học…</option>
          {courses.map((c) => (
            <option key={c.id} value={c.id}>
              {c.code} — {c.name}
            </option>
          ))}
        </select>
        <input
          type="text"
          placeholder="Tên lớp"
          value={form.name}
          onChange={(e) => onFormChange({ ...form, name: e.target.value })}
          required
        />
        <button type="submit" className="primary">
          Thêm lớp học
        </button>
      </form>

      {classes.length === 0 ? (
        <p className="dashboard-status">Chưa có lớp học nào.</p>
      ) : (
        <table className="academic-table">
          <thead>
            <tr>
              <th>Tên lớp</th>
              <th>Khóa học</th>
              <th>Giáo viên</th>
              <th>Học sinh</th>
              <th>Thao tác</th>
            </tr>
          </thead>
          <tbody>
            {classes.map((cl) => {
              const course = courses.find((c) => c.id === cl.course_id);
              return editingId === cl.id ? (
                <tr key={cl.id}>
                  <td>
                    <input
                      type="text"
                      value={editingDraft.name}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          name: e.target.value,
                        })
                      }
                    />
                  </td>
                  <td colSpan={3}>
                    <select
                      value={editingDraft.course_id}
                      onChange={(e) =>
                        onEditChange({
                          ...editingDraft,
                          course_id: e.target.value,
                        })
                      }
                    >
                      {courses.map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.code} — {c.name}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="row-actions">
                    <button type="button" onClick={onSave}>
                      Lưu
                    </button>
                    <button type="button" onClick={onCancelEdit}>
                      Hủy
                    </button>
                  </td>
                </tr>
              ) : (
                <tr key={cl.id}>
                  <td>{cl.name}</td>
                  <td>{course ? course.name : cl.course_id}</td>
                  <td>{cl.teacher_count}</td>
                  <td>{cl.student_count}</td>
                  <td className="row-actions">
                    <button type="button" onClick={() => onManage(cl.id)}>
                      Quản lý
                    </button>
                    <button type="button" onClick={() => onEdit(cl)}>
                      Sửa
                    </button>
                    <button type="button" onClick={() => onArchive(cl.id)}>
                      Lưu trữ
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {managingClass && (
        <div className="academic-class-detail">
          <h3>
            Quản lý lớp: {managingClass.name}
            <button
              type="button"
              className="text-button"
              onClick={() => onManage(null)}
            >
              Đóng
            </button>
          </h3>

          {detailLoading ? (
            <p className="dashboard-status">Đang tải…</p>
          ) : (
            <>
              <div className="academic-detail-column">
                <h4>Giáo viên ({teachers.length})</h4>
                <UserPicker
                  role="teacher"
                  onSelect={(userId) => onAddTeacher(userId, 'teacher')}
                  buttonLabel="Thêm giáo viên"
                />
                {teachers.length === 0 ? (
                  <p className="dashboard-status">Chưa có giáo viên.</p>
                ) : (
                  <ul className="academic-member-list">
                    {teachers.map((t) => (
                      <li key={t.id}>
                        <span>
                          {t.display_name}{' '}
                          <small>({t.role === 'assistant' ? 'Trợ giảng' : 'Giáo viên'})</small>
                        </span>
                        <button
                          type="button"
                          onClick={() => onRemoveTeacher(t.user_id)}
                        >
                          Gỡ
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="academic-detail-column">
                <h4>Học sinh ({students.length})</h4>
                <UserPicker
                  role="student"
                  onSelect={(userId) => onEnrollStudent(userId)}
                  buttonLabel="Thêm học sinh"
                />
                {students.length === 0 ? (
                  <p className="dashboard-status">Chưa có học sinh.</p>
                ) : (
                  <ul className="academic-member-list">
                    {students.map((s) => (
                      <li key={s.id}>
                        <span>{s.display_name}</span>
                        <button
                          type="button"
                          onClick={() => onUnenrollStudent(s.user_id)}
                        >
                          Rút
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </>
          )}

          <div className="academic-bulk-panel">
            <h4>Thao tác hàng loạt</h4>
            <div className="bulk-actions">
              <button
                type="button"
                className={bulkMode === 'students' ? 'active' : ''}
                onClick={() =>
                  onSetBulkMode(bulkMode === 'students' ? null : 'students')
                }
              >
                Ghi danh hàng loạt
              </button>
              <button
                type="button"
                className={bulkMode === 'teachers' ? 'active' : ''}
                onClick={() =>
                  onSetBulkMode(bulkMode === 'teachers' ? null : 'teachers')
                }
              >
                Phân công giáo viên hàng loạt
              </button>
            </div>

            {bulkMode === 'students' && (
              <div className="bulk-mode">
                <UserPicker
                  role="student"
                  onSelect={(userId) => onAddBulkStudent(userId)}
                  buttonLabel="Thêm học sinh vào danh sách"
                />
                {bulkStudentIds.length === 0 ? (
                  <p className="dashboard-status">
                    Chưa chọn học sinh nào.
                  </p>
                ) : (
                  <>
                    <p>Đã chọn {bulkStudentIds.length} học sinh.</p>
                    <ul className="academic-member-list compact">
                      {bulkStudentIds.map((id) => (
                        <li key={id}>
                          <code>{id}</code>
                          <button
                            type="button"
                            onClick={() => onRemoveBulkStudent(id)}
                          >
                            Xóa
                          </button>
                        </li>
                      ))}
                    </ul>
                    <div className="form-actions">
                      <button
                        type="button"
                        onClick={() => onBulkEnroll(true)}
                        disabled={bulkLoading}
                      >
                        {bulkLoading && !bulkPreview
                          ? 'Đang kiểm tra…'
                          : 'Kiểm tra'}
                      </button>
                      <button
                        type="button"
                        className="primary"
                        onClick={() => onBulkEnroll(false)}
                        disabled={bulkLoading}
                      >
                        {bulkLoading && bulkPreview
                          ? 'Đang ghi danh…'
                          : 'Xác nhận ghi danh'}
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}

            {bulkMode === 'teachers' && (
              <div className="bulk-mode">
                <div className="bulk-teacher-picker">
                  <UserPicker
                    role="teacher"
                    onSelect={(userId) =>
                      onAddBulkTeacher(
                        userId,
                        teacherRoleRef.current?.value as
                          | 'teacher'
                          | 'assistant'
                      )
                    }
                    buttonLabel="Thêm giáo viên vào danh sách"
                  />
                  <select ref={teacherRoleRef} defaultValue="teacher">
                    <option value="teacher">Giáo viên</option>
                    <option value="assistant">Trợ giảng</option>
                  </select>
                </div>
                {bulkTeacherItems.length === 0 ? (
                  <p className="dashboard-status">
                    Chưa chọn giáo viên nào.
                  </p>
                ) : (
                  <>
                    <p>Đã chọn {bulkTeacherItems.length} giáo viên.</p>
                    <ul className="academic-member-list compact">
                      {bulkTeacherItems.map((item) => (
                        <li key={item.user_id}>
                          <code>{item.user_id}</code>
                          <small>
                            {item.role === 'assistant'
                              ? 'Trợ giảng'
                              : 'Giáo viên'}
                          </small>
                          <button
                            type="button"
                            onClick={() => onRemoveBulkTeacher(item.user_id)}
                          >
                            Xóa
                          </button>
                        </li>
                      ))}
                    </ul>
                    <div className="form-actions">
                      <button
                        type="button"
                        onClick={() => onBulkAssignTeachers(true)}
                        disabled={bulkLoading}
                      >
                        {bulkLoading && !bulkPreview
                          ? 'Đang kiểm tra…'
                          : 'Kiểm tra'}
                      </button>
                      <button
                        type="button"
                        className="primary"
                        onClick={() => onBulkAssignTeachers(false)}
                        disabled={bulkLoading}
                      >
                        {bulkLoading && bulkPreview
                          ? 'Đang phân công…'
                          : 'Xác nhận phân công'}
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}

            {bulkPreview && (
              <div className="bulk-preview">
                <p>
                  Tổng: <strong>{bulkPreview.total}</strong> · Thành công:{' '}
                  <strong>
                    {'enrolled' in bulkPreview
                      ? bulkPreview.enrolled
                      : bulkPreview.assigned}
                  </strong>{' '}
                  · Lỗi: <strong>{bulkPreview.failed}</strong>
                  {bulkPreview.dry_run && (
                    <span className="dry-run-badge">Chế độ kiểm tra</span>
                  )}
                </p>
                <div className="table-wrap">
                  <table className="gradebook-table">
                    <thead>
                      <tr>
                        <th>User ID</th>
                        <th>Trạng thái</th>
                        <th>Lỗi</th>
                      </tr>
                    </thead>
                    <tbody>
                      {bulkPreview.rows.map((row, idx) => (
                        <tr key={idx}>
                          <td>
                            <code>{row.user_id}</code>
                          </td>
                          <td>
                            <span className={`status-badge ${row.status}`}>
                              {row.status}
                            </span>
                          </td>
                          <td>{row.error || '—'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function UserPicker({
  role,
  onSelect,
  buttonLabel,
}: {
  role: 'teacher' | 'student';
  onSelect: (userId: string) => void;
  buttonLabel: string;
}) {
  const [query, setQuery] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [selected, setSelected] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (!query.trim()) {
        setUsers([]);
        return;
      }
      setLoading(true);
      listUsers({ q: query.trim(), limit: 10 })
        .then((res) => {
          const filtered = res.data.filter((u) => u.roles.includes(role));
          setUsers(filtered);
        })
        .catch(() => setUsers([]))
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(timer);
  }, [query, role]);

  function handleSelect(user: User) {
    setSelected(user);
    setUsers([]);
    setQuery('');
  }

  function handleConfirm() {
    if (selected) {
      onSelect(selected.id);
      setSelected(null);
    }
  }

  return (
    <div className="academic-user-picker">
      {selected ? (
        <div className="inline-form">
          <span>
            {selected.display_name} ({selected.login_name})
          </span>
          <button type="button" className="primary" onClick={handleConfirm}>
            {buttonLabel}
          </button>
          <button type="button" onClick={() => setSelected(null)}>
            Hủy
          </button>
        </div>
      ) : (
        <>
          <input
            type="search"
            placeholder={`Tìm ${role === 'teacher' ? 'giáo viên' : 'học sinh'}…`}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          {loading && <p className="dashboard-status">Đang tìm…</p>}
          {users.length > 0 && (
            <ul className="academic-user-results">
              {users.map((u) => (
                <li key={u.id}>
                  <button type="button" onClick={() => handleSelect(u)}>
                    {u.display_name} ({u.login_name})
                  </button>
                </li>
              ))}
            </ul>
          )}
          {!loading && query && users.length === 0 && (
            <p className="dashboard-status">Không tìm thấy.</p>
          )}
        </>
      )}
    </div>
  );
}
