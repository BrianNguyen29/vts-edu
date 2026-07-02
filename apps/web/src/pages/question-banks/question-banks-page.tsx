import { useEffect, useState } from 'react';
import { ApiResponseError } from '@/shared/api/assessments';
import {
  createQuestionBank,
  createQuestionInBank,
  listQuestionBanks,
  listQuestionsInBank,
  publishQuestionVersion,
  type CreateQuestionRequest,
  type QuestionBank,
  type QuestionBankQuestion,
  type QuestionVersion,
} from '@/shared/api/assessments';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

type QuestionType = 'multiple_choice' | 'short_answer' | 'essay';

interface NewQuestionForm {
  questionType: QuestionType;
  promptText: string;
  choices: { id: string; text: string }[];
  correctOption: string;
  acceptedAnswers: string;
}

const blankForm = (): NewQuestionForm => ({
  questionType: 'multiple_choice',
  promptText: '',
  choices: [
    { id: 'A', text: '' },
    { id: 'B', text: '' },
    { id: 'C', text: '' },
    { id: 'D', text: '' },
  ],
  correctOption: 'A',
  acceptedAnswers: '',
});

function questionTypeLabel(t: string | undefined): string {
  switch (t) {
    case 'short_answer':
      return 'Trả lời ngắn';
    case 'essay':
      return 'Tự luận';
    case 'multiple_choice':
    default:
      return 'Trắc nghiệm';
  }
}

