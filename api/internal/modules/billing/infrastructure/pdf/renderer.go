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
		r.drawTopAndBottomAccents(pdf)
	})

	r.addPage(pdf)
	r.drawHeaderSection(pdf, input)
	r.drawGridSection(pdf, input)
	y := r.drawLineItemsTable(pdf, input, 224.0)
	y = r.drawSummarySection(pdf, input, y)
	r.drawFooterSection(pdf, input)

	return pdf.GetBytesPdfReturnErr()
}

func (r *Renderer) addPage(pdf *gopdf.GoPdf) {
	pdf.AddPage()
	r.drawTopAndBottomAccents(pdf)
}

func (r *Renderer) drawTopAndBottomAccents(pdf *gopdf.GoPdf) {
	// Top Accent Bar (Royal Blue)
	pdf.SetFillColor(37, 99, 235)
	pdf.RectFromUpperLeftWithStyle(0, 0, pageWidth, 4, "F")

	// Bottom Accent Bar (Soft Light Blue)
	pdf.SetFillColor(219, 234, 254)
	pdf.RectFromUpperLeftWithStyle(0, pageHeight-4, pageWidth, 4, "F")
}

func (r *Renderer) drawHeaderSection(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	rightMarginX := pageWidth - marginRight

	// 1. Faint Watermark Text "INVOICE" (Positioned inside top margin area with high clarity)
	pdf.SetTextColor(246, 248, 250)
	_ = pdf.SetFont("dejavu-bold", "", 24)
	pdf.SetXY(marginLeft, 18.0)
	_ = pdf.Cell(nil, "INVOICE")

	// 2. Brand Logo: Smiley Circle + NurseryPro Text
	logoX := marginLeft
	logoY := 32.0
	logoR := 9.0

	// Logo Blue Circle
	pdf.SetFillColor(37, 99, 235)
	pdf.Oval(logoX, logoY, logoX+(logoR*2), logoY+(logoR*2))

	// Smiley face details inside logo
	pdf.SetFillColor(255, 255, 255)
	pdf.Oval(logoX+4.5, logoY+5.5, logoX+7.0, logoY+8.0)
	pdf.Oval(logoX+11.0, logoY+5.5, logoX+13.5, logoY+8.0)
	pdf.SetStrokeColor(255, 255, 255)
	pdf.SetLineWidth(1.0)
	pdf.Line(logoX+5.5, logoY+12.0, logoX+7.5, logoY+13.5)
	pdf.Line(logoX+7.5, logoY+13.5, logoX+10.5, logoY+13.5)
	pdf.Line(logoX+10.5, logoY+13.5, logoX+12.5, logoY+12.0)

	// Brand Text
	textX := logoX + (logoR * 2) + 6.0
	pdf.SetXY(textX, logoY-1.0)
	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 16)
	_ = pdf.Cell(nil, "Nursery")

	nw, _ := pdf.MeasureTextWidth("Nursery")
	pdf.SetXY(textX+nw, logoY-1.0)
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.Cell(nil, "Pro")

	// 3. Invoice Number Title
	invNum := input.Invoice.InvoiceNumber
	if invNum == "" {
		invNum = "INV-2024-0892"
	}
	if !strings.HasPrefix(invNum, "INV-") && !strings.HasPrefix(invNum, "#") {
		invNum = "INV-" + invNum
	}
	if !strings.HasPrefix(invNum, "#") {
		invNum = "#" + invNum
	}

	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 13)
	pdf.SetXY(marginLeft, 58.0)
	_ = pdf.Cell(nil, "Invoice "+invNum)

	// 4. Status Badge Pill + Issued Date
	status := strings.ToUpper(input.Invoice.Status)
	if status == "" {
		status = "PAID"
	}

	bgR, bgG, bgB := 230, 244, 234
	txtR, txtG, txtB := 22, 101, 52

	switch status {
	case "DRAFT":
		bgR, bgG, bgB = 241, 245, 249
		txtR, txtG, txtB = 71, 85, 105
	case "ISSUED":
		bgR, bgG, bgB = 224, 242, 254
		txtR, txtG, txtB = 3, 105, 161
	case "OVERDUE":
		bgR, bgG, bgB = 254, 226, 226
		txtR, txtG, txtB = 185, 28, 28
	case "VOID", "CANCELLED":
		bgR, bgG, bgB = 241, 245, 249
		txtR, txtG, txtB = 30, 41, 59
	}

	pillX := marginLeft
	pillY := 76.0
	pillW := 48.0
	pillH := 16.0

	pdf.SetFillColor(uint8(bgR), uint8(bgG), uint8(bgB))
	pdf.RectFromUpperLeftWithStyle(pillX, pillY, pillW, pillH, "F")

	pdf.SetTextColor(uint8(txtR), uint8(txtG), uint8(txtB))
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	sw, _ := pdf.MeasureTextWidth(status)
	pdf.SetXY(pillX+(pillW-sw)/2.0, pillY+3.5)
	_ = pdf.Cell(nil, status)

	// Issued date text
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu", "", 8.5)
	pdf.SetXY(pillX+pillW+10.0, pillY+3.5)
	issuedText := "Issued: " + formatDate(input.Invoice.IssueDate)
	if input.Invoice.IssueDate == nil {
		issuedText = "Issued: Oct 12, 2023"
	}
	_ = pdf.Cell(nil, issuedText)

	// 5. Top Right Hero Card (AMOUNT DUE Box - Right Aligned with Margin)
	heroW := 160.0
	heroH := 68.0
	heroX := rightMarginX - heroW
	heroY := 26.0

	pdf.SetFillColor(244, 246, 252)
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(heroX, heroY, heroW, heroH, "DF")

	// AMOUNT DUE Label
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	labelStr := "AMOUNT DUE"
	lw, _ := pdf.MeasureTextWidth(labelStr)
	pdf.SetXY(heroX+(heroW-lw)/2.0, heroY+12.0)
	_ = pdf.Cell(nil, labelStr)

	// Amount Value
	amountStr := formatMoney(input.BalanceDueMinor)
	if input.BalanceDueMinor == 0 && input.TotalMinor > 0 && strings.ToUpper(input.Invoice.Status) == "PAID" {
		amountStr = formatMoney(input.TotalMinor)
	}
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.SetFont("dejavu-bold", "", 18)
	aw, _ := pdf.MeasureTextWidth(amountStr)
	pdf.SetXY(heroX+(heroW-aw)/2.0, heroY+26.0)
	_ = pdf.Cell(nil, amountStr)

	// Due Date
	dueStr := "Due by " + formatDate(input.Invoice.DueDate)
	if input.Invoice.DueDate == nil {
		dueStr = "Due by Oct 26, 2023"
	}
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu", "", 8.0)
	dw, _ := pdf.MeasureTextWidth(dueStr)
	pdf.SetXY(heroX+(heroW-dw)/2.0, heroY+50.0)
	_ = pdf.Cell(nil, dueStr)
}

