import { useEffect, useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import {
  createUser,
  getOrganization,
  importUsers,
  listUsers,
  resetUserPassword,
  updateOrganization,
  updateUserRoles,
  type CreateUserRequest,
  type ImportUsersResult,
  type UpdateRolesRequest,
  type User,
} from '@/shared/api/admin';
import { ApiResponseError } from '@/shared/api/attempts';
import { PasswordPolicyHints } from '@/shared/components/password-policy-hints';
import { validatePassword } from '@/shared/lib/password-policy';
import { AuditLogsPanel } from './audit-logs-panel';
import { AcademicManagementPanel } from './academic-management-panel';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

const AVAILABLE_ROLES = ['student', 'teacher', 'admin'] as const;

type ViewMode = 'list' | 'create' | 'edit-roles' | 'reset-password' | 'import-csv';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền truy cập chức năng quản trị.';
      case 404:
        return 'Không tìm thấy dữ liệu.';
      case 409:
        return 'Tên đăng nhập đã tồn tại.';
      default:
        return err.body.error.message || 'Yêu cầu thất bại.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function AdminDashboardPage() {
  const auth = useAuth();

  useDocumentTitle('Trang quản trị');

  const [orgName, setOrgName] = useState('');
  const [orgCode, setOrgCode] = useState('');
  const [orgLoading, setOrgLoading] = useState(true);

  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);
  const [userCursor, setUserCursor] = useState<string | undefined>();
  const [userHasMore, setUserHasMore] = useState(false);
  const [isLoadingMoreUsers, setIsLoadingMoreUsers] = useState(false);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [activeTab, setActiveTab] = useState<'org' | 'users' | 'audit' | 'academic'>('org');

  const [mode, setMode] = useState<ViewMode>('list');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const [orgEditing, setOrgEditing] = useState(false);
  const [orgDraft, setOrgDraft] = useState('');

  // create form
  const [loginName, setLoginName] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [tempPassword, setTempPassword] = useState('');
  const [newUserRoles, setNewUserRoles] = useState<string[]>(['student']);

  // edit roles
  const [editRoles, setEditRoles] = useState<string[]>([]);

  // reset password
  const [resetPassword, setResetPassword] = useState('');

  // import users
  const [importCsv, setImportCsv] = useState('');
  const [importPreview, setImportPreview] = useState<ImportUsersResult | null>(
    null
  );
  const [importLoading, setImportLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setUsersLoading(true);
      setUserCursor(undefined);
      setUserHasMore(false);
      try {
        const [orgData, usersData] = await Promise.all([
          getOrganization(),
          listUsers({ q: searchQuery || undefined, limit: 10 }),
        ]);
        if (cancelled) return;
        setOrgName(orgData.name);
        setOrgCode(orgData.code);
        setUsers(usersData.data);
        setUserCursor(usersData.page?.next_cursor ?? undefined);
        setUserHasMore(usersData.page?.has_more ?? false);
      } catch (err) {
        if (cancelled) return;
        setError(formatFriendlyError(err));
      } finally {
        if (!cancelled) {
          setOrgLoading(false);
          setUsersLoading(false);
        }
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, [searchQuery]);

  async function loadMoreUsers() {
    if (!userHasMore || !userCursor || isLoadingMoreUsers) return;
    setIsLoadingMoreUsers(true);
    try {
      const response = await listUsers({
        q: searchQuery || undefined,
        limit: 10,
        cursor: userCursor,
      });
      setUsers((prev) => [...prev, ...response.data]);
      setUserCursor(response.page?.next_cursor ?? undefined);
      setUserHasMore(response.page?.has_more ?? false);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setIsLoadingMoreUsers(false);
    }
  }

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  function clearMessages() {
    setError(null);
    setSuccess(null);
  }

  async function handleUpdateOrgName(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();
    const name = orgDraft.trim();
    if (!name) return;

    try {
      const updated = await updateOrganization({ name });
      setOrgName(updated.name);
      setOrgEditing(false);
      setSuccess('Đã cập nhật tên tổ chức.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleCreateUser(e: React.FormEvent) {
    e.preventDefault();
    clearMessages();

    if (newUserRoles.length === 0) {
      setError('Vui lòng chọn ít nhất một vai trò.');
      return;
    }

    const passwordCheck = validatePassword(tempPassword);
    if (!passwordCheck.valid) {
      setError('Mật khẩu tạm chưa đáp ứng yêu cầu bảo mật.');
      return;
    }

    try {
      const created = await createUser({
        login_name: loginName.trim(),
        display_name: displayName.trim(),
        email: email.trim(),
        temporary_password: tempPassword,
        roles: (newUserRoles as CreateUserRequest['roles']),
      });
      setUsers((prev) => [...prev, created]);
      setMode('list');
      resetCreateForm();
      setSuccess('Đã tạo người dùng mới.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  function resetCreateForm() {
    setLoginName('');
    setDisplayName('');
    setEmail('');
    setTempPassword('');
    setNewUserRoles(['student']);
  }

  async function handleUpdateRoles(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedUser) return;
    clearMessages();

    try {
      await updateUserRoles(selectedUser.id, { roles: editRoles as UpdateRolesRequest['roles'] });
      setUsers((prev) =>
        prev.map((u) => (u.id === selectedUser.id ? { ...u, roles: editRoles as UpdateRolesRequest['roles'] } : u))
      );
      setMode('list');
      setSelectedUser(null);
      setSuccess('Đã cập nhật vai trò.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleResetPassword(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedUser) return;
    clearMessages();

    const passwordCheck = validatePassword(resetPassword);
    if (!passwordCheck.valid) {
      setError('Mật khẩu tạm mới chưa đáp ứng yêu cầu bảo mật.');
      return;
    }

    try {
      await resetUserPassword(selectedUser.id, {
        temporary_password: resetPassword,
      });
      setMode('list');
      setSelectedUser(null);
      setResetPassword('');
      setSuccess('Đã đặt lại mật khẩu.');
    } catch (err) {
      setError(formatFriendlyError(err));
    }
  }

  async function handleImportUsers(dryRun: boolean) {
    clearMessages();
    setImportLoading(true);
    try {
      const result = await importUsers({ csv: importCsv, dry_run: dryRun });
      setImportPreview(result);
      if (!dryRun) {
        setSuccess(
          `Đã nhập ${result.created}/${result.total} người dùng.`
        );
        void loadUsers();
      }
    } catch (err) {
      setError(formatFriendlyError(err));
      setImportPreview(null);
    } finally {
      setImportLoading(false);
    }
  }

  async function loadUsers() {
    setUsersLoading(true);
    try {
      const usersData = await listUsers({ q: searchQuery || undefined, limit: 10 });
      setUsers(usersData.data);
      setUserCursor(usersData.page?.next_cursor ?? undefined);
      setUserHasMore(usersData.page?.has_more ?? false);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setUsersLoading(false);
    }
  }

  function toggleRole(roles: string[], role: string): string[] {
    return roles.includes(role)
      ? roles.filter((r) => r !== role)
      : [...roles, role];
  }

  function renderRoleCheckboxes(
    value: string[],
    onChange: (roles: string[]) => void
  ) {
    return (
      <div className="role-checkboxes">
        {AVAILABLE_ROLES.map((role) => (
          <label key={role} className="role-checkbox">
            <input
              type="checkbox"
              checked={value.includes(role)}
              onChange={() => onChange(toggleRole(value, role))}
            />
            <span className="role-label">{role}</span>
          </label>
        ))}
      </div>
    );
  }

  if (orgLoading || usersLoading) {
    return (
      <div className="dashboard-page">
        <h1>Trang quản trị</h1>
        <p className="dashboard-status">Đang tải dữ liệu…</p>
      </div>
    );
  }

  return (
    <div className="dashboard-page">
      <h1>Trang quản trị</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>

      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}
      {success && (
        <div className="success-banner" role="status">
          {success}
        </div>
      )}

      <div
        className="admin-tabs"
        role="tablist"
        aria-label="Quản lý quản trị"
      >
        <button
          type="button"
          role="tab"
          id="admin-tab-org"
          aria-controls="admin-panel-org"
          aria-selected={activeTab === 'org'}
          tabIndex={activeTab === 'org' ? 0 : -1}
          className={activeTab === 'org' ? 'active' : ''}
          onClick={() => setActiveTab('org')}
        >
          Tổ chức
        </button>
        <button
          type="button"
          role="tab"
          id="admin-tab-users"
          aria-controls="admin-panel-users"
          aria-selected={activeTab === 'users'}
          tabIndex={activeTab === 'users' ? 0 : -1}
          className={activeTab === 'users' ? 'active' : ''}
          onClick={() => setActiveTab('users')}
          data-testid="users-tab"
        >
          Người dùng
        </button>
        <button
          type="button"
          role="tab"
          id="admin-tab-audit"
          aria-controls="admin-panel-audit"
          aria-selected={activeTab === 'audit'}
          tabIndex={activeTab === 'audit' ? 0 : -1}
          className={activeTab === 'audit' ? 'active' : ''}
          onClick={() => setActiveTab('audit')}
        >
          Nhật ký hoạt động
        </button>
        <button
          type="button"
          role="tab"
          id="admin-tab-academic"
          aria-controls="admin-panel-academic"
          aria-selected={activeTab === 'academic'}
          tabIndex={activeTab === 'academic' ? 0 : -1}
          className={activeTab === 'academic' ? 'active' : ''}
          onClick={() => setActiveTab('academic')}
        >
          Học vụ
        </button>
      </div>

      {activeTab === 'org' && (
        <section
          className="admin-section"
          role="tabpanel"
          id="admin-panel-org"
          aria-labelledby="admin-tab-org"
        >
          <h2>Tổ chức</h2>
        <div className="org-card">
          <div className="org-info">
            <div>
              <strong>Tên:</strong>{' '}
              {orgEditing ? (
                <form onSubmit={handleUpdateOrgName} className="inline-form">
                  <label htmlFor="org-name-input" className="visually-hidden">
                    Tên tổ chức
                  </label>
                  <input
                    id="org-name-input"
                    type="text"
                    value={orgDraft}
                    onChange={(e) => setOrgDraft(e.target.value)}
                    required
                  />
                  <button type="submit" className="primary">
                    Lưu
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setOrgEditing(false);
                      setOrgDraft(orgName);
                    }}
                  >
                    Hủy
                  </button>
                </form>
              ) : (
                <>
                  {orgName}
                  <button
                    type="button"
                    className="text-button"
                    onClick={() => {
                      setOrgEditing(true);
                      setOrgDraft(orgName);
                    }}
                  >
                    Sửa
                  </button>
                </>
              )}
            </div>
            <div>
              <strong>Mã:</strong> {orgCode}
            </div>
          </div>
        </div>
      </section>
      )}

      {activeTab === 'users' && (
      <section
        className="admin-section"
        role="tabpanel"
        id="admin-panel-users"
        aria-labelledby="admin-tab-users"
      >
        <div className="section-header">
          <h2 id="admin-users-heading">Người dùng</h2>
          {mode === 'list' && (
            <div className="section-actions">
              <button
                type="button"
                onClick={() => {
                  clearMessages();
                  setMode('import-csv');
                  setImportCsv('');
                  setImportPreview(null);
                }}
                data-testid="import-csv-button"
              >
                Nhập CSV
              </button>
              <button
                type="button"
                className="primary"
                onClick={() => {
                  clearMessages();
                  setMode('create');
                }}
              >
                Thêm người dùng
              </button>
            </div>
          )}
        </div>

        {mode === 'list' && (
          <div className="search-bar">
            <label htmlFor="admin-user-search" className="visually-hidden">
              Tìm theo tên đăng nhập hoặc email
            </label>
            <input
              id="admin-user-search"
              type="search"
              placeholder="Tìm theo tên đăng nhập hoặc email…"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
            />
          </div>
        )}

        {mode === 'create' && (
          <form onSubmit={handleCreateUser} className="admin-form">
            <h3>Thêm người dùng mới</h3>
            <div className="field">
              <label htmlFor="loginName">Tên đăng nhập</label>
              <input
                id="loginName"
                type="text"
                required
                value={loginName}
                onChange={(e) => setLoginName(e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor="displayName">Tên hiển thị</label>
              <input
                id="displayName"
                type="text"
                required
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor="email">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor="tempPassword">Mật khẩu tạm</label>
              <input
                id="tempPassword"
                type="password"
                required
                value={tempPassword}
                onChange={(e) => setTempPassword(e.target.value)}
              />
              <PasswordPolicyHints password={tempPassword} />
            </div>
            <div className="field">
              <label>Vai trò</label>
              {renderRoleCheckboxes(newUserRoles, setNewUserRoles)}
            </div>
            <div className="form-actions">
              <button type="submit" className="primary">
                Tạo
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode('list');
                  resetCreateForm();
                }}
              >
                Hủy
              </button>
            </div>
          </form>
        )}

        {mode === 'edit-roles' && selectedUser && (
          <form onSubmit={handleUpdateRoles} className="admin-form">
            <h3>Cập nhật vai trò: {selectedUser.display_name}</h3>
            <div className="field">
              <label>Vai trò</label>
              {renderRoleCheckboxes(editRoles, setEditRoles)}
            </div>
            <div className="form-actions">
              <button type="submit" className="primary">
                Lưu
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode('list');
                  setSelectedUser(null);
                }}
              >
                Hủy
              </button>
            </div>
          </form>
        )}

        {mode === 'reset-password' && selectedUser && (
          <form onSubmit={handleResetPassword} className="admin-form">
            <h3>Đặt lại mật khẩu: {selectedUser.display_name}</h3>
            <div className="field">
              <label htmlFor="resetPassword">Mật khẩu tạm mới</label>
              <input
                id="resetPassword"
                type="password"
                required
                value={resetPassword}
                onChange={(e) => setResetPassword(e.target.value)}
              />
              <PasswordPolicyHints password={resetPassword} />
            </div>
            <div className="form-actions">
              <button type="submit" className="primary">
                Đặt lại
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode('list');
                  setSelectedUser(null);
                  setResetPassword('');
                }}
              >
                Hủy
              </button>
            </div>
          </form>
        )}

        {mode === 'import-csv' && (
          <div className="admin-form bulk-import-form">
            <h3>Nhập người dùng từ CSV</h3>
            <p className="hint">
              Dòng đầu tiên phải là:{' '}
              <code>login_name,display_name,email,temporary_password,roles</code>
              . Cột roles dùng dấu phẩy cho nhiều vai trò, ví dụ:{' '}
              <code>student,teacher</code>.
            </p>
            <div className="field">
              <label htmlFor="importCsv">Nội dung CSV</label>
              <textarea
                id="importCsv"
                rows={10}
                value={importCsv}
                onChange={(e) => {
                  setImportCsv(e.target.value);
                  setImportPreview(null);
                }}
                placeholder="login_name,display_name,email,temporary_password,roles"
                data-testid="import-csv-textarea"
              />
            </div>
            <div className="form-actions">
              <button
                type="button"
                className="primary"
                onClick={() => handleImportUsers(true)}
                disabled={importLoading || !importCsv.trim()}
                data-testid="dry-run-import-button"
              >
                {importLoading && importPreview === null
                  ? 'Đang kiểm tra…'
                  : 'Kiểm tra'}
              </button>
              <button
                type="button"
                className="primary"
                onClick={() => handleImportUsers(false)}
                disabled={
                  importLoading ||
                  !importCsv.trim() ||
                  (importPreview?.failed ?? 0) === importPreview?.total
                }
                data-testid="confirm-import-button"
              >
                {importLoading && importPreview !== null
                  ? 'Đang nhập…'
                  : 'Xác nhận nhập'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode('list');
                  setImportCsv('');
                  setImportPreview(null);
                }}
              >
                Hủy
              </button>
            </div>

            {importPreview && (
              <div className="bulk-preview" data-testid="import-preview">
                <p>
                  Tổng: <strong>{importPreview.total}</strong> · Đã tạo/ hợp lệ:{' '}
                  <strong>{importPreview.created}</strong> · Lỗi:{' '}
                  <strong>{importPreview.failed}</strong>
                  {importPreview.dry_run && (
                    <span className="dry-run-badge">Chế độ kiểm tra</span>
                  )}
                </p>
                <div className="table-wrap">
                  <table className="gradebook-table">
                    <thead>
                      <tr>
                        <th>Dòng</th>
                        <th>Tên đăng nhập</th>
                        <th>Trạng thái</th>
                        <th>Lỗi</th>
                      </tr>
                    </thead>
                    <tbody>
                      {importPreview.rows.map((row) => (
                        <tr key={row.row_number}>
                          <td>{row.row_number}</td>
                          <td>{row.login_name}</td>
                          <td>
                            <span
                              className={`status-badge ${row.status}`}
                            >
                              {row.status}
                            </span>
                          </td>
                          <td>{row.error || '—'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}

        {mode === 'list' && (
          <>
            {users.length === 0 ? (
              <p className="dashboard-status">
                {searchQuery
                  ? 'Không tìm thấy người dùng phù hợp.'
                  : 'Chưa có người dùng nào.'}
              </p>
            ) : (
              <>
              <div className="users-table-wrapper">
                <table className="users-table">
                  <caption className="visually-hidden">
                    Danh sách người dùng trong tổ chức
                  </caption>
                  <thead>
                    <tr>
                      <th>Tên đăng nhập</th>
                      <th>Tên hiển thị</th>
                      <th>Email</th>
                      <th>Vai trò</th>
                      <th>Đổi mật khẩu</th>
                      <th>Thao tác</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((user) => (
                      <tr key={user.id}>
                        <td>{user.login_name}</td>
                        <td>{user.display_name}</td>
                        <td>{user.email || '—'}</td>
                        <td>{user.roles.join(', ')}</td>
                        <td>{user.must_change_password ? 'Có' : 'Không'}</td>
                        <td>
                          <div className="row-actions">
                            <button
                              type="button"
                              onClick={() => {
                                clearMessages();
                                setSelectedUser(user);
                                setEditRoles(user.roles);
                                setMode('edit-roles');
                              }}
                            >
                              Sửa vai trò
                            </button>
                            <button
                              type="button"
                              onClick={() => {
                                clearMessages();
                                setSelectedUser(user);
                                setResetPassword('');
                                setMode('reset-password');
                              }}
                            >
                              Đặt lại mật khẩu
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {userHasMore && (
                <div className="load-more">
                  <button
                    type="button"
                    onClick={loadMoreUsers}
                    disabled={isLoadingMoreUsers}
                  >
                    {isLoadingMoreUsers ? 'Đang tải…' : 'Tải thêm'}
                  </button>
                </div>
              )}
              </>
            )}
          </>
        )}
      </section>
      )}

      {activeTab === 'audit' && (
        <section
          role="tabpanel"
          id="admin-panel-audit"
          aria-labelledby="admin-tab-audit"
        >
          <AuditLogsPanel />
        </section>
      )}
      {activeTab === 'academic' && (
        <section
          role="tabpanel"
          id="admin-panel-academic"
          aria-labelledby="admin-tab-academic"
        >
          <AcademicManagementPanel />
        </section>
      )}
    </div>
  );
}
