package pdf

import (
	"fmt"
	"strings"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
)

const (
	pageWidth    = 595.28 // A4
	pageHeight   = 841.89
	marginLeft   = 40.0
	marginRight  = 40.0
	contentWidth = pageWidth - marginLeft - marginRight

	headerY       = 35.0
	invoiceLabelY = 35.0
	separatorY    = 92.0
	detailsY      = 100.0
	billToY       = 142.0
	tableStartY   = 224.0

	footerReserve = 80.0
)

type InvoicePDFInput struct {
	SiteProfile     InvoicePDFSiteProfile
	Invoice         InvoicePDFMeta
	Parent          InvoicePDFParent
	Child           InvoicePDFChild
	Lines           []InvoicePDFLine
	SubtotalMinor   int
	DeductionMinor  int
	PaidMinor       int
	TotalMinor      int
	BalanceDueMinor int
	PaymentNote     string
}

type InvoicePDFSiteProfile struct {
	NurseryName     string
	Phone           string
	Email           string
	Website         string
	AddressStreet   string
	AddressCity     string
	AddressPostcode string
	RegNumber       string
	VATNumber       string
	BankName        string
	AccountName     string
	SortCode        string
	AccountNumber   string
}

type InvoicePDFMeta struct {
	InvoiceNumber   string
	BillingMonth    time.Time
	PeriodStartDate *time.Time
	PeriodEndDate   *time.Time
	IssueDate       *time.Time
	DueDate         *time.Time
	Status          string
}

type InvoicePDFParent struct {
	FullName        string
	AddressLine1    string
	AddressLine2    string
	AddressCity     string
	AddressPostcode string
	Email           string
	Phone           string
}

type InvoicePDFChild struct {
	ChildName string
	ChildID   string
	RoomName  string
}

type InvoicePDFLine struct {
	Description     string
	SubDescription  string
	SessionDate     string
	QuantityMinutes *int
	QuantityText    string
	SessionCount    *int
	UnitAmountMinor *int
	LineAmountMinor int
	IsFunded        bool
	IsDiscount      bool
	LineKind        string
}

func ManagerInput(sp *domain.InvoiceSiteProfile, inv domain.InvoiceReviewRow, lines []domain.InvoiceReviewLineRow, pc *domain.ParentContact, subtotal, deduction, total domain.Money) InvoicePDFInput {
	var startDate, endDate *time.Time
	if !inv.PeriodStartDate.IsZero() {
		t := inv.PeriodStartDate
		startDate = &t
	}
	if !inv.PeriodEndDate.IsZero() {
		t := inv.PeriodEndDate
		endDate = &t
	}

	childName := inv.ChildFirstName
	if inv.ChildMiddleName != nil && *inv.ChildMiddleName != "" {
		childName += " " + *inv.ChildMiddleName
	}
	if inv.ChildLastName != nil && *inv.ChildLastName != "" {
		childName += " " + *inv.ChildLastName
	}

	room := ""
	if inv.RoomName != nil {
		room = *inv.RoomName
	}

	totalVal := total.Minor()
	paidVal := inv.AmountPaid.Minor()
	balanceVal := totalVal - paidVal
	if balanceVal < 0 {
		balanceVal = 0
	}

	invNum := ptrStr(inv.InvoiceNumber)
	if invNum == "" {
		invNum = inv.ID.String()[:8]
	}

	inp := InvoicePDFInput{
		Invoice: InvoicePDFMeta{
			InvoiceNumber:   invNum,
			BillingMonth:    inv.BillingMonth,
			PeriodStartDate: startDate,
			PeriodEndDate:   endDate,
			IssueDate:       inv.IssuedAt,
			DueDate:         inv.DueAt,
			Status:          inv.Status,
		},
		Child: InvoicePDFChild{
			ChildName: childName,
			ChildID:   inv.ChildID.String()[:8],
			RoomName:  room,
		},
		Lines:           make([]InvoicePDFLine, len(lines)),
		SubtotalMinor:   subtotal.Minor(),
		DeductionMinor:  deduction.Minor(),
		PaidMinor:       paidVal,
		TotalMinor:      totalVal,
		BalanceDueMinor: balanceVal,
		PaymentNote:     inv.ParentNote,
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
			RegNumber:       "09876543",
			BankName:        "Barclays Bank UK",
			AccountName:     sp.NurseryName,
			SortCode:        "20-45-89",
			AccountNumber:   "83920147",
		}
	}
	if pc != nil {
		inp.Parent = InvoicePDFParent{
			FullName:        pc.FullName,
			AddressLine1:    pc.AddressLine1,
			AddressLine2:    pc.AddressLine2,
			AddressCity:     pc.AddressCity,
			AddressPostcode: pc.AddressPostcode,
			Email:           pc.Email,
			Phone:           pc.Telephone,
		}
	}
	for i, l := range lines {
		mainDesc, subDesc, isFunded, isDiscount := parseLineInfo(l.Description, l.LineKind, l.LineKind == "funded_deduction", l.LineAmount.Minor())
		inp.Lines[i] = InvoicePDFLine{
			Description:     mainDesc,
			SubDescription:  subDesc,
			QuantityMinutes: l.QuantityMinutes,
			SessionCount:    l.SessionCount,
			UnitAmountMinor: moneyPtrMinor(l.UnitAmount),
			LineAmountMinor: l.LineAmount.Minor(),
			IsFunded:        isFunded,
			IsDiscount:      isDiscount,
			LineKind:        l.LineKind,
		}
	}
	return inp
}