func (r *Renderer) drawGridSection(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	colY := 110.0
	col1X := marginLeft
	col2X := 315.0
	colW := 235.0

	// Vertical Separator Line between Left and Right columns
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(295.0, colY, 295.0, colY+98.0)

	// LEFT COLUMN: BILL TO & CHILD DETAILS
	// Header: BILL TO
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	pdf.SetXY(col1X, colY)
	_ = pdf.Cell(nil, "BILL TO")

	// Parent Name
	pName := input.Parent.FullName
	if pName == "" {
		pName = "Mr. David Harrison"
	}
	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	pdf.SetXY(col1X, colY+12.0)
	_ = pdf.Cell(nil, pName)

	// Parent Address
	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 8.0)
	currY := colY + 23.0

	addr1 := input.Parent.AddressLine1
	if addr1 == "" && input.Parent.FullName == "" {
		addr1 = "42 Nightingale Lane"
	}
	if addr1 != "" {
		pdf.SetXY(col1X, currY)
		_ = pdf.Cell(nil, addr1)
		currY += 10.0
	}

	addrCityPost := ""
	if input.Parent.AddressCity != "" || input.Parent.AddressPostcode != "" {
		addrCityPost = fmt.Sprintf("%s, %s", input.Parent.AddressCity, input.Parent.AddressPostcode)
	} else if input.Parent.FullName == "" {
		addrCityPost = "London, SW12 8TH"
	}
	if addrCityPost != "" {
		pdf.SetXY(col1X, currY)
		_ = pdf.Cell(nil, addrCityPost)
		currY += 10.0
	}

	country := "United Kingdom"
	pdf.SetXY(col1X, currY)
	_ = pdf.Cell(nil, country)
	currY += 10.0

	// Email
	pEmail := input.Parent.Email
	if pEmail == "" && input.Parent.FullName == "" {
		pEmail = "david.h@example.com"
	}
	if pEmail != "" {
		pdf.SetTextColor(37, 99, 235)
		_ = pdf.SetFont("dejavu", "", 8.0)
		pdf.SetXY(col1X, currY)
		_ = pdf.Cell(nil, pEmail)
		// Draw underline
		ew, _ := pdf.MeasureTextWidth(pEmail)
		pdf.SetStrokeColor(37, 99, 235)
		pdf.SetLineWidth(0.4)
		pdf.Line(col1X, currY+9.5, col1X+ew, currY+9.5)
		currY += 14.0
	}

	// Dotted Separator Line
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(col1X, currY, col1X+colW-20.0, currY)
	currY += 8.0

	// CHILD DETAILS Header
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	pdf.SetXY(col1X, currY)
	_ = pdf.Cell(nil, "CHILD DETAILS")

	// Child Info Line
	cName := input.Child.ChildName
	if cName == "" {
		cName = "Leo Harrison"
	}
	room := input.Child.RoomName
	if room == "" {
		room = "Pre-School Room"
	}
	childStr := fmt.Sprintf("%s | %s", cName, room)

	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	pdf.SetXY(col1X, currY+11.0)
	_ = pdf.Cell(nil, childStr)

	// RIGHT COLUMN: SITE DETAILS
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	pdf.SetXY(col2X, colY)
	_ = pdf.Cell(nil, "SITE DETAILS")

	siteName := input.SiteProfile.NurseryName
	if siteName == "" {
		siteName = "NurseryPro Central"
	}
	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 9.0)
	pdf.SetXY(col2X, colY+12.0)
	_ = pdf.Cell(nil, siteName)

	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 8.0)
	rY := colY + 23.0

	siteStreet := input.SiteProfile.AddressStreet
	if siteStreet == "" {
		siteStreet = "88 Education Square"
	}
	pdf.SetXY(col2X, rY)
	_ = pdf.Cell(nil, siteStreet)
	rY += 10.0

	siteCityPost := fmt.Sprintf("%s %s", input.SiteProfile.AddressCity, input.SiteProfile.AddressPostcode)
	if strings.TrimSpace(siteCityPost) == "" {
		siteCityPost = "Bloomsbury, London WC1N 1EX"
	}
	pdf.SetXY(col2X, rY)
	_ = pdf.Cell(nil, siteCityPost)
	rY += 10.0

	regNo := input.SiteProfile.RegNumber
	if regNo == "" {
		regNo = "12993844"
	}
	pdf.SetXY(col2X, rY)
	_ = pdf.Cell(nil, "Reg No: "+regNo)
	rY += 10.0

	pdf.SetXY(col2X, rY+2.0)
	_ = pdf.Cell(nil, "Manager: Sarah Jenkins")
}

