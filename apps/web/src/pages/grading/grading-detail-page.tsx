import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  useAttemptForReview,
  useGradeAttemptItem,
} from '@/shared/api/grading-queries';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function getPromptText(prompt: unknown): string {
  if (typeof prompt === 'string' && prompt.trim().length > 0) return prompt;
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

function getTextAnswer(payload: unknown): string {
  if (!payload || typeof payload !== 'object') return '';
  const p = payload as { text?: unknown };
  return typeof p.text === 'string' ? p.text : '';
}

function getAcceptedAnswers(correctAnswer: unknown): string[] {
  if (!correctAnswer || typeof correctAnswer !== 'object') return [];
  const c = correctAnswer as { accepted_answers?: unknown };
  if (!Array.isArray(c.accepted_answers)) return [];
  return c.accepted_answers.filter((v): v is string => typeof v === 'string');
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

interface GradeFormState {
  awarded: string;
  feedback: string;
}

const blankForm: GradeFormState = { awarded: '', feedback: '' };

export function GradingDetailPage() {
  useDocumentTitle('Chấm bài làm');
  const { attemptId = '' } = useParams<{ attemptId: string }>();
  const {
    data: review,
    isPending,
    error,
  } = useAttemptForReview(attemptId);
  const gradeMutation = useGradeAttemptItem(review?.assessment_id);

  const [forms, setForms] = useState<Record<string, GradeFormState>>({});

  useEffect(() => {
    if (!review) return;
    setForms((prev) => {
      const next: Record<string, GradeFormState> = {};
      for (const it of review.items) {
        if (it.question_type !== 'essay' && it.question_type !== 'short_answer') {
          continue;
        }
        const prevForm = prev[it.id];
        next[it.id] = prevForm ?? {
          awarded: it.item_grade?.awarded_score ?? '',
          feedback: it.item_grade?.feedback ?? '',
        };
      }
      return next;
    });
  }, [review]);

  if (isPending) {
    return <p className="loading">Đang tải bài làm…</p>;
  }
  if (error) {
    return (
      <ErrorState
        error={error}
        title="Không tải được bài làm"
      />
    );
  }
  if (!review) {
    return null;
  }

  function setForm(itemId: string, patch: Partial<GradeFormState>) {
    setForms((prev) => ({
      ...prev,
      [itemId]: { ...(prev[itemId] ?? blankForm), ...patch },
    }));
  }

  async function handleSave(itemId: string) {
    if (!review) return;
    const form = forms[itemId] ?? blankForm;
    const awarded = form.awarded.trim();
    if (!awarded) return;
    const feedback = form.feedback.trim();
    await gradeMutation.mutateAsync({
      attemptId: review.attempt_id,
      itemId,
      payload: {
        awarded_score: awarded,
        feedback: feedback || null,
      },
    });
  }

  return (
    <div className="dashboard-page">
      <p>
        <Link to="/app/grading">← Quay lại hàng chờ chấm</Link>
      </p>
      <h1>Chấm bài làm</h1>
      <p className="dashboard-hint">
        Học sinh: <strong>{review.student_name || review.student_user_id}</strong> —{' '}
        Trạng thái:{' '}
        <span className={`status-pill status-${review.grading_status.toLowerCase()}`}>
          {review.grading_status}
        </span>
        {review.submitted_at && (
          <>
            {' '}
            · Nộp lúc: {new Date(review.submitted_at).toLocaleString('vi-VN')}
          </>
        )}
      </p>

      {gradeMutation.isError && (
        <ErrorState
          error={gradeMutation.error}
          title="Lưu điểm thất bại"
        />
      )}
      {gradeMutation.isSuccess && (
        <div
          className="success-banner"
          role="status"
          data-testid="grading-save-success"
        >
          Đã lưu điểm. Tổng điểm hiện tại: {gradeMutation.data.attempt_score} /{' '}
          {gradeMutation.data.attempt_max_score} ({gradeMutation.data.grading_status}).
        </div>
      )}

      <ol className="review-list" data-testid="grading-items-list">
        {review.items.map((item) => {
          const promptText = getPromptText(item.prompt);
          const studentText = getTextAnswer(item.student_answer?.answer_payload);
          const accepted = getAcceptedAnswers(item.prompt);
          const isGradable =
            item.question_type === 'essay' || item.question_type === 'short_answer';
          const form = forms[item.id] ?? blankForm;
          return (
            <li
              key={item.id}
              className={`review-item ${item.item_grade ? 'correct' : isGradable ? 'pending' : ''}`}
              data-testid="grading-item"
            >
              <div className="review-item-header">
                <span className="review-item-number">Câu {item.position}</span>
                <span className="review-item-type">
                  {questionTypeLabel(item.question_type)}
                </span>
                <span className="review-item-points">{item.points} điểm</span>
                {item.item_grade && (
                  <span className="review-item-badge correct">
                    Đã chấm: {item.item_grade.awarded_score}
                  </span>
                )}
                {isGradable && !item.item_grade && (
                  <span className="review-item-badge pending">Chờ chấm</span>
                )}
              </div>
              <p className="review-prompt">
                {promptText || 'Câu hỏi chưa có nội dung'}
              </p>

              {isGradable && (
                <div className="review-short-answer">
                  <p className="review-answer">
                    <strong>Bài làm của học sinh:</strong>
                  </p>
                  <pre className="review-essay-text">{studentText || '—'}</pre>
                  {accepted.length > 0 && (
                    <p className="review-answer">
                      <strong>Đáp án tham khảo:</strong> {accepted.join(', ')}
                    </p>
                  )}
                  <div className="grade-form">
                    <label>
                      Điểm (0 – {item.points})
                      <input
                        type="text"
                        inputMode="decimal"
                        data-testid={`grade-score-${item.id}`}
                        value={form.awarded}
                        onChange={(e) =>
                          setForm(item.id, { awarded: e.target.value })
                        }
                        placeholder={item.points}
                      />
                    </label>
                    <label>
                      Nhận xét
                      <textarea
                        rows={3}
                        data-testid={`grade-feedback-${item.id}`}
                        value={form.feedback}
                        onChange={(e) =>
                          setForm(item.id, { feedback: e.target.value })
                        }
                      />
                    </label>
                    <button
                      type="button"
                      className="primary"
                      data-testid={`grade-save-${item.id}`}
                      disabled={gradeMutation.isPending || !form.awarded.trim()}
                      onClick={() => handleSave(item.id)}
                    >
                      {item.item_grade ? 'Cập nhật điểm' : 'Lưu điểm'}
                    </button>
                  </div>
                </div>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
