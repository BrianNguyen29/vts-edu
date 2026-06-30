import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ApiResponseError } from '@/shared/api/attempts';
import {
  createItem,
  createSection,
  createTarget,
  getAssessment,
  publishAssessment,
  updateAssessment,
  validateAssessment,
  type AssessmentDetail,
  type Section,
  type Target,
  type ValidationResult,
} from '@/shared/api/assessments';
import { listClasses, type ClassSection } from '@/shared/api/academics';

const DEMO_QUESTION_VERSION_ID = '00000000-0000-4000-8000-000000000002';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền thực hiện thao tác này.';
      case 404:
        return 'Không tìm thấy đề thi.';
      default:
        return err.body.error.message || 'Không thể thực hiện thao tác.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function AssessmentBuilderPage() {
  const { assessmentId } = useParams<{ assessmentId: string }>();

  const [assessment, setAssessment] = useState<AssessmentDetail | null>(null);
  const [classes, setClasses] = useState<ClassSection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [title, setTitle] = useState('');
  const [duration, setDuration] = useState('45');
  const [instructions, setInstructions] = useState('');
  const [savingSettings, setSavingSettings] = useState(false);

  const [sectionTitle, setSectionTitle] = useState('');
  const [addingSection, setAddingSection] = useState(false);

  const [targetClassId, setTargetClassId] = useState('');
  const [addingTarget, setAddingTarget] = useState(false);

  const [validation, setValidation] = useState<ValidationResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!assessmentId) return;
      setLoading(true);
      setError(null);
      try {
        const [assessmentData, classesData] = await Promise.all([
          getAssessment(assessmentId),
          listClasses(),
        ]);
        if (cancelled) return;
        setAssessment(assessmentData);
        setClasses(classesData.data);
        setTitle(assessmentData.title);
        setDuration(String(assessmentData.duration_minutes));
        setInstructions(assessmentData.instructions ?? '');
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
  }, [assessmentId]);

  function clearMessages() {
    setError(null);
    setSuccess(null);
  }

  async function handleSaveSettings(e: React.FormEvent) {
    e.preventDefault();
    if (!assessmentId) return;
    clearMessages();
    setSavingSettings(true);
    try {
      const updated = await updateAssessment(assessmentId, {
        title: title.trim(),
        duration_minutes: Number(duration),
        instructions: instructions.trim() || null,
      });
      setAssessment(updated);
      setSuccess('Đã cập nhật cài đặt đề thi.');
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setSavingSettings(false);
    }
  }

  async function handleAddSection(e: React.FormEvent) {
    e.preventDefault();
    if (!assessmentId || !sectionTitle.trim()) return;
    clearMessages();
    setAddingSection(true);
    try {
      const section = await createSection(assessmentId, {
        title: sectionTitle.trim(),
        position: (assessment?.sections.length ?? 0) + 1,
      });
      setAssessment((prev) =>
        prev
          ? { ...prev, sections: [...prev.sections, section as Section] }
          : prev
      );
      setSectionTitle('');
      setSuccess('Đã thêm phần mới.');
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setAddingSection(false);
    }
  }

  async function handleAddItem(sectionId: string) {
    if (!assessment) return;
    clearMessages();
    try {
      const section = assessment.sections.find((s) => s.id === sectionId);
      const item = await createItem(sectionId, {
        question_version_id: DEMO_QUESTION_VERSION_ID,
        position: (section?.items.length ?? 0) + 1,
        points: '1.00',
      });
      setAssessment((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          sections: prev.sections.map((s) =>
            s.id === sectionId
              ? { ...s, items: [...s.items, item] }
              : s
          ),
        };
      });
      setSuccess('Đã thêm câu hỏi.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleAddTarget(e: React.FormEvent) {
    e.preventDefault();
    if (!assessmentId || !targetClassId) return;
    clearMessages();
    setAddingTarget(true);
    try {
      const target = await createTarget(assessmentId, {
        class_section_id: targetClassId,
      });
      setAssessment((prev) =>
        prev ? { ...prev, targets: [...prev.targets, target as Target] } : prev
      );
      setTargetClassId('');
      setSuccess('Đã thêm lớp đích.');
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setAddingTarget(false);
    }
  }

  async function handleValidate() {
    if (!assessmentId) return;
    clearMessages();
    setValidating(true);
    try {
      const result = await validateAssessment(assessmentId);
      setValidation(result);
      if (result.valid) {
        setSuccess('Đề thi hợp lệ, sẵn sàng xuất bản.');
      }
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setValidating(false);
    }
  }

  async function handlePublish() {
    if (!assessmentId) return;
    clearMessages();
    setPublishing(true);
    try {
      const result = await publishAssessment(assessmentId);
      setAssessment((prev) =>
        prev
          ? { ...prev, status: result.status, revision: result.revision }
          : prev
      );
      setSuccess(`Đã xuất bản đề thi (bản thảo ${result.revision}).`);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setPublishing(false);
    }
  }

  if (loading) {
    return (
      <div className="dashboard-page">
        <p className="dashboard-status">Đang tải đề thi…</p>
      </div>
    );
  }

  if (error && !assessment) {
    return (
      <div className="dashboard-page">
        <div className="error-banner" role="alert">
          {error}
        </div>
        <p>
          <Link to="/app/teacher">← Quay lại trang giáo viên</Link>
        </p>
      </div>
    );
  }

  if (!assessment) {
    return (
      <div className="dashboard-page">
        <p className="dashboard-status">Không tìm thấy đề thi.</p>
      </div>
    );
  }

  const isPublished = assessment.status === 'PUBLISHED' || assessment.status === 'OPEN';

  return (
    <div className="dashboard-page">
      <div className="builder-header">
        <h1>{assessment.title || 'Đề thi chưa đặt tên'}</h1>
        <div className="builder-meta">
          <span className={`status-badge status-${assessment.status.toLowerCase()}`}>
            {assessment.status}
          </span>
          <span>Bản thảo {assessment.revision}</span>
          <span>{assessment.duration_minutes} phút</span>
        </div>
      </div>

      <p>
        <Link to="/app/teacher">← Quay lại trang giáo viên</Link>
      </p>

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

      <section className="admin-section">
        <h2>Cài đặt đề thi</h2>
        <form onSubmit={handleSaveSettings} className="admin-form">
          <div className="field">
            <label htmlFor="builder-title">Tên đề thi</label>
            <input
              id="builder-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              disabled={isPublished}
            />
          </div>
          <div className="field">
            <label htmlFor="builder-duration">Thời gian (phút)</label>
            <input
              id="builder-duration"
              type="number"
              min={1}
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              required
              disabled={isPublished}
            />
          </div>
          <div className="field">
            <label htmlFor="builder-instructions">Hướng dẫn</label>
            <textarea
              id="builder-instructions"
              rows={3}
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
              disabled={isPublished}
            />
          </div>
          <div className="form-actions">
            <button type="submit" className="primary" disabled={savingSettings || isPublished}>
              {savingSettings ? 'Đang lưu…' : 'Lưu cài đặt'}
            </button>
          </div>
        </form>
      </section>

      <section className="admin-section">
        <h2>Phần và câu hỏi</h2>

        {assessment.sections.length === 0 && (
          <p className="dashboard-status">Chưa có phần nào.</p>
        )}

        {assessment.sections.map((section) => (
          <div key={section.id} className="builder-section">
            <div className="section-header">
              <h3>{section.title}</h3>
              <span className="section-meta">
                {section.items.length} câu hỏi
              </span>
            </div>

            {section.items.length === 0 && (
              <p className="dashboard-status">Chưa có câu hỏi.</p>
            )}

            <ul className="item-list">
              {section.items.map((item, index) => (
                <li key={item.id} className="item-row">
                  <span>Câu {index + 1}</span>
                  <span className="item-meta">
                    {item.points} điểm · {item.question_version_id}
                  </span>
                </li>
              ))}
            </ul>

            {!isPublished && (
              <button
                type="button"
                onClick={() => handleAddItem(section.id)}
                className="secondary"
              >
                + Thêm câu hỏi
              </button>
            )}
          </div>
        ))}

        {!isPublished && (
          <form onSubmit={handleAddSection} className="admin-form">
            <h3>Thêm phần mới</h3>
            <div className="field">
              <label htmlFor="section-title">Tên phần</label>
              <input
                id="section-title"
                type="text"
                value={sectionTitle}
                onChange={(e) => setSectionTitle(e.target.value)}
                placeholder="Ví dụ: Phần trắc nghiệm"
                required
              />
            </div>
            <div className="form-actions">
              <button type="submit" className="primary" disabled={addingSection}>
                {addingSection ? 'Đang thêm…' : 'Thêm phần'}
              </button>
            </div>
          </form>
        )}
      </section>

      <section className="admin-section">
        <h2>Lớp đích</h2>

        {assessment.targets.length === 0 && (
          <p className="dashboard-status">Chưa gán lớp nào.</p>
        )}

        {assessment.targets.length > 0 && (
          <ul className="target-list">
            {assessment.targets.map((target) => {
              const cls = classes.find((c) => c.id === target.class_section_id);
              return (
                <li key={target.id} className="target-row">
                  {cls?.name ?? target.class_section_id}
                </li>
              );
            })}
          </ul>
        )}

        {!isPublished && (
          <form onSubmit={handleAddTarget} className="inline-form">
            <select
              value={targetClassId}
              onChange={(e) => setTargetClassId(e.target.value)}
              required
            >
              <option value="">Chọn lớp…</option>
              {classes.map((cls) => (
                <option key={cls.id} value={cls.id}>
                  {cls.name}
                </option>
              ))}
            </select>
            <button type="submit" className="primary" disabled={addingTarget}>
              {addingTarget ? 'Đang thêm…' : 'Thêm lớp'}
            </button>
          </form>
        )}
      </section>

      <section className="admin-section">
        <h2>Kiểm tra và xuất bản</h2>

        <div className="form-actions">
          <button
            type="button"
            onClick={handleValidate}
            disabled={validating || isPublished}
          >
            {validating ? 'Đang kiểm tra…' : 'Kiểm tra'}
          </button>
          <button
            type="button"
            className="primary"
            onClick={handlePublish}
            disabled={publishing || isPublished}
          >
            {publishing
              ? 'Đang xuất bản…'
              : isPublished
              ? 'Đã xuất bản'
              : 'Xuất bản'}
          </button>
        </div>

        {validation && !validation.valid && (
          <div className="error-banner" role="alert">
            <strong>Đề thi chưa hợp lệ:</strong>
            <ul className="validation-errors">
              {validation.errors?.map((err, idx) => (
                <li key={idx}>
                  {err.field}: {err.message}
                </li>
              ))}
            </ul>
          </div>
        )}
      </section>
    </div>
  );
}

