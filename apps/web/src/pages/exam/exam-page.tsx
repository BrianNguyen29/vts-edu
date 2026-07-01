import { useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react';
import { useParams } from 'react-router-dom';
import {
  ApiResponseError,
  getAttempt,
  saveAnswer,
  submitAttempt,
  type AttemptItem,
  type AttemptSnapshot,
  type AttemptSubmitted,
} from '@/shared/api/attempts';
import {
  createExamDraftStorage,
  shouldPreferDraft,
  type ExamDraft,
  type ExamDraftStorage,
} from '@/shared/lib/exam-draft-store';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

type SaveStatus =
  | { type: 'idle' }
  | { type: 'saving' }
  | { type: 'local' }
  | { type: 'syncing' }
  | { type: 'saved'; revision: number }
  | { type: 'error'; message: string };

function getChoiceFromPayload(payload: unknown): string | undefined {
  if (
    typeof payload === 'object' &&
    payload !== null &&
    'selected_option' in payload &&
    typeof (payload as { selected_option: unknown }).selected_option === 'string'
  ) {
    return (payload as { selected_option: string }).selected_option;
  }
  return undefined;
}

function formatTimeRemaining(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

function getPromptText(prompt: unknown): string {
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

interface NormalizedChoice {
  id: string;
  label: string;
}

function getChoices(choices: unknown): NormalizedChoice[] {
  if (!Array.isArray(choices)) return [];

  return choices
    .map((choice): NormalizedChoice | null => {
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
        if (id) {
          return { id, label: text || id };
        }
      }
      return null;
    })
    .filter((c): c is NormalizedChoice => c !== null);
}

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Không có quyền truy cập bài thi này.';
      case 404:
        return 'Không tìm thấy bài thi hoặc câu hỏi.';
      case 409:
        return err.body.error.code === 'attempt_expired'
          ? 'Bài thi đã hết thời gian.'
          : 'Bài thi không còn trong trạng thái làm bài.';
      default:
        return err.body.error.message || 'Không thể tải bài thi.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

async function syncPendingDrafts(
  attemptId: string,
  items: AttemptItem[],
  storage: ExamDraftStorage,
  setSaveStatuses: Dispatch<SetStateAction<Record<string, SaveStatus>>>,
  isCancelled: () => boolean
) {
  const pending = await storage.getPendingByAttempt(attemptId);
  for (const draft of pending) {
    if (isCancelled()) return;

    const item = items.find((i) => i.id === draft.item_id);
    if (!item) continue;

    setSaveStatuses((prev) => ({
      ...prev,
      [draft.item_id]: { type: 'syncing' },
    }));

    try {
      const saved = await saveAnswer(attemptId, draft.item_id, draft.payload);
      if (isCancelled()) return;

      await storage.setDraft({
        ...draft,
        pending: false,
        revision: saved.revision,
        updated_at: Date.now(),
      });

      setSaveStatuses((prev) => ({
        ...prev,
        [draft.item_id]: { type: 'saved', revision: saved.revision },
      }));
    } catch (err) {
      if (isCancelled()) return;
      setSaveStatuses((prev) => ({
        ...prev,
        [draft.item_id]: { type: 'error', message: formatFriendlyError(err) },
      }));
    }
  }
}

export function ExamPage() {
  const { attemptId } = useParams<{ attemptId: string }>();

  useDocumentTitle('Bài thi');

  const [snapshot, setSnapshot] = useState<AttemptSnapshot | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [saveStatuses, setSaveStatuses] = useState<Record<string, SaveStatus>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitResult, setSubmitResult] = useState<AttemptSubmitted | null>(null);
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [isOnline, setIsOnline] = useState<boolean>(() => {
    if (typeof navigator === 'undefined') return true;
    return navigator.onLine;
  });

  const storageRef = useRef<ExamDraftStorage | null>(null);
  const attemptIdRef = useRef(attemptId);
  const snapshotRef = useRef(snapshot);

  useEffect(() => {
    attemptIdRef.current = attemptId;
  }, [attemptId]);

  useEffect(() => {
    snapshotRef.current = snapshot;
  }, [snapshot]);

  useEffect(() => {
    function handleOnline() {
      setIsOnline(true);
      const id = attemptIdRef.current;
      const snap = snapshotRef.current;
      const storage = storageRef.current;
      if (id && snap?.status === 'IN_PROGRESS' && storage) {
        void syncPendingDrafts(id, snap.items, storage, setSaveStatuses, () => false);
      }
    }

    function handleOffline() {
      setIsOnline(false);
    }

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  useEffect(() => {
    if (!attemptId) {
      setError('Thiếu mã bài thi.');
      setLoading(false);
      return;
    }

    const id = attemptId;
    let cancelled = false;

    async function load() {
      try {
        const storage = await createExamDraftStorage();
        if (cancelled) return;
        storageRef.current = storage;

        const [data, drafts] = await Promise.all([
          getAttempt(id),
          storage.getPendingByAttempt(id),
        ]);
        if (cancelled) return;

        setSnapshot(data);

        const initialAnswers: Record<string, string> = {};
        const initialStatuses: Record<string, SaveStatus> = {};

        data.items.forEach((item) => {
          const choice = item.answer
            ? getChoiceFromPayload(item.answer.answer_payload)
            : undefined;
          if (choice) {
            initialAnswers[item.id] = choice;
          }
          if (item.answer) {
            initialStatuses[item.id] = {
              type: 'saved',
              revision: item.answer.revision,
            };
          }
        });

        drafts.forEach((draft) => {
          const item = data.items.find((i) => i.id === draft.item_id);
          if (!item) return;

          const serverRevision = item.answer?.revision;
          if (!shouldPreferDraft(draft, serverRevision)) return;

          const choice = getChoiceFromPayload(draft.payload);
          if (choice) {
            initialAnswers[item.id] = choice;
          }
          if (draft.pending) {
            initialStatuses[item.id] = { type: 'local' };
          } else if (typeof draft.revision === 'number') {
            initialStatuses[item.id] = {
              type: 'saved',
              revision: draft.revision,
            };
          }
        });

        setAnswers(initialAnswers);
        setSaveStatuses(initialStatuses);

        if (data.status !== 'IN_PROGRESS') {
          storage.deleteAllForAttempt(id).catch(() => {});
        } else {
          await syncPendingDrafts(id, data.items, storage, setSaveStatuses, () => cancelled);
        }
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
  }, [attemptId]);

  useEffect(() => {
    if (!snapshot?.expires_at) {
      setTimeLeft(null);
      return;
    }

    const expiresAt = new Date(snapshot.expires_at).getTime();

    function updateTime() {
      const remaining = Math.max(0, expiresAt - Date.now());
      setTimeLeft(remaining);
    }

    updateTime();
    const interval = setInterval(updateTime, 1000);
    return () => clearInterval(interval);
  }, [snapshot?.expires_at]);

  useEffect(() => {
    if (!attemptId || !storageRef.current || !snapshot) return;
    if (snapshot.status !== 'IN_PROGRESS') {
      storageRef.current.deleteAllForAttempt(attemptId).catch(() => {});
    }
  }, [attemptId, snapshot?.status]);

  const isExpired = timeLeft === 0;

  async function handleSelect(item: AttemptItem, choice: string) {
    if (!attemptId || snapshot?.status !== 'IN_PROGRESS' || isExpired) return;

    const currentAttemptId = attemptId;
    const storage = storageRef.current;

    setAnswers((prev) => ({ ...prev, [item.id]: choice }));
    setSaveStatuses((prev) => ({
      ...prev,
      [item.id]: { type: 'saving' },
    }));

    const draft: ExamDraft = {
      attempt_id: currentAttemptId,
      item_id: item.id,
      payload: { selected_option: choice },
      pending: true,
      updated_at: Date.now(),
    };

    try {
      await storage?.setDraft(draft);
      const saved = await saveAnswer(currentAttemptId, item.id, {
        selected_option: choice,
      });
      await storage?.setDraft({
        ...draft,
        pending: false,
        revision: saved.revision,
        updated_at: Date.now(),
      });
      setSaveStatuses((prev) => ({
        ...prev,
        [item.id]: { type: 'saved', revision: saved.revision },
      }));
    } catch (err) {
      // The local draft remains pending and will be retried when the
      // browser comes back online or the page reloads.
      setSaveStatuses((prev) => ({
        ...prev,
        [item.id]: isOnline
          ? { type: 'error', message: formatFriendlyError(err) }
          : { type: 'local' },
      }));
    }
  }

  async function handleSubmit() {
    if (!attemptId || snapshot?.status !== 'IN_PROGRESS' || isExpired) return;

    const confirmed = window.confirm(
      'Bạn có chắc chắn muốn nộp bài? Sau khi nộp không thể chỉnh sửa.'
    );
    if (!confirmed) return;

    const currentAttemptId = attemptId;

    setSubmitting(true);
    try {
      const result = await submitAttempt(currentAttemptId);
      await storageRef.current?.deleteAllForAttempt(currentAttemptId);
      setSubmitResult(result);
      setSnapshot((prev) =>
        prev
          ? { ...prev, status: 'SUBMITTED', submitted_at: result.submitted_at }
          : prev
      );
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <div className="exam-page">
        <p className="exam-status">Đang tải bài thi…</p>
      </div>
    );
  }

  if (error && !snapshot) {
    return (
      <div className="exam-page">
        <div className="error-banner" role="alert">
          {error}
        </div>
      </div>
    );
  }

  if (!snapshot) {
    return (
      <div className="exam-page">
        <p className="exam-status">Không có dữ liệu bài thi.</p>
      </div>
    );
  }

  if (snapshot.status === 'SUBMITTED' || submitResult) {
    const result = submitResult;
    const submittedAt = result?.submitted_at || snapshot.submitted_at || undefined;

    return (
      <div className="exam-page">
        <div className="exam-result" role="status">
          <h2>Đã nộp bài</h2>
          <p>
            Bài thi đã được nộp vào{' '}
            <time dateTime={submittedAt}>
              {submittedAt
                ? new Date(submittedAt).toLocaleString('vi-VN')
                : '—'}
            </time>
            .
          </p>
          {result && (
            <div className="exam-grading">
              <p>
                <strong>Trạng thái chấm điểm:</strong>{' '}
                {result.grading_status || '—'}
              </p>
              {result.score !== undefined && result.max_score !== undefined && (
                <p>
                  <strong>Điểm:</strong> {result.score} / {result.max_score}
                </p>
              )}
            </div>
          )}
        </div>
      </div>
    );
  }

  if (snapshot.status === 'EXPIRED') {
    return (
      <div className="exam-page">
        <div className="error-banner" role="alert">
          Bài thi đã hết thời gian và không thể tiếp tục làm.
        </div>
      </div>
    );
  }

  if (snapshot.status !== 'IN_PROGRESS') {
    return (
      <div className="exam-page">
        <div className="error-banner" role="alert">
          Bài thi không trong trạng thái làm bài.
        </div>
      </div>
    );
  }

  return (
    <div className="exam-page">
      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}

      {!isOnline && (
        <div className="exam-offline-banner" role="status" aria-live="polite">
          Mất kết nối. Câu trả lời vẫn được lưu cục bộ và sẽ đồng bộ khi có mạng.
        </div>
      )}

      <div className="exam-meta-bar">
        <span className={`exam-status-badge ${!isOnline ? 'offline' : ''}`}>
          {isExpired ? 'Đã hết thời gian' : isOnline ? 'Đang làm bài' : 'Ngoại tuyến'}
        </span>
        {timeLeft !== null && (
          <span
            className={`exam-timer ${isExpired ? 'expired' : ''}`}
            role="timer"
            aria-live="polite"
          >
            {isExpired ? '00:00' : `Còn lại: ${formatTimeRemaining(timeLeft)}`}
          </span>
        )}
        {snapshot.expires_at && timeLeft === null && (
          <span className="exam-expires">
            Hết hạn: {new Date(snapshot.expires_at).toLocaleString('vi-VN')}
          </span>
        )}
      </div>

      <form className="exam-form" onSubmit={(e) => e.preventDefault()}>
        {snapshot.items.map((item) => {
          const promptText = getPromptText(item.prompt);
          const choices = getChoices(item.choices);
          const hasContent = promptText.length > 0 && choices.length > 0;

          return (
            <fieldset key={item.id} className="exam-question" data-testid="exam-question">
              <legend>Câu {item.position}</legend>
              {hasContent ? (
                <>
                  <p className="exam-question-prompt">{promptText}</p>
                  <div className="exam-options">
                    {choices.map((choice) => (
                      <label key={choice.id} className="exam-option">
                        <input
                          type="radio"
                          name={`answer-${item.id}`}
                          value={choice.id}
                          checked={answers[item.id] === choice.id}
                          onChange={() => handleSelect(item, choice.id)}
                          disabled={submitting || isExpired}
                        />
                        <span className="exam-option-label">{choice.id}.</span>{' '}
                        {choice.label}
                      </label>
                    ))}
                  </div>
                </>
              ) : (
                <div className="exam-unsupported" role="alert">
                  <p>
                    Câu hỏi này chưa có nội dung đầy đủ (mã phiên bản:{' '}
                    {item.question_version_id}).
                  </p>
                </div>
              )}

              <SaveStatusMessage status={saveStatuses[item.id]} />
            </fieldset>
          );
        })}

        <div className="exam-actions">
          <button
            type="button"
            className="primary"
            onClick={handleSubmit}
            disabled={isExpired || submitting || snapshot.items.length === 0}
            aria-busy={submitting}
            data-testid="submit-exam-button"
          >
            {submitting ? 'Đang nộp bài…' : isExpired ? 'Đã hết thời gian' : 'Nộp bài'}
          </button>
        </div>
      </form>
    </div>
  );
}

function SaveStatusMessage({ status }: { status: SaveStatus | undefined }) {
  if (!status || status.type === 'idle') return null;

  if (status.type === 'saving') {
    return <p className="exam-save-status saving">Đang lưu…</p>;
  }

  if (status.type === 'local') {
    return <p className="exam-save-status local">Đã lưu cục bộ (chưa đồng bộ)</p>;
  }

  if (status.type === 'syncing') {
    return <p className="exam-save-status syncing">Đang đồng bộ…</p>;
  }

  if (status.type === 'saved') {
    return (
      <p className="exam-save-status saved">
        Đã lưu (phiên bản {status.revision})
      </p>
    );
  }

  return <p className="exam-save-status error">{status.message}</p>;
}
