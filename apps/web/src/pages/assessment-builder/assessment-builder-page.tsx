import { useEffect, useRef, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ApiResponseError } from '@/shared/api/attempts';
import {
  createItem,
  createSection,
  createTarget,
  deleteItem,
  deleteSection,
  deleteTarget,
  duplicateItem,
  duplicateSection,
  getAssessment,
  listPublications,
  listQuestions,
  previewAssessment,
  publishAssessment,
  reorderItems,
  reorderSections,
  updateAssessment,
  updateItem,
  updateSection,
  validateAssessment,
  type AssessmentDetail,
  type AssessmentPreview,
  type Item,
  type PreviewItem,
  type PublicationSummary,
  type QuestionPickerItem,
  type Section,
  type Target,
  type ValidationResult,
} from '@/shared/api/assessments';
import { listClasses, type ClassSection } from '@/shared/api/academics';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

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

  useDocumentTitle('Trình soạn đề thi');

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
  const [autosaveState, setAutosaveState] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle');
  const [autosaveMessage, setAutosaveMessage] = useState<string | null>(null);

  const [previewOpen, setPreviewOpen] = useState(false);
  const [preview, setPreview] = useState<AssessmentPreview | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);

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

  const previewCloseButtonRef = useRef<HTMLButtonElement | null>(null);
  const lastFocusedBeforePreview = useRef<HTMLElement | null>(null);

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

  useEffect(() => {
    if (!assessmentId || !assessment) return;
    if (assessment.status === 'PUBLISHED' || assessment.status === 'OPEN') return;

    const timeout = setTimeout(() => {
      let cancelled = false;

      async function save() {
        const durationNum = Number(duration);
        if (!title.trim() || Number.isNaN(durationNum) || durationNum < 1) {
          return;
        }
        setAutosaveState('saving');
        setAutosaveMessage(null);
        try {
          const updated = await updateAssessment(assessmentId!, {
            title: title.trim(),
            duration_minutes: durationNum,
            instructions: instructions.trim() || null,
          });
          if (cancelled) return;
          setAssessment(updated);
          setAutosaveState('saved');
          setAutosaveMessage('Đã tự động lưu');
          setTimeout(() => {
            if (!cancelled) {
              setAutosaveState('idle');
              setAutosaveMessage(null);
            }
          }, 2000);
        } catch (err) {
          if (cancelled) return;
          setAutosaveState('error');
          setAutosaveMessage(formatFriendlyError(err));
        }
      }

      void save();

      return () => {
        cancelled = true;
      };
    }, 800);

    return () => clearTimeout(timeout);
  }, [title, duration, instructions, assessmentId, assessment?.status]);

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

  async function handleDuplicateSection(sectionId: string) {
    if (!assessmentId) return;
    clearMessages();
    try {
      await duplicateSection(assessmentId, sectionId);
      await refreshAssessment();
      setSuccess('Đã nhân bản phần.');
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

  async function handleDuplicateItem(itemId: string, sectionId: string) {
    clearMessages();
    try {
      const duplicated = await duplicateItem(sectionId, itemId);
      setAssessment((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          sections: prev.sections.map((s) => {
            if (s.id !== sectionId) return s;
            const index = s.items.findIndex((it) => it.id === itemId);
            const newItems = [...s.items];
            newItems.splice(index + 1, 0, duplicated);
            return { ...s, items: newItems };
          }),
        };
      });
      setSuccess('Đã nhân bản câu hỏi.');
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

  async function handleOpenPreview() {
    if (!assessmentId) return;
    if (typeof document !== 'undefined' && typeof document.activeElement === 'object') {
      lastFocusedBeforePreview.current = document.activeElement as HTMLElement | null;
    }
    setPreviewOpen(true);
    setPreviewLoading(true);
    setPreview(null);
    try {
      const data = await previewAssessment(assessmentId);
      setPreview(data);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setPreviewLoading(false);
    }
  }

  function handleClosePreview() {
    setPreviewOpen(false);
    setPreview(null);
    // Return focus to the element that opened the dialog.
    const previous = lastFocusedBeforePreview.current;
    if (previous && typeof previous.focus === 'function') {
      previous.focus();
    }
  }

  // When the preview dialog opens, move focus to its close button.
  useEffect(() => {
    if (previewOpen) {
      previewCloseButtonRef.current?.focus();
    }
  }, [previewOpen]);

  function getPreviewPromptText(prompt: unknown): string {
    if (typeof prompt === 'string' && prompt.trim().length > 0) {
      return prompt;
    }
    if (
      typeof prompt === 'object' &&
      prompt !== null &&
      'text' in prompt &&
      typeof (prompt as { text: unknown }).text === 'string'
    ) {
      const text = (prompt as { text: string }).text;
      if (text.trim().length > 0) return text;
    }
    return '';
  }

  function getPreviewChoices(choices: unknown): { id: string; label: string }[] {
    if (!Array.isArray(choices)) return [];
    return choices
      .map((choice): { id: string; label: string } | null => {
        if (typeof choice === 'string') {
          return { id: choice, label: choice };
        }
        if (typeof choice === 'object' && choice !== null) {
          const id =
            'id' in choice && typeof choice.id === 'string' ? choice.id : '';
          const text =
            'text' in choice && typeof choice.text === 'string'
              ? choice.text
              : '';
          if (id) return { id, label: text || id };
        }
        return null;
      })
      .filter((c): c is { id: string; label: string } => c !== null);
  }

  const filteredQuestions = questions.filter((q) =>
    q.prompt.toLowerCase().includes(questionSearch.toLowerCase())
  );

  function getQuestionLabel(versionId: string): string {
    const q = questions.find((x) => x.question_version_id === versionId);
    return q?.prompt ?? versionId;
  }

  function questionTypeLabel(t: string | undefined): string {
    switch (t) {
      case 'short_answer':
        return 'TLN';
      case 'essay':
        return 'TL';
      case 'multiple_choice':
      default:
        return 'TN';
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
        <div className="error-banner" role="alert" data-testid="builder-error">
          {error}
        </div>
      )}
      {success && (
        <div className="success-banner" role="status" data-testid="builder-success">
          {success}
        </div>
      )}

      <section className="admin-section">
        <div className="section-header">
          <h2>Cài đặt đề thi</h2>
          <span className={`autosave-indicator autosave-${autosaveState}`}>
            {autosaveState === 'saving'
              ? 'Đang lưu…'
              : autosaveState === 'saved'
                ? 'Đã tự động lưu'
                : autosaveState === 'error'
                  ? autosaveMessage || 'Lỗi tự động lưu'
                  : ''}
          </span>
        </div>
        <form className="admin-form" onSubmit={(e) => e.preventDefault()}>
          <div className="field">
            <label htmlFor="builder-title">Tên đề thi</label>
            <input
              id="builder-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              disabled={isPublished}
              data-testid="builder-title"
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
              data-testid="builder-duration"
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
        </form>
      </section>

      <section className="admin-section">
        <h2>Phần và câu hỏi</h2>

        {assessment.sections.length === 0 && (
          <p className="dashboard-status">Chưa có phần nào.</p>
        )}

          {assessment.sections.map((section, sectionIndex) => (
          <div key={section.id} className="builder-section" data-testid="builder-section">
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
                  onClick={() => handleDuplicateSection(section.id)}
                >
                  Nhân bản
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
                              handleDuplicateItem(item.id, section.id)
                            }
                          >
                            Nhân bản
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
                    data-testid="picker-question-select"
                  >
                    <option value="">Chọn…</option>
                    {filteredQuestions.map((q) => (
                      <option key={q.question_version_id} value={q.question_version_id}>
                        {`[${questionTypeLabel(q.question_type)}] ${q.prompt}`}
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
                    data-testid="picker-add-button"
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
                data-testid="add-question-button"
              >
                + Thêm câu hỏi
              </button>
            )}
          </div>
        ))}

        {!isPublished && (
          <form onSubmit={handleAddSection} className="admin-form" data-testid="add-section-form">
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
                data-testid="section-title-input"
              />
            </div>
            <div className="form-actions">
              <button
                type="submit"
                className="primary"
                disabled={addingSection}
                data-testid="add-section-button"
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
              data-testid="target-class-select"
            >
              <option value="">Chọn lớp…</option>
              {classes.map((cls) => (
                <option key={cls.id} value={cls.id}>
                  {cls.name}
                </option>
              ))}
            </select>
            <button type="submit" className="primary" disabled={addingTarget} data-testid="add-target-button">
              {addingTarget ? 'Đang thêm…' : 'Thêm lớp'}
            </button>
          </form>
        )}
      </section>

      <section className="admin-section">
        <h2>Kiểm tra và xuất bản</h2>

        <div className="form-actions">
          <button type="button" onClick={handleOpenPreview} data-testid="preview-button">
            Xem trước
          </button>
          <button
            type="button"
            onClick={handleValidate}
            disabled={validating || isPublished}
            data-testid="validate-button"
          >
            {validating ? 'Đang kiểm tra…' : 'Kiểm tra'}
          </button>
          <button
            type="button"
            className="primary"
            onClick={handlePublish}
            disabled={publishing || isPublished}
            data-testid="publish-button"
          >
            {publishing
              ? 'Đang xuất bản…'
              : isPublished
              ? 'Đã xuất bản'
              : 'Xuất bản'}
          </button>
        </div>

        {validation && !validation.valid && (
          <div className="error-banner" role="alert" data-testid="validation-errors">
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
            <table className="publication-table" data-testid="publication-table">
              <caption className="visually-hidden">
                Lịch sử xuất bản của đề thi
              </caption>
              <thead>
                <tr>
                  <th scope="col">Phiên bản</th>
                  <th scope="col">Trạng thái</th>
                  <th scope="col">Thời gian</th>
                </tr>
              </thead>
              <tbody>
                {publications.map((pub) => (
                  <tr key={pub.id}>
                    <td>{pub.version}</td>
                    <td>{pub.status}</td>
                    <td>
                      <time dateTime={pub.published_at}>
                        {new Date(pub.published_at).toLocaleString('vi-VN')}
                      </time>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {previewOpen && (
        <div className="preview-overlay" role="dialog" aria-modal="true" aria-labelledby="preview-dialog-title">
          <div className="preview-panel">
            <div className="preview-header">
              <h2 id="preview-dialog-title">Xem trước đề thi</h2>
              <button
                type="button"
                onClick={handleClosePreview}
                ref={previewCloseButtonRef}
                aria-label="Đóng bản xem trước"
              >
                Đóng
              </button>
            </div>
            {previewLoading ? (
              <p className="dashboard-status">Đang tải bản xem trước…</p>
            ) : !preview ? (
              <p className="dashboard-status">Không thể tải bản xem trước.</p>
            ) : (
              <div className="preview-content">
                <div className="preview-meta">
                  <h3>{preview.title}</h3>
                  <p>Thời gian: {preview.duration_minutes} phút</p>
                  {preview.instructions && (
                    <p className="preview-instructions">{preview.instructions}</p>
                  )}
                </div>
                {preview.sections.map((section, sectionIndex) => (
                  <section key={section.id} className="preview-section">
                    <h4>
                      Phần {sectionIndex + 1}: {section.title}
                    </h4>
                    <ol className="preview-item-list">
                      {section.items.map((item, itemIndex) => (
                        <PreviewQuestion
                          key={item.id}
                          item={item}
                          number={itemIndex + 1}
                          getPromptText={getPreviewPromptText}
                          getChoices={getPreviewChoices}
                        />
                      ))}
                    </ol>
                  </section>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function PreviewQuestion({
  item,
  number,
  getPromptText,
  getChoices,
}: {
  item: PreviewItem;
  number: number;
  getPromptText: (prompt: unknown) => string;
  getChoices: (choices: unknown) => { id: string; label: string }[];
}) {
  const promptText = getPromptText(item.prompt);
  const choices = getChoices(item.choices);

  return (
    <li className="preview-item">
      <p className="preview-prompt">
        <strong>Câu {number}:</strong> {promptText || 'Câu hỏi chưa có nội dung'}
      </p>
      {choices.length > 0 && (
        <ul className="preview-options">
          {choices.map((choice) => (
            <li key={choice.id}>
              <label className="preview-option">
                <input type="radio" name={`preview-${item.id}`} value={choice.id} disabled />
                <span>{choice.id}.</span> {choice.label}
              </label>
            </li>
          ))}
        </ul>
      )}
      <p className="preview-points">{item.points} điểm</p>
    </li>
  );
}
