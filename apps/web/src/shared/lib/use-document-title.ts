import { useEffect } from 'react';

/**
 * Set the document title for the current page.
 *
 * Re-runs whenever the provided title changes. The "VTS EDU" suffix is
 * appended so the tab title stays recognisable when multiple pages are
 * open. If `title` is empty/undefined, the previous title is restored to
 * "VTS EDU".
 */
export function useDocumentTitle(title: string | undefined | null): void {
  useEffect(() => {
    const fallback = 'VTS EDU';
    if (typeof document === 'undefined') return;
    const previous = document.title;
    document.title = title && title.trim().length > 0 ? `${title} – ${fallback}` : fallback;
    return () => {
      document.title = previous;
    };
  }, [title]);
}
