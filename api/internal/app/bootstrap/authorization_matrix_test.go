package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "nursery-management-system/api/internal/modules/authentication/domain"
	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/dbtest"
	"nursery-management-system/api/internal/platform/email"
)

// ---------------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------------

type authzHarness struct {
	router *gin.Engine
	pool   *pgxpool.Pool
	tokens *authtokens.TokenManager
	logger *slog.Logger
	cfg    config.Config

	// Scope A
	tenantA uuid.UUID
	branchA uuid.UUID

	managerAUID   uuid.UUID
	managerAMID   uuid.UUID
	managerAToken string

	practitionerAUID   uuid.UUID
	practitionerAMID   uuid.UUID
	practitionerAToken string

	parentAUID   uuid.UUID
	parentAMID   uuid.UUID
	parentAToken string

	ownerAUID   uuid.UUID
	ownerAMID   uuid.UUID
	ownerAToken string

	// Scope B
	tenantB uuid.UUID
	branchB uuid.UUID

	managerBUID   uuid.UUID
	managerBMID   uuid.UUID
	managerBToken string

	parentBUID   uuid.UUID
	parentBMID   uuid.UUID
	parentBToken string

	// Seeded scope A entities
	childA         uuid.UUID
	inactiveChildA uuid.UUID
	guardianA      uuid.UUID
	linkA          uuid.UUID
	mappingA       uuid.UUID
	inviteA        uuid.UUID
	absenceA       uuid.UUID
	childAFunding  uuid.UUID
	invoiceARun    uuid.UUID
	invoiceA       uuid.UUID
	draftInvoiceA  uuid.UUID

	// Seeded scope B entities
	childB    uuid.UUID
	guardianB uuid.UUID
	linkB     uuid.UUID
	mappingB  uuid.UUID
	invoiceB  uuid.UUID
}

func setupAuthzHarness(t *testing.T) *authzHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &authzHarness{
		pool:   pool,
		tokens: authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720, 24),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),

		// Scope A — deterministic UUIDs in "a1" block
		tenantA: uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		branchA: uuid.MustParse("a1000000-0000-0000-0000-000000000002"),

		managerAUID: uuid.MustParse("a1000000-0000-0000-0000-000000000010"),
		managerAMID: uuid.MustParse("a1000000-0000-0000-0000-000000000011"),

		practitionerAUID: uuid.MustParse("a1000000-0000-0000-0000-000000000020"),
		practitionerAMID: uuid.MustParse("a1000000-0000-0000-0000-000000000021"),

		parentAUID: uuid.MustParse("a1000000-0000-0000-0000-000000000030"),
		parentAMID: uuid.MustParse("a1000000-0000-0000-0000-000000000031"),

		ownerAUID: uuid.MustParse("a1000000-0000-0000-0000-000000000040"),
		ownerAMID: uuid.MustParse("a1000000-0000-0000-0000-000000000041"),

		// Scope B — deterministic UUIDs in "b1" block
		tenantB: uuid.MustParse("b1000000-0000-0000-0000-000000000001"),
		branchB: uuid.MustParse("b1000000-0000-0000-0000-000000000002"),

		managerBUID: uuid.MustParse("b1000000-0000-0000-0000-000000000010"),
		managerBMID: uuid.MustParse("b1000000-0000-0000-0000-000000000011"),

		parentBUID: uuid.MustParse("b1000000-0000-0000-0000-000000000030"),
		parentBMID: uuid.MustParse("b1000000-0000-0000-0000-000000000031"),

		// Scope A entities
		childA:         uuid.MustParse("a1000000-0000-0000-0000-000000000100"),
		inactiveChildA: uuid.MustParse("a1000000-0000-0000-0000-000000000101"),
		guardianA:      uuid.MustParse("a1000000-0000-0000-0000-000000000200"),
		linkA:          uuid.MustParse("a1000000-0000-0000-0000-000000000300"),
		mappingA:       uuid.MustParse("a1000000-0000-0000-0000-000000000400"),
		inviteA:        uuid.MustParse("a1000000-0000-0000-0000-000000000500"),
		absenceA:       uuid.MustParse("a1000000-0000-0000-0000-000000000600"),
		childAFunding:  uuid.MustParse("a1000000-0000-0000-0000-000000000700"),
		invoiceARun:    uuid.MustParse("a1000000-0000-0000-0000-000000000800"),
		invoiceA:       uuid.MustParse("a1000000-0000-0000-0000-000000000801"),
		draftInvoiceA:  uuid.MustParse("a1000000-0000-0000-0000-000000000802"),

		// Scope B entities
		childB:    uuid.MustParse("b1000000-0000-0000-0000-000000000100"),
		guardianB: uuid.MustParse("b1000000-0000-0000-0000-000000000200"),
		linkB:     uuid.MustParse("b1000000-0000-0000-0000-000000000300"),
		mappingB:  uuid.MustParse("b1000000-0000-0000-0000-000000000400"),
		invoiceB:  uuid.MustParse("b1000000-0000-0000-0000-000000000801"),
	}

	h.cfg = testConfig()
	h.cfg.StripeSecretKey = ""

	h.router = BootstrapWithOptions(h.cfg, h.logger, pool, BootstrapOptions{
		EmailSender:      email.NewFakeSender(),
		CheckoutProvider: &fakeCheckoutProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://test", PaymentIntentID: "pi_test"}},
	})

	h.managerAToken = h.accessToken(h.managerAUID, h.managerAMID, h.tenantA, h.branchA, "manager")
	h.practitionerAToken = h.accessToken(h.practitionerAUID, h.practitionerAMID, h.tenantA, h.branchA, "practitioner")
	h.parentAToken = h.accessToken(h.parentAUID, h.parentAMID, h.tenantA, h.branchA, "parent")
	h.ownerAToken = h.accessToken(h.ownerAUID, h.ownerAMID, h.tenantA, uuid.Nil, "owner")
	h.managerBToken = h.accessToken(h.managerBUID, h.managerBMID, h.tenantB, h.branchB, "manager")
	h.parentBToken = h.accessToken(h.parentBUID, h.parentBMID, h.tenantB, h.branchB, "parent")

	h.seed(t)
	return h
}

func (h *authzHarness) accessToken(userID, membershipID, tenantID, branchID uuid.UUID, role string) string {
	branchIDStr := branchID.String()
	if role == "owner" {
		branchIDStr = ""
	}
	raw, _, err := h.tokens.GenerateAccessToken(userID, "test@example.com", authdomain.ScopeClaims{
		MembershipID: membershipID.String(),
		TenantID:     tenantID.String(),
		BranchID:     branchIDStr,
		Role:         role,
	})
	if err != nil {
		panic(fmt.Sprintf("generate access token: %v", err))
	}
	return raw
}

