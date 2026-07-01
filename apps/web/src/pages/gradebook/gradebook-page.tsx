import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { listAssessments } from '@/shared/api/assessments';
import { listClasses } from '@/shared/api/academics';
import {
  listAssessmentAttempts,
  getAssessmentResults,
  getClassGradebook,
  exportAssessmentAttemptsCSV,
  exportClassGradebookCSV,
  type AssessmentAttempt,
  type AssessmentResult,
  type ClassGradebookEntry,
} from '@/shared/api/gradebook';
import type { AssessmentListItem } from '@/shared/api/assessments';
import type { ClassSectionList } from '@/shared/api/academics';

function formatFriendlyError(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === 'string') return err;
  return 'Đã xảy ra lỗi.';
}

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

  const [assessments, setAssessments] = useState<AssessmentListItem[]>([]);
  const [classes, setClasses] = useState<ClassSectionList['data']>([]);
  const [metaLoading, setMetaLoading] = useState(true);
  const [metaError, setMetaError] = useState<string | null>(null);

  const [attempts, setAttempts] = useState<AssessmentAttempt[]>([]);
  const [results, setResults] = useState<AssessmentResult | null>(null);
  const [classEntries, setClassEntries] = useState<ClassGradebookEntry[]>([]);

  const [detailLoading, setDetailLoading] = useState(false);
  const [detailError, setDetailError] = useState<string | null>(null);

  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setMetaLoading(true);
      setMetaError(null);
      try {
        const [assessmentList, classList] = await Promise.all([
          listAssessments({ limit: 100 }),
          listClasses(),
        ]);
        if (cancelled) return;
        setAssessments((assessmentList.data ?? []) as AssessmentListItem[]);
        setClasses(classList.data ?? []);
      } catch (err) {
        if (cancelled) return;
        setMetaError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setMetaLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (tab !== 'assessment' || !assessmentId) {
      setAttempts([]);
      setResults(null);
      return;
    }
    let cancelled = false;
    async function load() {
      setDetailLoading(true);
      setDetailError(null);
      try {
        const [attemptList, result] = await Promise.all([
          listAssessmentAttempts(assessmentId),
          getAssessmentResults(assessmentId),
        ]);
        if (cancelled) return;
        setAttempts(attemptList);
        setResults(result);
      } catch (err) {
        if (cancelled) return;
        setDetailError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setDetailLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [tab, assessmentId]);

  useEffect(() => {
    if (tab !== 'class' || !classId) {
      setClassEntries([]);
      return;
    }
    let cancelled = false;
    async function load() {
      setDetailLoading(true);
      setDetailError(null);
      try {
        const entries = await getClassGradebook(classId);
        if (cancelled) return;
        setClassEntries(entries);
      } catch (err) {
        if (cancelled) return;
        setDetailError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setDetailLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [tab, classId]);

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
    setExporting(true);
    setExportError(null);
    try {
      await exportAssessmentAttemptsCSV(assessmentId);
    } catch (err) {
      setExportError(formatFriendlyError(err));
    } finally {
      setExporting(false);
    }
  }

  async function handleExportClass() {
    if (!classId) return;
    setExporting(true);
    setExportError(null);
    try {
      await exportClassGradebookCSV(classId);
    } catch (err) {
      setExportError(formatFriendlyError(err));
    } finally {
      setExporting(false);
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

      <div className="gradebook-tabs" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'assessment'}
          className={tab === 'assessment' ? 'active' : ''}
          onClick={() => switchTab('assessment')}
        >
          Theo đề thi
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'class'}
          className={tab === 'class' ? 'active' : ''}
          onClick={() => switchTab('class')}
        >
          Theo lớp
        </button>
      </div>

      {metaLoading && <p className="dashboard-status">Đang tải danh sách…</p>}
      {metaError && (
        <div className="error-banner" role="alert">
          {metaError}
        </div>
      )}

      {!metaLoading && !metaError && tab === 'assessment' && (
        <section className="dashboard-section">
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
                <div className="summary-cards">
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
                  data-testid="export-assessment-csv"
                >
                  {exporting ? 'Đang xuất…' : 'Xuất CSV'}
                </button>
              </div>

              {exportError && (
                <div className="error-banner" role="alert">
                  {exportError}
                </div>
              )}

              {detailLoading && (
                <p className="dashboard-status">Đang tải kết quả…</p>
              )}
              {detailError && (
                <div className="error-banner" role="alert">
                  {detailError}
                </div>
              )}

              {!detailLoading && !detailError && (
                <div className="table-wrap">
                  <table className="gradebook-table" data-testid="gradebook-table">
                    <thead>
                      <tr>
                        <th>Học sinh</th>
                        <th>Trạng thái</th>
                        <th>Bắt đầu</th>
                        <th>Nộp</th>
                        <th>Điểm</th>
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
                              <span className={`status-badge ${attempt.status.toLowerCase()}`}>
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
        <section className="dashboard-section">
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
                >
                  {exporting ? 'Đang xuất…' : 'Xuất CSV'}
                </button>
              </div>

              {exportError && (
                <div className="error-banner" role="alert">
                  {exportError}
                </div>
              )}

              {detailLoading && (
                <p className="dashboard-status">Đang tải sổ điểm…</p>
              )}
              {detailError && (
                <div className="error-banner" role="alert">
                  {detailError}
                </div>
              )}

              {!detailLoading && !detailError && (
                <div className="table-wrap">
                  <table className="gradebook-table">
                    <thead>
                      <tr>
                        <th>Học sinh</th>
                        {Array.from(classAssessments.entries()).map(([id, title]) => (
                          <th key={id}>{title}</th>
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
                                        >
                                          {entry.status ?? '—'}
                                        </span>
                                      </>
                                    ) : (
                                      '—'
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
