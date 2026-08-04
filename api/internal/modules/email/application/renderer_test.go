package application

import (
	"strings"
	"testing"
)

func sampleInvoiceData() map[string]interface{} {
	return map[string]interface{}{
		"NurseryName":   "Sunny Days Nursery",
		"ChildName":     "Leo Harrison",
		"InvoiceNumber": "INV-2026-08-0001",
		"BillingMonth":  "August 2026",
		"TotalDue":      "£420.00",
		"DueDate":       "2 September 2026",
		"PortalLink":    "https://app.example.com/parent/invoices/abc",
	}
}

func sampleInvoiceDataWithPayLink() map[string]interface{} {
	data := sampleInvoiceData()
	data["PayLink"] = "https://checkout.stripe.com/pay-inv-001"
	return data
}

func sampleReceiptData() map[string]interface{} {
	return map[string]interface{}{
		"NurseryName":   "Sunny Days Nursery",
		"ChildName":     "Leo Harrison",
		"InvoiceNumber": "INV-2026-08-0001",
		"TotalDue":      "£420.00",
		"AmountPaid":    "£420.00",
		"PaymentDate":   "1 August 2026",
		"PortalLink":    "https://app.example.com/parent/invoices/abc",
	}
}

func TestRenderer_V2HTMLTemplatesRenderOnLayoutShell(t *testing.T) {
	renderer := NewRenderer()

	// per-type message that must appear in its own template
	type case_ struct {
		name      string
		data      map[string]interface{}
		messaging string
	}

	cases := []case_{
		{name: "issued", data: sampleInvoiceData(), messaging: "New Invoice Issued"},
		{name: "overdue", data: sampleInvoiceData(), messaging: "Invoice Overdue"},
		{name: "due-soon", data: sampleInvoiceData(), messaging: "Payment Due Soon"},
		{name: "due-reminder", data: sampleInvoiceData(), messaging: "Payment Due Today"},
		{name: "receipt", data: sampleReceiptData(), messaging: "Payment Received"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			htmlBody, textBody, err := renderer.Render(tc.name, 2, tc.data)
			if err != nil {
				t.Fatalf("render %s v2: %v", tc.name, err)
			}

			// Layout shell: logo, hero amount, CTA, footer.
			for _, want := range []string{
				`aria-label="NurseryPro"`,
				"View Invoice &amp; Pay Online",
				"£420.00",
				"Amount Due",
				"&copy; 2026 NurseryPro",
				"Sunny Days Nursery",
			} {
				if !strings.Contains(htmlBody, want) {
					t.Errorf("html missing %q", want)
				}
			}

			// Per-type messaging.
			if !strings.Contains(htmlBody, tc.messaging) {
				t.Errorf("html missing per-type messaging %q", tc.messaging)
			}

			// Text alternative renders the same data fields without HTML tags.
			if strings.Contains(textBody, "<") || strings.Contains(textBody, ">") {
				t.Errorf("text contains HTML tags:\n%s", textBody)
			}
			if !strings.Contains(textBody, "INV-2026-08-0001") {
				t.Errorf("text missing invoice number:\n%s", textBody)
			}
			if !strings.Contains(textBody, "https://app.example.com/parent/invoices/abc") {
				t.Errorf("text missing portal link:\n%s", textBody)
			}
		})
	}
}