func (h *authzHarness) seed(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Scope A
	dbtest.InsertTenant(t, h.pool, h.tenantA, "Authz A")
	dbtest.InsertBranch(t, h.pool, h.tenantA, h.branchA, "Branch A")
	dbtest.InsertUser(t, h.pool, h.managerAUID, "authz-mgr-a@test.com", "hash", true)
	dbtest.InsertUser(t, h.pool, h.practitionerAUID, "authz-prac-a@test.com", "hash", true)
	dbtest.InsertUser(t, h.pool, h.parentAUID, "authz-parent-a@test.com", "hash", true)
	dbtest.InsertUser(t, h.pool, h.ownerAUID, "authz-owner-a@test.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, h.managerAMID, h.tenantA, h.branchA, h.managerAUID, "manager", true)
	dbtest.InsertMembership(t, h.pool, h.practitionerAMID, h.tenantA, h.branchA, h.practitionerAUID, "practitioner", true)
	dbtest.InsertMembership(t, h.pool, h.parentAMID, h.tenantA, h.branchA, h.parentAUID, "parent", true)
	// Owner membership (NULL branch_id)
	_, err := h.pool.Exec(ctx, "INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active) VALUES ($1, $2, NULL, $3, 'owner', true)",
		h.ownerAMID, h.tenantA, h.ownerAUID)
	if err != nil {
		t.Fatalf("insert owner membership: %v", err)
	}

	dbtest.InsertChild(t, h.pool, h.childA, h.tenantA, h.branchA, "Child A",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertChild(t, h.pool, h.inactiveChildA, h.tenantA, h.branchA, "Inactive Child A",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	_, err = h.pool.Exec(ctx, "UPDATE children SET is_active = false, updated_at = now() WHERE id = $1", h.inactiveChildA)
	if err != nil {
		t.Fatalf("deactivate child: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO child_leaving_records (id, tenant_id, branch_id, child_id, left_at, reason_code, reason_note)
		 VALUES (gen_random_uuid(), $1, $2, $3, now(), 'left_nursery', 'moved')`,
		h.tenantA, h.branchA, h.inactiveChildA)
	if err != nil {
		t.Fatalf("seed leaving record: %v", err)
	}
	dbtest.InsertGuardian(t, h.pool, h.guardianA, h.tenantA, h.branchA, "Guardian A", true)
	dbtest.InsertGuardianLink(t, h.pool, h.linkA, h.tenantA, h.branchA, h.guardianA, h.childA)
	dbtest.InsertParentMapping(t, h.pool, h.mappingA, h.tenantA, h.branchA, h.parentAMID, h.guardianA)

	// Invite A
	dbtest.InsertInvite(t, h.pool, h.inviteA, h.tenantA, h.branchA,
		"invite-a@test.com", "invite-a@test.com", "practitioner",
		"hash_aaa", time.Now().Add(7*24*time.Hour), h.managerAUID, h.managerAMID)

	// Absence marker A
	dbtest.InsertAbsenceMarker(t, h.pool, h.absenceA, h.tenantA, h.branchA, h.childA,
		h.managerAUID, h.managerAMID,
		dbtest.DateAt(2026, 6, 1), dbtest.TimestampAt(2026, 6, 1, 8, 0),
		nil, nil, nil)

	// Funding child A (same child used for funding)
	h.childAFunding = h.childA

	// Invoice run + issued invoice A
	_, err = h.pool.Exec(ctx, `INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		VALUES ($1, $2, $3, $4, 'issue', 'completed', now(), now(), $5, $6, 'authz-req')`,
		h.invoiceARun, h.tenantA, h.branchA, dbtest.DateAt(2026, 5, 1), h.managerAUID, h.managerAMID)
	if err != nil {
		t.Fatalf("insert invoice run A: %v", err)
	}
	_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','issued','INV-AZ-001',1,$6,$6,now(),$7,$8,now(),now() + interval '7 days',
		'GBP',5000,0,5000,0,$9,$10,'{}')`,
		h.invoiceA, h.tenantA, h.branchA, h.childA, dbtest.DateAt(2026, 5, 1),
		h.invoiceARun, h.managerAUID, h.managerAMID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert invoice A: %v", err)
	}

	// Draft invoice A (status = draft, not issued — no invoice_number/issued fields; different billing month to avoid unique constraint)
	_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		generated_run_id,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','draft',$6,
		'GBP',5000,0,5000,0,$7,$8,'{}')`,
		h.draftInvoiceA, h.tenantA, h.branchA, h.childA, dbtest.DateAt(2026, 6, 1),
		h.invoiceARun, dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 30))
	if err != nil {
		t.Fatalf("insert draft invoice A: %v", err)
	}

	// Scope B
	dbtest.InsertTenant(t, h.pool, h.tenantB, "Authz B")
	dbtest.InsertBranch(t, h.pool, h.tenantB, h.branchB, "Branch B")
	dbtest.InsertUser(t, h.pool, h.managerBUID, "authz-mgr-b@test.com", "hash", true)
	dbtest.InsertUser(t, h.pool, h.parentBUID, "authz-parent-b@test.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, h.managerBMID, h.tenantB, h.branchB, h.managerBUID, "manager", true)
	dbtest.InsertMembership(t, h.pool, h.parentBMID, h.tenantB, h.branchB, h.parentBUID, "parent", true)

	dbtest.InsertChild(t, h.pool, h.childB, h.tenantB, h.branchB, "Child B",
		dbtest.DateAt(2023, 2, 1), dbtest.DateAt(2026, 2, 1), true)
	dbtest.InsertGuardian(t, h.pool, h.guardianB, h.tenantB, h.branchB, "Guardian B", true)
	dbtest.InsertGuardianLink(t, h.pool, h.linkB, h.tenantB, h.branchB, h.guardianB, h.childB)
	dbtest.InsertParentMapping(t, h.pool, h.mappingB, h.tenantB, h.branchB, h.parentBMID, h.guardianB)

	// Invoice run + invoice B
	runB := uuid.MustParse("b1000000-0000-0000-0000-000000000800")
	_, err = h.pool.Exec(ctx, `INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		VALUES ($1, $2, $3, $4, 'issue', 'completed', now(), now(), $5, $6, 'authz-req-b')`,
		runB, h.tenantB, h.branchB, dbtest.DateAt(2026, 5, 1), h.managerBUID, h.managerBMID)
	if err != nil {
		t.Fatalf("insert invoice run B: %v", err)
	}
	_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','issued','INV-AZ-B01',1,$6,$6,now(),$7,$8,now(),now() + interval '7 days',
		'GBP',6000,0,6000,0,$9,$10,'{}')`,
		h.invoiceB, h.tenantB, h.branchB, h.childB, dbtest.DateAt(2026, 5, 1),
		runB, h.managerBUID, h.managerBMID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert invoice B: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *authzHarness) get(t *testing.T, path, token string) *httptest.ResponseRecorder {
	return h.doReq(t, http.MethodGet, path, token, "")
}

func (h *authzHarness) post(t *testing.T, path, token, body string) *httptest.ResponseRecorder {
	return h.doReq(t, http.MethodPost, path, token, body)
}

func (h *authzHarness) patch(t *testing.T, path, token, body string) *httptest.ResponseRecorder {
	return h.doReq(t, http.MethodPatch, path, token, body)
}

func (h *authzHarness) put(t *testing.T, path, token, body string) *httptest.ResponseRecorder {
	return h.doReq(t, http.MethodPut, path, token, body)
}

func (h *authzHarness) doReq(t *testing.T, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("expected status %d, got %d", want, w.Code)
	}
}

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()
	var resp struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != want {
		t.Fatalf("expected error code %q, got %q", want, resp.Code)
	}
}

func assertStatusAndCode(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	assertStatus(t, w, status)
	assertErrorCode(t, w, code)
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func extractIDs(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var resp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeBody(t, w, &resp)
	ids := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		ids[i] = item.ID
	}
	return ids
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Route classification
// ---------------------------------------------------------------------------

type routeClass string

const (
	classPublic            routeClass = "public"
	classProtectedBusiness routeClass = "protected_business"
)

type routeEntry struct {
	Method       string
	Path         string // exact Gin path pattern
	Class        routeClass
	AllowedRoles []string // empty = public
}

func allRouteEntries(h *authzHarness) []routeEntry {
	return []routeEntry{
		// Public
		{"GET", "/health", classPublic, nil},
		{"GET", "/api/v1/health", classPublic, nil},
		{"POST", "/api/v1/auth/login", classPublic, nil},
		{"POST", "/api/v1/auth/refresh", classPublic, nil},
		{"POST", "/api/v1/auth/logout", classPublic, nil},
		{"POST", "/api/v1/auth/switch-membership", classPublic, nil},
		{"POST", "/api/v1/auth/password-reset-requests", classPublic, nil},
		{"POST", "/api/v1/auth/password-resets", classPublic, nil},
		{"POST", "/api/v1/invites/accept", classPublic, nil},
		{"POST", "/api/v1/stripe/webhooks", classPublic, nil},

		// Children (manager)
		{"GET", "/api/v1/children", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/children", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/children/:child_id/actions/mark-inactive", classProtectedBusiness, []string{"manager"}},

		// Children attendance (manager + practitioner)
		{"GET", "/api/v1/children/attendance", classProtectedBusiness, []string{"manager", "practitioner"}},

		// Guardians (manager)
		{"GET", "/api/v1/guardians", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/guardians/:guardian_id", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/guardians", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/guardians/:guardian_id", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/guardians/:guardian_id/actions/deactivate", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/guardians/:guardian_id/actions/reactivate", classProtectedBusiness, []string{"manager"}},

		// Links (manager)
		{"GET", "/api/v1/children/:child_id/guardian-child-links", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/guardian-child-links", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/guardian-child-links/:link_id/actions/end", classProtectedBusiness, []string{"manager"}},

		// Mappings (manager)
		{"POST", "/api/v1/parent-membership-guardian-mappings", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/parent-membership-guardian-mappings/:mapping_id/actions/end", classProtectedBusiness, []string{"manager"}},

		// Attendance (manager + practitioner)
		{"POST", "/api/v1/attendance/check-ins", classProtectedBusiness, []string{"manager", "practitioner"}},
		{"POST", "/api/v1/attendance/check-outs", classProtectedBusiness, []string{"manager", "practitioner"}},

		// Attendance corrections (manager)
		{"POST", "/api/v1/attendance/corrections", classProtectedBusiness, []string{"manager"}},

		// Absence markers (manager + practitioner)
		{"POST", "/api/v1/attendance/absence-markers", classProtectedBusiness, []string{"manager", "practitioner"}},
		{"POST", "/api/v1/attendance/absence-markers/:absence_marker_id/clear", classProtectedBusiness, []string{"manager", "practitioner"}},

		// Invites (manager)
		{"POST", "/api/v1/invites", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/invites", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/invites/:invite_id/resend", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/invites/:invite_id/revoke", classProtectedBusiness, []string{"manager"}},

		// Funding (manager)
		{"GET", "/api/v1/funding/children/:child_id", classProtectedBusiness, []string{"manager"}},
		{"PUT", "/api/v1/funding/children/:child_id", classProtectedBusiness, []string{"manager"}},

		// Billing (manager)
		{"GET", "/api/v1/invoices/drafts/preflight", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/invoice-runs/drafts", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/invoices", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/invoices/:invoice_id", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/invoices/:invoice_id/issue", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/invoices/bulk-issue", classProtectedBusiness, []string{"manager"}},

		// Parent billing
		{"GET", "/api/v1/parent/invoices", classProtectedBusiness, []string{"parent"}},
		{"GET", "/api/v1/parent/invoices/:invoice_id", classProtectedBusiness, []string{"parent"}},
		{"POST", "/api/v1/parent/invoices/:invoice_id/checkout-sessions", classProtectedBusiness, []string{"parent"}},

		// Manager payment diagnostics
		{"GET", "/api/v1/invoices/:invoice_id/payment-status", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/invoices/:invoice_id/payment-events", classProtectedBusiness, []string{"manager"}},

		// Child Management sub-resources (manager)
		{"GET", "/api/v1/children/:child_id/profile", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id/profile", classProtectedBusiness, []string{"manager"}},
		{"PUT", "/api/v1/children/:child_id/collection-settings", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/health", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id/health", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/safeguarding", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id/safeguarding", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/consent", classProtectedBusiness, []string{"manager"}},
		{"PUT", "/api/v1/children/:child_id/consent", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/contacts", classProtectedBusiness, []string{"manager"}},
		{"PUT", "/api/v1/children/:child_id/contacts", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/funding", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id/funding", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/room-assignments", classProtectedBusiness, []string{"manager"}},
		{"POST", "/api/v1/children/:child_id/room-assignments", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/billing-profile", classProtectedBusiness, []string{"manager"}},
		{"PATCH", "/api/v1/children/:child_id/billing-profile", classProtectedBusiness, []string{"manager"}},
		{"GET", "/api/v1/children/:child_id/leaving-record", classProtectedBusiness, []string{"manager"}},

		// Owner
		{"GET", "/api/v1/owner/site-summaries", classProtectedBusiness, []string{"owner"}},
		{"GET", "/api/v1/owner/manager-access", classProtectedBusiness, []string{"owner"}},
		{"POST", "/api/v1/owner/sites/:site_id/manager-access", classProtectedBusiness, []string{"owner"}},
		{"POST", "/api/v1/owner/sites/:site_id/manager-access/:membership_id/actions/deactivate", classProtectedBusiness, []string{"owner"}},
		{"POST", "/api/v1/owner/sites/:site_id/manager-access/:membership_id/actions/activate", classProtectedBusiness, []string{"owner"}},
	}
}

// ---------------------------------------------------------------------------
// Task 3: Completeness guard
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixRouteClassificationIsComplete(t *testing.T) {
	h := setupAuthzHarness(t)

	registered := make(map[string]struct{})
	for _, r := range h.router.Routes() {
		registered[r.Method+" "+r.Path] = struct{}{}
	}

	classified := make(map[string]struct{})
	for _, e := range allRouteEntries(h) {
		key := e.Method + " " + e.Path
		classified[key] = struct{}{}
	}

	// Check all /api/v1 routes are classified
	for key := range registered {
		if !strings.HasPrefix(key, "/api/v1/") {
			continue
		}
		if _, ok := classified[key]; !ok {
			t.Errorf("unclassified /api/v1 route: %s", key)
		}
	}

	// Check no stale entries
	for key := range classified {
		if _, ok := registered[key]; !ok {
			t.Errorf("stale matrix entry not registered: %s", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Task 4: Public routes
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixPublicRoutesAreIntentional(t *testing.T) {
	h := setupAuthzHarness(t)

	t.Run("GET /health", func(t *testing.T) {
		w := h.get(t, "/health", "")
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("GET /api/v1/health", func(t *testing.T) {
		w := h.get(t, "/api/v1/health", "")
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("POST /api/v1/auth/login without body returns validation error", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/login", "", "{}")
		assertStatus(t, w, http.StatusBadRequest)
		assertErrorCode(t, w, "validation_error")
	})

	t.Run("POST /api/v1/auth/logout without cookie returns 204", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/logout", "", "")
		assertStatus(t, w, http.StatusNoContent)
	})

	t.Run("POST /api/v1/auth/refresh without cookie returns 401", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/refresh", "", "")
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/auth/switch-membership without cookie returns 401", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/switch-membership", "", `{"membership_id":"`+uuid.NewString()+`"}`)
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/auth/password-reset-requests unknown email returns 202", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/password-reset-requests", "", `{"email":"unknown@test.com"}`)
		assertStatus(t, w, http.StatusAccepted)
	})

	t.Run("POST /api/v1/auth/password-resets invalid token returns 400", func(t *testing.T) {
		w := h.post(t, "/api/v1/auth/password-resets", "", `{"token":"invalid","new_password":"Password123"}`)
		assertStatus(t, w, http.StatusBadRequest)
		assertErrorCode(t, w, "password_reset_token_invalid")
	})

	t.Run("POST /api/v1/invites/accept invalid token returns 422", func(t *testing.T) {
		w := h.post(t, "/api/v1/invites/accept", "", `{"token":"invalid","new_password":"Password123"}`)
		assertStatus(t, w, http.StatusUnprocessableEntity)
		assertErrorCode(t, w, "invite_token_invalid")
	})

	t.Run("POST /api/v1/stripe/webhooks without verifier returns 503", func(t *testing.T) {
		// Our harness uses empty StripeSecretKey and no WebhookVerifier → unconfigured
		w := h.post(t, "/api/v1/stripe/webhooks", "", "{}")
		assertStatus(t, w, http.StatusServiceUnavailable)
		assertErrorCode(t, w, "payment_provider_unconfigured")
	})
}

// ---------------------------------------------------------------------------
// Task 5: Protected route auth + role matrix
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixProtectedRoutesRequireAuthentication(t *testing.T) {
	h := setupAuthzHarness(t)

	protectedRoutes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		// Children
		{"list children", "GET", "/api/v1/children", ""},
		{"get child", "GET", "/api/v1/children/" + h.childA.String(), ""},
		{"create child", "POST", "/api/v1/children", `{"first_name":"X","date_of_birth":"2021-01-01","start_date":"2026-01-01"}`},
		{"update child", "PATCH", "/api/v1/children/" + h.childA.String(), `{"first_name":"Y"}`},
		{"mark inactive", "POST", "/api/v1/children/" + h.childA.String() + "/actions/mark-inactive", `{"reason_code":"other"}`},

		// Children attendance list
		{"attendance list", "GET", "/api/v1/children/attendance", ""},

		// Guardians
		{"list guardians", "GET", "/api/v1/guardians", ""},
		{"get guardian", "GET", "/api/v1/guardians/" + h.guardianA.String(), ""},
		{"create guardian", "POST", "/api/v1/guardians", `{"full_name":"G"}`},
		{"update guardian", "PATCH", "/api/v1/guardians/" + h.guardianA.String(), `{"full_name":"G2"}`},
		{"deactivate guardian", "POST", "/api/v1/guardians/" + h.guardianA.String() + "/actions/deactivate", `{"reason_code":"other"}`},
		{"reactivate guardian", "POST", "/api/v1/guardians/" + h.guardianA.String() + "/actions/reactivate", ""},

		// Links
		{"create link", "POST", "/api/v1/guardian-child-links", fmt.Sprintf(`{"guardian_id":"%s","child_id":"%s"}`, h.guardianA, h.childA)},
		{"end link", "POST", "/api/v1/guardian-child-links/" + h.linkA.String() + "/actions/end", `{"reason_code":"other"}`},

		// Mappings
		{"create mapping", "POST", "/api/v1/parent-membership-guardian-mappings", fmt.Sprintf(`{"membership_id":"%s","guardian_id":"%s"}`, h.parentAMID, h.guardianA)},
		{"end mapping", "POST", "/api/v1/parent-membership-guardian-mappings/" + h.mappingA.String() + "/actions/end", `{"reason_code":"other"}`},

		// Attendance
		{"check-in", "POST", "/api/v1/attendance/check-ins", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"check-out", "POST", "/api/v1/attendance/check-outs", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"correction", "POST", "/api/v1/attendance/corrections", fmt.Sprintf(`{"child_id":"%s","check_in_at":"2026-06-01T08:00:00Z","check_out_at":"2026-06-01T16:00:00Z"}`, h.childA)},

		// Absence
		{"mark absent", "POST", "/api/v1/attendance/absence-markers", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"clear marker", "POST", "/api/v1/attendance/absence-markers/" + h.absenceA.String() + "/clear", ""},

		// Invites
		{"create invite", "POST", "/api/v1/invites", `{"email":"new@test.com","role":"practitioner"}`},
		{"list invites", "GET", "/api/v1/invites", ""},
		{"resend invite", "POST", "/api/v1/invites/" + h.inviteA.String() + "/resend", ""},
		{"revoke invite", "POST", "/api/v1/invites/" + h.inviteA.String() + "/revoke", ""},

		// Funding
		{"get funding", "GET", "/api/v1/funding/children/" + h.childA.String() + "?billing_month=2026-05", ""},
		{"upsert funding", "PUT", "/api/v1/funding/children/" + h.childA.String(), `{"billing_month":"2026-05"}`},

		// Billing
		{"preflight", "GET", "/api/v1/invoices/drafts/preflight?billing_month=2026-05", ""},
		{"generate drafts", "POST", "/api/v1/invoice-runs/drafts", `{"billing_month":"2026-05"}`},
		{"list invoices", "GET", "/api/v1/invoices", ""},
		{"get invoice", "GET", "/api/v1/invoices/" + h.invoiceA.String(), ""},
		{"issue invoice", "POST", "/api/v1/invoices/" + h.invoiceA.String() + "/issue", `{"confirm":true}`},
		{"bulk issue", "POST", "/api/v1/invoices/bulk-issue", fmt.Sprintf(`{"invoice_ids":["%s"]}`, h.invoiceA)},

		// Parent billing
		{"parent list invoices", "GET", "/api/v1/parent/invoices", ""},
		{"parent get invoice", "GET", "/api/v1/parent/invoices/" + h.invoiceA.String(), ""},
		{"parent checkout", "POST", "/api/v1/parent/invoices/" + h.invoiceA.String() + "/checkout-sessions", ""},

		// Child Management sub-resources
		{"get child profile", "GET", "/api/v1/children/" + h.childA.String() + "/profile", ""},
		{"patch child profile", "PATCH", "/api/v1/children/" + h.childA.String() + "/profile", `{"disability_status":"unknown"}`},
		{"set collection settings", "PUT", "/api/v1/children/" + h.childA.String() + "/collection-settings", `{"over_18_collection_acknowledged":true}`},

		// Payment diagnostics
		{"payment status", "GET", "/api/v1/invoices/" + h.invoiceA.String() + "/payment-status", ""},
		{"payment events", "GET", "/api/v1/invoices/" + h.invoiceA.String() + "/payment-events", ""},
	}

	for _, tc := range protectedRoutes {
		t.Run(tc.name, func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, "", tc.body)
			assertStatusAndCode(t, w, http.StatusUnauthorized, "unauthorized")
		})
	}
}

func TestAuthorizationMatrixProtectedRoutesRejectWrongRoles(t *testing.T) {
	h := setupAuthzHarness(t)

	// Cases where route allows manager only → practitioner and parent forbidden
	managerOnlyRoutes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list children", "GET", "/api/v1/children", ""},
		{"get child", "GET", "/api/v1/children/" + h.childA.String(), ""},
		{"create child", "POST", "/api/v1/children", `{"first_name":"X","date_of_birth":"2021-01-01","start_date":"2026-01-01"}`},
		{"update child", "PATCH", "/api/v1/children/" + h.childA.String(), `{"first_name":"Y"}`},
		{"mark inactive", "POST", "/api/v1/children/" + h.childA.String() + "/actions/mark-inactive", `{"reason_code":"other"}`},

		{"list guardians", "GET", "/api/v1/guardians", ""},
		{"get guardian", "GET", "/api/v1/guardians/" + h.guardianA.String(), ""},
		{"create guardian", "POST", "/api/v1/guardians", `{"full_name":"G"}`},
		{"update guardian", "PATCH", "/api/v1/guardians/" + h.guardianA.String(), `{"full_name":"G2"}`},
		{"deactivate guardian", "POST", "/api/v1/guardians/" + h.guardianA.String() + "/actions/deactivate", `{"reason_code":"other"}`},
		{"reactivate guardian", "POST", "/api/v1/guardians/" + h.guardianA.String() + "/actions/reactivate", ""},

		{"create link", "POST", "/api/v1/guardian-child-links", fmt.Sprintf(`{"guardian_id":"%s","child_id":"%s"}`, h.guardianA, h.childA)},
		{"end link", "POST", "/api/v1/guardian-child-links/" + h.linkA.String() + "/actions/end", `{"reason_code":"other"}`},
		{"create mapping", "POST", "/api/v1/parent-membership-guardian-mappings", fmt.Sprintf(`{"membership_id":"%s","guardian_id":"%s"}`, h.parentAMID, h.guardianA)},
		{"end mapping", "POST", "/api/v1/parent-membership-guardian-mappings/" + h.mappingA.String() + "/actions/end", `{"reason_code":"other"}`},

		{"correction", "POST", "/api/v1/attendance/corrections", fmt.Sprintf(`{"child_id":"%s","check_in_at":"2026-06-01T08:00:00Z","check_out_at":"2026-06-01T16:00:00Z"}`, h.childA)},

		{"create invite", "POST", "/api/v1/invites", `{"email":"new@test.com","role":"practitioner"}`},
		{"list invites", "GET", "/api/v1/invites", ""},
		{"resend invite", "POST", "/api/v1/invites/" + h.inviteA.String() + "/resend", ""},
		{"revoke invite", "POST", "/api/v1/invites/" + h.inviteA.String() + "/revoke", ""},

		{"get funding", "GET", "/api/v1/funding/children/" + h.childA.String() + "?billing_month=2026-05", ""},
		{"upsert funding", "PUT", "/api/v1/funding/children/" + h.childA.String(), `{"billing_month":"2026-05"}`},

		{"preflight", "GET", "/api/v1/invoices/drafts/preflight?billing_month=2026-05", ""},
		{"generate drafts", "POST", "/api/v1/invoice-runs/drafts", `{"billing_month":"2026-05"}`},
		{"list invoices", "GET", "/api/v1/invoices", ""},
		{"get invoice", "GET", "/api/v1/invoices/" + h.invoiceA.String(), ""},
		{"issue invoice", "POST", "/api/v1/invoices/" + h.invoiceA.String() + "/issue", `{"confirm":true}`},
		{"bulk issue", "POST", "/api/v1/invoices/bulk-issue", fmt.Sprintf(`{"invoice_ids":["%s"]}`, h.invoiceA)},

		{"payment status", "GET", "/api/v1/invoices/" + h.invoiceA.String() + "/payment-status", ""},
		{"payment events", "GET", "/api/v1/invoices/" + h.invoiceA.String() + "/payment-events", ""},

		{"get child profile", "GET", "/api/v1/children/" + h.childA.String() + "/profile", ""},
		{"patch child profile", "PATCH", "/api/v1/children/" + h.childA.String() + "/profile", `{"disability_status":"unknown"}`},
		{"set collection settings", "PUT", "/api/v1/children/" + h.childA.String() + "/collection-settings", `{"over_18_collection_acknowledged":true}`},
	}

	for _, tc := range managerOnlyRoutes {
		t.Run(tc.name+"_practitioner_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.practitionerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_parent_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.parentAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_owner_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.ownerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
	}

	// Attendance check-in/check-out, absence: manager + practitioner allowed → parent forbidden
	managerPractitionerRoutes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"attendance list", "GET", "/api/v1/children/attendance", ""},
		{"check-in", "POST", "/api/v1/attendance/check-ins", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"check-out", "POST", "/api/v1/attendance/check-outs", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"mark absent", "POST", "/api/v1/attendance/absence-markers", fmt.Sprintf(`{"child_id":"%s"}`, h.childA)},
		{"clear marker", "POST", "/api/v1/attendance/absence-markers/" + h.absenceA.String() + "/clear", ""},
	}

	for _, tc := range managerPractitionerRoutes {
		t.Run(tc.name+"_parent_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.parentAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_owner_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.ownerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
	}

	// Parent-only routes → manager and practitioner forbidden
	parentOnlyRoutes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"parent list invoices", "GET", "/api/v1/parent/invoices", ""},
		{"parent get invoice", "GET", "/api/v1/parent/invoices/" + h.invoiceA.String(), ""},
		{"parent checkout", "POST", "/api/v1/parent/invoices/" + h.invoiceA.String() + "/checkout-sessions", ""},
	}

	for _, tc := range parentOnlyRoutes {
		t.Run(tc.name+"_manager_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.managerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_practitioner_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.practitionerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
			t.Run(tc.name+"_owner_forbidden", func(t *testing.T) {
				w := h.doReq(t, tc.method, tc.path, h.ownerAToken, tc.body)
				assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
			})
		})
	}

}

// ---------------------------------------------------------------------------
// Task 6: Tenant/branch scope matrix
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixTenantBranchScope(t *testing.T) {
	h := setupAuthzHarness(t)

	// Object routes: scope B resource IDs → not found from scope A
	t.Run("child B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/children/"+h.childB.String(), h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "child_not_found")
	})

	t.Run("guardian B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/guardians/"+h.guardianB.String(), h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "guardian_not_found")
	})

	t.Run("link B end not found from scope A", func(t *testing.T) {
		w := h.post(t, "/api/v1/guardian-child-links/"+h.linkB.String()+"/actions/end", h.managerAToken, `{"reason_code":"duplicate_record"}`)
		assertStatusAndCode(t, w, http.StatusNotFound, "guardian_child_link_not_found")
	})

	t.Run("mapping B end not found from scope A", func(t *testing.T) {
		w := h.post(t, "/api/v1/parent-membership-guardian-mappings/"+h.mappingB.String()+"/actions/end", h.managerAToken, `{"reason_code":"duplicate_record"}`)
		assertStatusAndCode(t, w, http.StatusNotFound, "parent_mapping_not_found")
	})

	t.Run("funding child B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/funding/children/"+h.childB.String()+"?billing_month=2026-05", h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "funding_profile_not_found")
	})

	t.Run("invoice B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/invoices/"+h.invoiceB.String(), h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})

	t.Run("invoice B issue not found from scope A", func(t *testing.T) {
		w := h.post(t, "/api/v1/invoices/"+h.invoiceB.String()+"/issue", h.managerAToken, `{"confirm":true}`)
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})

	t.Run("payment status B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/invoices/"+h.invoiceB.String()+"/payment-status", h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})

	t.Run("payment events B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/invoices/"+h.invoiceB.String()+"/payment-events", h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})

	t.Run("registration profile child B not found from scope A", func(t *testing.T) {
		w := h.get(t, "/api/v1/children/"+h.childB.String()+"/profile", h.managerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "child_not_found")
	})

	t.Run("invite A not found from scope B manager", func(t *testing.T) {
		w := h.post(t, "/api/v1/invites/"+h.inviteA.String()+"/resend", h.managerBToken, "")
		assertStatusAndCode(t, w, http.StatusNotFound, "invite_not_found")
	})

	// Create routes with cross-scope body IDs
	t.Run("create link with scope B guardian and scope A child rejects", func(t *testing.T) {
		body := fmt.Sprintf(`{"guardian_id":"%s","child_id":"%s"}`, h.guardianB, h.childA)
		w := h.post(t, "/api/v1/guardian-child-links", h.managerAToken, body)
		// Guardian B belongs to scope B → not found in scope A
		assertStatusAndCode(t, w, http.StatusNotFound, "guardian_not_found")
	})

	t.Run("create link with scope A guardian and scope B child rejects", func(t *testing.T) {
		body := fmt.Sprintf(`{"guardian_id":"%s","child_id":"%s"}`, h.guardianA, h.childB)
		w := h.post(t, "/api/v1/guardian-child-links", h.managerAToken, body)
		assertStatusAndCode(t, w, http.StatusNotFound, "child_not_found")
	})

	t.Run("create mapping with scope B membership rejects", func(t *testing.T) {
		body := fmt.Sprintf(`{"membership_id":"%s","guardian_id":"%s"}`, h.parentBMID, h.guardianA)
		w := h.post(t, "/api/v1/parent-membership-guardian-mappings", h.managerAToken, body)
		assertStatusAndCode(t, w, http.StatusNotFound, "membership_not_found")
	})

	// List routes: scope B records excluded when listing from scope A
	t.Run("children list excludes scope B", func(t *testing.T) {
		w := h.get(t, "/api/v1/children", h.managerAToken)
		assertStatus(t, w, http.StatusOK)
		ids := extractIDs(t, w)
		if containsString(ids, h.childB.String()) {
			t.Fatal("scope B child should not appear in scope A list")
		}
	})

	t.Run("guardians list excludes scope B", func(t *testing.T) {
		w := h.get(t, "/api/v1/guardians", h.managerAToken)
		assertStatus(t, w, http.StatusOK)
		ids := extractIDs(t, w)
		if containsString(ids, h.guardianB.String()) {
			t.Fatal("scope B guardian should not appear in scope A list")
		}
	})

	t.Run("invoices list excludes scope B", func(t *testing.T) {
		w := h.get(t, "/api/v1/invoices", h.managerAToken)
		assertStatus(t, w, http.StatusOK)
		ids := extractIDs(t, w)
		if containsString(ids, h.invoiceB.String()) {
			t.Fatal("scope B invoice should not appear in scope A list")
		}
	})

	t.Run("attendance list excludes scope B", func(t *testing.T) {
		w := h.get(t, "/api/v1/children/attendance", h.managerAToken)
		assertStatus(t, w, http.StatusOK)
		var resp struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		decodeBody(t, w, &resp)
		for _, item := range resp.Items {
			if item.ID == h.childB.String() {
				t.Fatal("scope B child should not appear in attendance list")
			}
		}
	})

	// Parent invoice routes with scope B invoice
	t.Run("parent invoice detail scope B not found", func(t *testing.T) {
		w := h.get(t, "/api/v1/parent/invoices/"+h.invoiceB.String(), h.parentAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})

	t.Run("parent checkout scope B not found", func(t *testing.T) {
		w := h.post(t, "/api/v1/parent/invoices/"+h.invoiceB.String()+"/checkout-sessions", h.parentAToken, "")
		assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
	})
}

