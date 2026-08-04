package pdf

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"

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
		r.drawFooterBar(pdf, input)
	})

	r.addPage(pdf, input)
	r.drawHeaderSection(pdf, input)
	r.drawSupplierCustomerSection(pdf, input)
	y := r.drawLineItemsTable(pdf, input, tableStartY)
	y = r.drawTotalsSummarySection(pdf, input, y)
	r.drawPaymentAndNotesSection(pdf, input, y)

	return pdf.GetBytesPdfReturnErr()
}

func (r *Renderer) addPage(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	pdf.AddPage()
	r.drawFooterBar(pdf, input)
}

func (r *Renderer) drawFooterBar(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	// Bottom fixed footer bar with light grey background (#f1f5f9) across full width
	pdf.SetFillColor(241, 245, 249)
	pdf.RectFromUpperLeftWithStyle(0, pageHeight-28.0, pageWidth, 28.0, "F")

	// Centered footer text: Supplier Name | Email | Phone
	sName := input.SiteProfile.NurseryName
	if sName == "" {
		sName = "NurseryPro Site"
	}
	sEmail := input.SiteProfile.Email
	if sEmail == "" {
		sEmail = "info@nurserypro.co.uk"
	}
	sPhone := input.SiteProfile.Phone
	if sPhone == "" {
		sPhone = "+44 20 7123 4567"
	}

	footerText := fmt.Sprintf("%s   |   %s   |   %s", sName, sEmail, sPhone)
	pdf.SetTextColor(100, 116, 139) // slate-500
	_ = pdf.SetFont("dejavu", "", 8.0)
	tw, _ := pdf.MeasureTextWidth(footerText)
	pdf.SetXY((pageWidth-tw)/2.0, pageHeight-18.0)
	_ = pdf.Cell(nil, footerText)
}

func (r *Renderer) drawHeaderSection(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	rightMarginX := pageWidth - marginRight

	// 1. Left Brand Header: Logo / Title
	logoX := marginLeft
	logoY := headerY + 2.0
	logoR := 8.0

	// Logo Circle (#5c6ac4)
	pdf.SetFillColor(92, 106, 196)
	pdf.Oval(logoX, logoY, logoX+(logoR*2), logoY+(logoR*2))

	// Smiley face details
	pdf.SetFillColor(255, 255, 255)
	pdf.Oval(logoX+4.0, logoY+5.0, logoX+6.2, logoY+7.2)
	pdf.Oval(logoX+9.8, logoY+5.0, logoX+12.0, logoY+7.2)
	pdf.SetStrokeColor(255, 255, 255)
	pdf.SetLineWidth(0.9)
	pdf.Line(logoX+4.8, logoY+10.8, logoX+6.8, logoY+12.2)
	pdf.Line(logoX+6.8, logoY+12.2, logoX+9.2, logoY+12.2)
	pdf.Line(logoX+9.2, logoY+12.2, logoX+11.2, logoY+10.8)

	// Brand Text
	textX := logoX + (logoR * 2) + 6.0
	pdf.SetXY(textX, logoY-2.0)
	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 15)
	_ = pdf.Cell(nil, "Nursery")

	nw, _ := pdf.MeasureTextWidth("Nursery")
	pdf.SetXY(textX+nw, logoY-2.0)
	pdf.SetTextColor(92, 106, 196) // #5c6ac4
	_ = pdf.Cell(nil, "Pro")

	// 2. Right Header Box: Date & Invoice # (matching Templid layout with border-r divider)
	invNum := input.Invoice.InvoiceNumber
	if invNum == "" {
		invNum = "INV-2024-0892"
	}

	dateStr := formatDate(input.Invoice.IssueDate)
	if input.Invoice.IssueDate == nil {
		dateStr = "26 Apr 2023"
	}

	divX := rightMarginX - 95.0

	// Date Column (left of divider)
	pdf.SetTextColor(148, 163, 184) // text-slate-400
	_ = pdf.SetFont("dejavu", "", 8.0)
	lblDateW, _ := pdf.MeasureTextWidth("Date")
	pdf.SetXY(divX-12.0-lblDateW, headerY)
	_ = pdf.Cell(nil, "Date")

	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 9.5)
	valDateW, _ := pdf.MeasureTextWidth(dateStr)
	pdf.SetXY(divX-12.0-valDateW, headerY+12.0)
	_ = pdf.Cell(nil, dateStr)

	// Vertical Divider Line (border-r)
	pdf.SetStrokeColor(203, 213, 225) // border-slate-300
	pdf.SetLineWidth(0.5)
	pdf.Line(divX, headerY+2.0, divX, headerY+24.0)

	// Invoice # Column (right of divider)
	pdf.SetTextColor(148, 163, 184) // text-slate-400
	_ = pdf.SetFont("dejavu", "", 8.0)
	lblInvW, _ := pdf.MeasureTextWidth("Invoice #")
	pdf.SetXY(rightMarginX-lblInvW, headerY)
	_ = pdf.Cell(nil, "Invoice #")

	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 9.5)
	valInvW, _ := pdf.MeasureTextWidth(invNum)
	pdf.SetXY(rightMarginX-valInvW, headerY+12.0)
	_ = pdf.Cell(nil, invNum)
}