func ParentInput(sp *domain.ParentInvoiceSiteProfile, inv domain.ParentInvoiceRow, lines []domain.ParentInvoiceLineRow, subtotal, deduction, total domain.Money) InvoicePDFInput {
	var startDate, endDate *time.Time
	if !inv.PeriodStartDate.IsZero() {
		t := inv.PeriodStartDate
		startDate = &t
	}
	if !inv.PeriodEndDate.IsZero() {
		t := inv.PeriodEndDate
		endDate = &t
	}

	childName := inv.ChildFirstName
	if inv.ChildMiddleName != nil && *inv.ChildMiddleName != "" {
		childName += " " + *inv.ChildMiddleName
	}
	if inv.ChildLastName != nil && *inv.ChildLastName != "" {
		childName += " " + *inv.ChildLastName
	}

	totalVal := total.Minor()
	paidVal := inv.AmountPaid.Minor()
	balanceVal := totalVal - paidVal
	if balanceVal < 0 {
		balanceVal = 0
	}

	invNum := ptrStr(inv.InvoiceNumber)
	if invNum == "" {
		invNum = inv.ID.String()[:8]
	}

	inp := InvoicePDFInput{
		Invoice: InvoicePDFMeta{
			InvoiceNumber:   invNum,
			BillingMonth:    inv.BillingMonth,
			PeriodStartDate: startDate,
			PeriodEndDate:   endDate,
			IssueDate:       inv.IssuedAt,
			DueDate:         inv.DueAt,
			Status:          inv.Status,
		},
		Child: InvoicePDFChild{
			ChildName: childName,
			ChildID:   inv.ChildID.String()[:8],
		},
		Lines:           make([]InvoicePDFLine, len(lines)),
		SubtotalMinor:   subtotal.Minor(),
		DeductionMinor:  deduction.Minor(),
		PaidMinor:       paidVal,
		TotalMinor:      totalVal,
		BalanceDueMinor: balanceVal,
		PaymentNote:     "Please settle outstanding balances by the due date.",
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
			RegNumber:       "09876543",
			BankName:        "Barclays Bank UK",
			AccountName:     sp.NurseryName,
			SortCode:        "20-45-89",
			AccountNumber:   "83920147",
		}
	}
	for i, l := range lines {
		mainDesc, subDesc, isFunded, isDiscount := parseLineInfo(l.Description, l.LineKind, l.LineKind == "funded_deduction", l.LineAmount.Minor())
		inp.Lines[i] = InvoicePDFLine{
			Description:     mainDesc,
			SubDescription:  subDesc,
			QuantityMinutes: l.QuantityMinutes,
			UnitAmountMinor: moneyPtrMinor(l.UnitAmount),
			LineAmountMinor: l.LineAmount.Minor(),
			IsFunded:        isFunded,
			IsDiscount:      isDiscount,
			LineKind:        l.LineKind,
		}
	}
	return inp
}

func parseLineInfo(desc, lineKind string, isFunded bool, lineAmountMinor int) (mainDesc, subDesc string, funded, discount bool) {
	funded = isFunded || lineKind == "funded_deduction" || strings.Contains(strings.ToLower(desc), "funded") || strings.Contains(strings.ToLower(desc), "funding")
	discount = lineKind == "discount" || (lineAmountMinor < 0 && !funded) || strings.Contains(strings.ToLower(desc), "discount")

	mainDesc = desc
	subDesc = ""
	if idx := strings.Index(desc, "\n"); idx != -1 {
		mainDesc = strings.TrimSpace(desc[:idx])
		subDesc = strings.TrimSpace(desc[idx+1:])
	}
	return mainDesc, subDesc, funded, discount
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
