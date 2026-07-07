package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	"nursery-management-system/api/internal/platform/dbtest"
)

type siteProfileHarness struct {
	router            *gin.Engine
	pool              *pgxpool.Pool
	tokens            *authtokens.TokenManager
	tenantID          uuid.UUID
	branchID          uuid.UUID
	ownerUID          uuid.UUID
	ownerMID          uuid.UUID
	managerUID        uuid.UUID
	managerMID        uuid.UUID
	practitionerUID   uuid.UUID
	practitionerMID   uuid.UUID
	parentUID         uuid.UUID
	parentMID         uuid.UUID
	ownerToken        string
	managerToken      string
	practitionerToken string
	parentToken       string
}

func setupSiteProfileHarness(t *testing.T) *siteProfileHarness {
	t.Helper()

	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &siteProfileHarness{
		router:          Bootstrap(testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), pool),
		pool:            pool,
		tokens:          authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720, 24),
		tenantID:        uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		branchID:        uuid.MustParse("c2000000-0000-0000-0000-000000000001"),
		ownerUID:        uuid.MustParse("c3000000-0000-0000-0000-000000000001"),
		ownerMID:        uuid.MustParse("c4000000-0000-0000-0000-000000000001"),
		managerUID:      uuid.MustParse("c3000000-0000-0000-0000-000000000002"),
		managerMID:      uuid.MustParse("c4000000-0000-0000-0000-000000000002"),
		practitionerUID: uuid.MustParse("c3000000-0000-0000-0000-000000000003"),
		practitionerMID: uuid.MustParse("c4000000-0000-0000-0000-000000000003"),
		parentUID:       uuid.MustParse("c3000000-0000-0000-0000-000000000004"),
		parentMID:       uuid.MustParse("c4000000-0000-0000-0000-000000000004"),
	}

	dbtest.InsertTenant(t, pool, h.tenantID, "SiteProfile Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "SiteProfile Branch")
	dbtest.InsertUser(t, pool, h.ownerUID, "sp-owner@example.com", "hash", true)
	dbtest.InsertUser(t, pool, h.managerUID, "sp-mgr@example.com", "hash", true)
	dbtest.InsertUser(t, pool, h.practitionerUID, "sp-prac@example.com", "hash", true)
	dbtest.InsertUser(t, pool, h.parentUID, "sp-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, h.ownerMID, h.tenantID, h.branchID, h.ownerUID, "owner", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, h.practitionerMID, h.tenantID, h.branchID, h.practitionerUID, "practitioner", true)
	dbtest.InsertMembership(t, pool, h.parentMID, h.tenantID, h.branchID, h.parentUID, "parent", true)

	h.ownerToken = mustAccessToken(t, h.tokens, h.ownerUID, h.ownerMID, h.tenantID, h.branchID, "owner")
	h.managerToken = mustAccessToken(t, h.tokens, h.managerUID, h.managerMID, h.tenantID, h.branchID, "manager")
	h.practitionerToken = mustAccessToken(t, h.tokens, h.practitionerUID, h.practitionerMID, h.tenantID, h.branchID, "practitioner")
	h.parentToken = mustAccessToken(t, h.tokens, h.parentUID, h.parentMID, h.tenantID, h.branchID, "parent")

	return h
}

func TestSiteProfileRouteInventory(t *testing.T) {
	h := setupSiteProfileHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/site-profile",
		"PUT /api/v1/site-profile",
	} {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestSiteProfileUnauthenticated(t *testing.T) {
	h := setupSiteProfileHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", "", "")
	requireStatus(t, w, http.StatusUnauthorized)

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", "", `{}`)
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestSiteProfileManagerGetEmpty(t *testing.T) {
	h := setupSiteProfileHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp struct {
		SiteProfile *struct {
			NurseryName string `json:"nursery_name"`
		} `json:"site_profile"`
	}
	decodeJSON(t, w, &resp)
	if resp.SiteProfile != nil {
		t.Fatal("expected site_profile to be null when no profile exists")
	}
}

func TestSiteProfileManagerPutValid(t *testing.T) {
	h := setupSiteProfileHarness(t)

	body := `{
		"nursery_name": "Little Stars Nursery",
		"description": "A warm and caring nursery",
		"phone": "+44 161 555 0100",
		"email": "hello@littlestars.example",
		"website": "https://littlestars.example",
		"address_street": "12 Acacia Ave",
		"address_city": "Manchester",
		"address_postcode": "M1 4BT"
	}`

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp struct {
		NurseryName     string `json:"nursery_name"`
		Description     string `json:"description"`
		Phone           string `json:"phone"`
		Email           string `json:"email"`
		Website         string `json:"website"`
		AddressStreet   string `json:"address_street"`
		AddressCity     string `json:"address_city"`
		AddressPostcode string `json:"address_postcode"`
	}
	decodeJSON(t, w, &resp)
	if resp.NurseryName != "Little Stars Nursery" {
		t.Fatalf("got nursery_name %q, want %q", resp.NurseryName, "Little Stars Nursery")
	}

	ctx := context.Background()
	var auditCount int
	err := h.pool.QueryRow(ctx,
		"SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action_type = 'site_profile_updated'",
		h.tenantID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected 1 audit log entry, got %d", auditCount)
	}
}

func TestSiteProfileGetAfterPut(t *testing.T) {
	h := setupSiteProfileHarness(t)

	body := `{
		"nursery_name": "Little Stars Nursery",
		"description": "",
		"phone": "+44 161 555 0100",
		"email": "hello@littlestars.example",
		"website": "https://littlestars.example",
		"address_street": "12 Acacia Ave",
		"address_city": "Manchester",
		"address_postcode": "M1 4BT"
	}`

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	w = doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp struct {
		SiteProfile *struct {
			NurseryName string `json:"nursery_name"`
		} `json:"site_profile"`
	}
	decodeJSON(t, w, &resp)
	if resp.SiteProfile == nil {
		t.Fatal("expected site_profile to be non-null after PUT")
	}
	if resp.SiteProfile.NurseryName != "Little Stars Nursery" {
		t.Fatalf("got nursery_name %q, want %q", resp.SiteProfile.NurseryName, "Little Stars Nursery")
	}
}

func TestSiteProfileOwnerGetPut(t *testing.T) {
	h := setupSiteProfileHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", h.ownerToken, "")
	requireStatus(t, w, http.StatusOK)

	body := `{
		"nursery_name": "Owner Nursery",
		"description": "",
		"phone": "+44 161 555 0101",
		"email": "owner@nursery.example",
		"website": "https://owner.example",
		"address_street": "1 High Street",
		"address_city": "London",
		"address_postcode": "SW1A 1AA"
	}`
	w = doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.ownerToken, body)
	requireStatus(t, w, http.StatusOK)
}

func TestSiteProfilePractitionerGetPut(t *testing.T) {
	h := setupSiteProfileHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", h.practitionerToken, "")
	requireStatus(t, w, http.StatusOK)

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.practitionerToken, `{
		"nursery_name": "Test",
		"description": "",
		"phone": "+44 161 555 0100",
		"email": "test@example.com",
		"website": "https://example.com",
		"address_street": "1 Street",
		"address_city": "City",
		"address_postcode": "SW1A 1AA"
	}`)
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")
}

func TestSiteProfileParentGetPut(t *testing.T) {
	h := setupSiteProfileHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/site-profile", h.parentToken, "")
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.parentToken, `{}`)
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")
}