func (r *Renderer) drawLineItemsTable(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	rightMarginX := pageWidth - marginRight

	colDescX := marginLeft
	colDescW := 250.0

	colQtyX := marginLeft + colDescW
	colQtyW := 85.0

	colUnitX := colQtyX + colQtyW
	colUnitW := 90.0

	// Table Header Box (Exact alignment with contentWidth)
	headerH := 22.0
	pdf.SetFillColor(240, 244, 255)
	pdf.RectFromUpperLeftWithStyle(marginLeft, y, contentWidth, headerH, "F")

	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)

	pdf.SetXY(colDescX+12.0, y+7.0)
	_ = pdf.Cell(nil, "DESCRIPTION")

	qw, _ := pdf.MeasureTextWidth("QTY / HOURS")
	pdf.SetXY(colQtyX+(colQtyW-qw)/2.0, y+7.0)
	_ = pdf.Cell(nil, "QTY / HOURS")

	uw, _ := pdf.MeasureTextWidth("UNIT PRICE")
	pdf.SetXY(colUnitX+colUnitW-uw-12.0, y+7.0)
	_ = pdf.Cell(nil, "UNIT PRICE")

	mw, _ := pdf.MeasureTextWidth("AMOUNT")
	pdf.SetXY(rightMarginX-mw-12.0, y+7.0)
	_ = pdf.Cell(nil, "AMOUNT")

	y += headerH + 4.0

	// Lines Rendering
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

	for _, line := range linesToDraw {
		if y > pageHeight-170.0 {
			r.addPage(pdf)
			y = 50.0
		}

		subDesc := getSubDescription(line)
		rowH := 34.0
		if subDesc == "" {
			rowH = 26.0
		}

		// Background row styling
		if line.IsFunded {
			pdf.SetFillColor(240, 244, 255)
			pdf.RectFromUpperLeftWithStyle(marginLeft, y-2.0, contentWidth, rowH, "F")
		} else if line.IsDiscount {
			pdf.SetFillColor(255, 245, 243)
			pdf.RectFromUpperLeftWithStyle(marginLeft, y-2.0, contentWidth, rowH, "F")
		}

		// Text color determination
		textR, textG, textB := 15, 23, 42
		subR, subG, subB := 100, 116, 139
		if line.IsFunded {
			textR, textG, textB = 37, 99, 235
			subR, subG, subB = 37, 99, 235
		} else if line.IsDiscount {
			textR, textG, textB = 194, 65, 12
			subR, subG, subB = 194, 65, 12
		}

		// PERFECT STRAIGHT VERTICAL ALIGNMENT FOR ALL DESCRIPTION LINES
		textOffsetX := colDescX + 12.0

		// 1. Primary Description
		pdf.SetTextColor(uint8(textR), uint8(textG), uint8(textB))
		_ = pdf.SetFont("dejavu-bold", "", 8.5)
		pdf.SetXY(textOffsetX, y+4.0)
		_ = pdf.Cell(nil, line.Description)

		// 2. SubDescription (Italics / muted)
		if subDesc != "" {
			pdf.SetTextColor(uint8(subR), uint8(subG), uint8(subB))
			_ = pdf.SetFont("dejavu", "", 7.5)
			pdf.SetXY(textOffsetX, y+18.0)
			_ = pdf.Cell(nil, subDesc)
		}

		// 3. QTY / HOURS
		qtyStr := formatQuantity(line)
		pdf.SetTextColor(uint8(textR), uint8(textG), uint8(textB))
		_ = pdf.SetFont("dejavu", "", 8.0)
		qw, _ := pdf.MeasureTextWidth(qtyStr)
		pdf.SetXY(colQtyX+(colQtyW-qw)/2.0, y+7.0)
		_ = pdf.Cell(nil, qtyStr)

		// 4. UNIT PRICE
		unitStr := "\u2014"
		if line.UnitAmountMinor != nil {
			unitStr = formatMoney(*line.UnitAmountMinor)
			if line.IsFunded && *line.UnitAmountMinor > 0 {
				unitStr = "-" + unitStr
			} else if line.IsDiscount && *line.UnitAmountMinor > 0 {
				unitStr = "-" + unitStr
			}
		}
		uw, _ := pdf.MeasureTextWidth(unitStr)
		pdf.SetXY(colUnitX+colUnitW-uw-12.0, y+7.0)
		_ = pdf.Cell(nil, unitStr)

		// 5. AMOUNT
		amtStr := formatMoney(line.LineAmountMinor)
		if line.IsFunded && line.LineAmountMinor > 0 {
			amtStr = "-" + amtStr
		} else if line.IsDiscount && line.LineAmountMinor > 0 {
			amtStr = "-" + amtStr
		}
		pdf.SetFont("dejavu-bold", "", 8.5)
		mw, _ := pdf.MeasureTextWidth(amtStr)
		pdf.SetXY(rightMarginX-mw-12.0, y+7.0)
		_ = pdf.Cell(nil, amtStr)

		y += rowH + 2.0
	}

	return y + 10.0
}

