import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useDocumentTitle } from './use-document-title';

describe('useDocumentTitle', () => {
  const ORIGINAL_TITLE = document.title;

  beforeEach(() => {
    document.title = ORIGINAL_TITLE;
  });

  afterEach(() => {
    document.title = ORIGINAL_TITLE;
  });

  it('appends the title with the brand suffix on mount', () => {
    renderHook(() => useDocumentTitle('Đăng nhập'));
    expect(document.title).toBe('Đăng nhập – VTS EDU');
  });

  it('falls back to the brand name when title is empty', () => {
    renderHook(() => useDocumentTitle(''));
    expect(document.title).toBe('VTS EDU');
  });

  it('falls back to the brand name when title is null', () => {
    renderHook(() => useDocumentTitle(null));
    expect(document.title).toBe('VTS EDU');
  });

  it('updates the title when the value changes', () => {
    const { rerender } = renderHook(({ title }) => useDocumentTitle(title), {
      initialProps: { title: 'Đăng nhập' },
    });
    expect(document.title).toBe('Đăng nhập – VTS EDU');
    rerender({ title: 'Trang làm việc' });
    expect(document.title).toBe('Trang làm việc – VTS EDU');
  });

  it('restores the previous title on unmount', () => {
    document.title = 'Original';
    const { unmount } = renderHook(() => useDocumentTitle('Exam'));
    expect(document.title).toBe('Exam – VTS EDU');
    unmount();
    expect(document.title).toBe('Original');
  });
});
