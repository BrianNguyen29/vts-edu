import { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { useCreateAssessment, useInfiniteAssessments } from '@/shared/api/assessments-queries';
import { useClasses } from '@/shared/api/academics-queries';
import type { AssessmentListItem } from '@/shared/api/assessments';
import type { ClassSection } from '@/shared/api/academics';
import { ErrorState } from '@/shared/components/error-state';
import { ClassRosterPanel } from './class-roster-panel';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

export function TeacherDashboardPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  useDocumentTitle('Trang giáo viên');

  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const {
    data: assessmentsData,
    isPending: assessmentsLoading,
    isFetchingNextPage,
    error: assessmentsError,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteAssessments(searchQuery);

  const assessments = useMemo(
    () => (assessmentsData?.pages.flatMap((p) => p.data) ?? []) as AssessmentListItem[],
    [assessmentsData]
  );

  const {
    data: classesData,
    isPending: classesLoading,
    error: classesError,
  } = useClasses();
  const classes = (classesData?.data ?? []) as ClassSection[];

  const [selectedClass, setSelectedClass] = useState<ClassSection | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createTitle, setCreateTitle] = useState('');
  const [createClassId, setCreateClassId] = useState('');
  const [createError, setCreateError] = useState<unknown | null>(null);
  const createAssessment = useCreateAssessment();

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  async function handleCreateAssessment(e: React.FormEvent) {
    e.preventDefault();
    if (!createClassId || !createTitle.trim()) return;
    setCreateError(null);
    try {
      const assessment = await createAssessment.mutateAsync({
        classSectionId: createClassId,
        title: createTitle.trim(),
        duration_minutes: 45,
        max_attempts: 1,
      });
      navigate(`/app/teacher/assessments/${assessment.id}`);
    } catch (err) {
      setCreateError(err);
    }
  }

  return (
    <div className="dashboard-page">
      <h1>Trang giáo viên</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>

      <section className="dashboard-section" aria-labelledby="teacher-assessments-heading">
        <h2 id="teacher-assessments-heading">Đề thi</h2>

        <div className="search-bar">
          <label htmlFor="teacher-assessment-search" className="visually-hidden">
            Tìm theo tên đề thi
          </label>
          <input
            id="teacher-assessment-search"
            type="search"
            placeholder="Tìm theo tên đề thi…"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            data-testid="teacher-assessment-search"
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
          <form onSubmit={handleCreateAssessment} className="admin-form" data-testid="create-assessment-form" aria-labelledby="create-assessment-heading">
            <h3 id="create-assessment-heading">Tạo đề thi mới</h3>
            {!!createError && (
              <ErrorState
                error={createError}
                overrides={{
                  403: 'Bạn không có quyền tạo đề thi.',
                }}
              />
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
              <button type="submit" className="primary" disabled={createAssessment.isPending} aria-busy={createAssessment.isPending} data-testid="create-assessment-submit">
                {createAssessment.isPending ? 'Đang tạo…' : 'Tạo đề thi'}
              </button>
              <button
                type="button"
                onClick={() => setShowCreateForm(false)}
                disabled={createAssessment.isPending}
              >
                Hủy
              </button>
            </div>
          </form>
        )}

        {assessmentsLoading && (
          <p className="dashboard-status" role="status" aria-live="polite">
            Đang tải danh sách đề thi…
          </p>
        )}

        {!!assessmentsError && !assessmentsLoading && (
          <ErrorState error={assessmentsError} />
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
            <ul className="assessment-list" aria-label="Danh sách đề thi">
              {assessments.map((assessment) => (
                <li key={assessment.id} className="assessment-item">
                  <div className="assessment-info">
                    <div className="assessment-title">{assessment.title}</div>
                    <div className="assessment-meta">
                      <span className="assessment-status" aria-label={`Trạng thái ${assessment.status}`}>
                        {assessment.status}
                      </span>
                      <span aria-hidden="true">·</span>
                      <span className="assessment-duration">
                        {assessment.duration_minutes} phút
                      </span>
                    </div>
                  </div>
                  <Link
                    to={`/app/teacher/gradebook?tab=assessment&assessment=${assessment.id}`}
                    className="button-link"
                    aria-label={`Mở sổ điểm cho đề thi ${assessment.title}`}
                  >
                    Sổ điểm
                  </Link>
                </li>
              ))}
            </ul>
            {hasNextPage && (
              <div className="load-more">
                <button
                  type="button"
                  onClick={() => fetchNextPage()}
                  disabled={isFetchingNextPage}
                  aria-busy={isFetchingNextPage}
                >
                  {isFetchingNextPage ? 'Đang tải…' : 'Tải thêm'}
                </button>
              </div>
            )}
          </>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="teacher-classes-heading">
        <h2 id="teacher-classes-heading">Lớp học</h2>

        {selectedClass ? (
          <ClassRosterPanel
            classSection={selectedClass}
            onClose={() => setSelectedClass(null)}
          />
        ) : (
          <>
            {classesLoading && (
              <p className="dashboard-status" role="status" aria-live="polite">
                Đang tải danh sách lớp…
              </p>
            )}

            {!!classesError && !classesLoading && (
              <ErrorState error={classesError} />
            )}

            {!classesLoading && !classesError && classes.length === 0 && (
              <p className="dashboard-status">Bạn chưa được phân công lớp nào.</p>
            )}

            {!classesLoading && !classesError && classes.length > 0 && (
              <ul className="class-list" aria-label="Danh sách lớp">
                {classes.map((classSection) => (
                  <li key={classSection.id} className="class-item">
                    <button
                      type="button"
                      className="class-button"
                      onClick={() => setSelectedClass(classSection)}
                    >
                      <span className="class-name">{classSection.name}</span>
                      <span className="class-meta" aria-hidden="true">
                        {classSection.student_count} học sinh ·{' '}
                        {classSection.teacher_count} giáo viên
                      </span>
                      <span className="visually-hidden">
                        {`${classSection.student_count} học sinh, ${classSection.teacher_count} giáo viên`}
                      </span>
                    </button>
                    <Link
                      to={`/app/teacher/gradebook?tab=class&class=${classSection.id}`}
                      className="button-link"
                      aria-label={`Mở sổ điểm cho lớp ${classSection.name}`}
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
          <span className="card-link-action" aria-hidden="true">Mở sổ điểm →</span>
          <span className="visually-hidden">Mở trang sổ điểm</span>
        </Link>
      </div>
    </div>
  );
}