func (r *Renderer) drawSupplierCustomerSection(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	boxY := detailsY
	boxH := 92.0
	boxW := contentWidth

	// 1. Background Box: Slate-100 (#f1f5f9)
	pdf.SetFillColor(241, 245, 249)
	pdf.RectFromUpperLeftWithStyle(marginLeft, boxY, boxW, boxH, "F")

	// 2. Left Column: Supplier / Nursery Site Company Info
	leftX := marginLeft + 14.0
	currY := boxY + 12.0

	sName := input.SiteProfile.NurseryName
	if sName == "" {
		sName = "Supplier Company INC"
	}
	pdf.SetTextColor(30, 41, 59) // text-neutral-700 / slate-800
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, sName)
	currY += 12.0

	pdf.SetTextColor(82, 82, 82) // text-neutral-600
	_ = pdf.SetFont("dejavu", "", 8.0)

	regNo := input.SiteProfile.RegNumber
	if regNo == "" {
		regNo = "12993844"
	}
	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, "Number: "+regNo)
	currY += 10.0

	vatNo := input.SiteProfile.VATNumber
	if vatNo == "" {
		vatNo = "GB 992 1122 33"
	}
	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, "VAT: "+vatNo)
	currY += 10.0

	street := input.SiteProfile.AddressStreet
	if street == "" {
		street = "88 Education Square"
	}
	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, street)
	currY += 10.0

	cityPost := fmt.Sprintf("%s, %s", input.SiteProfile.AddressCity, input.SiteProfile.AddressPostcode)
	if strings.Trim(cityPost, ", ") == "" {
		cityPost = "London, WC1N 1EX"
	}
	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, cityPost)
	currY += 10.0

	pdf.SetXY(leftX, currY)
	_ = pdf.Cell(nil, "United Kingdom")

	// 3. Right Column: Customer / Parent & Child Info (right-aligned)
	rightX := marginLeft + contentWidth - 14.0
	rY := boxY + 12.0

	pName := input.Parent.FullName
	if pName == "" {
		pName = "Customer / Parent Name"
	}
	pdf.SetTextColor(30, 41, 59)
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	pnW, _ := pdf.MeasureTextWidth(pName)
	pdf.SetXY(rightX-pnW, rY)
	_ = pdf.Cell(nil, pName)
	rY += 12.0

	pdf.SetTextColor(82, 82, 82)
	_ = pdf.SetFont("dejavu", "", 8.0)

	// Child Info
	cName := input.Child.ChildName
	if cName != "" {
		childStr := cName
		if input.Child.RoomName != "" {
			childStr += " (" + input.Child.RoomName + ")"
		}
		csW, _ := pdf.MeasureTextWidth(childStr)
		pdf.SetXY(rightX-csW, rY)
		_ = pdf.Cell(nil, childStr)
		rY += 10.0
	}

	pAddr1 := input.Parent.AddressLine1
	if pAddr1 == "" && input.Parent.FullName == "" {
		pAddr1 = "9552 Vandervort Spurs"
	}
	if pAddr1 != "" {
		pa1W, _ := pdf.MeasureTextWidth(pAddr1)
		pdf.SetXY(rightX-pa1W, rY)
		_ = pdf.Cell(nil, pAddr1)
		rY += 10.0
	}

	pCityPost := ""
	if input.Parent.AddressCity != "" || input.Parent.AddressPostcode != "" {
		pCityPost = fmt.Sprintf("%s, %s", input.Parent.AddressCity, input.Parent.AddressPostcode)
	} else if input.Parent.FullName == "" {
		pCityPost = "Paradise, 43325"
	}
	if pCityPost != "" {
		pcpW, _ := pdf.MeasureTextWidth(pCityPost)
		pdf.SetXY(rightX-pcpW, rY)
		_ = pdf.Cell(nil, pCityPost)
		rY += 10.0
	}

	pEmail := input.Parent.Email
	if pEmail == "" && input.Parent.FullName == "" {
		pEmail = "customer@example.com"
	}
	if pEmail != "" {
		peW, _ := pdf.MeasureTextWidth(pEmail)
		pdf.SetXY(rightX-peW, rY)
		_ = pdf.Cell(nil, pEmail)
		rY += 10.0
	}

	countryStr := "United Kingdom"
	ctW, _ := pdf.MeasureTextWidth(countryStr)
	pdf.SetXY(rightX-ctW, rY)
	_ = pdf.Cell(nil, countryStr)
}

