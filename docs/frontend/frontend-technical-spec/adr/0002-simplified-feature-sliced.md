# ADR-0002 — Simplified Feature-Sliced Frontend Architecture

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Frontend có nhiều domain và role. Một thư mục `components/` hoặc tổ chức theo technical type dễ mất ownership. Full Feature-Sliced Design có thể quá nghiêm cho dự án solo.

## Decision

Dùng các layer `app/pages/widgets/features/entities/shared` với import direction một chiều, nhưng không áp dụng mọi ceremony của FSD.

## Consequences

- Domain/user action dễ định vị.
- AI agents có boundary rõ.
- Cần kiểm soát public `index.ts` và circular import.
- Một số component chỉ dùng một lần vẫn có thể ở page, không buộc thành feature.
