For a UK Nursery Management System (NMS), I would design the invoice and payment workflow as an **event-driven financial pipeline**, not as a synchronous "click button → send email → create Stripe payment" flow.

This makes the system reliable, idempotent, auditable, and easy to recover from failures.

---

# High-Level Architecture

```text
Nursery Manager
      │
      ▼
Create Invoice
      │
      ▼
Invoice Service
      │
      ├──────────────┐
      │              │
      ▼              ▼
PostgreSQL      Outbox Event
      │              │
      └──────┬───────┘
             ▼
     Background Worker
             │
   ┌─────────┴─────────┐
   ▼                   ▼
Stripe Service     Email Service
   │                   │
Create Payment     Send Invoice
Link                  Email
   │                   │
   └─────────┬─────────┘
             ▼
Update Invoice Status
             │
             ▼
Webhook (Stripe)
             │
             ▼
Payment Completed
             │
             ▼
Invoice Paid
             │
             ▼
Receipt Email
```

---

# Invoice Lifecycle

```text
DRAFT
   │
   ▼
ISSUED
   │
   ▼
EMAIL_SENT
   │
   ▼
PAYMENT_PENDING
   │
   ▼
PAID
```

Other states

```text
VOID

OVERDUE

PARTIALLY_PAID

FAILED
```

---

# Step 1 — Create Invoice

Manager clicks

> Generate July Invoice

Backend

```text
BEGIN TRANSACTION

Insert invoice

Insert invoice_items

Calculate total

Insert Outbox:
InvoiceCreated

COMMIT
```

Never call Stripe here.

Never send email here.

Return immediately.

Response

```json
{
  "invoice_id": "...",
  "status":"ISSUED"
}
```

Fast.

Reliable.

---

# Step 2 — Background Worker

Worker polls

```text
InvoiceCreated
```

Then

```text
Create Stripe Checkout Session
```

or

```text
Create Stripe Payment Link
```

Store

```text
stripe_payment_intent_id

checkout_url

payment_link

expires_at
```

If failed

Retry

```
1 min

5 min

15 min

1 hour

```

No user impact.

---

# Step 3 — Send Email

After Stripe succeeds

Publish

```text
StripePaymentCreated
```

Worker

Generate email

```
Invoice

Amount

Due date

Pay button
```

Button

```
Pay Invoice
```

↓

Stripe Checkout

Store

```text
email_sent=true

sent_at
```

---

# Step 4 — Parent Opens Email

Clicks

```
Pay Invoice
```

↓

Stripe Checkout

Parent pays

---

# Step 5 — Stripe Webhook

Stripe calls

```
POST /webhooks/stripe
```

Example

```
checkout.session.completed

payment_intent.succeeded

invoice.paid
```

Never trust frontend.

Always trust webhook.

Webhook

```text
Verify Signature

Find invoice

Mark Paid

Store payment

Publish PaymentCompleted
```

---

# Step 6 — Receipt Email

Worker

```
PaymentCompleted
```

↓

Send

```
Thank you

Receipt

PDF

Balance
```

---

# Suggested Database Tables

## invoices

```text
id

tenant_id

parent_id

number

status

subtotal

tax

discount

total

due_date

issued_at

paid_at

stripe_checkout_id

stripe_payment_intent

payment_url
```

---

## invoice_items

```text
invoice_id

description

quantity

unit_price

amount
```

---

## payments

```text
id

invoice_id

provider

stripe_payment_intent

amount

currency

status

paid_at

metadata
```

---

## outbox

```text
id

event

payload

status

retry_count

next_retry

created_at
```

---

# Email Timeline

```text
Invoice Created

        │

        ▼

Generate Stripe Link

        │

        ▼

Email Invoice

        │

        ▼

Reminder 7 Days

        │

        ▼

Reminder 3 Days

        │

        ▼

Reminder Due Today

        │

        ▼

Reminder Overdue
```

---

# Stripe Objects

I recommend using

```
Invoice
        │
        ▼
Stripe Checkout Session
        │
        ▼
Payment Intent
```

instead of manually creating Payment Intents.

Checkout handles:

* 3D Secure
* Apple Pay
* Google Pay
* Cards
* Link
* Authentication

Automatically.

---

# Idempotency

Every operation should have an idempotency key.

Example

```
InvoiceCreated
```

↓

```
stripe-idempotency-key

invoice-{invoiceId}
```

If worker crashes

```
Retry
```

Stripe returns the same checkout session.

No duplicate payments.

---

# Failure Recovery

Suppose:

```
Invoice Created
```

↓

Stripe succeeds

↓

Email server crashes

Invoice status

```
PAYMENT_PENDING
```

Worker retries

↓

Email eventually sent.

No duplicate invoice.

---

Suppose

```
Stripe timeout
```

Worker

```
Retry

Retry

Retry
```

Eventually succeeds.

---

Suppose

Email fails

```
Retry
```

No invoice recreation.

---

# Recommended Services

```text
Billing Service
│
├── Invoice Generator
├── Invoice Calculator
├── PDF Generator
├── Reminder Scheduler
├── Payment Processor
├── Stripe Integration
├── Email Sender
└── Webhook Handler
```

Each should be independent and communicate through events.

---

# Production Workflow

```text
Monthly Billing Job
        │
        ▼
Generate Invoice
        │
        ▼
Store Invoice (DB)
        │
        ▼
Publish InvoiceCreated
        │
        ▼
──────────────────────────────────────
Background Worker
──────────────────────────────────────
        │
        ▼
Create Stripe Checkout Session
        │
        ▼
Store Checkout URL
        │
        ▼
Publish PaymentLinkCreated
        │
        ▼
──────────────────────────────────────
Email Worker
──────────────────────────────────────
        │
        ▼
Generate PDF Invoice
        │
        ▼
Send Email with Payment Link
        │
        ▼
Mark Email Sent
        │
        ▼
──────────────────────────────────────
Parent
──────────────────────────────────────
        │
        ▼
Open Email
        │
        ▼
Stripe Checkout
        │
        ▼
Payment Success
        │
        ▼
Stripe Webhook
        │
        ▼
Verify Signature
        │
        ▼
Record Payment
        │
        ▼
Invoice → PAID
        │
        ▼
Publish PaymentCompleted
        │
        ▼
Send Receipt Email
        │
        ▼
Update Child Account Balance
```

This design follows common patterns used in SaaS billing systems because it separates invoice creation, payment setup, email delivery, and payment confirmation into independent, retryable steps. Combined with an outbox pattern, background workers, Stripe idempotency keys, and webhook-driven payment confirmation, it provides resilience against transient failures while maintaining a complete audit trail suitable for financial operations.