func (r *Renderer) drawLineItemsTable(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	rightMarginX := pageWidth - marginRight // 559.28

	// Column Coordinates (Total width = 523.0 pt)
	colIdxX := marginLeft         // 36.0, w = 25.0
	colDescX := colIdxX + 25.0    // 61.0, w = 195.0
	colPriceX := colDescX + 195.0 // 256.0, w = 60.0 (right aligned)
	colQtyX := colPriceX + 60.0   // 316.0, w = 45.0 (center aligned)
	colVatX := colQtyX + 45.0     // 361.0, w = 45.0 (center aligned)
	colSubX := colVatX + 45.0     // 406.0, w = 75.0 (right aligned)
	colTotX := colSubX + 75.0     // 481.0, w = 78.0 (right aligned to rightMarginX 559.28)

	// 1. Table Header Row with 2px #5c6ac4 bottom border
	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 8.0)

	pdf.SetXY(colIdxX+3.0, y)
	_ = pdf.Cell(nil, "#")

	pdf.SetXY(colDescX+2.0, y)
	_ = pdf.Cell(nil, "Product / Details")

	pw, _ := pdf.MeasureTextWidth("Price")
	pdf.SetXY(colPriceX+60.0-pw-2.0, y)
	_ = pdf.Cell(nil, "Price")

	qw, _ := pdf.MeasureTextWidth("Qty.")
	pdf.SetXY(colQtyX+(45.0-qw)/2.0, y)
	_ = pdf.Cell(nil, "Qty.")

	vw, _ := pdf.MeasureTextWidth("VAT")
	pdf.SetXY(colVatX+(45.0-vw)/2.0, y)
	_ = pdf.Cell(nil, "VAT")

	sw, _ := pdf.MeasureTextWidth("Subtotal")
	pdf.SetXY(colSubX+75.0-sw-2.0, y)
	_ = pdf.Cell(nil, "Subtotal")

	tw, _ := pdf.MeasureTextWidth("Subtotal + VAT")
	pdf.SetXY(colTotX+78.0-tw-3.0, y)
	_ = pdf.Cell(nil, "Subtotal + VAT")

	y += 12.0

	// 2px solid #5c6ac4 line under header
	pdf.SetStrokeColor(92, 106, 196)
	pdf.SetLineWidth(2.0)
	pdf.Line(marginLeft, y, rightMarginX, y)

	y += 6.0

	// 2. Table Rows
	linesToDraw := input.Lines
	if len(linesToDraw) == 0 {
		linesToDraw = []InvoicePDFLine{
			{
				Description:     "Full-Time Childcare",
				SubDescription:  "Standard Weekly Rate - October Week 1-4",
				QuantityMinutes: intPtr(9600),
				UnitAmountMinor: intPtr(22500),
				LineAmountMinor: 90000,
			},
		}
	}

	for i, line := range linesToDraw {
		if y > pageHeight-140.0 {
			r.addPage(pdf, input)
			y = 45.0
		}

		subDesc := getSubDescription(line)
		rowH := 24.0
		if subDesc != "" {
			rowH = 34.0
		}

		// Text color determination
		textR, textG, textB := 30, 41, 59
		if line.IsFunded {
			textR, textG, textB = 37, 99, 235
		} else if line.IsDiscount {
			textR, textG, textB = 194, 65, 12
		}

		// 1. Index #
		pdf.SetTextColor(100, 116, 139)
		_ = pdf.SetFont("dejavu", "", 8.0)
		pdf.SetXY(colIdxX+3.0, y+4.0)
		_ = pdf.Cell(nil, fmt.Sprintf("%d.", i+1))

		// 2. Product / Details
		pdf.SetTextColor(uint8(textR), uint8(textG), uint8(textB))
		_ = pdf.SetFont("dejavu-bold", "", 8.5)
		pdf.SetXY(colDescX+2.0, y+4.0)
		_ = pdf.Cell(nil, line.Description)

		if subDesc != "" {
			pdf.SetTextColor(100, 116, 139)
			_ = pdf.SetFont("dejavu", "", 7.5)
			pdf.SetXY(colDescX+2.0, y+18.0)
			_ = pdf.Cell(nil, subDesc)
		}

		// 3. Price
		unitStr := "—"
		if line.UnitAmountMinor != nil {
			unitStr = formatMoney(*line.UnitAmountMinor)
		}
		pdf.SetTextColor(uint8(textR), uint8(textG), uint8(textB))
		_ = pdf.SetFont("dejavu", "", 8.0)
		upw, _ := pdf.MeasureTextWidth(unitStr)
		pdf.SetXY(colPriceX+60.0-upw-2.0, y+4.0)
		_ = pdf.Cell(nil, unitStr)

		// 4. Qty
		qtyStr := formatQuantity(line)
		qw, _ := pdf.MeasureTextWidth(qtyStr)
		pdf.SetXY(colQtyX+(45.0-qw)/2.0, y+4.0)
		_ = pdf.Cell(nil, qtyStr)

		// 5. VAT
		vatStr := "0%"
		if line.IsFunded {
			vatStr = "Funded"
		} else if line.IsDiscount {
			vatStr = "Disc"
		}
		vw, _ := pdf.MeasureTextWidth(vatStr)
		pdf.SetXY(colVatX+(45.0-vw)/2.0, y+4.0)
		_ = pdf.Cell(nil, vatStr)

		// 6. Subtotal
		subStr := formatMoney(line.LineAmountMinor)
		if line.IsFunded || line.IsDiscount {
			if line.LineAmountMinor > 0 {
				subStr = "-" + subStr
			}
		}
		sw, _ := pdf.MeasureTextWidth(subStr)
		pdf.SetXY(colSubX+75.0-sw-2.0, y+4.0)
		_ = pdf.Cell(nil, subStr)

		// 7. Subtotal + VAT (Total Line Amount)
		totLineStr := subStr
		_ = pdf.SetFont("dejavu-bold", "", 8.5)
		tlw, _ := pdf.MeasureTextWidth(totLineStr)
		pdf.SetXY(colTotX+78.0-tlw-3.0, y+4.0)
		_ = pdf.Cell(nil, totLineStr)

		y += rowH

		// Bottom border line for row (#e5e7eb)
		pdf.SetStrokeColor(229, 231, 235)
		pdf.SetLineWidth(0.5)
		pdf.Line(marginLeft, y, rightMarginX, y)
	}

	return y + 10.0
}