func TestSiteProfilePutEmptyName(t *testing.T) {
	h := setupSiteProfileHarness(t)

	body := `{
		"nursery_name": "",
		"description": "",
		"phone": "+44 161 555 0100",
		"email": "hello@littlestars.example",
		"website": "https://littlestars.example",
		"address_street": "12 Acacia Ave",
		"address_city": "Manchester",
		"address_postcode": "M1 4BT"
	}`

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusBadRequest)

	var errResp struct {
		Code    string `json:"code"`
		Details []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"details"`
	}
	decodeJSON(t, w, &errResp)
	if errResp.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errResp.Code)
	}
	if len(errResp.Details) == 0 {
		t.Fatal("expected details with field errors")
	}
	found := false
	for _, f := range errResp.Details {
		if f.Field == "nursery_name" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected nursery_name in field errors")
	}
}

func TestSiteProfilePutAllEmpty(t *testing.T) {
	h := setupSiteProfileHarness(t)

	body := `{
		"nursery_name": "",
		"description": "",
		"phone": "",
		"email": "",
		"website": "",
		"address_street": "",
		"address_city": "",
		"address_postcode": ""
	}`

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusBadRequest)

	var errResp struct {
		Code    string `json:"code"`
		Details []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"details"`
	}
	decodeJSON(t, w, &errResp)
	if errResp.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errResp.Code)
	}
	if len(errResp.Details) < 7 {
		t.Fatalf("expected at least 7 field errors, got %d", len(errResp.Details))
	}
}

func TestSiteProfileRepeatedPutAudit(t *testing.T) {
	h := setupSiteProfileHarness(t)

	body := `{
		"nursery_name": "Little Stars Nursery",
		"description": "",
		"phone": "+44 161 555 0100",
		"email": "hello@littlestars.example",
		"website": "https://littlestars.example",
		"address_street": "12 Acacia Ave",
		"address_city": "Manchester",
		"address_postcode": "M1 4BT"
	}`

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/site-profile", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	ctx := context.Background()
	var auditCount int
	err := h.pool.QueryRow(ctx,
		"SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action_type = 'site_profile_updated'",
		h.tenantID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if auditCount != 2 {
		t.Fatalf("expected 2 audit log entries for 2 PUTs, got %d", auditCount)
	}
}
