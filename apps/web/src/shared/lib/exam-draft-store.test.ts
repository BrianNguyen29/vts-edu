import { describe, it, expect, beforeEach } from 'vitest';
import {
  InMemoryExamDraftStorage,
  shouldPreferDraft,
  type ExamDraft,
} from './exam-draft-store';

describe('InMemoryExamDraftStorage', () => {
  let storage: InMemoryExamDraftStorage;

  beforeEach(() => {
    storage = new InMemoryExamDraftStorage();
  });

  const sampleDraft = (overrides: Partial<ExamDraft> = {}): ExamDraft => ({
    attempt_id: 'attempt-1',
    item_id: 'item-1',
    payload: { selected_option: 'A' },
    pending: true,
    updated_at: Date.now(),
    ...overrides,
  });

  it('stores and retrieves a draft', async () => {
    const draft = sampleDraft();
    await storage.setDraft(draft);
    const found = await storage.getDraft(draft.attempt_id, draft.item_id);
    expect(found).toEqual(draft);
  });

  it('returns undefined for missing drafts', async () => {
    const found = await storage.getDraft('attempt-1', 'missing');
    expect(found).toBeUndefined();
  });

  it('updates an existing draft', async () => {
    await storage.setDraft(sampleDraft({ payload: { selected_option: 'A' } }));
    await storage.setDraft(sampleDraft({ payload: { selected_option: 'B' } }));
    const found = await storage.getDraft('attempt-1', 'item-1');
    expect(found?.payload).toEqual({ selected_option: 'B' });
  });

  it('lists only pending drafts for an attempt', async () => {
    await storage.setDraft(sampleDraft({ attempt_id: 'attempt-1', item_id: 'item-1' }));
    await storage.setDraft(
      sampleDraft({ attempt_id: 'attempt-1', item_id: 'item-2', pending: false, revision: 3 })
    );
    await storage.setDraft(sampleDraft({ attempt_id: 'attempt-2', item_id: 'item-1' }));

    const pending = await storage.getPendingByAttempt('attempt-1');
    expect(pending).toHaveLength(1);
    expect(pending[0].item_id).toBe('item-1');
  });

  it('deletes a single draft', async () => {
    await storage.setDraft(sampleDraft());
    await storage.deleteDraft('attempt-1', 'item-1');
    const found = await storage.getDraft('attempt-1', 'item-1');
    expect(found).toBeUndefined();
  });

  it('deletes all drafts for an attempt', async () => {
    await storage.setDraft(sampleDraft({ item_id: 'item-1' }));
    await storage.setDraft(sampleDraft({ item_id: 'item-2' }));
    await storage.setDraft(sampleDraft({ attempt_id: 'attempt-2', item_id: 'item-1' }));

    await storage.deleteAllForAttempt('attempt-1');

    expect(await storage.getDraft('attempt-1', 'item-1')).toBeUndefined();
    expect(await storage.getDraft('attempt-1', 'item-2')).toBeUndefined();
    expect(await storage.getDraft('attempt-2', 'item-1')).toBeDefined();
  });
});

describe('shouldPreferDraft', () => {
  it('prefers pending drafts regardless of revision', () => {
    expect(shouldPreferDraft({ pending: true, revision: 1 }, 5)).toBe(true);
  });

  it('prefers a draft when the server has no revision', () => {
    expect(shouldPreferDraft({ pending: false, revision: 1 }, undefined)).toBe(true);
  });

  it('prefers a draft when it has no revision', () => {
    expect(shouldPreferDraft({ pending: false, revision: undefined }, 5)).toBe(true);
  });

  it('prefers a draft with a newer or equal revision', () => {
    expect(shouldPreferDraft({ pending: false, revision: 5 }, 4)).toBe(true);
    expect(shouldPreferDraft({ pending: false, revision: 5 }, 5)).toBe(true);
  });

  it('defers to the server when the draft revision is older', () => {
    expect(shouldPreferDraft({ pending: false, revision: 3 }, 5)).toBe(false);
  });
});