export function QuestionBanksPage() {
  useDocumentTitle('Bộ câu hỏi');

  const [banks, setBanks] = useState<QuestionBank[] | null>(null);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [newBankTitle, setNewBankTitle] = useState('');
  const [creatingBank, setCreatingBank] = useState(false);
  const [selectedBankId, setSelectedBankId] = useState<string | null>(null);
  const [questions, setQuestions] = useState<QuestionBankQuestion[] | null>(null);
  const [form, setForm] = useState<NewQuestionForm>(blankForm());
  const [formError, setFormError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  async function refreshBanks() {
    try {
      const data = await listQuestionBanks({ limit: 50 });
      setBanks(data);
      if (!selectedBankId && data.length > 0) {
        setSelectedBankId(data[0].id);
      }
    } catch (err) {
      setLoadError(err);
    }
  }

  async function refreshQuestions(bankID: string) {
    try {
      const data = await listQuestionsInBank(bankID, { limit: 50 });
      setQuestions(data);
    } catch (err) {
      setLoadError(err);
    }
  }

  useEffect(() => {
    void refreshBanks();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (selectedBankId) {
      void refreshQuestions(selectedBankId);
    } else {
      setQuestions(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedBankId]);

  async function handleCreateBank(e: React.FormEvent) {
    e.preventDefault();
    setCreatingBank(true);
    setFormError(null);
    try {
      const bank = await createQuestionBank({ title: newBankTitle.trim() });
      setNewBankTitle('');
      setSelectedBankId(bank.id);
      await refreshBanks();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Không thể tạo bộ câu hỏi.');
    } finally {
      setCreatingBank(false);
    }
  }

  function buildPayload(): CreateQuestionRequest | null {
    if (!form.promptText.trim()) {
      setFormError('Vui lòng nhập đề bài.');
      return null;
    }
    if (form.questionType === 'multiple_choice') {
      if (form.choices.some((c) => !c.text.trim())) {
        setFormError('Mỗi lựa chọn phải có nội dung.');
        return null;
      }
      return {
        question_type: 'multiple_choice',
        prompt: { text: form.promptText.trim() },
        choices: form.choices.map((c) => ({ id: c.id, text: c.text.trim() })),
        answer_key: { correct_option: form.correctOption },
        max_score: '1.00',
      };
    }
    if (form.questionType === 'short_answer') {
      const accepted = form.acceptedAnswers
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean);
      if (accepted.length === 0) {
        setFormError('Vui lòng nhập ít nhất một đáp án hợp lệ (phân cách bằng dấu phẩy).');
        return null;
      }
      return {
        question_type: 'short_answer',
        prompt: { text: form.promptText.trim() },
        answer_key: { accepted_answers: accepted },
        max_score: '1.00',
      };
    }
    return {
      question_type: 'essay',
      prompt: { text: form.promptText.trim() },
      answer_key: { rubric: '' },
      max_score: '2.00',
    };
  }

  async function handleCreateQuestion(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedBankId) return;
    setFormError(null);
    const payload = buildPayload();
    if (!payload) return;
    setCreating(true);
    try {
      await createQuestionInBank(selectedBankId, payload);
      setForm(blankForm());
      await refreshQuestions(selectedBankId);
    } catch (err) {
      setFormError(
        err instanceof ApiResponseError
          ? err.body.error.message
          : err instanceof Error
          ? err.message
          : 'Không thể tạo câu hỏi.'
      );
    } finally {
      setCreating(false);
    }
  }

  async function handlePublish(questionID: string, version: QuestionVersion) {
    if (!selectedBankId) return;
    try {
      await publishQuestionVersion(selectedBankId, questionID, version.id);
      await refreshQuestions(selectedBankId);
    } catch (err) {
      setFormError(
        err instanceof ApiResponseError
          ? err.body.error.message
          : err instanceof Error
          ? err.message
          : 'Không thể xuất bản.'
      );
    }
  }

  if (loadError && !banks) {
    return (
      <div className="dashboard-page">
        <ErrorState error={loadError} />
      </div>
    );
  }

  return (
    <div className="dashboard-page">
      <h1>Bộ câu hỏi</h1>

      <section className="qb-section">
        <h2>Tạo bộ câu hỏi mới</h2>
        <form onSubmit={handleCreateBank} className="qb-create-bank-form">
          <label htmlFor="qb-title">Tên bộ câu hỏi</label>
          <input
            id="qb-title"
            type="text"
            value={newBankTitle}
            onChange={(e) => setNewBankTitle(e.target.value)}
            required
            data-testid="qb-title"
          />
          <button type="submit" disabled={creatingBank || !newBankTitle.trim()}>
            {creatingBank ? 'Đang tạo…' : 'Tạo bộ câu hỏi'}
          </button>
        </form>
      </section>

      <section className="qb-section">
        <h2>Danh sách bộ câu hỏi</h2>
        {banks === null ? (
          <p>Đang tải…</p>
        ) : banks.length === 0 ? (
          <p>Chưa có bộ câu hỏi nào.</p>
        ) : (
          <ul className="qb-bank-list">
            {banks.map((bank) => (
              <li key={bank.id}>
                <button
                  type="button"
                  className={bank.id === selectedBankId ? 'qb-bank active' : 'qb-bank'}
                  onClick={() => setSelectedBankId(bank.id)}
                >
                  {bank.title}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      {selectedBankId && (
        <section className="qb-section">
          <h2>Tạo câu hỏi mới</h2>
          <form onSubmit={handleCreateQuestion} className="qb-create-question-form">
            <div className="qb-field">
              <label htmlFor="qb-qtype">Loại câu hỏi</label>
              <select
                id="qb-qtype"
                value={form.questionType}
                onChange={(e) =>
                  setForm({ ...form, questionType: e.target.value as QuestionType })
                }
              >
                <option value="multiple_choice">Trắc nghiệm</option>
                <option value="short_answer">Trả lời ngắn</option>
                <option value="essay">Tự luận</option>
              </select>
            </div>
            <div className="qb-field">
              <label htmlFor="qb-prompt">Đề bài</label>
              <textarea
                id="qb-prompt"
                rows={3}
                value={form.promptText}
                onChange={(e) => setForm({ ...form, promptText: e.target.value })}
                data-testid="qb-prompt"
              />
            </div>
            {form.questionType === 'multiple_choice' && (
              <div className="qb-field">
                <label>Các lựa chọn</label>
                {form.choices.map((c, idx) => (
                  <div key={c.id} className="qb-choice-row">
                    <span>{c.id}.</span>
                    <input
                      type="text"
                      value={c.text}
                      onChange={(e) => {
                        const next = [...form.choices];
                        next[idx] = { ...c, text: e.target.value };
                        setForm({ ...form, choices: next });
                      }}
                    />
                    <label className="qb-correct-radio">
                      <input
                        type="radio"
                        name="qb-correct"
                        value={c.id}
                        checked={form.correctOption === c.id}
                        onChange={() => setForm({ ...form, correctOption: c.id })}
                      />
                      Đúng
                    </label>
                  </div>
                ))}
              </div>
            )}
            {form.questionType === 'short_answer' && (
              <div className="qb-field">
                <label htmlFor="qb-accepted">Đáp án hợp lệ (phân cách bằng dấu phẩy)</label>
                <input
                  id="qb-accepted"
                  type="text"
                  value={form.acceptedAnswers}
                  onChange={(e) => setForm({ ...form, acceptedAnswers: e.target.value })}
                  placeholder="ví dụ: 7, bảy, seven"
                />
              </div>
            )}
            {formError && (
              <p role="alert" className="error-banner">
                {formError}
              </p>
            )}
            <button type="submit" disabled={creating} data-testid="qb-create">
              {creating ? 'Đang tạo…' : 'Tạo và xuất bản'}
            </button>
          </form>
        </section>
      )}

      {selectedBankId && (
        <section className="qb-section">
          <h2>Câu hỏi trong bộ</h2>
          {questions === null ? (
            <p>Đang tải…</p>
          ) : questions.length === 0 ? (
            <p>Chưa có câu hỏi nào.</p>
          ) : (
            <ul className="qb-question-list">
              {questions.map((q) => (
                <li key={q.id} className="qb-question-row">
                  <div>
                    <span className={`qb-type-badge qb-type-${q.question_type ?? ''}`}>
                      {questionTypeLabel(q.question_type ?? undefined)}
                    </span>
                    {q.latest_version_status && (
                      <span className={`qb-status-badge qb-status-${q.latest_version_status.toLowerCase()}`}>
                        {q.latest_version_status}
                      </span>
                    )}
                    {q.latest_version && (
                      <span className="qb-version">v{q.latest_version}</span>
                    )}
                  </div>
                  {q.latest_version_status === 'DRAFT' && q.latest_version_id && (
                    <button
                      type="button"
                      onClick={() =>
                        q.latest_version_id &&
                        handlePublish(q.id, {
                          id: q.latest_version_id,
                          question_id: q.id,
                          version: q.latest_version ?? 1,
                          question_type: (q.question_type as QuestionType) ?? 'multiple_choice',
                          prompt: {},
                          max_score: '1.00',
                          status: 'DRAFT',
                          created_at: q.created_at,
                        })
                      }
                    >
                      Xuất bản
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  );
}
