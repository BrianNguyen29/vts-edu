import { useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useAssessments } from '@/shared/api/assessments-queries';
import { useClasses } from '@/shared/api/academics-queries';
import {
  useAssessmentAttempts,
  useAssessmentResults,
  useClassGradebook,
  useExportAssessmentAttemptsCSV,
  useExportClassGradebookCSV,
} from '@/shared/api/gradebook-queries';
import type { ClassGradebookEntry } from '@/shared/api/gradebook';
import type { AssessmentListItem } from '@/shared/api/assessments';
import type { ClassSectionList } from '@/shared/api/academics';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function formatDateTime(iso: string | undefined | null): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString('vi-VN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  });
}

function formatScore(score: string | undefined | null): string {
  if (score === undefined || score === null || score === '') return '—';
  return score;
}

export function GradebookPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const assessmentId = searchParams.get('assessment') ?? '';
  const classId = searchParams.get('class') ?? '';
  const initialTab = searchParams.get('tab') ?? (classId ? 'class' : 'assessment');
  const [tab, setTab] = useState<'assessment' | 'class'>(
    initialTab === 'class' ? 'class' : 'assessment'
  );

  useDocumentTitle('Sổ điểm');

  const {
    data: assessmentsData,
    isPending: assessmentsLoading,
    error: assessmentsError,
  } = useAssessments({ limit: 100 });
  const {
    data: classesData,
    isPending: classesLoading,
    error: classesError,
  } = useClasses();

  const assessments = (assessmentsData?.data ?? []) as AssessmentListItem[];
  const classes = (classesData?.data ?? []) as ClassSectionList['data'];

  const metaLoading = assessmentsLoading || classesLoading;
  const metaError = assessmentsError || classesError;

  const {
    data: attempts = [],
    isPending: attemptsLoading,
    error: attemptsError,
  } = useAssessmentAttempts(tab === 'assessment' ? assessmentId : undefined);
  const {
    data: results = null,
    isPending: resultsLoading,
    error: resultsError,
  } = useAssessmentResults(tab === 'assessment' ? assessmentId : undefined);

  const {
    data: classEntries = [],
    isPending: classGradebookLoading,
    error: classGradebookError,
  } = useClassGradebook(tab === 'class' ? classId : undefined);

  const detailLoading =
    tab === 'assessment'
      ? attemptsLoading || resultsLoading
      : classGradebookLoading;
  const detailError =
    tab === 'assessment' ? attemptsError || resultsError : classGradebookError;

  const exportAssessment = useExportAssessmentAttemptsCSV();
  const exportClass = useExportClassGradebookCSV();
  const exporting = exportAssessment.isPending || exportClass.isPending;
  const exportError = exportAssessment.error || exportClass.error;

  function selectAssessment(id: string) {
    setSearchParams({ tab: 'assessment', assessment: id }, { replace: true });
  }

  function selectClass(id: string) {
    setSearchParams({ tab: 'class', class: id }, { replace: true });
  }

  function switchTab(next: 'assessment' | 'class') {
    setTab(next);
    if (next === 'assessment') {
      const first = assessments[0]?.id;
      setSearchParams(
        { tab: 'assessment', assessment: assessmentId || first || '' },
        { replace: true }
      );
    } else {
      const first = classes[0]?.id;
      setSearchParams(
        { tab: 'class', class: classId || first || '' },
        { replace: true }
      );
    }
  }

  async function handleExportAssessment() {
    if (!assessmentId) return;
    try {
      await exportAssessment.mutateAsync(assessmentId);
    } catch {
      // Error surfaced via exportError below.
    }
  }

  async function handleExportClass() {
    if (!classId) return;
    try {
      await exportClass.mutateAsync(classId);
    } catch {
      // Error surfaced via exportError below.
    }
  }

  const selectedAssessment = useMemo(
    () => assessments.find((a) => a.id === assessmentId),
    [assessments, assessmentId]
  );

  const selectedClass = useMemo(
    () => classes.find((c) => c.id === classId),
    [classes, classId]
  );

  const classGradebookByStudent = useMemo(() => {
    const map = new Map<string, { name: string; entries: ClassGradebookEntry[] }>();
    for (const entry of classEntries) {
      const row = map.get(entry.student_user_id);
      if (row) {
        row.entries.push(entry);
      } else {
        map.set(entry.student_user_id, { name: entry.student_name, entries: [entry] });
      }
    }
    return map;
  }, [classEntries]);

  const classAssessments = useMemo(() => {
    const titles = new Map<string, string>();
    for (const entry of classEntries) {
      if (!titles.has(entry.assessment_id)) {
        titles.set(entry.assessment_id, entry.assessment_title);
      }
    }
    return titles;
  }, [classEntries]);

  return (
    <div className="dashboard-page gradebook-page">
      <Link to="/app/teacher" className="back-link">
        ← Quay lại trang giáo viên
      </Link>
      <h1>Sổ điểm</h1>

      <div className="gradebook-tabs" role="tablist" aria-label="Chế độ xem sổ điểm">
        <button
          type="button"
          role="tab"
          id="gradebook-tab-assessment"
          aria-controls="gradebook-panel-assessment"
          aria-selected={tab === 'assessment'}
          tabIndex={tab === 'assessment' ? 0 : -1}
          className={tab === 'assessment' ? 'active' : ''}
          onClick={() => switchTab('assessment')}
        >
          Theo đề thi
        </button>
        <button
          type="button"
          role="tab"
          id="gradebook-tab-class"
          aria-controls="gradebook-panel-class"
          aria-selected={tab === 'class'}
          tabIndex={tab === 'class' ? 0 : -1}
          className={tab === 'class' ? 'active' : ''}
          onClick={() => switchTab('class')}
        >
          Theo lớp
        </button>
      </div>

      {metaLoading && (
        <p className="dashboard-status" role="status" aria-live="polite">
          Đang tải danh sách…
        </p>
      )}
      {!!metaError && <ErrorState error={metaError} />}

      {!metaLoading && !metaError && tab === 'assessment' && (
        <section
          className="dashboard-section"
          role="tabpanel"
          id="gradebook-panel-assessment"
          aria-labelledby="gradebook-tab-assessment"
        >
          <div className="gradebook-selector">
            <label htmlFor="gradebook-assessment">Chọn đề thi</label>
            <select
              id="gradebook-assessment"
              value={assessmentId}
              onChange={(e) => selectAssessment(e.target.value)}
              data-testid="gradebook-assessment-select"
            >
              <option value="">Chọn đề thi…</option>
              {assessments.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.title}
                </option>
              ))}
            </select>
          </div>

          {assessmentId && selectedAssessment && (
            <div className="gradebook-summary">
              <h2>{selectedAssessment.title}</h2>
              {results && (
                <div className="summary-cards" aria-label="Tóm tắt kết quả đề thi">
                  <div className="summary-card">
                    <span className="summary-value">{results.total_attempts}</span>
                    <span className="summary-label">Tổng lần làm</span>
                  </div>
                  <div className="summary-card">
                    <span className="summary-value">{results.submitted_count}</span>
                    <span className="summary-label">Đã nộp</span>
                  </div>
                  <div className="summary-card">
                    <span className="summary-value">{results.in_progress_count}</span>
                    <span className="summary-label">Đang làm</span>
                  </div>
                  <div className="summary-card">
                    <span className="summary-value">{results.expired_count}</span>
                    <span className="summary-label">Hết hạn</span>
                  </div>
                  <div className="summary-card">
                    <span className="summary-value">
                      {formatScore(results.average_score)}
                    </span>
                    <span className="summary-label">Điểm trung bình</span>
                  </div>
                  <div className="summary-card">
                    <span className="summary-value">{formatScore(results.max_score)}</span>
                    <span className="summary-label">Điểm tối đa</span>
                  </div>
                </div>
              )}

              <div className="gradebook-actions">
                <button
                  type="button"
                  className="primary"
                  onClick={handleExportAssessment}
                  disabled={exporting || attempts.length === 0}
                  aria-busy={exporting}
                  data-testid="export-assessment-csv"
                >
                  {exporting ? 'Đang xuất…' : 'Xuất CSV'}
                </button>
              </div>

              {!!exportError && <ErrorState error={exportError} />}

              {detailLoading && (
                <p className="dashboard-status" role="status" aria-live="polite">
                  Đang tải kết quả…
                </p>
              )}
              {!!detailError && <ErrorState error={detailError} />}

              {!detailLoading && !detailError && (
                <div className="table-wrap">
                  <table className="gradebook-table" data-testid="gradebook-table">
                    <caption className="visually-hidden">
                      Bảng điểm theo học sinh cho đề thi {selectedAssessment.title}
                    </caption>
                    <thead>
                      <tr>
                        <th scope="col">Học sinh</th>
                        <th scope="col">Trạng thái</th>
                        <th scope="col">Bắt đầu</th>
                        <th scope="col">Nộp</th>
                        <th scope="col">Điểm</th>
                      </tr>
                    </thead>
                    <tbody>
                      {attempts.length === 0 ? (
                        <tr>
                          <td colSpan={5} className="empty-cell">
                            Chưa có bài làm nào.
                          </td>
                        </tr>
                      ) : (
                        attempts.map((attempt) => (
                          <tr key={attempt.id}>
                            <td>{attempt.student_name || attempt.student_user_id}</td>
                            <td>
                              <span
                                className={`status-badge ${attempt.status.toLowerCase()}`}
                                aria-label={`Trạng thái ${attempt.status}`}
                              >
                                {attempt.status}
                              </span>
                            </td>
                            <td>{formatDateTime(attempt.started_at)}</td>
                            <td>{formatDateTime(attempt.submitted_at)}</td>
                            <td>
                              {formatScore(attempt.score)}
                              {attempt.max_score ? ` / ${formatScore(attempt.max_score)}` : ''}
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {!assessmentId && (
            <p className="dashboard-status">Vui lòng chọn một đề thi để xem sổ điểm.</p>
          )}
        </section>
      )}

      {!metaLoading && !metaError && tab === 'class' && (
        <section
          className="dashboard-section"
          role="tabpanel"
          id="gradebook-panel-class"
          aria-labelledby="gradebook-tab-class"
        >
          <div className="gradebook-selector">
            <label htmlFor="gradebook-class">Chọn lớp</label>
            <select
              id="gradebook-class"
              value={classId}
              onChange={(e) => selectClass(e.target.value)}
            >
              <option value="">Chọn lớp…</option>
              {classes.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.student_count} học sinh)
                </option>
              ))}
            </select>
          </div>

          {classId && selectedClass && (
            <div className="gradebook-summary">
              <h2>Lớp {selectedClass.name}</h2>

              <div className="gradebook-actions">
                <button
                  type="button"
                  className="primary"
                  onClick={handleExportClass}
                  disabled={exporting || classEntries.length === 0}
                  aria-busy={exporting}
                >
                  {exporting ? 'Đang xuất…' : 'Xuất CSV'}
                </button>
              </div>

              {!!exportError && <ErrorState error={exportError} />}

              {detailLoading && (
                <p className="dashboard-status" role="status" aria-live="polite">
                  Đang tải sổ điểm…
                </p>
              )}
              {!!detailError && <ErrorState error={detailError} />}

              {!detailLoading && !detailError && (
                <div className="table-wrap">
                  <table className="gradebook-table">
                    <caption className="visually-hidden">
                      Bảng điểm tổng hợp theo học sinh cho lớp {selectedClass.name}
                    </caption>
                    <thead>
                      <tr>
                        <th scope="col">Học sinh</th>
                        {Array.from(classAssessments.entries()).map(([id, title]) => (
                          <th scope="col" key={id}>
                            {title}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {classGradebookByStudent.size === 0 ? (
                        <tr>
                          <td
                            colSpan={Math.max(classAssessments.size + 1, 2)}
                            className="empty-cell"
                          >
                            Chưa có dữ liệu điểm cho lớp này.
                          </td>
                        </tr>
                      ) : (
                        Array.from(classGradebookByStudent.entries()).map(
                          ([studentId, { name, entries }]) => (
                            <tr key={studentId}>
                              <td>{name}</td>
                              {Array.from(classAssessments.entries()).map(([assessmentID]) => {
                                const entry = entries.find((e) => e.assessment_id === assessmentID);
                                return (
                                  <td key={assessmentID}>
                                    {entry ? (
                                      <>
                                        {formatScore(entry.score)}
                                        {entry.max_score ? ` / ${formatScore(entry.max_score)}` : ''}
                                        <span
                                          className={`status-badge small ${
                                            entry.status?.toLowerCase() ?? 'empty'
                                          }`}
                                          aria-label={`Trạng thái ${entry.status ?? 'chưa có'}`}
                                        >
                                          {entry.status ?? '—'}
                                        </span>
                                      </>
                                    ) : (
                                      <span aria-label="Chưa có điểm">—</span>
                                    )}
                                  </td>
                                );
                              })}
                            </tr>
                          )
                        )
                      )}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {!classId && (
            <p className="dashboard-status">Vui lòng chọn một lớp để xem sổ điểm.</p>
          )}
        </section>
      )}
    </div>
  );
}
