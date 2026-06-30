# VTS EDU Web

Minimal Vite + React 19 scaffold for the MVP demo.

## Local development

```bash
cd apps/web
cp .env.example .env
# Edit .env with local API URL, then:
pnpm install
pnpm dev
```

## Build

```bash
pnpm build
```

## Notes

- `apiClient` fetches CSRF token before unsafe requests and sends `X-CSRF-Token` header.
- `runtime-config.ts` falls back to `import.meta.env.VITE_API_BASE_URL` if `/app-config.json` is unavailable.