func (r *Renderer) drawTotalsSummarySection(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	if y > pageHeight-150.0 {
		r.addPage(pdf, input)
		y = 45.0
	}

	rightMarginX := pageWidth - marginRight // 559.28
	summaryWidth := 220.0
	summaryX := rightMarginX - summaryWidth

	// 1. Net Total Row
	pdf.SetTextColor(148, 163, 184) // text-slate-400
	_ = pdf.SetFont("dejavu", "", 8.5)
	pdf.SetXY(summaryX+10.0, y)
	_ = pdf.Cell(nil, "Net total:")

	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	netStr := formatMoney(input.SubtotalMinor)
	nw, _ := pdf.MeasureTextWidth(netStr)
	pdf.SetXY(rightMarginX-nw-10.0, y)
	_ = pdf.Cell(nil, netStr)

	y += 14.0
	pdf.SetStrokeColor(229, 231, 235)
	pdf.SetLineWidth(0.5)
	pdf.Line(summaryX, y, rightMarginX, y)
	y += 6.0

	// 2. VAT Total / Deductions Row
	dedMinor := input.DeductionMinor
	vatLabel := "VAT total:"
	vatValStr := "£0.00"
	if dedMinor > 0 {
		vatLabel = "Deductions total:"
		vatValStr = "-" + formatMoney(dedMinor)
	}

	pdf.SetTextColor(148, 163, 184) // text-slate-400
	_ = pdf.SetFont("dejavu", "", 8.5)
	pdf.SetXY(summaryX+10.0, y)
	_ = pdf.Cell(nil, vatLabel)

	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	vw, _ := pdf.MeasureTextWidth(vatValStr)
	pdf.SetXY(rightMarginX-vw-10.0, y)
	_ = pdf.Cell(nil, vatValStr)

	y += 14.0
	pdf.SetStrokeColor(229, 231, 235)
	pdf.SetLineWidth(0.5)
	pdf.Line(summaryX, y, rightMarginX, y)
	y += 6.0

	// 3. Filled #5c6ac4 Total Box (bg-main p-3 font-bold text-white)
	boxH := 28.0
	pdf.SetFillColor(92, 106, 196)
	pdf.RectFromUpperLeftWithStyle(summaryX, y, summaryWidth, boxH, "F")

	pdf.SetTextColor(255, 255, 255)
	_ = pdf.SetFont("dejavu-bold", "", 10.0)
	pdf.SetXY(summaryX+12.0, y+8.0)
	_ = pdf.Cell(nil, "Total:")

	totStr := formatMoney(input.TotalMinor)
	if input.TotalMinor == 0 && input.BalanceDueMinor > 0 {
		totStr = formatMoney(input.BalanceDueMinor)
	}
	_ = pdf.SetFont("dejavu-bold", "", 12.0)
	tw, _ := pdf.MeasureTextWidth(totStr)
	pdf.SetXY(rightMarginX-tw-12.0, y+7.0)
	_ = pdf.Cell(nil, totStr)

	return y + boxH + 20.0
}

