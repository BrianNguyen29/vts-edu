import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient, getCsrfToken } from '@/shared/api/api-client';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

export function DiagnosticsPage() {
  const [health, setHealth] = useState<string>('checking…');
  const [csrfStatus, setCsrfStatus] = useState<string>('not fetched');

  useDocumentTitle('Chẩn đoán hệ thống');

  useEffect(() => {
    apiClient('/healthz')
      .then((res) => res.json())
      .then((data) => setHealth(JSON.stringify(data)))
      .catch((err) => setHealth(`error: ${err.message}`));
  }, []);

  const handleFetchCsrf = async () => {
    try {
      const token = await getCsrfToken();
      setCsrfStatus(token ? `ok: ${token.slice(0, 8)}…` : 'no token');
    } catch (err) {
      setCsrfStatus(`error: ${(err as Error).message}`);
    }
  };

  const handleTestUnsafe = async () => {
    try {
      const res = await apiClient('/attempts/demo/submit', {
        method: 'POST',
      });
      const data = await res.json();
      alert(JSON.stringify(data));
    } catch (err) {
      alert(`error: ${(err as Error).message}`);
    }
  };

  return (
    <main className="diagnostics-page">
      <h1>Chẩn đoán hệ thống</h1>
      <p>
        <Link to="/app">← Quay lại trang làm việc</Link>
      </p>
      <section className="diagnostics-section" aria-labelledby="diagnostics-health">
        <h2 id="diagnostics-health">API health</h2>
        <p>
          <strong>Health:</strong>{' '}
          <span role="status" aria-live="polite">
            {health}
          </span>
        </p>
      </section>
      <section className="diagnostics-section" aria-labelledby="diagnostics-csrf">
        <h2 id="diagnostics-csrf">CSRF / unsafe request</h2>
        <p>
          <strong>CSRF:</strong>{' '}
          <span role="status" aria-live="polite">
            {csrfStatus}
          </span>
        </p>
        <div className="button-row">
          <button type="button" onClick={handleFetchCsrf}>
            Fetch CSRF token
          </button>
          <button type="button" onClick={handleTestUnsafe}>
            Test unsafe request
          </button>
        </div>
      </section>
    </main>
  );
}
