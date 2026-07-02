import { useEffect, useRef, useState } from 'react';
import {
  useMarkReadMutation,
  useNotificationsQuery,
  useUnreadCountQuery,
} from '@/shared/api/notifications-queries';
import type { Notification } from '@/shared/api/notifications';

function formatRelativeTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function eventLabel(eventType: string): string {
  switch (eventType) {
    case 'attempt.graded':
      return 'Bài thi';
    case 'assessment.published':
      return 'Đề thi';
    case 'resource.published':
      return 'Tài liệu';
    default:
      return eventType;
  }
}

function NotificationItemView({
  item,
  onMarkRead,
  isPending,
}: {
  item: Notification;
  onMarkRead: (id: string) => void;
  isPending: boolean;
}) {
  return (
    <li
      className={`notification-item${item.is_read ? ' is-read' : ''}`}
      data-testid={`notification-item-${item.id}`}
    >
      <div className="notification-item__header">
        <span className="notification-item__kind" aria-hidden="true">
          {eventLabel(item.event_type)}
        </span>
        <time
          className="notification-item__time"
          dateTime={item.created_at}
          aria-label={`Lúc ${formatRelativeTime(item.created_at)}`}
        >
          {formatRelativeTime(item.created_at)}
        </time>
      </div>
      <p className="notification-item__title">{item.title}</p>
      <p className="notification-item__body">{item.body}</p>
      <div className="notification-item__actions">
        {!item.is_read && (
          <button
            type="button"
            onClick={() => onMarkRead(item.id)}
            disabled={isPending}
            aria-label={`Đánh dấu đã đọc: ${item.title}`}
          >
            Đánh dấu đã đọc
          </button>
        )}
      </div>
    </li>
  );
}

// The id is reused by aria-controls on the trigger button and aria-labelledby
// on the dropdown panel so screen readers can announce the relationship.
const DROPDOWN_ID = 'notification-bell-dropdown';
const DROPDOWN_TITLE_ID = 'notification-bell-dropdown-title';

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const lastFocusedBeforeOpenRef = useRef<HTMLElement | null>(null);

  const unreadQuery = useUnreadCountQuery();
  const listQuery = useNotificationsQuery({ limit: 20, enabled: open });
  const markRead = useMarkReadMutation();

  const unread = unreadQuery.data ?? 0;
  const items = listQuery.data ?? [];

  useEffect(() => {
    if (!open) return;
    // Remember the trigger so we can return focus to it when the dropdown
    // closes (Escape / click-outside / item click).
    if (typeof document !== 'undefined' && document.activeElement instanceof HTMLElement) {
      lastFocusedBeforeOpenRef.current = document.activeElement;
    }
    const handler = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setOpen(false);
      }
    };
    const keyHandler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setOpen(false);
        // Restore focus to the trigger so keyboard users do not lose context.
        const previous = lastFocusedBeforeOpenRef.current;
        if (previous && typeof previous.focus === 'function') {
          previous.focus();
        }
      }
    };
    document.addEventListener('mousedown', handler);
    document.addEventListener('keydown', keyHandler);
    return () => {
      document.removeEventListener('mousedown', handler);
      document.removeEventListener('keydown', keyHandler);
    };
  }, [open]);

  return (
    <div
      className={`notification-bell${open ? ' is-open' : ''}`}
      ref={containerRef}
    >
      <button
        type="button"
        className="notification-bell__button"
        aria-label="Thông báo"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-controls={open ? DROPDOWN_ID : undefined}
        onClick={() => setOpen((value) => !value)}
        data-testid="notification-bell"
      >
        <span aria-hidden="true" className="notification-bell__icon">
          🔔
        </span>
        <span
          className="notification-bell__badge"
          aria-hidden="true"
          data-testid="notification-bell-badge"
          style={{ display: unread > 0 ? 'inline-flex' : 'none' }}
        >
          {unread > 99 ? '99+' : unread}
        </span>
        <span
          className="visually-hidden"
          aria-live="polite"
          data-testid="notification-unread-count"
        >
          {unread > 0
            ? `${unread} thông báo chưa đọc`
            : 'Không có thông báo chưa đọc'}
        </span>
      </button>
      {open && (
        <div
          className="notification-bell__dropdown"
          id={DROPDOWN_ID}
          role="dialog"
          aria-labelledby={DROPDOWN_TITLE_ID}
          data-testid="notification-dropdown"
        >
          <h2
            className="notification-bell__dropdown-header"
            id={DROPDOWN_TITLE_ID}
          >
            Thông báo
          </h2>
          {listQuery.isLoading && (
            <p className="notification-bell__empty">Đang tải…</p>
          )}
          {!listQuery.isLoading && items.length === 0 && (
            <p className="notification-bell__empty" data-testid="notification-empty">
              Bạn chưa có thông báo nào.
            </p>
          )}
          {items.length > 0 && (
            <ul className="notification-bell__list">
              {items.map((item) => (
                <NotificationItemView
                  key={item.id}
                  item={item}
                  isPending={markRead.isPending}
                  onMarkRead={(id) => markRead.mutate(id)}
                />
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
