package bootstrap

import (
	"context"
	"encoding/json"
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
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/dbtest"
)

type peopleHarness struct {
	router            *gin.Engine
	pool              *pgxpool.Pool
	tokens            *authtokens.TokenManager
	scopeA            peopleScope
	scopeB            peopleScopeB
	managerToken      string
	practitionerToken string
	parentToken       string
}

type peopleScope struct {
	tenantID uuid.UUID
	branchID uuid.UUID

	managerUserID       uuid.UUID
	managerMembershipID uuid.UUID
	practitionerUserID  uuid.UUID
	practitionerMID     uuid.UUID
	parentUserID        uuid.UUID
	parentMembershipID  uuid.UUID
	parentUserID2       uuid.UUID
	parentMembershipID2 uuid.UUID

	activeChildID   uuid.UUID
	inactiveChildID uuid.UUID

	activeGuardianID   uuid.UUID
	inactiveGuardianID uuid.UUID
	otherGuardianID    uuid.UUID

	activeLinkID    uuid.UUID
	activeMappingID uuid.UUID
}

type peopleScopeB struct {
	tenantID uuid.UUID
	branchID uuid.UUID

	parentUserID       uuid.UUID
	parentMembershipID uuid.UUID

	childID    uuid.UUID
	guardianID uuid.UUID
	linkID     uuid.UUID
	mappingID  uuid.UUID
}

func TestPeopleRoutesInventory(t *testing.T) {
	h := setupPeopleHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/children",
		"GET /api/v1/children/:child_id",
		"POST /api/v1/children",
		"PATCH /api/v1/children/:child_id",
		"POST /api/v1/children/:child_id/actions/mark-inactive",
		"GET /api/v1/children/attendance",
		"GET /api/v1/guardians",
		"GET /api/v1/guardians/:guardian_id",
		"POST /api/v1/guardians",
		"PATCH /api/v1/guardians/:guardian_id",
		"POST /api/v1/guardians/:guardian_id/actions/deactivate",
		"POST /api/v1/guardians/:guardian_id/actions/reactivate",
		"POST /api/v1/guardian-child-links",
		"POST /api/v1/guardian-child-links/:link_id/actions/end",
		"POST /api/v1/parent-membership-guardian-mappings",
		"POST /api/v1/parent-membership-guardian-mappings/:mapping_id/actions/end",
	}

	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}

	for _, route := range h.router.Routes() {
		if route.Method != http.MethodDelete {
			continue
		}
		if isPeopleRoute(route.Path) {
			t.Fatalf("unexpected DELETE people route registered: %s %s", route.Method, route.Path)
		}
	}

	forbiddenDeletes := []string{
		"DELETE /api/v1/children/:id",
		"DELETE /api/v1/guardians/:id",
		"DELETE /api/v1/guardian-child-links/:id",
		"DELETE /api/v1/parent-membership-guardian-mappings/:id",
	}
	for _, want := range forbiddenDeletes {
		if _, ok := have[want]; ok {
			t.Fatalf("unexpected hard-delete route registered: %s", want)
		}
	}
}