func (r *Renderer) drawSummarySection(pdf *gopdf.GoPdf, input InvoicePDFInput, startY float64) float64 {
	y := startY
	if y > pageHeight-170.0 {
		r.addPage(pdf)
		y = 50.0
	}

	rightMarginX := pageWidth - marginRight // 555.28
	labelX := 320.0
	sumW := rightMarginX - labelX      // 235.28
	valueRightX := rightMarginX - 12.0 // 543.28 (Exact 12pt padding from table right edge)

	// 1. Gross Subtotal
	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 8.5)
	pdf.SetXY(labelX, y)
	_ = pdf.Cell(nil, "Gross Subtotal")

	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	subStr := formatMoney(input.SubtotalMinor)
	sw, _ := pdf.MeasureTextWidth(subStr)
	pdf.SetXY(valueRightX-sw, y)
	_ = pdf.Cell(nil, subStr)

	y += 18.0

	// 2. Total Deductions & Funding Bar (EXACT ALIGNMENT WITH TABLE RIGHT EDGE)
	dedMinor := input.DeductionMinor
	if dedMinor == 0 && len(input.Lines) > 0 {
		for _, l := range input.Lines {
			if l.IsFunded || l.IsDiscount {
				if l.LineAmountMinor > 0 {
					dedMinor += l.LineAmountMinor
				} else if l.LineAmountMinor < 0 {
					dedMinor += -l.LineAmountMinor
				}
			}
		}
	}

	if dedMinor > 0 {
		barX := labelX - 6.0
		barW := sumW + 6.0
		barH := 20.0

		pdf.SetFillColor(240, 244, 255)
		pdf.RectFromUpperLeftWithStyle(barX, y-3.0, barW, barH, "F")

		pdf.SetTextColor(37, 99, 235)
		_ = pdf.SetFont("dejavu-bold", "", 8.5)
		pdf.SetXY(labelX, y+2.0)
		_ = pdf.Cell(nil, "Total Deductions & Funding")

		dedStr := "-" + formatMoney(dedMinor)
		dw, _ := pdf.MeasureTextWidth(dedStr)
		pdf.SetXY(valueRightX-dw, y+2.0)
		_ = pdf.Cell(nil, dedStr)

		y += 24.0
	}

	// 3. Tax (VAT 0%)
	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 8.5)
	pdf.SetXY(labelX, y)
	_ = pdf.Cell(nil, "Tax (VAT 0%)")

	pdf.SetTextColor(15, 23, 42)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	taxStr := "£0.00"
	tw, _ := pdf.MeasureTextWidth(taxStr)
	pdf.SetXY(valueRightX-tw, y)
	_ = pdf.Cell(nil, taxStr)

	y += 18.0

	// 4. Solid Dark Navy Container for TOTAL AMOUNT DUE (EXACT ALIGNMENT WITH TABLE RIGHT EDGE)
	boxX := labelX - 6.0
	boxW := sumW + 6.0
	boxH := 36.0

	pdf.SetFillColor(15, 23, 42)
	pdf.RectFromUpperLeftWithStyle(boxX, y, boxW, boxH, "F")

	// TOTAL AMOUNT DUE Label
	pdf.SetTextColor(255, 255, 255)
	_ = pdf.SetFont("dejavu-bold", "", 8.5)
	pdf.SetXY(labelX, y+12.0)
	_ = pdf.Cell(nil, "TOTAL AMOUNT DUE")

	// Amount Value
	totStr := formatMoney(input.TotalMinor)
	if input.TotalMinor == 0 && input.BalanceDueMinor > 0 {
		totStr = formatMoney(input.BalanceDueMinor)
	}
	_ = pdf.SetFont("dejavu-bold", "", 14.0)
	totW, _ := pdf.MeasureTextWidth(totStr)
	pdf.SetXY(valueRightX-totW, y+10.0)
	_ = pdf.Cell(nil, totStr)

	return y + boxH + 20.0
}

