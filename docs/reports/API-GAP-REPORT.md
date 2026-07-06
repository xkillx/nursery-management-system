# API Design Gap Analysis

Generated against [API Design Principles](../../.agents/skills/api-design-principles/SKILL.md) checklist.

---

## Executive Summary

The API follows Clean Architecture with well-structured modules. It has solid foundations (JWT auth, role-based access, domain error mapping, request IDs, audit trail). The gaps are primarily in **documentation, pagination, versioning, rate limiting coverage, and developer experience**.

---

## 1. Resource Design

### Strengths
- Resources are nouns (`/children`, `/rooms`, `/invites`)
- Plural naming used consistently
- Clear hierarchy (`/children/:child_id/profile`, `/children/:child_id/contacts`)

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **Action endpoints use non-RESTful verbs** | Low | `POST /attendance/check-ins`, `POST /attendance/check-outs`, `POST /attendance/corrections` — acceptable for domain commands, but should be documented as RPC-style |
| **Inconsistent action patterns** | Low | Some use `POST .../actions/mark-inactive`, others use `POST .../actions/archive`, `POST .../actions/activate`, `POST .../actions/deactivate` — naming is mostly consistent but `activate` vs `reactivate` varies |
| **Deep nesting not present** | OK | Max nesting is 2 levels (`/children/:child_id/room-assignments/:assignment_id`) — good |

---

## 2. HTTP Methods

### Strengths
- GET for retrieval, POST for creation, PATCH for partial updates, PUT for full replacement, DELETE for removal — all correctly used
- Idempotent operations use appropriate methods

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No PUT for full replacement** | Low | PUT is used for upsert-style operations (`PUT /funding/children/:child_id`, `PUT /children/:child_id/contacts`) but these are full replacements, which is correct |
| **POST for idempotent actions** | Low | `POST .../actions/archive` and `POST .../actions/activate` could be idempotent — consider if retry semantics are safe |

---

## 3. Status Codes

### Strengths
- 200 for successful GET/PATCH/PUT
- 201 for POST (children, rooms, invites, attendance check-in)
- 204 for DELETE and stateless actions (invite accept, manager access deactivate)
- 400 for validation errors
- 401 for missing auth
- 403 for insufficient permissions
- 404 for missing resources
- 409 for conflict states
- 429 for rate limiting (invite accept, password reset)
- 500 for internal errors

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **422 not used** | Medium | Validation failures return 400 instead of 422 Unprocessable Entity. While 400 is acceptable, 422 is more semantically correct for validation errors on well-formed JSON |
| **Inconsistent 200 vs 201 for upserts** | Low | `POST /invites` returns 201 for new, 200 for existing — good pattern. `POST /attendance/absence-markers` does the same. But `PUT /funding/children/:child_id` returns 201/200 — this is fine |

---

## 4. Pagination

### Gaps (Critical)

| Issue | Severity | Details |
|-------|----------|---------|
| **No server-side pagination on collection endpoints** | **High** | `/children`, `/invites`, `/attendance/sessions`, `/sites/:site_id/rooms`, `/sites/:site_id/session-types`, `/sites/:site_id/session-templates`, `/funding/overview`, billing invoice lists — none accept `page`/`page_size` or return pagination metadata |
| **No cursor-based pagination** | **High** | For real-time data like attendance sessions, cursor-based pagination would prevent missed/duplicate records |
| **Frontend implements client-side "load more"** | Medium | The Angular UI uses `visibleCount` slicing (e.g., `owner-rooms.component.ts:68`) — this means the API returns all records and the client paginates, which won't scale |

---

## 5. Filtering & Sorting

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **Limited query parameter filtering** | Medium | Only a few endpoints support filtering: `/invites?status=`, `/owner/manager-access?status=`, `/owner/site-summaries?billing_month=&site_id=`. Most collection endpoints have no filter support |
| **No sort parameter** | Medium | No endpoint accepts a `sort` or `order_by` query parameter |
| **No search parameter** | Low | No full-text search on collection endpoints |
| **No sparse fieldsets** | Low | No `fields` parameter to limit response payload |

---

## 6. Versioning

### Gaps (High)

| Issue | Severity | Details |
|-------|----------|---------|
| **No API versioning strategy** | **High** | All routes are under a single base path (e.g., `/api/v1/...` or just `/api/...`). No version prefix, no `Accept` header versioning, no deprecation policy |
| **No deprecation mechanism** | **High** | When breaking changes are needed, there's no way to communicate them to consumers |

---

## 7. Error Handling

### Strengths
- Consistent `ErrorResponse` struct: `{ code, message, details?, request_id }`
- `MapDomainError()` maps domain errors to HTTP status codes with specific error codes
- Field-level validation details in some handlers
- Request ID in every error response

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **Inconsistent `writeError` signatures** | Medium | Some modules define local `writeError(c, status, code, message)` (3-arg details omitted), while the platform provides `WriteError(c, status, code, message, details)` (4-arg). The local wrappers drop `details` |
| **No timestamp in error responses** | Low | The checklist recommends `timestamp` in errors; only `/health` includes one |
| **Mixed error detail formats** | Low | Some handlers return `details` as `map[string]string{"field": ..., "message": ...}`, others return `err.Error()` raw string, others return `gin.H{"field": ...}` — should standardize |
| **Generic validation messages** | Medium | Many handlers return `"Invalid request payload."` without specifying which field failed. The `details` field is optional and inconsistently populated |

