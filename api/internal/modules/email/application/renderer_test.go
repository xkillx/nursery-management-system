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
		"PortalLink":    "https://app.example.com/parent/billing/abc",
	}
}

func sampleReceiptData() map[string]interface{} {
	return map[string]interface{}{
		"NurseryName":   "Sunny Days Nursery",
		"ChildName":     "Leo Harrison",
		"InvoiceNumber": "INV-2026-08-0001",
		"TotalDue":      "£420.00",
		"AmountPaid":    "£420.00",
		"PaymentDate":   "1 August 2026",
		"PortalLink":    "https://app.example.com/parent/billing/abc",
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
			if !strings.Contains(textBody, "https://app.example.com/parent/billing/abc") {
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
		"PortalLink":    "https://app.example.com/parent/billing/abc",
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