func TestRenderer_V2HTMLEscapesUserData(t *testing.T) {
	renderer := NewRenderer()

	data := map[string]interface{}{
		"NurseryName":   "A & B Nursery <Care>",
		"ChildName":     "Bobby <Chandler> & Co",
		"InvoiceNumber": "INV-2026-08-0001",
		"BillingMonth":  "August 2026",
		"TotalDue":      "£10.00",
		"DueDate":       "2 September 2026",
		"PortalLink":    "https://app.example.com/parent/invoices/abc",
	}

	htmlBody, textBody, err := renderer.Render("issued", 2, data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if strings.Contains(htmlBody, "Bobby <Chandler>") {
		t.Errorf("html contains unescaped child name:\n%s", htmlBody)
	}
	if !strings.Contains(htmlBody, "Bobby &lt;Chandler&gt; &amp; Co") {
		t.Errorf("html missing escaped child name:\n%s", htmlBody)
	}
	if !strings.Contains(htmlBody, "A &amp; B Nursery &lt;Care&gt;") {
		t.Errorf("html missing escaped nursery name:\n%s", htmlBody)
	}

	if !strings.Contains(textBody, "Bobby <Chandler> & Co") {
		t.Errorf("text missing literal child name:\n%s", textBody)
	}
	if !strings.Contains(textBody, "A & B Nursery <Care>") {
		t.Errorf("text missing literal nursery name:\n%s", textBody)
	}
}

func TestRenderer_V2LayoutContentBlockIsMandatory(t *testing.T) {
	renderer := NewRenderer()

	_, _, err := renderer.Render("missing_content", 2, sampleInvoiceData())
	if err == nil {
		t.Fatal("expected render to fail when the content template omits {{define \"content\"}}")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Errorf("expected error to mention the missing content block, got: %v", err)
	}
}

func TestRenderer_V1TemplateStillRenders(t *testing.T) {
	renderer := NewRenderer()

	htmlBody, textBody, err := renderer.Render("issued", 1, sampleInvoiceData())
	if err != nil {
		t.Fatalf("render issued v1: %v", err)
	}

	if !strings.Contains(htmlBody, "New Invoice") {
		t.Errorf("v1 html missing heading:\n%s", htmlBody)
	}
	if !strings.Contains(textBody, "INV-2026-08-0001") {
		t.Errorf("v1 text missing invoice number:\n%s", textBody)
	}
}

func TestRenderer_V2InvoiceCTAHonorsPayLink(t *testing.T) {
	renderer := NewRenderer()

	for _, name := range []string{"issued", "overdue", "due-soon", "due-reminder"} {
		t.Run(name, func(t *testing.T) {
			htmlBody, textBody, err := renderer.Render(name, 2, sampleInvoiceDataWithPayLink())
			if err != nil {
				t.Fatalf("render %s v2 with PayLink: %v", name, err)
			}

			// Primary CTA points at the Stripe pay link, labelled "Pay Now".
			if !strings.Contains(htmlBody, `href="https://checkout.stripe.com/pay-inv-001"`) {
				t.Errorf("html missing pay-link href:\n%s", htmlBody)
			}
			if !strings.Contains(htmlBody, ">Pay Now</a>") {
				t.Errorf("html missing Pay Now button:\n%s", htmlBody)
			}
			// Secondary "view invoice details" link uses the portal link.
			if !strings.Contains(htmlBody, `href="https://app.example.com/parent/invoices/abc"`) {
				t.Errorf("html missing portal view-details href:\n%s", htmlBody)
			}
			if !strings.Contains(htmlBody, "View invoice details") {
				t.Errorf("html missing secondary view-details link:\n%s", htmlBody)
			}
			// The fallback single-CTA label must not appear.
			if strings.Contains(htmlBody, "View Invoice &amp; Pay Online") {
				t.Errorf("html still renders the fallback CTA when PayLink present:\n%s", htmlBody)
			}

			// Text variant lists both links.
			if !strings.Contains(textBody, "Pay Now: https://checkout.stripe.com/pay-inv-001") {
				t.Errorf("text missing pay link:\n%s", textBody)
			}
			if !strings.Contains(textBody, "View Invoice Details: https://app.example.com/parent/invoices/abc") {
				t.Errorf("text missing view-details link:\n%s", textBody)
			}
		})
	}
}

func TestRenderer_V2InvoiceCTAFallsBackToPortalLink(t *testing.T) {
	renderer := NewRenderer()

	for _, name := range []string{"issued", "overdue", "due-soon", "due-reminder"} {
		t.Run(name, func(t *testing.T) {
			htmlBody, textBody, err := renderer.Render(name, 2, sampleInvoiceData())
			if err != nil {
				t.Fatalf("render %s v2 without PayLink: %v", name, err)
			}

			if !strings.Contains(htmlBody, `href="https://app.example.com/parent/invoices/abc"`) {
				t.Errorf("html missing portal-link href:\n%s", htmlBody)
			}
			if !strings.Contains(htmlBody, "View Invoice &amp; Pay Online") {
				t.Errorf("html missing fallback CTA label:\n%s", htmlBody)
			}
			if strings.Contains(htmlBody, "https://checkout.stripe.com/pay-inv-001") {
				t.Errorf("html must not contain a pay link when absent:\n%s", htmlBody)
			}

			// Text variant uses the single portal link without a pay link line.
			if !strings.Contains(textBody, "View Invoice & Pay Online: https://app.example.com/parent/invoices/abc") {
				t.Errorf("text missing fallback portal link:\n%s", textBody)
			}
			if strings.Contains(textBody, "Pay Now:") {
				t.Errorf("text must not contain a pay link when absent:\n%s", textBody)
			}
		})
	}
}

func TestRenderer_V2ReceiptUnchangedWithoutPayLink(t *testing.T) {
	renderer := NewRenderer()

	htmlBody, textBody, err := renderer.Render("receipt", 2, sampleReceiptData())
	if err != nil {
		t.Fatalf("render receipt v2: %v", err)
	}

	if !strings.Contains(htmlBody, `href="https://app.example.com/parent/invoices/abc"`) {
		t.Errorf("html missing single app link:\n%s", htmlBody)
	}
	if !strings.Contains(htmlBody, "View Invoice &amp; Pay Online") {
		t.Errorf("html missing receipt CTA:\n%s", htmlBody)
	}
	if strings.Contains(htmlBody, "Pay Now") || strings.Contains(textBody, "Pay Now:") {
		t.Errorf("receipt must not contain a pay link:\n%s\n%s", htmlBody, textBody)
	}
}
