# Cloudflare Container Workers Design

**Date:** 2026-09-01  
**Author:** Claude Sonnet 4.5  
**Status:** Design Approved  
**Approach:** Unified Abstraction - Single Entrypoint

---

## Overview

This design adds Cloudflare Container Workers deployment support to mycart (Go Fiber v3 e-commerce backend) while maintaining full compatibility with the existing local CLI server. The solution uses a unified abstraction layer that seamlessly switches between local backends (SQLite, filesystem) and Cloudflare backends (D1, R2) based on environment detection.

### Goals

1. **Dual Deployment** - Same binary runs locally (CLI) or on Cloudflare Container Workers
2. **Clean Architecture** - Single codebase with proper abstractions (no duplication)
3. **Auto-Detection** - Automatically uses correct backends based on environment
4. **Zero Breaking Changes** - Existing local CLI server continues to work unchanged
5. **TDD Approach** - Test-driven development for all new abstractions

### Key Decisions

| Component | Local Mode | Cloudflare Mode | Strategy |
|-----------|------------|-----------------|----------|
| **Database** | SQLite | Cloudflare D1 | Abstraction interface with auto-detection |
| **Storage** | Filesystem (`./lc_uploads`) | Cloudflare R2 | Abstraction interface with auto-detection |
| **Static Assets** | Embedded in binary | Workers Assets (CDN) | Separate deployment for Cloudflare |
| **Email** | SMTP | SMTP (unchanged) | Works from container |
| **Migrations** | Existing SQLite migrations | Compatible SQL subset | Single migration files |
| **Configuration** | Auto-detect + env var override | Auto-detect + env var override | Hybrid approach |

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     mycart Binary                        │
│  ┌────────────────────────────────────────────────────┐ │
│  │           cmd/main.go (single entrypoint)          │ │
│  │  - Detects environment (CLI vs Cloudflare)         │ │
│  │  - Initializes appropriate backends                │ │
│  └────────────────────────────────────────────────────┘ │
│                          │                               │
│                          ▼                               │
│  ┌────────────────────────────────────────────────────┐ │
│  │         internal/app.go (Fiber v3 server)          │ │
│  │  - Uses database.Database interface                │ │
│  │  - Uses storage.Storage interface                  │ │
│  └────────────────────────────────────────────────────┘ │
│           │                              │               │
│           ▼                              ▼               │
│  ┌──────────────────┐         ┌──────────────────┐     │
│  │ Database Layer   │         │ Storage Layer    │     │
│  │ ┌──────┐┌──────┐│         │ ┌──────┐┌──────┐ │     │
│  │ │SQLite││  D1  ││         │ │ File ││  R2  │ │     │
│  │ └──────┘└──────┘│         │ └──────┘└──────┘ │     │
│  └──────────────────┘         └──────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

For complete architectural details, interface designs, implementation phases, and testing strategy, see the full document.

---

## Quick Reference

**Branch:** `feature/cloudflare-container-workers`  
**Merge Target:** `main_dure`  
**Approach:** Unified Abstraction (Approach 2)  
**TDD:** Required for all new abstractions  
**Coverage:** Minimum 80%  

**Implementation Order:**
1. Database abstraction (pkg/database)
2. Storage abstraction (pkg/storage)
3. Config & detection (internal/config)
4. Application integration
5. Cloudflare infrastructure
6. Testing & validation
7. Documentation

---

## References

- [Cloudflare Container Workers](https://developers.cloudflare.com/workers/runtime-apis/bindings/service-bindings/)
- [Cloudflare D1 Database](https://developers.cloudflare.com/d1/)
- [Cloudflare R2 Storage](https://developers.cloudflare.com/r2/)
- [Go Fiber v3](https://docs.gofiber.io/)
