package pdf

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"github.com/signintech/gopdf"

	"nursery-management-system/api/internal/modules/billing/domain"
)

//go:embed fonts/DejaVuSans.ttf
var dejaVuSans []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejaVuSansBold []byte

type Renderer struct{}

func NewRenderer() (*Renderer, error) {
	return &Renderer{}, nil
}

func (r *Renderer) RenderManagerInvoice(ctx context.Context, sp *domain.InvoiceSiteProfile, inv domain.InvoiceReviewRow, lines []domain.InvoiceReviewLineRow, pc *domain.ParentContact, subtotal, deduction, total domain.Money) ([]byte, error) {
	input := ManagerInput(sp, inv, lines, pc, subtotal, deduction, total)
	return r.Render(ctx, input)
}

func (r *Renderer) RenderParentInvoice(ctx context.Context, sp *domain.ParentInvoiceSiteProfile, inv domain.ParentInvoiceRow, lines []domain.ParentInvoiceLineRow, subtotal, deduction, total domain.Money) ([]byte, error) {
	input := ParentInput(sp, inv, lines, subtotal, deduction, total)
	return r.Render(ctx, input)
}

func (r *Renderer) Render(ctx context.Context, input InvoicePDFInput) ([]byte, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	if err := pdf.AddTTFFontByReader("dejavu", bytes.NewReader(dejaVuSans)); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	if err := pdf.AddTTFFontByReader("dejavu-bold", bytes.NewReader(dejaVuSansBold)); err != nil {
		return nil, fmt.Errorf("load bold font: %w", err)
	}

	pdf.AddHeader(func() {
		r.drawHeader(pdf, input)
	})

	r.addPage(pdf)
	r.drawDetails(pdf, input)
	y := r.drawBillTo(pdf, input)
	y = r.drawLineItemsTable(pdf, input, y)
	y = r.drawCalculation(pdf, input, y)
	r.drawPaymentNote(pdf, input, y)

	return pdf.GetBytesPdfReturnErr()
}

func (r *Renderer) addPage(pdf *gopdf.GoPdf) {
	pdf.AddPage()
}

func (r *Renderer) drawHeader(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	pdf.SetTextColor(31, 41, 55)
	_ = pdf.SetFont("dejavu-bold", "", 18)
	pdf.SetXY(marginLeft, headerY)
	_ = pdf.Cell(nil, input.SiteProfile.NurseryName)

	pdf.SetTextColor(107, 114, 128)
	_ = pdf.SetFont("dejavu", "", 9)
	pdf.SetXY(marginLeft, headerY+24)
	_ = pdf.Cell(nil, input.SiteProfile.AddressStreet)
	pdf.SetXY(marginLeft, headerY+35)
	_ = pdf.Cell(nil, fmt.Sprintf("%s, %s", input.SiteProfile.AddressCity, input.SiteProfile.AddressPostcode))
	pdf.SetXY(marginLeft, headerY+46)
	_ = pdf.Cell(nil, fmt.Sprintf("Tel: %s", input.SiteProfile.Phone))
	pdf.SetXY(marginLeft, headerY+57)
	_ = pdf.Cell(nil, input.SiteProfile.Email)

	pdf.SetTextColor(70, 95, 255)
	_ = pdf.SetFont("dejavu-bold", "", 14)
	pdf.SetXY(pageWidth-marginRight-60, invoiceLabelY)
	_ = pdf.Cell(nil, "INVOICE")

	pdf.SetStrokeColor(229, 231, 235)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginLeft, separatorY, pageWidth-marginRight, separatorY)
}

func (r *Renderer) drawDetails(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	inv := input.Invoice
	_ = pdf.SetFont("dejavu", "", 9)

	fields := []struct {
		label string
		value string
	}{
		{"Invoice No.", inv.InvoiceNumber},
		{"Billing Month", formatMonth(inv.BillingMonth)},
		{"Issue Date", formatDate(inv.IssueDate)},
		{"Due Date", formatDate(inv.DueDate)},
		{"Status", inv.Status},
	}

	col1X := marginLeft
	col2X := marginLeft + 150.0
	y := detailsY

	for i, f := range fields {
		pdf.SetTextColor(107, 114, 128)
		pdf.SetXY(col1X, y)
		_ = pdf.Cell(nil, f.label)

		pdf.SetTextColor(31, 41, 55)
		_ = pdf.SetFont("dejavu-bold", "", 9)
		pdf.SetXY(col1X+80, y)
		_ = pdf.Cell(nil, f.value)
		_ = pdf.SetFont("dejavu", "", 9)

		if i%2 == 0 {
			y = detailsY
			col1X = col2X
		} else {
			y += 16
			col1X = marginLeft
		}
	}
}