func TestPeopleWriteRoleGuards(t *testing.T) {
	h := setupPeopleHarness(t)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "create child",
			method:     http.MethodPost,
			path:       "/api/v1/children",
			body:       `{"full_name":"New Child","date_of_birth":"2021-05-01","start_date":"2024-01-01","core_hourly_rate_minor":450}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "update child",
			method:     http.MethodPatch,
			path:       "/api/v1/children/" + h.scopeA.activeChildID.String(),
			body:       `{"full_name":"Updated Child"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "mark child inactive",
			method:     http.MethodPost,
			path:       "/api/v1/children/" + h.scopeA.activeChildID.String() + "/actions/mark-inactive",
			body:       `{"reason_code":"left_nursery"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "create guardian",
			method:     http.MethodPost,
			path:       "/api/v1/guardians",
			body:       `{"full_name":"New Guardian"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "update guardian",
			method:     http.MethodPatch,
			path:       "/api/v1/guardians/" + h.scopeA.activeGuardianID.String(),
			body:       `{"notes":"Updated notes"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "deactivate guardian",
			method:     http.MethodPost,
			path:       "/api/v1/guardians/" + h.scopeA.otherGuardianID.String() + "/actions/deactivate",
			body:       `{"reason_code":"access_revoked"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "reactivate guardian",
			method:     http.MethodPost,
			path:       "/api/v1/guardians/" + h.scopeA.inactiveGuardianID.String() + "/actions/reactivate",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create guardian-child link",
			method:     http.MethodPost,
			path:       "/api/v1/guardian-child-links",
			body:       `{"guardian_id":"` + h.scopeA.activeGuardianID.String() + `","child_id":"` + h.scopeA.activeChildID.String() + `"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "end guardian-child link",
			method:     http.MethodPost,
			path:       "/api/v1/guardian-child-links/" + h.scopeA.activeLinkID.String() + "/actions/end",
			body:       `{"reason_code":"contact_update"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "create parent mapping",
			method:     http.MethodPost,
			path:       "/api/v1/parent-membership-guardian-mappings",
			body:       `{"membership_id":"` + h.scopeA.parentMembershipID2.String() + `","guardian_id":"` + h.scopeA.activeGuardianID.String() + `"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "end parent mapping",
			method:     http.MethodPost,
			path:       "/api/v1/parent-membership-guardian-mappings/" + h.scopeA.activeMappingID.String() + "/actions/end",
			body:       `{"reason_code":"contact_update"}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(t, h.router, tc.method, tc.path, "", tc.body)
			requireStatus(t, w, http.StatusUnauthorized)
			requireErrorCode(t, w, "unauthorized")

			w = doRequest(t, h.router, tc.method, tc.path, h.practitionerToken, tc.body)
			requireStatus(t, w, http.StatusForbidden)
			requireErrorCode(t, w, "forbidden_role")

			w = doRequest(t, h.router, tc.method, tc.path, h.parentToken, tc.body)
			requireStatus(t, w, http.StatusForbidden)
			requireErrorCode(t, w, "forbidden_role")

			w = doRequest(t, h.router, tc.method, tc.path, h.managerToken, tc.body)
			requireStatus(t, w, tc.wantStatus)
		})
	}
}

func TestPeopleReadRoleGuards(t *testing.T) {
	h := setupPeopleHarness(t)

	t.Run("attendance role access", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/children/attendance", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children/attendance", h.practitionerToken, "")
		requireStatus(t, w, http.StatusOK)

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children/attendance", h.parentToken, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	})

	paths := []string{
		"/api/v1/children",
		"/api/v1/children/" + h.scopeA.activeChildID.String(),
		"/api/v1/guardians",
		"/api/v1/guardians/" + h.scopeA.activeGuardianID.String(),
	}

	for _, path := range paths {
		t.Run("practitioner rejected "+path, func(t *testing.T) {
			w := doRequest(t, h.router, http.MethodGet, path, h.practitionerToken, "")
			requireStatus(t, w, http.StatusForbidden)
			requireErrorCode(t, w, "forbidden_role")
		})
		t.Run("parent rejected "+path, func(t *testing.T) {
			w := doRequest(t, h.router, http.MethodGet, path, h.parentToken, "")
			requireStatus(t, w, http.StatusForbidden)
			requireErrorCode(t, w, "forbidden_role")
		})
	}
}

func TestPeopleScopeBoundaries(t *testing.T) {
	h := setupPeopleHarness(t)

	cases := []struct {
		name    string
		method  string
		path    string
		body    string
		errCode string
	}{
		{
			name:    "get child",
			method:  http.MethodGet,
			path:    "/api/v1/children/" + h.scopeB.childID.String(),
			errCode: "child_not_found",
		},
		{
			name:    "update child",
			method:  http.MethodPatch,
			path:    "/api/v1/children/" + h.scopeB.childID.String(),
			body:    `{"full_name":"Updated"}`,
			errCode: "child_not_found",
		},
		{
			name:    "mark child inactive",
			method:  http.MethodPost,
			path:    "/api/v1/children/" + h.scopeB.childID.String() + "/actions/mark-inactive",
			body:    `{"reason_code":"left_nursery"}`,
			errCode: "child_not_found",
		},
		{
			name:    "get guardian",
			method:  http.MethodGet,
			path:    "/api/v1/guardians/" + h.scopeB.guardianID.String(),
			errCode: "guardian_not_found",
		},
		{
			name:    "update guardian",
			method:  http.MethodPatch,
			path:    "/api/v1/guardians/" + h.scopeB.guardianID.String(),
			body:    `{"notes":"Updated"}`,
			errCode: "guardian_not_found",
		},
		{
			name:    "deactivate guardian",
			method:  http.MethodPost,
			path:    "/api/v1/guardians/" + h.scopeB.guardianID.String() + "/actions/deactivate",
			body:    `{"reason_code":"access_revoked"}`,
			errCode: "guardian_not_found",
		},
		{
			name:    "end guardian-child link",
			method:  http.MethodPost,
			path:    "/api/v1/guardian-child-links/" + h.scopeB.linkID.String() + "/actions/end",
			body:    `{"reason_code":"contact_update"}`,
			errCode: "guardian_child_link_not_found",
		},
		{
			name:    "end parent mapping",
			method:  http.MethodPost,
			path:    "/api/v1/parent-membership-guardian-mappings/" + h.scopeB.mappingID.String() + "/actions/end",
			body:    `{"reason_code":"contact_update"}`,
			errCode: "parent_mapping_not_found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(t, h.router, tc.method, tc.path, h.managerToken, tc.body)
			requireStatus(t, w, http.StatusNotFound)
			requireErrorCode(t, w, tc.errCode)
		})
	}

	t.Run("create link rejects cross-scope guardian", func(t *testing.T) {
		body := `{"guardian_id":"` + h.scopeB.guardianID.String() + `","child_id":"` + h.scopeA.activeChildID.String() + `"}`
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/guardian-child-links", h.managerToken, body)
		requireStatus(t, w, http.StatusNotFound)
		requireErrorCode(t, w, "guardian_not_found")

		var count int
		err := h.pool.QueryRow(context.Background(),
			"SELECT count(*) FROM guardian_child_links WHERE tenant_id = $1 AND branch_id = $2 AND guardian_id = $3 AND child_id = $4",
			h.scopeA.tenantID, h.scopeA.branchID, h.scopeB.guardianID, h.scopeA.activeChildID,
		).Scan(&count)
		if err != nil {
			t.Fatalf("count cross-scope links: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected no cross-scope links, found %d", count)
		}
	})

	t.Run("create mapping rejects cross-scope membership", func(t *testing.T) {
		body := `{"membership_id":"` + h.scopeB.parentMembershipID.String() + `","guardian_id":"` + h.scopeA.activeGuardianID.String() + `"}`
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/parent-membership-guardian-mappings", h.managerToken, body)
		requireStatus(t, w, http.StatusNotFound)
		requireErrorCode(t, w, "membership_not_found")

		var count int
		err := h.pool.QueryRow(context.Background(),
			"SELECT count(*) FROM parent_membership_guardians WHERE tenant_id = $1 AND branch_id = $2 AND membership_id = $3",
			h.scopeA.tenantID, h.scopeA.branchID, h.scopeB.parentMembershipID,
		).Scan(&count)
		if err != nil {
			t.Fatalf("count cross-scope mappings: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected no cross-scope mappings, found %d", count)
		}
	})
}

