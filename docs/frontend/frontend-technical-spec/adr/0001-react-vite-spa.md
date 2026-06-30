# ADR-0001 — React SPA with Vite

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Ứng dụng LMS chủ yếu sau đăng nhập, không phụ thuộc SEO. Dự án solo cần giảm hạ tầng và không muốn vận hành Node.js server production bên cạnh Go backend.

## Decision

Dùng React 19 SPA build bằng Vite 8. Output static được Go hoặc CDN phục vụ. React Router quản lý client routing.

## Consequences

### Positive

- Một production runtime backend Go.
- Build/dev nhanh.
- Route-level code splitting.
- Triển khai static đơn giản.

### Negative

- Không SSR application pages.
- Session bootstrap phía client.
- Cần SPA fallback và runtime config.

## Rejected alternatives

- Next.js full-stack: thêm Node runtime và overlap backend Go.
- Server-rendered Go templates: khó đáp ứng builder/exam interaction.
- Micro-frontends: không phù hợp solo/MVP.