func (r *Renderer) drawPaymentAndNotesSection(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) {
	y := startY
	if y > pageHeight-120.0 {
		r.addPage(pdf, input)
		y = 45.0
	}

	// Left side: PAYMENT DETAILS
	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	pdf.SetXY(marginLeft, y)
	_ = pdf.Cell(nil, "PAYMENT DETAILS")

	pdf.SetTextColor(64, 64, 64) // text-neutral-700
	_ = pdf.SetFont("dejavu", "", 8.0)
	py := y + 12.0

	bank := input.SiteProfile.BankName
	if bank == "" {
		bank = "Barclays Bank UK"
	}
	sort := input.SiteProfile.SortCode
	if sort == "" {
		sort = "20-45-89"
	}
	acc := input.SiteProfile.AccountNumber
	if acc == "" {
		acc = "88220033"
	}
	ref := input.Invoice.InvoiceNumber
	if ref == "" {
		ref = "INV-2024-0892"
	}

	pdf.SetXY(marginLeft, py)
	_ = pdf.Cell(nil, bank)
	py += 10.0
	pdf.SetXY(marginLeft, py)
	_ = pdf.Cell(nil, "Bank/Sort Code: "+sort)
	py += 10.0
	pdf.SetXY(marginLeft, py)
	_ = pdf.Cell(nil, "Account Number: "+acc)
	py += 10.0
	pdf.SetXY(marginLeft, py)
	_ = pdf.Cell(nil, "Payment Reference: "+ref)

	// Right / Next section: Notes
	ny := py + 16.0
	pdf.SetTextColor(92, 106, 196) // text-main (#5c6ac4)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	pdf.SetXY(marginLeft, ny)
	_ = pdf.Cell(nil, "Notes")

	noteText := input.PaymentNote
	if noteText == "" {
		noteText = "Please quote invoice number as payment reference. Monthly fees are payable in advance."
	}

	pdf.SetTextColor(64, 64, 64) // text-neutral-700
	_ = pdf.SetFont("dejavu", "", 8.0)
	pdf.SetXY(marginLeft, ny+12.0)
	_ = pdf.Cell(nil, noteText)
}

func formatQuantity(line InvoicePDFLine) string {
	if line.QuantityText != "" {
		return line.QuantityText
	}
	if line.QuantityMinutes != nil && *line.QuantityMinutes > 0 {
		mins := *line.QuantityMinutes
		if mins%60 == 0 {
			hours := mins / 60
			return fmt.Sprintf("%d hrs", hours)
		}
		hours := float64(mins) / 60.0
		return fmt.Sprintf("%.1f hrs", hours)
	}
	if line.SessionCount != nil && *line.SessionCount > 0 {
		return fmt.Sprintf("%d", *line.SessionCount)
	}
	return "1"
}

func getSubDescription(line InvoicePDFLine) string {
	if line.SubDescription != "" {
		return line.SubDescription
	}
	lk := strings.ToLower(line.LineKind)
	desc := strings.ToLower(line.Description)
	if line.IsFunded || lk == "funded_deduction" || strings.Contains(desc, "funded") || strings.Contains(desc, "funding") {
		return "Early Years Funding Offset"
	}
	if line.IsDiscount || lk == "discount" || strings.Contains(desc, "discount") {
		return "Special Fee Reduction"
	}
	if lk == "attendance_fee" || lk == "core" || strings.Contains(desc, "core") || strings.Contains(desc, "childcare") {
		return "Standard Weekly / Monthly Session Rate"
	}
	if lk == "extra_hours" || strings.Contains(desc, "extra") || strings.Contains(desc, "club") || strings.Contains(desc, "after-school") {
		return "Extended Hours & Ad-hoc Care"
	}
	return ""
}

func intPtr(i int) *int {
	return &i
}