---

## 8. Authentication & Authorization

### Strengths
- JWT-based authentication with access + refresh tokens
- Role-based access control (owner, manager, practitioner, parent)
- CSRF protection on state-changing endpoints
- Refresh token rotation with revocation
- Password reset with rate limiting
- Invite system with token-based acceptance

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No API key authentication** | Low | No support for service-to-service API keys (only relevant if third-party integrations are planned) |
| **No token introspection endpoint** | Low | No `GET /auth/me` or similar endpoint to verify current session details |

---

## 9. Rate Limiting

### Strengths
- Rate limiting on invite accept (`FixedWindowLimiter(10, 15min)`)
- Rate limiting on password reset (email: 5/15min, IP: 20/15min)

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No global rate limiting** | **High** | Only 2 endpoints have rate limits. All other endpoints are unprotected |
| **No rate limit headers** | **High** | Responses don't include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers |
| **No Retry-After header on 429** | Medium | When rate limited, no `Retry-After` header is provided |
| **No per-user rate limiting** | Medium | Rate limits are IP-based only; authenticated users have no per-user limits |

---

## 10. Documentation

### Gaps (Critical)

| Issue | Severity | Details |
|-------|----------|---------|
| **No OpenAPI/Swagger specification** | **Critical** | No `openapi.yaml` or Swagger UI. API documentation exists only in code |
| **No generated API docs** | **Critical** | No interactive API documentation for frontend developers or third-party consumers |
| **No request/response examples** | **High** | No documented examples for any endpoint |
| **No authentication flow documentation** | **High** | JWT flow, CSRF handling, refresh token lifecycle — all undocumented outside code |

---

## 11. CORS

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No visible CORS configuration** | Medium | No CORS middleware found in the bootstrap code. If the Angular frontend runs on a different origin in development, this would cause issues (may be handled by a reverse proxy) |

---

## 12. Health & Monitoring

### Strengths
- `/health` endpoint with database ping
- `/metrics` endpoint (Prometheus) when enabled
- Request ID middleware on all requests
- Access log middleware with optional metrics
- Recovery middleware for panic handling

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No readiness/liveness distinction** | Low | Single `/health` endpoint — no separate readiness probe for Kubernetes-style orchestration |
| **No version endpoint** | Low | No `/version` or build info endpoint |

---

## 13. Security

### Strengths
- Input validation with binding tags
- SQL injection prevented by sqlc (parameterized queries)
- Password hashing (bcrypt assumed)
- Token hashing for storage
- CSRF protection
- Request ID for tracing

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No HTTPS enforcement middleware** | Medium | No redirect from HTTP to HTTPS in the application layer (may be handled by infrastructure) |
| **No Content-Security-Policy headers** | Low | No security headers middleware |
| **No request body size limits** | Medium | No visible `MaxBodySize` middleware — large payloads could cause memory issues |

---

## 14. Testing

### Strengths
- Integration tests for several handlers (invites, password reset, billing routes, attendance routes, etc.)
- Domain unit tests for billing, attendance, rooms
- Test harness with `BootstrapOptions` for dependency injection

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **No contract tests** | Medium | No API contract tests to ensure response shape stability |
| **Missing handler tests** | Medium | Some handlers noted as having no covering tests (absence, funding, children) |

---

## 15. Response Envelope Consistency

### Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| **Inconsistent collection response wrapping** | Medium | `/invites` returns `{"items": [...]}`, `/owner/manager-access` returns a bare array `[...]`, `/children` returns a bare array, `/funding/overview` returns a structured object. Should standardize on `{ "items": [...], "total": N }` or similar |
| **No consistent timestamp format** | Low | Some timestamps use RFC3339, others use `"2006-01-02T15:04:05Z"` — functionally equivalent but should be documented |

---

## Prioritised Action Plan

### P0 — Critical (blocks developer adoption)
1. **Generate OpenAPI specification** — add `swag` or `oapi-codegen` to produce `openapi.yaml` from handler annotations or struct tags
2. **Add pagination to all collection endpoints** — implement offset-based pagination with `page`, `page_size`, `total` metadata

### P1 — High (production readiness)
3. **Add API versioning** — prefix all routes with `/api/v1/`
4. **Add global rate limiting** — apply `FixedWindowLimiter` or token-bucket to all endpoints, not just 2
5. **Include rate limit headers** in responses
6. **Standardize error `details` format** — always use `map[string]string{"field": ..., "message": ...}` for validation errors

### P2 — Medium (quality & consistency)
7. **Standardize collection response envelope** — `{ "items": [...], "total": N, "page": N, "page_size": N }`
8. **Add filtering/sorting** to key collection endpoints (`/children`, `/invites`, invoice lists)
9. **Add CORS middleware** if not handled by reverse proxy
10. **Add request body size limits**
11. **Standardize local `writeError` wrappers** to pass `details`

### P3 — Low (polish)
12. Add `timestamp` to error responses
13. Add `/version` endpoint
14. Add security headers middleware
15. Add contract tests for response shape stability
