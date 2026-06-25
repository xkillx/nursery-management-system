# Testing Patterns

| Layer | Approach |
|---|---|
| Domain / Application | Mock repositories |
| Integration | Real PostgreSQL |
| Handler | `httptest` + Gin context |
| Repository | Real PostgreSQL via `TEST_DATABASE_URL` |

## Requirements

- Repository tests require `TEST_DATABASE_URL` pointing to a disposable test database
- Migration verification requires `VERIFY_DATABASE_URL` (a different database)
