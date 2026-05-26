package bootstrap

import (
	"net/http"
	"testing"
)

func TestAttendanceRouteInventory(t *testing.T) {
	h := setupPeopleHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"POST /api/v1/attendance/check-ins",
		"POST /api/v1/attendance/check-outs",
		"POST /api/v1/attendance/corrections",
	}

	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected attendance route %s to be registered", want)
		}
	}
}

func TestAttendanceCheckInOutRoleGuards(t *testing.T) {
	h := setupPeopleHarness(t)

	t.Run("check-in requires auth", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-ins", "", `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusUnauthorized)
		requireErrorCode(t, w, "unauthorized")
	})

	t.Run("check-in rejects parent", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-ins", h.parentToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	})

	t.Run("check-out rejects parent", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-outs", h.parentToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	})

	t.Run("check-in accepts manager", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-ins", h.managerToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusCreated)
	})

	t.Run("check-out accepts manager", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-outs", h.managerToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusOK)
	})
}

func TestAttendanceCheckInOutPractitioner(t *testing.T) {
	h := setupPeopleHarness(t)

	t.Run("check-in accepts practitioner", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-ins", h.practitionerToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusCreated)
	})

	t.Run("check-out accepts practitioner", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/check-outs", h.practitionerToken, `{"child_id":"`+h.scopeA.activeChildID.String()+`"}`)
		requireStatus(t, w, http.StatusOK)
	})
}

func TestAttendanceCorrectionRoleGuards(t *testing.T) {
	h := setupPeopleHarness(t)

	correctionBody := `{"session_id":"` + h.scopeA.activeChildID.String() + `","check_in_at":"2026-05-26T08:00:00Z","check_out_at":"2026-05-26T16:00:00Z","reason_code":"incorrect_time"}`

	t.Run("correction requires auth", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/corrections", "", correctionBody)
		requireStatus(t, w, http.StatusUnauthorized)
		requireErrorCode(t, w, "unauthorized")
	})

	t.Run("correction rejects practitioner", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/corrections", h.practitionerToken, correctionBody)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	})

	t.Run("correction rejects parent", func(t *testing.T) {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/attendance/corrections", h.parentToken, correctionBody)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	})
}
