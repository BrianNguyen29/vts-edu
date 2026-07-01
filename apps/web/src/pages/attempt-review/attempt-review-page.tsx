import { Link, useParams } from 'react-router-dom';
import { useAttemptResult } from '@/shared/api/attempts-queries';
import { type AttemptResultItem } from '@/shared/api/attempts';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

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
  const {
    data: result,
    isPending: loading,
    error,
  } = useAttemptResult(attemptId);

  useDocumentTitle('Kết quả bài làm');

  if (!attemptId) {
    return (
      <div className="dashboard-page">
        <div className="error-banner" role="alert">
          Thiếu mã bài làm.
        </div>
      </div>
    );
  }

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
        <ErrorState
          error={error}
          overrides={{
            403: 'Không có quyền truy cập kết quả này.',
            404: 'Không tìm thấy kết quả bài làm.',
            409: 'Bài làm chưa được nộp hoặc chưa chấm điểm.',
          }}
        />
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