func TestPeopleListingBehavior(t *testing.T) {
	h := setupPeopleHarness(t)

	t.Run("children listing filters", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/children", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeChildID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children?status=active", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeChildID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children?status=inactive", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.inactiveChildID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children?status=all", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeChildID.String(), h.scopeA.inactiveChildID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/children?status=unknown", h.managerToken, "")
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "validation_error")

		for _, path := range []string{
			"/api/v1/children?limit=0",
			"/api/v1/children?limit=201",
			"/api/v1/children?offset=-1",
		} {
			w = doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
			requireStatus(t, w, http.StatusBadRequest)
			requireErrorCode(t, w, "validation_error")
		}
	})

	t.Run("guardian listing filters", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/guardians", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeGuardianID.String(), h.scopeA.otherGuardianID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/guardians?status=active", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeGuardianID.String(), h.scopeA.otherGuardianID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/guardians?status=inactive", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.inactiveGuardianID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/guardians?status=all", h.managerToken, "")
		requireStatus(t, w, http.StatusOK)
		requireListIDs(t, w, []string{h.scopeA.activeGuardianID.String(), h.scopeA.otherGuardianID.String(), h.scopeA.inactiveGuardianID.String()})

		w = doRequest(t, h.router, http.MethodGet, "/api/v1/guardians?status=unknown", h.managerToken, "")
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "validation_error")

		for _, path := range []string{
			"/api/v1/guardians?limit=0",
			"/api/v1/guardians?limit=201",
			"/api/v1/guardians?offset=-1",
		} {
			w = doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
			requireStatus(t, w, http.StatusBadRequest)
			requireErrorCode(t, w, "validation_error")
		}
	})
}

