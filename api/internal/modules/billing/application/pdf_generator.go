package application

import (
	"context"
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"nursery-management-system/api/internal/modules/billing/domain"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
)

type InvoicePDFData struct {
	Invoice     domain.Invoice
	SiteProfile siteprofiledomain.SiteProfile
	ParentName  string
	ChildName   string
	CheckoutURL string
}

type InvoicePDFGenerator struct{}

func NewInvoicePDFGenerator() *InvoicePDFGenerator {
	return &InvoicePDFGenerator{}
}

func (g *InvoicePDFGenerator) Generate(ctx context.Context, data InvoicePDFData) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	g.addHeader(m, data.SiteProfile)
	g.addInvoiceMeta(m, data.Invoice)
	g.addBillTo(m, data.ParentName, data.ChildName)
	g.addLineItemsTable(m, data.Invoice.Lines)
	g.addTotal(m, data.Invoice.TotalMinor)
	g.addPaymentLink(m, data.CheckoutURL)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

func (g *InvoicePDFGenerator) addHeader(m core.Maroto, sp siteprofiledomain.SiteProfile) {
	m.AddRows(
		text.NewRow(10, sp.NurseryName, props.Text{Size: 18, Style: fontstyle.Bold, Align: align.Left}),
	)
	if sp.AddressStreet != "" || sp.AddressCity != "" {
		addr := sp.AddressStreet
		if sp.AddressCity != "" {
			if addr != "" {
				addr += ", "
			}
			addr += sp.AddressCity
		}
		if sp.AddressPostcode != "" {
			if addr != "" {
				addr += " "
			}
			addr += sp.AddressPostcode
		}
		m.AddRows(
			text.NewRow(6, addr, props.Text{Size: 9, Align: align.Left}),
		)
	}
	if sp.Phone != "" || sp.Email != "" {
		contact := ""
		if sp.Phone != "" {
			contact = sp.Phone
		}
		if sp.Email != "" {
			if contact != "" {
				contact += "  |  "
			}
			contact += sp.Email
		}
		m.AddRows(
			text.NewRow(6, contact, props.Text{Size: 9, Align: align.Left}),
		)
	}
}

func (g *InvoicePDFGenerator) addInvoiceMeta(m core.Maroto, inv domain.Invoice) {
	m.AddRows(
		text.NewRow(8, "INVOICE", props.Text{Size: 14, Style: fontstyle.Bold, Align: align.Left}),
	)

	issuedStr := inv.IssuedAt.Format("2 January 2006")
	dueStr := inv.DueDate.Format("2 January 2006")
	meta := fmt.Sprintf("Invoice # %s    |    Issued: %s    |    Due: %s", inv.InvoiceNumber, issuedStr, dueStr)
	m.AddRows(
		text.NewRow(6, meta, props.Text{Size: 9, Align: align.Left}),
	)
}

func (g *InvoicePDFGenerator) addBillTo(m core.Maroto, parentName, childName string) {
	m.AddRows(
		text.NewRow(6, "Bill To", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Left}),
	)
	if parentName != "" {
		m.AddRows(
			text.NewRow(5, parentName, props.Text{Size: 9, Align: align.Left}),
		)
	}
	if childName != "" {
		m.AddRows(
			text.NewRow(5, "Child: "+childName, props.Text{Size: 9, Align: align.Left}),
		)
	}
}

func (g *InvoicePDFGenerator) addLineItemsTable(m core.Maroto, lines []domain.InvoiceLine) {
	m.AddRows(
		text.NewRow(7, "Description                                Qty       Rate      Amount", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
	)

	for _, line := range lines {
		desc := line.Description
		if desc == "" {
			desc = line.LineKind
		}
		qty := fmt.Sprintf("%d min", line.QuantityMinutes)
		rate := formatPDFMoney(line.UnitAmount.Minor())
		amount := formatPDFMoney(line.LineAmount.Minor())

		rowText := fmt.Sprintf("%-40s %8s %10s %10s", desc, qty, rate, amount)
		m.AddRows(
			text.NewRow(5, rowText, props.Text{Size: 8, Align: align.Left}),
		)
	}
}

func (g *InvoicePDFGenerator) addTotal(m core.Maroto, totalMinor int) {
	totalText := fmt.Sprintf("Total Due: %s", formatPDFMoney(totalMinor))
	m.AddRows(
		text.NewRow(8, totalText, props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Right}),
	)
}

func (g *InvoicePDFGenerator) addPaymentLink(m core.Maroto, checkoutURL string) {
	if checkoutURL == "" {
		return
	}
	m.AddRows(
		text.NewRow(6, "Pay online:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
	)
	m.AddRows(
		text.NewRow(5, checkoutURL, props.Text{Size: 8, Align: align.Left}),
	)
}

func formatPDFMoney(minor int) string {
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
