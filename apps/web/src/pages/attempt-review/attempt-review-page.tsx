import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  ApiResponseError,
  getAttemptResult,
  type AttemptResult,
  type AttemptResultItem,
} from '@/shared/api/attempts';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Không có quyền truy cập kết quả này.';
      case 404:
        return 'Không tìm thấy kết quả bài làm.';
      case 409:
        return 'Bài làm chưa được nộp hoặc chưa chấm điểm.';
      default:
        return err.body.error.message || 'Không thể tải kết quả.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
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

function getSelectedOption(payload: unknown): string | undefined {
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

function getCorrectOption(payload: unknown): string | undefined {
  if (
    typeof payload === 'object' &&
    payload !== null &&
    'correct_option' in payload &&
    typeof (payload as { correct_option: unknown }).correct_option === 'string'
  ) {
    return (payload as { correct_option: string }).correct_option;
  }
  return undefined;
}

export function AttemptReviewPage() {
  const { attemptId } = useParams<{ attemptId: string }>();

  const [result, setResult] = useState<AttemptResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!attemptId) {
      setError('Thiếu mã bài làm.');
      setLoading(false);
      return;
    }

    let cancelled = false;

    async function load() {
      try {
        const data = await getAttemptResult(attemptId!);
        if (cancelled) return;
        setResult(data);
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

  if (loading) {
    return (
      <div className="dashboard-page">
        <p className="dashboard-status">Đang tải kết quả…</p>
      </div>
    );
  }

  if (error && !result) {
    return (
      <div className="dashboard-page">
        <div className="error-banner" role="alert">
          {error}
        </div>
        <p>
          <Link to="/app/student">← Quay lại trang làm việc</Link>
        </p>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="dashboard-page">
        <p className="dashboard-status">Không có dữ liệu kết quả.</p>
      </div>
    );
  }

  const scoreText =
    result.score !== undefined && result.max_score !== undefined
      ? `${result.score} / ${result.max_score}`
      : '—';

  return (
    <div className="dashboard-page">
      <p>
        <Link to="/app/student">← Quay lại trang làm việc</Link>
      </p>

      <div className="review-header">
        <h1>Kết quả bài làm</h1>
        <div className="review-score" role="status">
          <span className="review-score-value">{scoreText}</span>
          <span className="review-score-label">Điểm</span>
        </div>
      </div>

      <div className="review-meta">
        <p>
          <strong>Trạng thái:</strong>{' '}
          {result.status === 'SUBMITTED' ? 'Đã nộp' : 'Hết hạn'}
        </p>
        {result.submitted_at && (
          <p>
            <strong>Nộp lúc:</strong>{' '}
            {new Date(result.submitted_at).toLocaleString('vi-VN')}
          </p>
        )}
        {result.grading_status && (
          <p>
            <strong>Chấm điểm:</strong> {result.grading_status}
          </p>
        )}
      </div>

      <h2>Chi tiết các câu</h2>

      {result.items.length === 0 ? (
        <p className="dashboard-status">Không có câu hỏi nào.</p>
      ) : (
        <ol className="review-item-list">
          {result.items.map((item) => (
            <ReviewItemRow key={item.id} item={item} />
          ))}
        </ol>
      )}
    </div>
  );
}

function ReviewItemRow({ item }: { item: AttemptResultItem }) {
  const promptText = getPromptText(item.prompt);
  const choices = getChoices(item.correct_answer);
  const studentChoice = getSelectedOption(item.student_answer?.answer_payload);
  const correctChoice = getCorrectOption(item.correct_answer);

  return (
    <li
      className={`review-item ${item.is_correct ? 'correct' : 'incorrect'}`}
      role="listitem"
    >
      <div className="review-item-header">
        <span className="review-item-number">Câu {item.position}</span>
        <span
          className={`review-item-badge ${item.is_correct ? 'correct' : 'incorrect'}`}
        >
          {item.is_correct ? 'Đúng' : 'Sai'}
        </span>
        <span className="review-item-points">{item.points} điểm</span>
      </div>

      <p className="review-prompt">
        {promptText || 'Câu hỏi chưa có nội dung'}
      </p>

      {choices.length > 0 && (
        <ul className="review-options">
          {choices.map((choice) => {
            const isSelected = studentChoice === choice.id;
            const isCorrect = correctChoice === choice.id;
            return (
              <li
                key={choice.id}
                className={`review-option ${isSelected ? 'selected' : ''} ${isCorrect ? 'correct' : ''}`}
              >
                <span className="review-option-label">{choice.id}.</span>{' '}
                {choice.label}
                {isSelected && <span className="review-option-note"> (bạn chọn)</span>}
                {isCorrect && <span className="review-option-note"> (đáp án đúng)</span>}
              </li>
            );
          })}
        </ul>
      )}

      {choices.length === 0 && (
        <p className="review-answer">
          <strong>Đáp án đúng:</strong>{' '}
          {correctChoice || '—'}
        </p>
      )}

      {studentChoice && choices.length === 0 && (
        <p className="review-answer">
          <strong>Bạn chọn:</strong> {studentChoice}
        </p>
      )}
    </li>
  );
}