func TestPeopleLifecycleReasonHandling(t *testing.T) {
	h := setupPeopleHarness(t)

	t.Run("child mark-inactive reasons", func(t *testing.T) {
		path := "/api/v1/children/" + h.scopeA.activeChildID.String() + "/actions/mark-inactive"

		w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "child_lifecycle_reason_required")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"bad"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "lifecycle_reason_invalid")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"other"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "reason_note_required_for_other")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"left_nursery","reason_note":"moving"}`)
		requireStatus(t, w, http.StatusOK)

		var resp struct {
			IsActive       bool    `json:"is_active"`
			LeftAt         *string `json:"left_at"`
			LeftReasonCode *string `json:"left_reason_code"`
			LeftReasonNote *string `json:"left_reason_note"`
		}
		decodeJSON(t, w, &resp)
		if resp.IsActive {
			t.Fatal("expected child to be inactive")
		}
		if resp.LeftAt == nil || *resp.LeftAt == "" {
			t.Fatal("expected left_at to be set")
		}
		if resp.LeftReasonCode == nil || *resp.LeftReasonCode != "left_nursery" {
			t.Fatalf("expected left_reason_code left_nursery, got %v", resp.LeftReasonCode)
		}
		if resp.LeftReasonNote == nil || *resp.LeftReasonNote != "moving" {
			t.Fatalf("expected left_reason_note moving, got %v", resp.LeftReasonNote)
		}
	})

	t.Run("guardian deactivation reasons", func(t *testing.T) {
		path := "/api/v1/guardians/" + h.scopeA.otherGuardianID.String() + "/actions/deactivate"

		w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "guardian_deactivation_reason_required")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"bad"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "lifecycle_reason_invalid")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"other"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "reason_note_required_for_other")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"access_revoked","reason_note":"staff change"}`)
		requireStatus(t, w, http.StatusOK)

		var resp struct {
			IsActive               bool    `json:"is_active"`
			DeactivatedAt          *string `json:"deactivated_at"`
			DeactivationReasonCode *string `json:"deactivation_reason_code"`
			DeactivationReasonNote *string `json:"deactivation_reason_note"`
		}
		decodeJSON(t, w, &resp)
		if resp.IsActive {
			t.Fatal("expected guardian to be inactive")
		}
		if resp.DeactivatedAt == nil || *resp.DeactivatedAt == "" {
			t.Fatal("expected deactivated_at to be set")
		}
		if resp.DeactivationReasonCode == nil || *resp.DeactivationReasonCode != "access_revoked" {
			t.Fatalf("expected deactivation_reason_code access_revoked, got %v", resp.DeactivationReasonCode)
		}
		if resp.DeactivationReasonNote == nil || *resp.DeactivationReasonNote != "staff change" {
			t.Fatalf("expected deactivation_reason_note staff change, got %v", resp.DeactivationReasonNote)
		}
	})

	t.Run("guardian-child link end reasons", func(t *testing.T) {
		path := "/api/v1/guardian-child-links/" + h.scopeA.activeLinkID.String() + "/actions/end"

		w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "relationship_reason_required")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"bad"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "lifecycle_reason_invalid")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"other"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "reason_note_required_for_other")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"contact_update","reason_note":"updated"}`)
		requireStatus(t, w, http.StatusOK)

		var resp struct {
			EndedAt         *string `json:"ended_at"`
			EndedReasonCode *string `json:"ended_reason_code"`
			EndedReasonNote *string `json:"ended_reason_note"`
		}
		decodeJSON(t, w, &resp)
		if resp.EndedAt == nil || *resp.EndedAt == "" {
			t.Fatal("expected ended_at to be set")
		}
		if resp.EndedReasonCode == nil || *resp.EndedReasonCode != "contact_update" {
			t.Fatalf("expected ended_reason_code contact_update, got %v", resp.EndedReasonCode)
		}
		if resp.EndedReasonNote == nil || *resp.EndedReasonNote != "updated" {
			t.Fatalf("expected ended_reason_note updated, got %v", resp.EndedReasonNote)
		}
	})

	t.Run("parent mapping end reasons", func(t *testing.T) {
		path := "/api/v1/parent-membership-guardian-mappings/" + h.scopeA.activeMappingID.String() + "/actions/end"

		w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "relationship_reason_required")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"bad"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "lifecycle_reason_invalid")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"other"}`)
		requireStatus(t, w, http.StatusBadRequest)
		requireErrorCode(t, w, "reason_note_required_for_other")

		w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"contact_update","reason_note":"updated"}`)
		requireStatus(t, w, http.StatusOK)

		var resp struct {
			EndedAt         *string `json:"ended_at"`
			EndedReasonCode *string `json:"ended_reason_code"`
			EndedReasonNote *string `json:"ended_reason_note"`
		}
		decodeJSON(t, w, &resp)
		if resp.EndedAt == nil || *resp.EndedAt == "" {
			t.Fatal("expected ended_at to be set")
		}
		if resp.EndedReasonCode == nil || *resp.EndedReasonCode != "contact_update" {
			t.Fatalf("expected ended_reason_code contact_update, got %v", resp.EndedReasonCode)
		}
		if resp.EndedReasonNote == nil || *resp.EndedReasonNote != "updated" {
			t.Fatalf("expected ended_reason_note updated, got %v", resp.EndedReasonNote)
		}
	})
}

func TestPeopleIdempotentChildMarkInactive(t *testing.T) {
	h := setupPeopleHarness(t)

	path := "/api/v1/children/" + h.scopeA.activeChildID.String() + "/actions/mark-inactive"
	w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"left_nursery","reason_note":"left"}`)
	requireStatus(t, w, http.StatusOK)

	var first struct {
		LeftAt         *string `json:"left_at"`
		LeftReasonCode *string `json:"left_reason_code"`
	}
	decodeJSON(t, w, &first)

	w = doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"duplicate_record"}`)
	requireStatus(t, w, http.StatusOK)

	var second struct {
		LeftAt         *string `json:"left_at"`
		LeftReasonCode *string `json:"left_reason_code"`
	}
	decodeJSON(t, w, &second)

	if first.LeftAt == nil || second.LeftAt == nil || *first.LeftAt != *second.LeftAt {
		t.Fatalf("expected left_at unchanged on idempotent call")
	}
	if first.LeftReasonCode == nil || second.LeftReasonCode == nil || *first.LeftReasonCode != *second.LeftReasonCode {
		t.Fatalf("expected left_reason_code unchanged on idempotent call")
	}
}

func TestPeopleIdempotentGuardianLifecycle(t *testing.T) {
	h := setupPeopleHarness(t)

	deactivatePath := "/api/v1/guardians/" + h.scopeA.activeGuardianID.String() + "/actions/deactivate"
	w := doRequest(t, h.router, http.MethodPost, deactivatePath, h.managerToken, `{"reason_code":"access_revoked"}`)
	requireStatus(t, w, http.StatusOK)

	firstLinkEnded := fetchEndedAt(t, h.pool, "guardian_child_links", h.scopeA.activeLinkID)
	firstMappingEnded := fetchEndedAt(t, h.pool, "parent_membership_guardians", h.scopeA.activeMappingID)

	w = doRequest(t, h.router, http.MethodPost, deactivatePath, h.managerToken, `{"reason_code":"access_revoked"}`)
	requireStatus(t, w, http.StatusOK)

	secondLinkEnded := fetchEndedAt(t, h.pool, "guardian_child_links", h.scopeA.activeLinkID)
	secondMappingEnded := fetchEndedAt(t, h.pool, "parent_membership_guardians", h.scopeA.activeMappingID)

	if !firstLinkEnded.Equal(secondLinkEnded) {
		t.Fatalf("expected link ended_at unchanged on repeated deactivate")
	}
	if !firstMappingEnded.Equal(secondMappingEnded) {
		t.Fatalf("expected mapping ended_at unchanged on repeated deactivate")
	}

	reactivatePath := "/api/v1/guardians/" + h.scopeA.activeGuardianID.String() + "/actions/reactivate"
	w = doRequest(t, h.router, http.MethodPost, reactivatePath, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	linkEnded := fetchEndedAt(t, h.pool, "guardian_child_links", h.scopeA.activeLinkID)
	mappingEnded := fetchEndedAt(t, h.pool, "parent_membership_guardians", h.scopeA.activeMappingID)
	if linkEnded.IsZero() || mappingEnded.IsZero() {
		t.Fatalf("expected links/mappings to remain ended after reactivation")
	}

	auditCount := countGuardianReactivatedAudits(t, h.pool, h.scopeA.activeGuardianID)
	if auditCount != 1 {
		t.Fatalf("expected 1 reactivation audit for inactive guardian, got %d", auditCount)
	}

	auditCount = countGuardianReactivatedAudits(t, h.pool, h.scopeA.otherGuardianID)
	if auditCount != 0 {
		t.Fatalf("expected no reactivation audits for active guardian, got %d", auditCount)
	}

	reactivateOther := "/api/v1/guardians/" + h.scopeA.otherGuardianID.String() + "/actions/reactivate"
	w = doRequest(t, h.router, http.MethodPost, reactivateOther, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	auditCount = countGuardianReactivatedAudits(t, h.pool, h.scopeA.otherGuardianID)
	if auditCount != 0 {
		t.Fatalf("expected no reactivation audits for already active guardian, got %d", auditCount)
	}
}

func TestPeopleIdempotentLinkAndMappingEnd(t *testing.T) {
	h := setupPeopleHarness(t)

	idempotentLinkID := uuid.MustParse("a7000000-0000-0000-0000-000000000002")
	dbtest.InsertGuardianLink(t, h.pool, idempotentLinkID, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.otherGuardianID, h.scopeA.activeChildID)

	idempotentMappingID := uuid.MustParse("a8000000-0000-0000-0000-000000000002")
	dbtest.InsertParentMapping(t, h.pool, idempotentMappingID, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID2, h.scopeA.otherGuardianID)

	endLinkPath := "/api/v1/guardian-child-links/" + idempotentLinkID.String() + "/actions/end"
	w := doRequest(t, h.router, http.MethodPost, endLinkPath, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)
	firstEnded := fetchEndedAt(t, h.pool, "guardian_child_links", idempotentLinkID)

	w = doRequest(t, h.router, http.MethodPost, endLinkPath, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)
	secondEnded := fetchEndedAt(t, h.pool, "guardian_child_links", idempotentLinkID)
	if !firstEnded.Equal(secondEnded) {
		t.Fatalf("expected link ended_at unchanged on repeated end")
	}

	endMappingPath := "/api/v1/parent-membership-guardian-mappings/" + idempotentMappingID.String() + "/actions/end"
	w = doRequest(t, h.router, http.MethodPost, endMappingPath, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)
	firstMappingEnded := fetchEndedAt(t, h.pool, "parent_membership_guardians", idempotentMappingID)

	w = doRequest(t, h.router, http.MethodPost, endMappingPath, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)
	secondMappingEnded := fetchEndedAt(t, h.pool, "parent_membership_guardians", idempotentMappingID)
	if !firstMappingEnded.Equal(secondMappingEnded) {
		t.Fatalf("expected mapping ended_at unchanged on repeated end")
	}
}

func TestPeopleIdempotentCreateLinkAndMapping(t *testing.T) {
	h := setupPeopleHarness(t)

	linkBody := `{"guardian_id":"` + h.scopeA.activeGuardianID.String() + `","child_id":"` + h.scopeA.activeChildID.String() + `"}`
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/guardian-child-links", h.managerToken, linkBody)
	requireStatus(t, w, http.StatusCreated)

	var linkResp struct {
		ID string `json:"id"`
	}
	decodeJSON(t, w, &linkResp)
	if linkResp.ID != h.scopeA.activeLinkID.String() {
		t.Fatalf("expected existing link id %s, got %s", h.scopeA.activeLinkID, linkResp.ID)
	}

	mappingBody := `{"membership_id":"` + h.scopeA.parentMembershipID.String() + `","guardian_id":"` + h.scopeA.activeGuardianID.String() + `"}`
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/parent-membership-guardian-mappings", h.managerToken, mappingBody)
	requireStatus(t, w, http.StatusCreated)

	var mappingResp struct {
		ID string `json:"id"`
	}
	decodeJSON(t, w, &mappingResp)
	if mappingResp.ID != h.scopeA.activeMappingID.String() {
		t.Fatalf("expected existing mapping id %s, got %s", h.scopeA.activeMappingID, mappingResp.ID)
	}

	conflictBody := `{"membership_id":"` + h.scopeA.parentMembershipID.String() + `","guardian_id":"` + h.scopeA.otherGuardianID.String() + `"}`
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/parent-membership-guardian-mappings", h.managerToken, conflictBody)
	requireStatus(t, w, http.StatusConflict)
	requireErrorCode(t, w, "parent_mapping_active_conflict")
}

func TestPeopleAccessChainLinkEnd(t *testing.T) {
	h := setupPeopleHarness(t)

	if !accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to exist before link end")
	}

	path := "/api/v1/guardian-child-links/" + h.scopeA.activeLinkID.String() + "/actions/end"
	w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)

	if accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to be broken after link end")
	}
}

func TestPeopleAccessChainMappingEnd(t *testing.T) {
	h := setupPeopleHarness(t)

	if !accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to exist before mapping end")
	}

	path := "/api/v1/parent-membership-guardian-mappings/" + h.scopeA.activeMappingID.String() + "/actions/end"
	w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"contact_update"}`)
	requireStatus(t, w, http.StatusOK)

	if accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to be broken after mapping end")
	}
}