// ---------------------------------------------------------------------------
// Task 7: Parent relationship matrix
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixParentRelationship(t *testing.T) {
	h := setupAuthzHarness(t)

	// Seed parent relationship test data
	unlinkedChild := uuid.MustParse("a1000000-0000-0000-0000-000000000900")
	unlinkedInvoice := uuid.MustParse("a1000000-0000-0000-0000-000000000901")
	endedLinkChild := uuid.MustParse("a1000000-0000-0000-0000-000000000910")
	endedLinkGuardian := uuid.MustParse("a1000000-0000-0000-0000-000000000911")
	endedLinkID := uuid.MustParse("a1000000-0000-0000-0000-000000000912")
	endedLinkMapping := uuid.MustParse("a1000000-0000-0000-0000-000000000913")
	endedLinkInvoice := uuid.MustParse("a1000000-0000-0000-0000-000000000914")
	endedMappingChild := uuid.MustParse("a1000000-0000-0000-0000-000000000920")
	endedMappingGuardian := uuid.MustParse("a1000000-0000-0000-0000-000000000921")
	endedMappingLink := uuid.MustParse("a1000000-0000-0000-0000-000000000922")
	endedMappingID := uuid.MustParse("a1000000-0000-0000-0000-000000000923")
	endedMappingInvoice := uuid.MustParse("a1000000-0000-0000-0000-000000000924")

	ctx := context.Background()
	runID := uuid.MustParse("a1000000-0000-0000-0000-000000000990")

	_, err := h.pool.Exec(ctx, `INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		VALUES ($1, $2, $3, $4, 'issue', 'completed', now(), now(), $5, $6, 'parent-req')`,
		runID, h.tenantA, h.branchA, dbtest.DateAt(2026, 5, 1), h.managerAUID, h.managerAMID)
	if err != nil {
		t.Fatalf("insert parent test run: %v", err)
	}

	// Unlinked child + invoice (no guardian link to parent A)
	dbtest.InsertChild(t, h.pool, unlinkedChild, h.tenantA, h.branchA, "Unlinked Child",
		dbtest.DateAt(2023, 3, 1), dbtest.DateAt(2026, 3, 1), true)
	_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','issued','INV-UL-001',1,$6,$6,now(),$7,$8,now(),now() + interval '7 days',
		'GBP',5000,0,5000,0,$9,$10,'{}')`,
		unlinkedInvoice, h.tenantA, h.branchA, unlinkedChild, dbtest.DateAt(2026, 5, 1),
		runID, h.managerAUID, h.managerAMID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert unlinked invoice: %v", err)
	}

	// Ended link child
	dbtest.InsertChild(t, h.pool, endedLinkChild, h.tenantA, h.branchA, "Ended Link Child",
		dbtest.DateAt(2023, 3, 1), dbtest.DateAt(2026, 3, 1), true)
	dbtest.InsertGuardian(t, h.pool, endedLinkGuardian, h.tenantA, h.branchA, "Ended Link Guardian", true)
	endedAt := time.Now().UTC()
	_, err = h.pool.Exec(ctx, `INSERT INTO guardian_child_links (id, tenant_id, branch_id, guardian_id, child_id, ended_at, ended_reason_code)
		VALUES ($1, $2, $3, $4, $5, $6, 'duplicate_record')`,
		endedLinkID, h.tenantA, h.branchA, endedLinkGuardian, endedLinkChild, endedAt)
	if err != nil {
		t.Fatalf("insert ended link: %v", err)
		dbtest.InsertParentMapping(t, h.pool, endedLinkMapping, h.tenantA, h.branchA, h.parentAMID, endedLinkGuardian)
		_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','issued','INV-EL-001',1,$6,$6,now(),$7,$8,now(),now() + interval '7 days',
		'GBP',5000,0,5000,0,$9,$10,'{}')`,
			endedLinkInvoice, h.tenantA, h.branchA, endedLinkChild, dbtest.DateAt(2026, 5, 1),
			runID, h.managerAUID, h.managerAMID,
			dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
		if err != nil {
			t.Fatalf("insert ended link invoice: %v", err)
		}

		// Ended mapping child
		dbtest.InsertChild(t, h.pool, endedMappingChild, h.tenantA, h.branchA, "Ended Mapping Child",
			dbtest.DateAt(2023, 3, 1), dbtest.DateAt(2026, 3, 1), true)
		dbtest.InsertGuardian(t, h.pool, endedMappingGuardian, h.tenantA, h.branchA, "Ended Mapping Guardian", true)
		dbtest.InsertGuardianLink(t, h.pool, endedMappingLink, h.tenantA, h.branchA, endedMappingGuardian, endedMappingChild)
		_, err = h.pool.Exec(ctx, `INSERT INTO parent_membership_guardians (id, tenant_id, branch_id, membership_id, guardian_id, ended_at, ended_reason_code)
		VALUES ($1, $2, $3, $4, $5, $6, 'duplicate_record')`,
			endedMappingID, h.tenantA, h.branchA, h.parentAMID, endedMappingGuardian, endedAt)
		if err != nil {
			t.Fatalf("insert ended mapping: %v", err)
		}
		_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details)
		VALUES ($1,$2,$3,$4,$5,'monthly','issued','INV-EM-001',1,$6,$6,now(),$7,$8,now(),now() + interval '7 days',
		'GBP',5000,0,5000,0,$9,$10,'{}')`,
			endedMappingInvoice, h.tenantA, h.branchA, endedMappingChild, dbtest.DateAt(2026, 5, 1),
			runID, h.managerAUID, h.managerAMID,
			dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
		if err != nil {
			t.Fatalf("insert ended mapping invoice: %v", err)
		}

		// --- Assertions ---

		t.Run("parent list includes linked issued invoice", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices", h.parentAToken)
			assertStatus(t, w, http.StatusOK)
			ids := extractIDs(t, w)
			if !containsString(ids, h.invoiceA.String()) {
				t.Fatal("linked issued invoice should appear in parent list")
			}
		})

		t.Run("parent list excludes unlinked invoice", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices", h.parentAToken)
			assertStatus(t, w, http.StatusOK)
			ids := extractIDs(t, w)
			if containsString(ids, unlinkedInvoice.String()) {
				t.Fatal("unlinked invoice should not appear in parent list")
			}
		})

		t.Run("parent list excludes ended-link invoice", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices", h.parentAToken)
			assertStatus(t, w, http.StatusOK)
			ids := extractIDs(t, w)
			if containsString(ids, endedLinkInvoice.String()) {
				t.Fatal("ended-link invoice should not appear in parent list")
			}
		})

		t.Run("parent list excludes ended-mapping invoice", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices", h.parentAToken)
			assertStatus(t, w, http.StatusOK)
			ids := extractIDs(t, w)
			if containsString(ids, endedMappingInvoice.String()) {
				t.Fatal("ended-mapping invoice should not appear in parent list")
			}
		})

		t.Run("parent list excludes draft invoice", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices", h.parentAToken)
			assertStatus(t, w, http.StatusOK)
			ids := extractIDs(t, w)
			if containsString(ids, h.draftInvoiceA.String()) {
				t.Fatal("draft invoice should not appear in parent list")
			}
		})

		t.Run("parent detail unlinked returns 404", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices/"+unlinkedInvoice.String(), h.parentAToken)
			assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
		})

		t.Run("parent detail ended-link returns 404", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices/"+endedLinkInvoice.String(), h.parentAToken)
			assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
		})

		t.Run("parent detail ended-mapping returns 404", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices/"+endedMappingInvoice.String(), h.parentAToken)
			assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
		})

		t.Run("parent detail linked returns 200", func(t *testing.T) {
			w := h.get(t, "/api/v1/parent/invoices/"+h.invoiceA.String(), h.parentAToken)
			assertStatus(t, w, http.StatusOK)
		})

		t.Run("parent checkout unlinked returns 404", func(t *testing.T) {
			w := h.post(t, "/api/v1/parent/invoices/"+unlinkedInvoice.String()+"/checkout-sessions", h.parentAToken, "")
			assertStatusAndCode(t, w, http.StatusNotFound, "invoice_not_found")
		})

		t.Run("parent checkout linked returns 201", func(t *testing.T) {
			w := h.post(t, "/api/v1/parent/invoices/"+h.invoiceA.String()+"/checkout-sessions", h.parentAToken, "")
			assertStatus(t, w, http.StatusCreated)
		})

	}
}

// ---------------------------------------------------------------------------
// Owner role boundaries
// ---------------------------------------------------------------------------

func TestAuthorizationMatrixOwnerRoleBoundaries(t *testing.T) {
	h := setupAuthzHarness(t)

	// Owner CAN access owner endpoints
	t.Run("owner can get site summaries", func(t *testing.T) {
		w := h.get(t, "/api/v1/owner/site-summaries", h.ownerAToken)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("owner can list manager access", func(t *testing.T) {
		w := h.get(t, "/api/v1/owner/manager-access?site_id="+h.branchA.String(), h.ownerAToken)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("owner can grant manager access", func(t *testing.T) {
		body := fmt.Sprintf(`{"email":"new-grant@test.com"}`)
		w := h.post(t, "/api/v1/owner/sites/"+h.branchA.String()+"/manager-access", h.ownerAToken, body)
		// 200 (invite created) or other success — just not 403
		if w.Code == http.StatusForbidden {
			t.Fatalf("owner should access grant endpoint, got 403")
		}
	})

	// Non-owner roles CANNOT access owner endpoints
	ownerRoutes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"site summaries", "GET", "/api/v1/owner/site-summaries", ""},
		{"manager access list", "GET", "/api/v1/owner/manager-access?site_id=" + h.branchA.String(), ""},
		{"grant access", "POST", "/api/v1/owner/sites/" + h.branchA.String() + "/manager-access", `{"email":"x@test.com"}`},
		{"deactivate access", "POST", "/api/v1/owner/sites/" + h.branchA.String() + "/manager-access/" + h.managerAMID.String() + "/actions/deactivate", ""},
		{"reactivate access", "POST", "/api/v1/owner/sites/" + h.branchA.String() + "/manager-access/" + h.managerAMID.String() + "/actions/activate", ""},
	}

	for _, tc := range ownerRoutes {
		t.Run(tc.name+"_manager_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.managerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_practitioner_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.practitionerAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
		t.Run(tc.name+"_parent_forbidden", func(t *testing.T) {
			w := h.doReq(t, tc.method, tc.path, h.parentAToken, tc.body)
			assertStatusAndCode(t, w, http.StatusForbidden, "forbidden_role")
		})
	}

	// Owner unauthenticated → 401
	t.Run("owner site summaries requires auth", func(t *testing.T) {
		w := h.get(t, "/api/v1/owner/site-summaries", "")
		assertStatusAndCode(t, w, http.StatusUnauthorized, "unauthorized")
	})
}

func TestAuthorizationMatrixOwnerTenantIsolation(t *testing.T) {
	h := setupAuthzHarness(t)

	// Owner A cannot see tenant B sites in summaries
	t.Run("owner A site summaries excludes tenant B", func(t *testing.T) {
		w := h.get(t, "/api/v1/owner/site-summaries", h.ownerAToken)
		assertStatus(t, w, http.StatusOK)
		var resp struct {
			Sites []struct {
				SiteID string `json:"site_id"`
			} `json:"sites"`
		}
		decodeBody(t, w, &resp)
		for _, s := range resp.Sites {
			if s.SiteID == h.branchB.String() {
				t.Fatal("owner A should not see tenant B sites")
			}
		}
	})

	// Owner A querying tenant B site → site_not_found
	t.Run("owner A manager access for tenant B site not found", func(t *testing.T) {
		w := h.get(t, "/api/v1/owner/manager-access?site_id="+h.branchB.String(), h.ownerAToken)
		assertStatusAndCode(t, w, http.StatusNotFound, "site_not_found")
	})

	t.Run("owner A grant access for tenant B site not found", func(t *testing.T) {
		w := h.post(t, "/api/v1/owner/sites/"+h.branchB.String()+"/manager-access", h.ownerAToken, `{"email":"x@test.com"}`)
		assertStatusAndCode(t, w, http.StatusNotFound, "site_not_found")
	})

	t.Run("owner A deactivate for tenant B site not found", func(t *testing.T) {
		w := h.post(t, "/api/v1/owner/sites/"+h.branchB.String()+"/manager-access/"+h.managerBMID.String()+"/actions/deactivate", h.ownerAToken, "")
		assertStatusAndCode(t, w, http.StatusNotFound, "site_not_found")
	})
}
