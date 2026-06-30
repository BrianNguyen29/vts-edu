import { useEffect, useState } from 'react';
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

type SaveStatus =
  | { type: 'idle' }
  | { type: 'saving' }
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

export function ExamPage() {
  const { attemptId } = useParams<{ attemptId: string }>();

  const [snapshot, setSnapshot] = useState<AttemptSnapshot | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [saveStatuses, setSaveStatuses] = useState<Record<string, SaveStatus>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitResult, setSubmitResult] = useState<AttemptSubmitted | null>(null);
  const [timeLeft, setTimeLeft] = useState<number | null>(null);

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
        const data = await getAttempt(id);
        if (cancelled) return;

        setSnapshot(data);

        const initialAnswers: Record<string, string> = {};
        data.items.forEach((item) => {
          const choice = item.answer
            ? getChoiceFromPayload(item.answer.answer_payload)
            : undefined;
          if (choice) {
            initialAnswers[item.id] = choice;
          }
          if (item.answer) {
            setSaveStatuses((prev) => ({
              ...prev,
              [item.id]: { type: 'saved', revision: item.answer!.revision },
            }));
          }
        });
        setAnswers(initialAnswers);
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

  const isExpired = timeLeft === 0;

  async function handleSelect(item: AttemptItem, choice: string) {
    if (!attemptId || snapshot?.status !== 'IN_PROGRESS' || isExpired) return;

    const currentAttemptId = attemptId;

    setAnswers((prev) => ({ ...prev, [item.id]: choice }));
    setSaveStatuses((prev) => ({
      ...prev,
      [item.id]: { type: 'saving' },
    }));

    try {
      const saved = await saveAnswer(currentAttemptId, item.id, {
        selected_option: choice,
      });
      setSaveStatuses((prev) => ({
        ...prev,
        [item.id]: { type: 'saved', revision: saved.revision },
      }));
    } catch (err) {
      setSaveStatuses((prev) => ({
        ...prev,
        [item.id]: { type: 'error', message: formatFriendlyError(err) },
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

      <div className="exam-meta-bar">
        <span className="exam-status-badge">
          {isExpired ? 'Đã hết thời gian' : 'Đang làm bài'}
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
            <fieldset key={item.id} className="exam-question">
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

  if (status.type === 'saved') {
    return (
      <p className="exam-save-status saved">
        Đã lưu (phiên bản {status.revision})
      </p>
    );
  }

  return <p className="exam-save-status error">{status.message}</p>;
}