func TestPeopleAccessChainGuardianDeactivate(t *testing.T) {
	h := setupPeopleHarness(t)

	if !accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to exist before guardian deactivation")
	}

	path := "/api/v1/guardians/" + h.scopeA.activeGuardianID.String() + "/actions/deactivate"
	w := doRequest(t, h.router, http.MethodPost, path, h.managerToken, `{"reason_code":"access_revoked"}`)
	requireStatus(t, w, http.StatusOK)

	if accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to be broken after guardian deactivation")
	}

	linkReasonCode, linkReasonNote := fetchEndReason(t, h.pool, "guardian_child_links", h.scopeA.activeLinkID)
	if linkReasonCode != "access_revoked" {
		t.Fatalf("expected link ended_reason_code access_revoked, got %s", linkReasonCode)
	}
	if linkReasonNote != "guardian_deactivated_cascade" {
		t.Fatalf("expected link ended_reason_note guardian_deactivated_cascade, got %s", linkReasonNote)
	}

	mappingReasonCode, mappingReasonNote := fetchEndReason(t, h.pool, "parent_membership_guardians", h.scopeA.activeMappingID)
	if mappingReasonCode != "access_revoked" {
		t.Fatalf("expected mapping ended_reason_code access_revoked, got %s", mappingReasonCode)
	}
	if mappingReasonNote != "guardian_deactivated_cascade" {
		t.Fatalf("expected mapping ended_reason_note guardian_deactivated_cascade, got %s", mappingReasonNote)
	}

	reactivate := "/api/v1/guardians/" + h.scopeA.activeGuardianID.String() + "/actions/reactivate"
	w = doRequest(t, h.router, http.MethodPost, reactivate, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	if accessChainExists(t, h.pool, h.scopeA.tenantID, h.scopeA.branchID, h.scopeA.parentMembershipID, h.scopeA.activeChildID) {
		t.Fatal("expected access chain to remain absent after reactivation")
	}
}

