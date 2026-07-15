package httpbilling

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/billing/application"
	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type stubExportUseCase struct {
	err error
}

func (s *stubExportUseCase) Execute(_ context.Context, _ tenant.ActorContext, w io.Writer, _ application.ExportInvoicesParams) error {
	if s.err != nil {
		return s.err
	}
	csvWriter := csv.NewWriter(w)
	csvWriter.Write([]string{"invoice_number", "child_name", "billing_month", "status", "subtotal_minor", "funded_deduction_minor", "total_due_minor", "issued_at", "due_date", "paid_at"})
	csvWriter.Write([]string{"INV-001", "Test Child", "2026-07", "issued", "10000", "0", "10000", "2026-07-01T00:00:00Z", "2026-07-15T00:00:00Z", ""})
	csvWriter.Flush()
	return nil
}

type stubSummaryUseCase struct {
	result application.InvoiceSummaryResult
	err    error
}

func (s *stubSummaryUseCase) Execute(_ context.Context, _ tenant.ActorContext, _ application.InvoiceSummaryParams) (application.InvoiceSummaryResult, error) {
	return s.result, s.err
}

func setupExportRouter(exportUC *stubExportUseCase, summaryUC *stubSummaryUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(tenant.AuthContextKey, tenant.AuthorizationContext{
			UserID:       "00000000-0000-0000-0000-000000000001",
			MembershipID: "00000000-0000-0000-0000-000000000002",
			TenantID:     "00000000-0000-0000-0000-000000000003",
			BranchID:     "00000000-0000-0000-0000-000000000004",
			Role:         "manager",
		})
		c.Next()
	})
	r.Use(httpserver.RequestIDMiddleware())

	manager := r.Group("/api/v1/manager")
	manager.GET("/invoices/export", func(c *gin.Context) {
		actor, ok := tenant.ActorFromGinContext(c)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}
		format := c.DefaultQuery("format", "csv")
		if format != "csv" && format != "csv-detail" {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Format must be 'csv' or 'csv-detail'.", nil)
			return
		}
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=\"invoices-export.csv\"")
		err := exportUC.Execute(c.Request.Context(), actor, c.Writer, application.ExportInvoicesParams{
			BillingMonthFrom: queryParamPtr(c, "billing_month_from"),
			BillingMonthTo:   queryParamPtr(c, "billing_month_to"),
			Statuses:         queryParamPtr(c, "status"),
			Format:           format,
		})
		if err != nil {
			httpserver.WriteMappedError(c, nil, err)
		}
	})

	manager.GET("/invoices/summary", func(c *gin.Context) {
		actor, ok := tenant.ActorFromGinContext(c)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}
		result, err := summaryUC.Execute(c.Request.Context(), actor, application.InvoiceSummaryParams{
			BillingMonthFrom: queryParamPtr(c, "billing_month_from"),
			BillingMonthTo:   queryParamPtr(c, "billing_month_to"),
		})
		if err != nil {
			httpserver.WriteMappedError(c, nil, err)
			return
		}
		type monthSummaryResponse struct {
			BillingMonth          string `json:"billing_month"`
			TotalInvoicedMinor    int    `json:"total_invoiced_minor"`
			TotalCollectedMinor   int    `json:"total_collected_minor"`
			TotalOutstandingMinor int    `json:"total_outstanding_minor"`
			TotalOverdueMinor     int    `json:"total_overdue_minor"`
			InvoiceCount          int    `json:"invoice_count"`
		}
		months := make([]monthSummaryResponse, 0, len(result.Months))
		for _, m := range result.Months {
			months = append(months, monthSummaryResponse{
				BillingMonth:          m.BillingMonth.Format("2006-01"),
				TotalInvoicedMinor:    m.TotalInvoicedMinor,
				TotalCollectedMinor:   m.TotalCollectedMinor,
				TotalOutstandingMinor: m.TotalOutstandingMinor,
				TotalOverdueMinor:     m.TotalOverdueMinor,
				InvoiceCount:          m.InvoiceCount,
			})
		}
		c.JSON(http.StatusOK, gin.H{"months": months})
	})

	return r
}

func TestExport_CSV_Success(t *testing.T) {
	r := setupExportRouter(&stubExportUseCase{}, &stubSummaryUseCase{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/export?format=csv", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("expected text/csv, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "invoices-export.csv") {
		t.Fatalf("expected attachment filename, got %s", cd)
	}

	body := w.Body.String()
	if !strings.Contains(body, "invoice_number") {
		t.Fatal("expected CSV header in body")
	}
	if !strings.Contains(body, "INV-001") {
		t.Fatal("expected CSV data row in body")
	}
}

func TestExport_InvalidFormat_400(t *testing.T) {
	r := setupExportRouter(&stubExportUseCase{}, &stubSummaryUseCase{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/export?format=pdf", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestExport_ValidationError_400(t *testing.T) {
	exportUC := &stubExportUseCase{
		err: domainerrors.Validation("Invalid billing_month_from format. Use YYYY-MM.", "billing_month_from"),
	}
	r := setupExportRouter(exportUC, &stubSummaryUseCase{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/export?billing_month_from=invalid", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSummary_Success(t *testing.T) {
	summaryUC := &stubSummaryUseCase{
		result: application.InvoiceSummaryResult{
			Months: []domain.InvoiceMonthSummary{
				{
					BillingMonth:          time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
					TotalInvoicedMinor:    50000,
					TotalCollectedMinor:   30000,
					TotalOutstandingMinor: 20000,
					TotalOverdueMinor:     5000,
					InvoiceCount:          10,
				},
			},
		},
	}
	r := setupExportRouter(&stubExportUseCase{}, summaryUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/summary", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Months []struct {
			BillingMonth          string `json:"billing_month"`
			TotalInvoicedMinor    int    `json:"total_invoiced_minor"`
			TotalCollectedMinor   int    `json:"total_collected_minor"`
			TotalOutstandingMinor int    `json:"total_outstanding_minor"`
			TotalOverdueMinor     int    `json:"total_overdue_minor"`
			InvoiceCount          int    `json:"invoice_count"`
		} `json:"months"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(resp.Months) != 1 {
		t.Fatalf("expected 1 month, got %d", len(resp.Months))
	}
	if resp.Months[0].BillingMonth != "2026-07" {
		t.Fatalf("expected 2026-07, got %s", resp.Months[0].BillingMonth)
	}
	if resp.Months[0].TotalInvoicedMinor != 50000 {
		t.Fatalf("expected 50000, got %d", resp.Months[0].TotalInvoicedMinor)
	}
	if resp.Months[0].InvoiceCount != 10 {
		t.Fatalf("expected 10, got %d", resp.Months[0].InvoiceCount)
	}
}

func TestSummary_EmptyResult(t *testing.T) {
	r := setupExportRouter(&stubExportUseCase{}, &stubSummaryUseCase{
		result: application.InvoiceSummaryResult{Months: []domain.InvoiceMonthSummary{}},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/summary", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Months []domain.InvoiceMonthSummary `json:"months"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(resp.Months) != 0 {
		t.Fatalf("expected 0 months, got %d", len(resp.Months))
	}
}

func TestSummary_ValidationError_400(t *testing.T) {
	summaryUC := &stubSummaryUseCase{
		err: domainerrors.Validation("Date range exceeds maximum of 24 months.", "billing_month_from"),
	}
	r := setupExportRouter(&stubExportUseCase{}, summaryUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/manager/invoices/summary?billing_month_from=2024-01&billing_month_to=2026-12", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}
