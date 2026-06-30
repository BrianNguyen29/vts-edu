import { useEffect, useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import {
  createUser,
  getOrganization,
  listUsers,
  resetUserPassword,
  updateOrganization,
  updateUserRoles,
  type User,
} from '@/shared/api/admin';
import { ApiResponseError } from '@/shared/api/attempts';
import { PasswordPolicyHints } from '@/shared/components/password-policy-hints';
import { validatePassword } from '@/shared/lib/password-policy';

const AVAILABLE_ROLES = ['student', 'teacher', 'admin'] as const;

type ViewMode = 'list' | 'create' | 'edit-roles' | 'reset-password';

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

  const [orgName, setOrgName] = useState('');
  const [orgCode, setOrgCode] = useState('');
  const [orgLoading, setOrgLoading] = useState(true);

  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const [orgData, usersData] = await Promise.all([
          getOrganization(),
          listUsers({ q: searchQuery || undefined, limit: 50 }),
        ]);
        if (cancelled) return;
        setOrgName(orgData.name);
        setOrgCode(orgData.code);
        setUsers(usersData.data);
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
        roles: newUserRoles,
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
      await updateUserRoles(selectedUser.id, { roles: editRoles });
      setUsers((prev) =>
        prev.map((u) => (u.id === selectedUser.id ? { ...u, roles: editRoles } : u))
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

      <section className="admin-section">
        <h2>Tổ chức</h2>
        <div className="org-card">
          <div className="org-info">
            <div>
              <strong>Tên:</strong>{' '}
              {orgEditing ? (
                <form onSubmit={handleUpdateOrgName} className="inline-form">
                  <input
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

      <section className="admin-section">
        <div className="section-header">
          <h2>Ngườii dùng</h2>
          {mode === 'list' && (
            <button
              type="button"
              className="primary"
              onClick={() => {
                clearMessages();
                setMode('create');
              }}
            >
              Thêm ngườii dùng
            </button>
          )}
        </div>

        {mode === 'list' && (
          <div className="search-bar">
            <input
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

        {mode === 'list' && (
          <>
            {users.length === 0 ? (
              <p className="dashboard-status">
                {searchQuery
                  ? 'Không tìm thấy ngườii dùng phù hợp.'
                  : 'Chưa có ngườii dùng nào.'}
              </p>
            ) : (
              <div className="users-table-wrapper">
                <table className="users-table">
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
            )}
          </>
        )}
      </section>
    </div>
  );
}