func setupPeopleHarness(t *testing.T) *peopleHarness {
	t.Helper()

	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	scopeA, scopeB := seedPeopleData(t, pool)

	cfg := testConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := Bootstrap(cfg, logger, pool)

	tokens := authtokens.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHours)

	h := &peopleHarness{
		router: router,
		pool:   pool,
		tokens: tokens,
		scopeA: scopeA,
		scopeB: scopeB,
	}
	h.managerToken = mustAccessToken(t, tokens, scopeA.managerUserID, scopeA.managerMembershipID, scopeA.tenantID, scopeA.branchID, "manager")
	h.practitionerToken = mustAccessToken(t, tokens, scopeA.practitionerUserID, scopeA.practitionerMID, scopeA.tenantID, scopeA.branchID, "practitioner")
	h.parentToken = mustAccessToken(t, tokens, scopeA.parentUserID, scopeA.parentMembershipID, scopeA.tenantID, scopeA.branchID, "parent")

	return h
}

func seedPeopleData(t *testing.T, pool *pgxpool.Pool) (peopleScope, peopleScopeB) {
	t.Helper()

	scopeA := peopleScope{
		tenantID:            uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		branchID:            uuid.MustParse("a2000000-0000-0000-0000-000000000001"),
		managerUserID:       uuid.MustParse("a3000000-0000-0000-0000-000000000001"),
		managerMembershipID: uuid.MustParse("a4000000-0000-0000-0000-000000000001"),
		practitionerUserID:  uuid.MustParse("a3000000-0000-0000-0000-000000000002"),
		practitionerMID:     uuid.MustParse("a4000000-0000-0000-0000-000000000002"),
		parentUserID:        uuid.MustParse("a3000000-0000-0000-0000-000000000003"),
		parentMembershipID:  uuid.MustParse("a4000000-0000-0000-0000-000000000003"),
		parentUserID2:       uuid.MustParse("a3000000-0000-0000-0000-000000000004"),
		parentMembershipID2: uuid.MustParse("a4000000-0000-0000-0000-000000000004"),
		activeChildID:       uuid.MustParse("a5000000-0000-0000-0000-000000000001"),
		inactiveChildID:     uuid.MustParse("a5000000-0000-0000-0000-000000000002"),
		activeGuardianID:    uuid.MustParse("a6000000-0000-0000-0000-000000000001"),
		inactiveGuardianID:  uuid.MustParse("a6000000-0000-0000-0000-000000000002"),
		otherGuardianID:     uuid.MustParse("a6000000-0000-0000-0000-000000000003"),
		activeLinkID:        uuid.MustParse("a7000000-0000-0000-0000-000000000001"),
		activeMappingID:     uuid.MustParse("a8000000-0000-0000-0000-000000000001"),
	}

	scopeB := peopleScopeB{
		tenantID:           uuid.MustParse("b1000000-0000-0000-0000-000000000001"),
		branchID:           uuid.MustParse("b2000000-0000-0000-0000-000000000001"),
		parentUserID:       uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
		parentMembershipID: uuid.MustParse("b4000000-0000-0000-0000-000000000001"),
		childID:            uuid.MustParse("b5000000-0000-0000-0000-000000000001"),
		guardianID:         uuid.MustParse("b6000000-0000-0000-0000-000000000001"),
		linkID:             uuid.MustParse("b7000000-0000-0000-0000-000000000001"),
		mappingID:          uuid.MustParse("b8000000-0000-0000-0000-000000000001"),
	}

	dbtest.InsertTenant(t, pool, scopeA.tenantID, "Scope A")
	dbtest.InsertBranch(t, pool, scopeA.tenantID, scopeA.branchID, "Scope A Branch")

	dbtest.InsertUser(t, pool, scopeA.managerUserID, "manager@example.com", "hash", true)
	dbtest.InsertUser(t, pool, scopeA.practitionerUserID, "practitioner@example.com", "hash", true)
	dbtest.InsertUser(t, pool, scopeA.parentUserID, "parent@example.com", "hash", true)
	dbtest.InsertUser(t, pool, scopeA.parentUserID2, "parent2@example.com", "hash", true)

	dbtest.InsertMembership(t, pool, scopeA.managerMembershipID, scopeA.tenantID, scopeA.branchID, scopeA.managerUserID, "manager", true)
	dbtest.InsertMembership(t, pool, scopeA.practitionerMID, scopeA.tenantID, scopeA.branchID, scopeA.practitionerUserID, "practitioner", true)
	dbtest.InsertMembership(t, pool, scopeA.parentMembershipID, scopeA.tenantID, scopeA.branchID, scopeA.parentUserID, "parent", true)
	dbtest.InsertMembership(t, pool, scopeA.parentMembershipID2, scopeA.tenantID, scopeA.branchID, scopeA.parentUserID2, "parent", true)

	dbtest.InsertChild(t, pool, scopeA.activeChildID, scopeA.tenantID, scopeA.branchID, "Active Child", dbtest.DateAt(2022, 3, 10), dbtest.DateAt(2024, 1, 1), 450, true)
	dbtest.InsertChild(t, pool, scopeA.inactiveChildID, scopeA.tenantID, scopeA.branchID, "Inactive Child", dbtest.DateAt(2021, 3, 10), dbtest.DateAt(2023, 1, 1), 450, true)

	_, err := pool.Exec(context.Background(),
		"UPDATE children SET is_active = false, left_at = now(), left_reason_code = 'left_nursery', left_reason_note = 'moved', updated_at = now() WHERE id = $1",
		scopeA.inactiveChildID,
	)
	if err != nil {
		t.Fatalf("seed inactive child: %v", err)
	}

	dbtest.InsertGuardian(t, pool, scopeA.activeGuardianID, scopeA.tenantID, scopeA.branchID, "Active Guardian", true)
	dbtest.InsertGuardian(t, pool, scopeA.inactiveGuardianID, scopeA.tenantID, scopeA.branchID, "Inactive Guardian", true)
	dbtest.InsertGuardian(t, pool, scopeA.otherGuardianID, scopeA.tenantID, scopeA.branchID, "Other Guardian", true)

	_, err = pool.Exec(context.Background(),
		"UPDATE guardians SET is_active = false, deactivated_at = now(), deactivation_reason_code = 'access_revoked', deactivation_reason_note = 'seeded', updated_at = now() WHERE id = $1",
		scopeA.inactiveGuardianID,
	)
	if err != nil {
		t.Fatalf("seed inactive guardian: %v", err)
	}

	dbtest.InsertGuardianLink(t, pool, scopeA.activeLinkID, scopeA.tenantID, scopeA.branchID, scopeA.activeGuardianID, scopeA.activeChildID)
	dbtest.InsertParentMapping(t, pool, scopeA.activeMappingID, scopeA.tenantID, scopeA.branchID, scopeA.parentMembershipID, scopeA.activeGuardianID)

	dbtest.InsertTenant(t, pool, scopeB.tenantID, "Scope B")
	dbtest.InsertBranch(t, pool, scopeB.tenantID, scopeB.branchID, "Scope B Branch")
	dbtest.InsertUser(t, pool, scopeB.parentUserID, "scopeb-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, scopeB.parentMembershipID, scopeB.tenantID, scopeB.branchID, scopeB.parentUserID, "parent", true)
	dbtest.InsertChild(t, pool, scopeB.childID, scopeB.tenantID, scopeB.branchID, "Scope B Child", dbtest.DateAt(2021, 4, 1), dbtest.DateAt(2024, 1, 1), 450, true)
	dbtest.InsertGuardian(t, pool, scopeB.guardianID, scopeB.tenantID, scopeB.branchID, "Scope B Guardian", true)
	dbtest.InsertGuardianLink(t, pool, scopeB.linkID, scopeB.tenantID, scopeB.branchID, scopeB.guardianID, scopeB.childID)
	dbtest.InsertParentMapping(t, pool, scopeB.mappingID, scopeB.tenantID, scopeB.branchID, scopeB.parentMembershipID, scopeB.guardianID)

	return scopeA, scopeB
}

func testConfig() config.Config {
	return config.Config{
		AppEnv:                       "local",
		APIPort:                      "8080",
		APIBasePath:                  "/api/v1",
		DatabaseURL:                  "postgres://unused",
		JWTAccessSecret:              "access-secret",
		JWTRefreshSecret:             "refresh-secret",
		JWTAccessTTLMin:              15,
		JWTRefreshTTLHours:           720,
		WebBaseURL:                   "http://example.local",
		EmailProvider:                "smtp",
		SMTPHost:                     "smtp.example.local",
		SMTPPort:                     1025,
		SMTPUser:                     "smtp-user",
		SMTPPass:                     "smtp-pass",
		SMTPFrom:                     "no-reply@example.local",
		PasswordResetTokenSecret:     "reset-secret",
		PasswordResetTokenTTLMinutes: 60,
		InviteTokenSecret:            "invite-secret",
		InviteTokenTTLHours:          168,
	}
}

func mustAccessToken(t *testing.T, tokens *authtokens.TokenManager, userID, membershipID, tenantID, branchID uuid.UUID, role string) string {
	t.Helper()

	raw, _, err := tokens.GenerateAccessToken(userID, "test@example.com", authdomain.ScopeClaims{
		MembershipID: membershipID.String(),
		TenantID:     tenantID.String(),
		BranchID:     branchID.String(),
		Role:         role,
	})
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	return raw
}

func doRequest(t *testing.T, router *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
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
	router.ServeHTTP(w, req)
	return w
}

func requireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("expected status %d, got %d", want, w.Code)
	}
}

