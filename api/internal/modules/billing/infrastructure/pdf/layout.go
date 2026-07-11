package pdf

import (
	"fmt"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
)

const (
	pageWidth    = 595.28 // A4
	pageHeight   = 841.89
	marginLeft   = 40.0
	marginRight  = 40.0
	contentWidth = pageWidth - marginLeft - marginRight

	headerY       = 40.0
	invoiceLabelY = 50.0
	separatorY    = 80.0
	detailsY      = 95.0
	billToY       = 145.0
	tableStartY   = 210.0

	footerReserve = 80.0
)

type InvoicePDFInput struct {
	SiteProfile    InvoicePDFSiteProfile
	Invoice        InvoicePDFMeta
	Parent         InvoicePDFParent
	Lines          []InvoicePDFLine
	SubtotalMinor  int
	DeductionMinor int
	TotalMinor     int
	PaymentNote    string
}

type InvoicePDFSiteProfile struct {
	NurseryName     string
	Phone           string
	Email           string
	Website         string
	AddressStreet   string
	AddressCity     string
	AddressPostcode string
}

type InvoicePDFMeta struct {
	InvoiceNumber string
	BillingMonth  time.Time
	IssueDate     *time.Time
	DueDate       *time.Time
	Status        string
}

type InvoicePDFParent struct {
	FullName        string
	AddressLine1    string
	AddressLine2    string
	AddressCity     string
	AddressPostcode string
}

type InvoicePDFLine struct {
	Description     string
	QuantityMinutes *int
	SessionCount    *int
	UnitAmountMinor *int
	LineAmountMinor int
	IsFunded        bool
}

func ManagerInput(sp *domain.InvoiceSiteProfile, inv domain.InvoiceReviewRow, lines []domain.InvoiceReviewLineRow, pc *domain.ParentContact, subtotal, deduction, total domain.Money) InvoicePDFInput {
	inp := InvoicePDFInput{
		Invoice: InvoicePDFMeta{
			InvoiceNumber: ptrStr(inv.InvoiceNumber),
			BillingMonth:  inv.BillingMonth,
			IssueDate:     inv.IssuedAt,
			DueDate:       inv.DueAt,
			Status:        inv.Status,
		},
		Lines:          make([]InvoicePDFLine, len(lines)),
		SubtotalMinor:  subtotal.Minor(),
		DeductionMinor: deduction.Minor(),
		TotalMinor:     total.Minor(),
		PaymentNote:    "Please settle outstanding balances by the due date.",
	}
	if sp != nil {
		inp.SiteProfile = InvoicePDFSiteProfile{
			NurseryName:     sp.NurseryName,
			Phone:           sp.Phone,
			Email:           sp.Email,
			Website:         sp.Website,
			AddressStreet:   sp.AddressStreet,
			AddressCity:     sp.AddressCity,
			AddressPostcode: sp.AddressPostcode,
		}
	}
	if pc != nil {
		inp.Parent = InvoicePDFParent{
			FullName:        pc.FullName,
			AddressLine1:    pc.AddressLine1,
			AddressLine2:    pc.AddressLine2,
			AddressCity:     pc.AddressCity,
			AddressPostcode: pc.AddressPostcode,
		}
	}
	for i, l := range lines {
		inp.Lines[i] = InvoicePDFLine{
			Description:     l.Description,
			QuantityMinutes: l.QuantityMinutes,
			SessionCount:    l.SessionCount,
			UnitAmountMinor: moneyPtrMinor(l.UnitAmount),
			LineAmountMinor: l.LineAmount.Minor(),
			IsFunded:        l.LineKind == "funded_deduction",
		}
	}
	return inp
}

func ParentInput(sp *domain.ParentInvoiceSiteProfile, inv domain.ParentInvoiceRow, lines []domain.ParentInvoiceLineRow, subtotal, deduction, total domain.Money) InvoicePDFInput {
	inp := InvoicePDFInput{
		Invoice: InvoicePDFMeta{
			InvoiceNumber: ptrStr(inv.InvoiceNumber),
			BillingMonth:  inv.BillingMonth,
			IssueDate:     inv.IssuedAt,
			DueDate:       inv.DueAt,
			Status:        inv.Status,
		},
		Lines:          make([]InvoicePDFLine, len(lines)),
		SubtotalMinor:  subtotal.Minor(),
		DeductionMinor: deduction.Minor(),
		TotalMinor:     total.Minor(),
		PaymentNote:    "Please settle outstanding balances by the due date.",
	}
	if sp != nil {
		inp.SiteProfile = InvoicePDFSiteProfile{
			NurseryName:     sp.NurseryName,
			Phone:           sp.Phone,
			Email:           sp.Email,
			Website:         sp.Website,
			AddressStreet:   sp.AddressStreet,
			AddressCity:     sp.AddressCity,
			AddressPostcode: sp.AddressPostcode,
		}
	}
	for i, l := range lines {
		inp.Lines[i] = InvoicePDFLine{
			Description:     l.Description,
			QuantityMinutes: l.QuantityMinutes,
			UnitAmountMinor: moneyPtrMinor(l.UnitAmount),
			LineAmountMinor: l.LineAmount.Minor(),
			IsFunded:        l.LineKind == "funded_deduction",
		}
	}
	return inp
}

func formatMoney(minor int) string {
	negative := minor < 0
	if negative {
		minor = -minor
	}
	pounds := minor / 100
	pence := minor % 100
	if negative {
		return fmt.Sprintf("-\u00a3%d.%02d", pounds, pence)
	}
	return fmt.Sprintf("\u00a3%d.%02d", pounds, pence)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "\u2014"
	}
	return t.Format("02 Jan 2006")
}

func formatMonth(t time.Time) string {
	return t.Format("January 2006")
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func moneyPtrMinor(m *domain.Money) *int {
	if m == nil {
		return nil
	}
	v := m.Minor()
	return &v
}
