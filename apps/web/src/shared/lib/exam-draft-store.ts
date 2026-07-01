export interface ExamDraft {
  attempt_id: string;
  item_id: string;
  payload: unknown;
  pending: boolean;
  revision?: number;
  updated_at: number;
}

export interface ExamDraftStorage {
  getDraft(attemptId: string, itemId: string): Promise<ExamDraft | undefined>;
  setDraft(draft: ExamDraft): Promise<void>;
  getPendingByAttempt(attemptId: string): Promise<ExamDraft[]>;
  deleteDraft(attemptId: string, itemId: string): Promise<void>;
  deleteAllForAttempt(attemptId: string): Promise<void>;
}

function draftKey(attemptId: string, itemId: string): string {
  return `${attemptId}:${itemId}`;
}

/**
 * Decide whether a local draft should override the server answer.
 * - Pending drafts always win (they represent unsynced user input).
 * - If either side has no revision, prefer the draft.
 * - Otherwise prefer the draft only when its revision is at least the server's.
 */
export function shouldPreferDraft(
  draft: Pick<ExamDraft, 'pending' | 'revision'>,
  serverRevision?: number
): boolean {
  if (draft.pending) return true;
  if (typeof draft.revision !== 'number') return true;
  if (typeof serverRevision !== 'number') return true;
  return draft.revision >= serverRevision;
}

export class InMemoryExamDraftStorage implements ExamDraftStorage {
  private drafts = new Map<string, ExamDraft>();

  async getDraft(attemptId: string, itemId: string): Promise<ExamDraft | undefined> {
    return this.drafts.get(draftKey(attemptId, itemId));
  }

  async setDraft(draft: ExamDraft): Promise<void> {
    this.drafts.set(draftKey(draft.attempt_id, draft.item_id), { ...draft });
  }

  async getPendingByAttempt(attemptId: string): Promise<ExamDraft[]> {
    const result: ExamDraft[] = [];
    for (const draft of this.drafts.values()) {
      if (draft.attempt_id === attemptId && draft.pending) {
        result.push({ ...draft });
      }
    }
    return result;
  }

  async deleteDraft(attemptId: string, itemId: string): Promise<void> {
    this.drafts.delete(draftKey(attemptId, itemId));
  }

  async deleteAllForAttempt(attemptId: string): Promise<void> {
    for (const key of this.drafts.keys()) {
      if (key.startsWith(`${attemptId}:`)) {
        this.drafts.delete(key);
      }
    }
  }
}

const DB_NAME = 'vts-exam-drafts';
const DB_VERSION = 1;
const STORE_NAME = 'drafts';

export class IndexedDBExamDraftStorage implements ExamDraftStorage {
  private constructor(private db: IDBDatabase) {}

  static async create(): Promise<IndexedDBExamDraftStorage> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => reject(request.error ?? new Error('Failed to open IndexedDB'));
      request.onsuccess = () => resolve(new IndexedDBExamDraftStorage(request.result));

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const store = db.createObjectStore(STORE_NAME, { keyPath: 'id' });
          store.createIndex('byAttempt', 'attempt_id', { unique: false });
        }
      };
    });
  }

  async getDraft(attemptId: string, itemId: string): Promise<ExamDraft | undefined> {
    const id = draftKey(attemptId, itemId);
    return this.requestToPromise<ExamDraft | undefined>((transaction) =>
      transaction.objectStore(STORE_NAME).get(id)
    );
  }

  async setDraft(draft: ExamDraft): Promise<void> {
    const record = { id: draftKey(draft.attempt_id, draft.item_id), ...draft };
    await this.requestToPromise<void>((transaction) =>
      transaction.objectStore(STORE_NAME).put(record)
    );
  }

  async getPendingByAttempt(attemptId: string): Promise<ExamDraft[]> {
    const all = await this.requestToPromise<ExamDraft[]>((transaction) => {
      const store = transaction.objectStore(STORE_NAME);
      const index = store.index('byAttempt');
      return index.getAll(attemptId);
    });
    return all.filter((draft) => draft.pending);
  }

  async deleteDraft(attemptId: string, itemId: string): Promise<void> {
    const id = draftKey(attemptId, itemId);
    await this.requestToPromise<void>((transaction) =>
      transaction.objectStore(STORE_NAME).delete(id)
    );
  }

  async deleteAllForAttempt(attemptId: string): Promise<void> {
    const transaction = this.db.transaction(STORE_NAME, 'readwrite');
    const store = transaction.objectStore(STORE_NAME);
    const index = store.index('byAttempt');
    const request = index.openCursor(attemptId);

    return new Promise((resolve, reject) => {
      request.onerror = () => reject(request.error ?? new Error('Failed to delete drafts'));
      request.onsuccess = () => {
        const cursor = request.result;
        if (cursor) {
          cursor.delete();
          cursor.continue();
        } else {
          resolve();
        }
      };
    });
  }

  private requestToPromise<T>(
    execute: (transaction: IDBTransaction) => IDBRequest
  ): Promise<T> {
    const transaction = this.db.transaction(STORE_NAME, 'readwrite');
    const request = execute(transaction);

    return new Promise((resolve, reject) => {
      request.onerror = () => reject(request.error ?? new Error('IndexedDB request failed'));
      request.onsuccess = () => resolve(request.result as T);
      transaction.onerror = () => reject(transaction.error ?? new Error('IndexedDB transaction failed'));
    });
  }
}

/**
 * Create the best available draft storage for the current environment.
 * Falls back to an in-memory store when IndexedDB is unavailable so the
 * exam page never crashes, but in-memory drafts do not survive reloads.
 */
export async function createExamDraftStorage(): Promise<ExamDraftStorage> {
  if (typeof window === 'undefined' || !('indexedDB' in window)) {
    return new InMemoryExamDraftStorage();
  }
  try {
    return await IndexedDBExamDraftStorage.create();
  } catch {
    return new InMemoryExamDraftStorage();
  }
}