func requireErrorCode(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()

	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, w, &resp)
	if resp.Code != want {
		t.Fatalf("expected error code %s, got %s", want, resp.Code)
	}
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func requireListIDs(t *testing.T, w *httptest.ResponseRecorder, want []string) {
	t.Helper()

	var resp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSON(t, w, &resp)

	have := make(map[string]struct{}, len(resp.Items))
	for _, item := range resp.Items {
		have[item.ID] = struct{}{}
	}

	if len(have) != len(want) {
		t.Fatalf("expected %d items, got %d", len(want), len(have))
	}
	for _, id := range want {
		if _, ok := have[id]; !ok {
			t.Fatalf("expected id %s in response", id)
		}
	}
}

func isPeopleRoute(path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/v1/children"),
		strings.HasPrefix(path, "/api/v1/guardians"),
		strings.HasPrefix(path, "/api/v1/guardian-child-links"),
		strings.HasPrefix(path, "/api/v1/parent-membership-guardian-mappings"):
		return true
	default:
		return false
	}
}

func fetchEndedAt(t *testing.T, pool *pgxpool.Pool, table string, id uuid.UUID) time.Time {
	t.Helper()

	var endedAt *time.Time
	err := pool.QueryRow(context.Background(), "SELECT ended_at FROM "+table+" WHERE id = $1", id).Scan(&endedAt)
	if err != nil {
		t.Fatalf("fetch ended_at from %s: %v", table, err)
	}
	if endedAt == nil {
		return time.Time{}
	}
	return endedAt.UTC()
}

