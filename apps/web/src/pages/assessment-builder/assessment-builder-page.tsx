import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ApiResponseError } from '@/shared/api/attempts';
import {
  createItem,
  createSection,
  createTarget,
  deleteItem,
  deleteSection,
  deleteTarget,
  getAssessment,
  listPublications,
  listQuestions,
  publishAssessment,
  reorderItems,
  reorderSections,
  updateAssessment,
  updateItem,
  updateSection,
  validateAssessment,
  type AssessmentDetail,
  type Item,
  type PublicationSummary,
  type QuestionPickerItem,
  type Section,
  type Target,
  type ValidationResult,
} from '@/shared/api/assessments';
import { listClasses, type ClassSection } from '@/shared/api/academics';

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
  const [questions, setQuestions] = useState<QuestionPickerItem[]>([]);
  const [publications, setPublications] = useState<PublicationSummary[]>([]);
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

  const [editingSectionId, setEditingSectionId] = useState<string | null>(null);
  const [editingSectionTitle, setEditingSectionTitle] = useState('');

  const [editingItemId, setEditingItemId] = useState<string | null>(null);
  const [editingItemPoints, setEditingItemPoints] = useState('');
  const [editingItemQuestionId, setEditingItemQuestionId] = useState('');

  const [pickerSectionId, setPickerSectionId] = useState<string | null>(null);
  const [pickerQuestionId, setPickerQuestionId] = useState('');
  const [pickerPoints, setPickerPoints] = useState('1.00');
  const [questionSearch, setQuestionSearch] = useState('');

  useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!assessmentId) return;
      setLoading(true);
      setError(null);
      try {
        const [assessmentData, classesData, questionsData, pubsData] =
          await Promise.all([
            getAssessment(assessmentId),
            listClasses(),
            listQuestions({ limit: 100 }),
            listPublications(assessmentId),
          ]);
        if (cancelled) return;
        setAssessment(assessmentData);
        setClasses(classesData.data);
        setQuestions(questionsData.data);
        setPublications(pubsData);
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

  async function refreshAssessment() {
    if (!assessmentId) return;
    try {
      const data = await getAssessment(assessmentId);
      setAssessment(data);
    } catch (err) {
      setError(formatFriendlyError(err));
    }
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

  async function handleUpdateSection(sectionId: string) {
    if (!editingSectionTitle.trim()) return;
    clearMessages();
    try {
      const updated = await updateSection(sectionId, {
        title: editingSectionTitle.trim(),
      });
      setAssessment((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          sections: prev.sections.map((s) =>
            s.id === sectionId ? (updated as Section) : s
          ),
        };
      });
      setEditingSectionId(null);
      setSuccess('Đã cập nhật phần.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleDeleteSection(sectionId: string) {
    if (!window.confirm('Bạn có chắc muốn xóa phần này?')) return;
    clearMessages();
    try {
      await deleteSection(sectionId);
      await refreshAssessment();
      setSuccess('Đã xóa phần.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleMoveSection(sectionId: string, direction: -1 | 1) {
    if (!assessment) return;
    const index = assessment.sections.findIndex((s) => s.id === sectionId);
    const newIndex = index + direction;
    if (index < 0 || newIndex < 0 || newIndex >= assessment.sections.length) {
      return;
    }
    const newOrder = [...assessment.sections];
    const [moved] = newOrder.splice(index, 1);
    newOrder.splice(newIndex, 0, moved);
    clearMessages();
    try {
      await reorderSections(
        assessment.id,
        { section_ids: newOrder.map((s) => s.id) }
      );
      await refreshAssessment();
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleAddItem(sectionId: string) {
    if (!pickerQuestionId) return;
    clearMessages();
    try {
      const section = assessment?.sections.find((s) => s.id === sectionId);
      const item = await createItem(sectionId, {
        question_version_id: pickerQuestionId,
        position: (section?.items.length ?? 0) + 1,
        points: pickerPoints || '1.00',
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
      setPickerSectionId(null);
      setPickerQuestionId('');
      setPickerPoints('1.00');
      setQuestionSearch('');
      setSuccess('Đã thêm câu hỏi.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleUpdateItem(itemId: string, sectionId: string) {
    clearMessages();
    try {
      const updated = await updateItem(itemId, {
        question_version_id: editingItemQuestionId || undefined,
        points: editingItemPoints,
      });
      setAssessment((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          sections: prev.sections.map((s) =>
            s.id === sectionId
              ? {
                  ...s,
                  items: s.items.map((it) =>
                    it.id === itemId ? (updated as Item) : it
                  ),
                }
              : s
          ),
        };
      });
      setEditingItemId(null);
      setSuccess('Đã cập nhật câu hỏi.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleDeleteItem(itemId: string, sectionId: string) {
    if (!window.confirm('Bạn có chắc muốn xóa câu hỏi này?')) return;
    clearMessages();
    try {
      await deleteItem(itemId);
      setAssessment((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          sections: prev.sections.map((s) =>
            s.id === sectionId
              ? { ...s, items: s.items.filter((it) => it.id !== itemId) }
              : s
          ),
        };
      });
      setSuccess('Đã xóa câu hỏi.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleMoveItem(
    itemId: string,
    sectionId: string,
    direction: -1 | 1
  ) {
    const section = assessment?.sections.find((s) => s.id === sectionId);
    if (!section) return;
    const index = section.items.findIndex((it) => it.id === itemId);
    const newIndex = index + direction;
    if (index < 0 || newIndex < 0 || newIndex >= section.items.length) return;
    const newOrder = [...section.items];
    const [moved] = newOrder.splice(index, 1);
    newOrder.splice(newIndex, 0, moved);
    clearMessages();
    try {
      await reorderItems(sectionId, { item_ids: newOrder.map((it) => it.id) });
      await refreshAssessment();
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

  async function handleDeleteTarget(targetId: string) {
    if (!assessmentId) return;
    if (!window.confirm('Bạn có chắc muốn gỡ lớp đích này?')) return;
    clearMessages();
    try {
      await deleteTarget(assessmentId, targetId);
      setAssessment((prev) =>
        prev
          ? { ...prev, targets: prev.targets.filter((t) => t.id !== targetId) }
          : prev
      );
      setSuccess('Đã gỡ lớp đích.');
    } catch (err) {
      setError(formatFriendlyError(err));
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
      await refreshAssessment();
      const pubs = await listPublications(assessmentId);
      setPublications(pubs);
      setSuccess(`Đã xuất bản đề thi (bản thảo ${result.revision}).`);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setPublishing(false);
    }
  }

  const filteredQuestions = questions.filter((q) =>
    q.prompt.toLowerCase().includes(questionSearch.toLowerCase())
  );

  function getQuestionLabel(versionId: string): string {
    const q = questions.find((x) => x.question_version_id === versionId);
    return q?.prompt ?? versionId;
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

  const isPublished =
    assessment.status === 'PUBLISHED' || assessment.status === 'OPEN';

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
            <button
              type="submit"
              className="primary"
              disabled={savingSettings || isPublished}
            >
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

        {assessment.sections.map((section, sectionIndex) => (
          <div key={section.id} className="builder-section">
            <div className="section-header">
              {editingSectionId === section.id ? (
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    void handleUpdateSection(section.id);
                  }}
                  className="inline-form"
                >
                  <input
                    type="text"
                    value={editingSectionTitle}
                    onChange={(e) => setEditingSectionTitle(e.target.value)}
                    autoFocus
                  />
                  <button type="submit" className="primary">
                    Lưu
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditingSectionId(null)}
                  >
                    Hủy
                  </button>
                </form>
              ) : (
                <h3>{section.title}</h3>
              )}
              <span className="section-meta">
                {section.items.length} câu hỏi
              </span>
            </div>

            {!isPublished && editingSectionId !== section.id && (
              <div className="row-actions" style={{ marginBottom: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => {
                    setEditingSectionId(section.id);
                    setEditingSectionTitle(section.title);
                  }}
                >
                  Sửa tên
                </button>
                <button
                  type="button"
                  onClick={() => handleMoveSection(section.id, -1)}
                  disabled={sectionIndex === 0}
                >
                  Lên
                </button>
                <button
                  type="button"
                  onClick={() => handleMoveSection(section.id, 1)}
                  disabled={sectionIndex === assessment.sections.length - 1}
                >
                  Xuống
                </button>
                <button
                  type="button"
                  onClick={() => handleDeleteSection(section.id)}
                >
                  Xóa
                </button>
              </div>
            )}

            {section.items.length === 0 && (
              <p className="dashboard-status">Chưa có câu hỏi.</p>
            )}

            <ul className="item-list">
              {section.items.map((item, itemIndex) => (
                <li key={item.id} className="item-row">
                  {editingItemId === item.id ? (
                    <form
                      onSubmit={(e) => {
                        e.preventDefault();
                        void handleUpdateItem(item.id, section.id);
                      }}
                      className="inline-form"
                      style={{ flex: 1 }}
                    >
                      <label htmlFor={`edit-item-question-${item.id}`} className="sr-only">
                        Câu hỏi
                      </label>
                      <select
                        id={`edit-item-question-${item.id}`}
                        value={editingItemQuestionId}
                        onChange={(e) =>
                          setEditingItemQuestionId(e.target.value)
                        }
                        required
                      >
                        <option value="">Chọn câu hỏi…</option>
                        {questions.map((q) => (
                          <option
                            key={q.question_version_id}
                            value={q.question_version_id}
                          >
                            {q.prompt}
                          </option>
                        ))}
                      </select>
                      <input
                        type="text"
                        value={editingItemPoints}
                        onChange={(e) => setEditingItemPoints(e.target.value)}
                        placeholder="Điểm"
                        style={{ width: '6rem' }}
                        required
                      />
                      <button type="submit" className="primary">
                        Lưu
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingItemId(null)}
                      >
                        Hủy
                      </button>
                    </form>
                  ) : (
                    <>
                      <div>
                        <div className="item-title">
                          Câu {itemIndex + 1}: {getQuestionLabel(item.question_version_id)}
                        </div>
                        <div className="item-meta">
                          {item.points} điểm · {item.question_version_id}
                        </div>
                      </div>
                      {!isPublished && (
                        <div className="row-actions">
                          <button
                            type="button"
                            onClick={() => {
                              setEditingItemId(item.id);
                              setEditingItemQuestionId(item.question_version_id);
                              setEditingItemPoints(item.points);
                            }}
                          >
                            Sửa
                          </button>
                          <button
                            type="button"
                            onClick={() =>
                              handleMoveItem(item.id, section.id, -1)
                            }
                            disabled={itemIndex === 0}
                          >
                            Lên
                          </button>
                          <button
                            type="button"
                            onClick={() =>
                              handleMoveItem(item.id, section.id, 1)
                            }
                            disabled={itemIndex === section.items.length - 1}
                          >
                            Xuống
                          </button>
                          <button
                            type="button"
                            onClick={() =>
                              handleDeleteItem(item.id, section.id)
                            }
                          >
                            Xóa
                          </button>
                        </div>
                      )}
                    </>
                  )}
                </li>
              ))}
            </ul>

            {!isPublished && pickerSectionId === section.id && (
              <div className="item-picker">
                <div className="field">
                  <label>Tìm câu hỏi</label>
                  <input
                    type="search"
                    placeholder="Nhập từ khóa…"
                    value={questionSearch}
                    onChange={(e) => setQuestionSearch(e.target.value)}
                  />
                </div>
                <div className="field">
                  <label htmlFor="picker-question">Chọn câu hỏi</label>
                  <select
                    id="picker-question"
                    value={pickerQuestionId}
                    onChange={(e) => setPickerQuestionId(e.target.value)}
                    required
                    size={Math.min(5, filteredQuestions.length || 1)}
                  >
                    <option value="">Chọn…</option>
                    {filteredQuestions.map((q) => (
                      <option key={q.question_version_id} value={q.question_version_id}>
                        {q.prompt}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="field">
                  <label>Điểm</label>
                  <input
                    type="text"
                    value={pickerPoints}
                    onChange={(e) => setPickerPoints(e.target.value)}
                    style={{ width: '6rem' }}
                  />
                </div>
                <div className="form-actions">
                  <button
                    type="button"
                    className="primary"
                    onClick={() => handleAddItem(section.id)}
                    disabled={!pickerQuestionId}
                  >
                    Thêm câu hỏi
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setPickerSectionId(null);
                      setPickerQuestionId('');
                      setQuestionSearch('');
                    }}
                  >
                    Hủy
                  </button>
                </div>
              </div>
            )}

            {!isPublished && pickerSectionId !== section.id && (
              <button
                type="button"
                onClick={() => {
                  setPickerSectionId(section.id);
                  setPickerQuestionId('');
                  setPickerPoints('1.00');
                  setQuestionSearch('');
                }}
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
              <button
                type="submit"
                className="primary"
                disabled={addingSection}
              >
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
                  <span>{cls?.name ?? target.class_section_id}</span>
                  {!isPublished && (
                    <button
                      type="button"
                      onClick={() => handleDeleteTarget(target.id)}
                    >
                      Gỡ
                    </button>
                  )}
                </li>
              );
            })}
          </ul>
        )}

        {!isPublished && (
          <form onSubmit={handleAddTarget} className="inline-form">
            <label htmlFor="target-class" className="sr-only">
              Lớp đích
            </label>
            <select
              id="target-class"
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

      {publications.length > 0 && (
        <section className="admin-section">
          <h2>Lịch sử xuất bản</h2>
          <div className="publication-table-wrapper">
            <table className="publication-table">
              <thead>
                <tr>
                  <th>Phiên bản</th>
                  <th>Trạng thái</th>
                  <th>Thời gian</th>
                </tr>
              </thead>
              <tbody>
                {publications.map((pub) => (
                  <tr key={pub.id}>
                    <td>{pub.version}</td>
                    <td>{pub.status}</td>
                    <td>{new Date(pub.published_at).toLocaleString('vi-VN')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  );
}