func (r *Renderer) drawBillTo(pdf *gopdf.GoPdf, input InvoicePDFInput) float64 {
	y := billToY
	p := input.Parent

	pdf.SetTextColor(70, 95, 255)
	_ = pdf.SetFont("dejavu-bold", "", 8)
	pdf.SetXY(marginLeft, y)
	_ = pdf.Cell(nil, "BILL TO")
	y += 14

	pdf.SetTextColor(31, 41, 55)
	_ = pdf.SetFont("dejavu-bold", "", 10)
	pdf.SetXY(marginLeft, y)
	_ = pdf.Cell(nil, p.FullName)
	y += 14

	_ = pdf.SetFont("dejavu", "", 9)
	pdf.SetTextColor(107, 114, 128)

	if p.AddressLine1 != "" {
		pdf.SetXY(marginLeft, y)
		_ = pdf.Cell(nil, p.AddressLine1)
		y += 12
	}
	if p.AddressLine2 != "" {
		pdf.SetXY(marginLeft, y)
		_ = pdf.Cell(nil, p.AddressLine2)
		y += 12
	}
	if p.AddressCity != "" || p.AddressPostcode != "" {
		pdf.SetXY(marginLeft, y)
		_ = pdf.Cell(nil, fmt.Sprintf("%s, %s", p.AddressCity, p.AddressPostcode))
		y += 12
	}

	return y + 10
}

func (r *Renderer) drawLineItemsTable(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	maxRowsPerPage := int((pageHeight - y - footerReserve) / 20)
	if maxRowsPerPage < 1 {
		maxRowsPerPage = 1
	}

	table := pdf.NewTableLayout(marginLeft, y, 20, maxRowsPerPage)

	table.AddColumn("Description", contentWidth*0.45, "left")
	table.AddColumn("Qty", contentWidth*0.15, "center")
	table.AddColumn("Unit Price", contentWidth*0.2, "right")
	table.AddColumn("Total", contentWidth*0.2, "right")

	headerStyle := gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Top: true, Left: false, Right: false, Bottom: true,
			Width: 0.5, RGBColor: gopdf.RGBColor{R: 229, G: 231, B: 235},
		},
		FillColor: gopdf.RGBColor{R: 249, G: 250, B: 251},
		TextColor: gopdf.RGBColor{R: 107, G: 114, B: 128},
		Font:      "dejavu-bold",
		FontSize:  8,
	}
	table.SetHeaderStyle(headerStyle)

	cellStyle := gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Top: false, Left: false, Right: false, Bottom: true,
			Width: 0.3, RGBColor: gopdf.RGBColor{R: 243, G: 244, B: 246},
		},
		TextColor: gopdf.RGBColor{R: 31, G: 41, B: 55},
		Font:      "dejavu",
		FontSize:  9,
	}
	table.SetCellStyle(cellStyle)

	table.SetTableStyle(gopdf.CellStyle{
		Font:     "dejavu",
		FontSize: 9,
	})

	for _, line := range input.Lines {
		qty := ""
		if line.QuantityMinutes != nil {
			qty = fmt.Sprintf("%d min", *line.QuantityMinutes)
		} else if line.SessionCount != nil {
			qty = fmt.Sprintf("%d", *line.SessionCount)
		}

		unitStr := "\u2014"
		if line.UnitAmountMinor != nil {
			unitStr = formatMoney(*line.UnitAmountMinor)
		}

		totalStr := formatMoney(line.LineAmountMinor)
		if line.IsFunded && line.LineAmountMinor > 0 {
			totalStr = "-" + totalStr
		}

		table.AddRow([]string{line.Description, qty, unitStr, totalStr})
	}

	if err := table.DrawTable(); err != nil {
		return y + float64(len(input.Lines))*20 + 20
	}

	return y + float64(len(input.Lines)+1)*20 + 10
}

func (r *Renderer) drawCalculation(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	rightX := pageWidth - marginRight - 160

	drawCalcRow := func(label, value string, isBold, isHighlight bool) {
		if isBold {
			_ = pdf.SetFont("dejavu-bold", "", 10)
		} else {
			_ = pdf.SetFont("dejavu", "", 9)
		}

		if isHighlight {
			pdf.SetTextColor(70, 95, 255)
		} else {
			pdf.SetTextColor(107, 114, 128)
		}
		pdf.SetXY(rightX, y)
		_ = pdf.Cell(nil, label)

		if isHighlight {
			pdf.SetTextColor(70, 95, 255)
		} else {
			pdf.SetTextColor(31, 41, 55)
		}
		pdf.SetXY(rightX+100, y)
		_ = pdf.Cell(nil, value)
		y += 18
	}

	drawCalcRow("Subtotal", formatMoney(input.SubtotalMinor), false, false)

	if input.DeductionMinor > 0 {
		drawCalcRow("Funded Deduction", "-"+formatMoney(input.DeductionMinor), false, false)
	}

	y += 4
	pdf.SetStrokeColor(229, 231, 235)
	pdf.SetLineWidth(0.5)
	pdf.Line(rightX, y, rightX+160, y)
	y += 10

	drawCalcRow("Total Due", formatMoney(input.TotalMinor), true, true)

	return y
}

func (r *Renderer) drawPaymentNote(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) {
	if input.PaymentNote == "" {
		return
	}
	y := startY + 20
	if y > pageHeight-60 {
		return
	}

	pdf.SetTextColor(107, 114, 128)
	_ = pdf.SetFont("dejavu", "", 8)
	pdf.SetXY(marginLeft, y)
	_ = pdf.Cell(nil, input.PaymentNote)
}