func fetchEndReason(t *testing.T, pool *pgxpool.Pool, table string, id uuid.UUID) (string, string) {
	t.Helper()

	var reasonCode, reasonNote *string
	err := pool.QueryRow(context.Background(), "SELECT ended_reason_code::text, ended_reason_note FROM "+table+" WHERE id = $1", id).Scan(&reasonCode, &reasonNote)
	if err != nil {
		t.Fatalf("fetch end reason from %s: %v", table, err)
	}
	code := ""
	note := ""
	if reasonCode != nil {
		code = *reasonCode
	}
	if reasonNote != nil {
		note = *reasonNote
	}
	return code, note
}

func accessChainExists(t *testing.T, pool *pgxpool.Pool, tenantID, branchID, membershipID, childID uuid.UUID) bool {
	t.Helper()

	var exists bool
	err := pool.QueryRow(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM memberships m
    JOIN parent_membership_guardians pmg
      ON pmg.membership_id = m.id
     AND pmg.tenant_id = m.tenant_id
     AND pmg.branch_id = m.branch_id
     AND pmg.ended_at IS NULL
    JOIN guardian_child_links gcl
      ON gcl.guardian_id = pmg.guardian_id
     AND gcl.child_id = $4
     AND gcl.tenant_id = m.tenant_id
     AND gcl.branch_id = m.branch_id
     AND gcl.ended_at IS NULL
    WHERE m.id = $3
      AND m.tenant_id = $1
      AND m.branch_id = $2
      AND m.is_active = true
      AND m.role = 'parent'
)`, tenantID, branchID, membershipID, childID).Scan(&exists)
	if err != nil {
		t.Fatalf("access chain query: %v", err)
	}
	return exists
}

func countGuardianReactivatedAudits(t *testing.T, pool *pgxpool.Pool, guardianID uuid.UUID) int {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM audit_logs WHERE action_type = 'guardian_reactivated' AND action_entity_id = $1",
		guardianID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("audit log count: %v", err)
	}
	return count
}