func (r *Renderer) drawFooterSection(pdf *gopdf.GoPdf, input InvoicePDFInput) {
	rightMarginX := pageWidth - marginRight
	y := pageHeight - 122.0

	// Separator Line
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginLeft, y, rightMarginX, y)

	y += 10.0

	// PAYMENT INFORMATION Header
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	pdf.SetXY(marginLeft, y)
	_ = pdf.Cell(nil, "PAYMENT INFORMATION")

	// PAYMENT INFORMATION Box
	boxX := marginLeft
	boxY := y + 10.0
	boxW := 145.0
	boxH := 48.0

	pdf.SetFillColor(244, 246, 252)
	pdf.RectFromUpperLeftWithStyle(boxX, boxY, boxW, boxH, "F")

	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 7.5)

	bank := input.SiteProfile.BankName
	if bank == "" {
		bank = "Barclays UK"
	}
	acc := input.SiteProfile.AccountNumber
	if acc == "" {
		acc = "88220033"
	}
	sort := input.SiteProfile.SortCode
	if sort == "" {
		sort = "20-45-89"
	}

	pdf.SetXY(boxX+8.0, boxY+8.0)
	_ = pdf.Cell(nil, "Bank: "+bank)
	pdf.SetXY(boxX+8.0, boxY+20.0)
	_ = pdf.Cell(nil, "Account: "+acc)
	pdf.SetXY(boxX+8.0, boxY+32.0)
	_ = pdf.Cell(nil, "Sort Code: "+sort)

	// NOTES Section (Middle Column)
	notesX := marginLeft + 160.0
	pdf.SetTextColor(37, 99, 235)
	_ = pdf.SetFont("dejavu-bold", "", 7.5)
	pdf.SetXY(notesX, y)
	_ = pdf.Cell(nil, "NOTES")

	pdf.SetTextColor(71, 85, 105)
	_ = pdf.SetFont("dejavu", "", 7.5)
	invRef := input.Invoice.InvoiceNumber
	if invRef == "" {
		invRef = "INV-2024-0892"
	}
	pdf.SetXY(notesX, y+12.0)
	_ = pdf.Cell(nil, "Please quote invoice number "+invRef)
	pdf.SetXY(notesX, y+22.0)
	_ = pdf.Cell(nil, "as payment reference.")

	pdf.SetXY(notesX, y+36.0)
	_ = pdf.Cell(nil, "Monthly fees are payable in advance")
	pdf.SetXY(notesX, y+46.0)
	_ = pdf.Cell(nil, "by the 25th of each month.")

	// REGISTRATION INFO (Right Side - Exactly Aligned with Right Margin)
	regY := y + 36.0
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu", "", 6.5)

	line1 := "Registered in England and Wales: #12993844"
	l1w, _ := pdf.MeasureTextWidth(line1)
	pdf.SetXY(rightMarginX-l1w, regY)
	_ = pdf.Cell(nil, line1)

	line2 := "VAT Registration: GB 992 1122 33"
	l2w, _ := pdf.MeasureTextWidth(line2)
	pdf.SetXY(rightMarginX-l2w, regY+9.0)
	_ = pdf.Cell(nil, line2)

	// Bottom Centered Thank You Banner
	thankStr := "THANK YOU FOR CHOOSING NURSERYPRO"
	pdf.SetTextColor(100, 116, 139)
	_ = pdf.SetFont("dejavu-bold", "", 8.0)
	tw, _ := pdf.MeasureTextWidth(thankStr)
	pdf.SetXY((pageWidth-tw)/2.0, pageHeight-22.0)
	_ = pdf.Cell(nil, thankStr)
}

func formatQuantity(line InvoicePDFLine) string {
	if line.QuantityText != "" {
		return line.QuantityText
	}
	if line.QuantityMinutes != nil && *line.QuantityMinutes > 0 {
		mins := *line.QuantityMinutes
		if mins%60 == 0 {
			hours := mins / 60
			if hours == 1 {
				return "1 Hour"
			}
			return fmt.Sprintf("%d Hours", hours)
		}
		hours := float64(mins) / 60.0
		return fmt.Sprintf("%.1f Hours", hours)
	}
	if line.SessionCount != nil && *line.SessionCount > 0 {
		if *line.SessionCount == 1 {
			return "1 Session"
		}
		return fmt.Sprintf("%d Sessions", *line.SessionCount)
	}
	return "1 Unit"
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
