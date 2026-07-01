import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { ApiResponseError } from '@/shared/api/attempts';
import {
  createAssessment,
  listAssessments,
  type AssessmentListItem,
} from '@/shared/api/assessments';
import { listClasses, type ClassSection } from '@/shared/api/academics';
import { ClassRosterPanel } from './class-roster-panel';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền xem nội dung này.';
      case 404:
        return 'Không tìm thấy dữ liệu.';
      default:
        return err.body.error.message || 'Không thể tải dữ liệu.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function TeacherDashboardPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  const [assessments, setAssessments] = useState<AssessmentListItem[]>([]);
  const [assessmentsLoading, setAssessmentsLoading] = useState(true);
  const [assessmentsError, setAssessmentsError] = useState<string | null>(null);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [assessmentCursor, setAssessmentCursor] = useState<string | undefined>();
  const [assessmentHasMore, setAssessmentHasMore] = useState(false);
  const [isLoadingMoreAssessments, setIsLoadingMoreAssessments] = useState(false);

  const [classes, setClasses] = useState<ClassSection[]>([]);
  const [classesLoading, setClassesLoading] = useState(true);
  const [classesError, setClassesError] = useState<string | null>(null);
  const [selectedClass, setSelectedClass] = useState<ClassSection | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createTitle, setCreateTitle] = useState('');
  const [createClassId, setCreateClassId] = useState('');
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setAssessmentsLoading(true);
      setAssessmentsError(null);
      setAssessmentCursor(undefined);
      setAssessmentHasMore(false);
      try {
        const response = await listAssessments({
          q: searchQuery || undefined,
          limit: 10,
        });
        if (cancelled) return;
        setAssessments(response.data);
        setAssessmentCursor(response.page?.next_cursor ?? undefined);
        setAssessmentHasMore(response.page?.has_more ?? false);
      } catch (err) {
        if (cancelled) return;
        setAssessmentsError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setAssessmentsLoading(false);
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, [searchQuery]);

  async function loadMoreAssessments() {
    if (!assessmentHasMore || !assessmentCursor || isLoadingMoreAssessments) return;
    setIsLoadingMoreAssessments(true);
    try {
      const response = await listAssessments({
        q: searchQuery || undefined,
        limit: 10,
        cursor: assessmentCursor,
      });
      setAssessments((prev) => [...prev, ...response.data]);
      setAssessmentCursor(response.page?.next_cursor ?? undefined);
      setAssessmentHasMore(response.page?.has_more ?? false);
    } catch (err) {
      setAssessmentsError(formatFriendlyError(err));
    } finally {
      setIsLoadingMoreAssessments(false);
    }
  }

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  async function handleCreateAssessment(e: React.FormEvent) {
    e.preventDefault();
    if (!createClassId || !createTitle.trim()) return;
    setCreating(true);
    setCreateError(null);
    try {
      const assessment = await createAssessment(createClassId, {
        title: createTitle.trim(),
        duration_minutes: 45,
        max_attempts: 1,
      });
      navigate(`/app/teacher/assessments/${assessment.id}`);
    } catch (err) {
      setCreating(false);
      setCreateError(formatFriendlyError(err));
    }
  }

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setClassesLoading(true);
      setClassesError(null);
      try {
        const response = await listClasses();
        if (cancelled) return;
        setClasses(response.data);
      } catch (err) {
        if (cancelled) return;
        setClassesError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setClassesLoading(false);
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="dashboard-page">
      <h1>Trang giáo viên</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>

      <section className="dashboard-section">
        <h2>Đề thi</h2>

        <div className="search-bar">
          <input
            type="search"
            placeholder="Tìm theo tên đề thi…"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>

        {!showCreateForm && (
          <button
            type="button"
            className="primary"
            onClick={() => setShowCreateForm(true)}
            style={{ marginBottom: '1rem' }}
            data-testid="create-assessment-button"
          >
            + Tạo đề thi
          </button>
        )}

        {showCreateForm && (
          <form onSubmit={handleCreateAssessment} className="admin-form" data-testid="create-assessment-form">
            <h3>Tạo đề thi mới</h3>
            {createError && (
              <div className="error-banner" role="alert">
                {createError}
              </div>
            )}
            <div className="field">
              <label htmlFor="create-title">Tên đề thi</label>
              <input
                id="create-title"
                type="text"
                value={createTitle}
                onChange={(e) => setCreateTitle(e.target.value)}
                placeholder="Ví dụ: Kiểm tra 15 phút"
                required
                data-testid="create-assessment-title"
              />
            </div>
            <div className="field">
              <label htmlFor="create-class">Lớp</label>
              <select
                id="create-class"
                value={createClassId}
                onChange={(e) => setCreateClassId(e.target.value)}
                required
                disabled={classesLoading || classes.length === 0}
                data-testid="create-assessment-class"
              >
                <option value="">
                  {classesLoading
                    ? 'Đang tải…'
                    : classes.length === 0
                    ? 'Không có lớp'
                    : 'Chọn lớp…'}
                </option>
                {classes.map((cls) => (
                  <option key={cls.id} value={cls.id}>
                    {cls.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-actions">
              <button type="submit" className="primary" disabled={creating} data-testid="create-assessment-submit">
                {creating ? 'Đang tạo…' : 'Tạo đề thi'}
              </button>
              <button
                type="button"
                onClick={() => setShowCreateForm(false)}
                disabled={creating}
              >
                Hủy
              </button>
            </div>
          </form>
        )}

        {assessmentsLoading && (
          <p className="dashboard-status">Đang tải danh sách đề thi…</p>
        )}

        {assessmentsError && !assessmentsLoading && (
          <div className="error-banner" role="alert">
            {assessmentsError}
          </div>
        )}

        {!assessmentsLoading &&
          !assessmentsError &&
          assessments.length === 0 &&
          (searchQuery ? (
            <p className="dashboard-status">Không tìm thấy đề thi phù hợp.</p>
          ) : (
            <p className="dashboard-status">Chưa có đề thi nào.</p>
          ))}

        {!assessmentsLoading && !assessmentsError && assessments.length > 0 && (
          <>
            <ul className="assessment-list">
              {assessments.map((assessment) => (
                <li key={assessment.id} className="assessment-item">
                  <div className="assessment-info">
                    <div className="assessment-title">{assessment.title}</div>
                    <div className="assessment-meta">
                      <span className="assessment-status">{assessment.status}</span>
                      <span className="assessment-duration">
                        {assessment.duration_minutes} phút
                      </span>
                    </div>
                  </div>
                  <Link
                    to={`/app/teacher/gradebook?tab=assessment&assessment=${assessment.id}`}
                    className="button-link"
                  >
                    Sổ điểm
                  </Link>
                </li>
              ))}
            </ul>
            {assessmentHasMore && (
              <div className="load-more">
                <button
                  type="button"
                  onClick={loadMoreAssessments}
                  disabled={isLoadingMoreAssessments}
                >
                  {isLoadingMoreAssessments ? 'Đang tải…' : 'Tải thêm'}
                </button>
              </div>
            )}
          </>
        )}
      </section>

      <section className="dashboard-section">
        <h2>Lớp học</h2>

        {selectedClass ? (
          <ClassRosterPanel
            classSection={selectedClass}
            onClose={() => setSelectedClass(null)}
          />
        ) : (
          <>
            {classesLoading && (
              <p className="dashboard-status">Đang tải danh sách lớp…</p>
            )}

            {classesError && !classesLoading && (
              <div className="error-banner" role="alert">
                {classesError}
              </div>
            )}

            {!classesLoading && !classesError && classes.length === 0 && (
              <p className="dashboard-status">Bạn chưa được phân công lớp nào.</p>
            )}

            {!classesLoading && !classesError && classes.length > 0 && (
              <ul className="class-list">
                {classes.map((classSection) => (
                  <li key={classSection.id} className="class-item">
                    <button
                      type="button"
                      className="class-button"
                      onClick={() => setSelectedClass(classSection)}
                    >
                      <span className="class-name">{classSection.name}</span>
                      <span className="class-meta">
                        {classSection.student_count} học sinh ·{' '}
                        {classSection.teacher_count} giáo viên
                      </span>
                    </button>
                    <Link
                      to={`/app/teacher/gradebook?tab=class&class=${classSection.id}`}
                      className="button-link"
                    >
                      Sổ điểm
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
      </section>

      <div className="dashboard-cards">
        <Link to="/app/teacher/gradebook" className="dashboard-card card-link">
          <h2>Chấm điểm</h2>
          <p>Xem bài làm và xuất điểm theo đề thi hoặc lớp.</p>
          <span className="card-link-action">Mở sổ điểm →</span>
        </Link>
      </div>
    </div>
  );
}
