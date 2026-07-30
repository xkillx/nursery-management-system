For a production SaaS like your UK Nursery Management System (NMS), email is not just a "send email" feature. It is a distributed system. The SMTP server, email provider, network, DNS, and recipient mailbox can all fail independently.

The architecture should follow one fundamental principle:

> **The business transaction must never depend on email delivery.**

For example:

* ✅ Invoice created → Success even if email fails
* ✅ User invited → User account created even if email fails
* ✅ Password reset → Token generated even if email provider is temporarily down

Email is a side effect, not part of the core transaction.

---

# Recommended Architecture

```
                HTTP Request
                     │
                     ▼
              Application Service
                     │
      ┌──────────────┴──────────────┐
      │                             │
 Save Business Data          Insert Email Job
 (Invoice/User/etc)          (Outbox table)
      │                             │
      └──────────────┬──────────────┘
                     │
                Commit Transaction
                     │
                     ▼
              Return 200 OK

----------------------------------------------------

            Background Email Worker

      Poll Pending Emails
               │
               ▼
        Send via Provider
               │
      ┌────────┴────────┐
      │                 │
 Success             Failure
      │                 │
 Mark Sent      Retry with Backoff
```

Never send SMTP directly inside the HTTP request.

---

# Use an Outbox Pattern

Inside the same database transaction:

```sql
BEGIN;

INSERT INTO invoices ...

INSERT INTO email_outbox (...)

COMMIT;
```

This guarantees:

* invoice exists
* email request exists

Never one without the other.

---

## Example Outbox Table

```sql
email_outbox

id
event_type
recipient
subject
template
payload_json

status
attempts
next_retry_at

created_at
sent_at
last_error
provider_message_id
```

Status

```
Pending

Processing

Sent

Failed

Dead Letter
```

---

# Email Worker

Worker runs every few seconds.

```
SELECT *
FROM email_outbox
WHERE status='Pending'
AND next_retry_at <= now()
LIMIT 100
FOR UPDATE SKIP LOCKED;
```

`SKIP LOCKED` allows many workers simultaneously.

---

# Retry Strategy

Don't retry immediately.

Example:

Attempt 1

```
5 sec
```

Attempt 2

```
30 sec
```

Attempt 3

```
2 min
```

Attempt 4

```
10 min
```

Attempt 5

```
1 hour
```

Attempt 6

```
6 hours
```

Then move to Dead Letter.

---

# Dead Letter Queue

Never retry forever.

Example

```
Max Attempts = 8
```

If exceeded:

```
status = DEAD_LETTER

last_error =
"Mailbox unavailable"

attempts = 8
```

Admin can manually retry later.

---

# Idempotency

Suppose worker crashes after provider accepted email.

Without idempotency:

```
Invoice email sent twice.
```

Each email should have:

```
message_id

or

idempotency_key
```

Example

```
invoice_9843
```

Worker checks:

Already sent?

Yes

Don't send again.

---

# Email Templates

Never hardcode HTML.

```
templates/

invoice.html

invite.html

forgot_password.html

receipt.html

welcome.html
```

Variables:

```
{{parent_name}}

{{invoice_number}}

{{amount}}

{{payment_link}}
```

---

# Version Templates

```
template_name

version

published_at
```

Useful when invoice design changes.

---

# Audit Log

Store every attempt.

```
email_log

email_id

attempt

provider

status

response

created_at
```

Very useful for debugging.

---

# Webhooks

Providers like SendGrid, Mailgun, or Postmark provide delivery events.

Example:

```
Queued

Delivered

Opened

Clicked

Bounced

Spam

Dropped
```

Store:

```
email_delivery

provider_message_id

status

timestamp
```

Now support can answer:

> "Did parent receive invoice?"

instead of

> "I don't know."

---

# Provider Abstraction

Don't tie yourself to one provider.

```go
type EmailProvider interface {

    Send(ctx context.Context, Email) error
}
```

Implementations:

```
SMTP

AWS SES

Postmark

Mailgun

SendGrid

Resend
```

Switch providers without changing business logic.

---

# Rate Limiting

Suppose:

```
1,000 invoices
```

Don't send all simultaneously.

Worker configuration:

```
20 emails/sec
```

or

```
200/minute
```

depending on provider limits.

---

# Batch Sending

For monthly invoices:

```
10,000 invoices
```

Don't create one HTTP request.

Instead:

```
Generate invoices

↓

Create 10,000 outbox rows

↓

Workers process gradually
```

---

# Password Reset

Never store reset code.

Store only hash.

```
token_hash

expires_at

used_at
```

Flow

```
Generate random token

↓

Hash(token)

↓

Store hash

↓

Email raw token

↓

User clicks link

↓

Hash incoming token

↓

Compare hash
```

---

# Invite User

Flow

```
Admin

↓

Create User

↓

Create Invite Token

↓

Insert Outbox

↓

Return Success

↓

Worker Sends Email
```

Invitation expires:

```
72 hours
```

---

# Invoice Email

Flow

```
Generate Invoice

↓

PDF

↓

Upload PDF (optional)

↓

Insert Outbox

↓

Worker

↓

Send Email

↓

Webhook

↓

Delivered
```

---

# Monitoring

Useful metrics include:

```
Emails Sent

Emails Failed

Retry Count

Average Delivery Time

Bounce Rate

Spam Rate

Queue Size

Dead Letters
```

If queue size increases rapidly:

```
Pending:
12,491
```

You know the worker is unhealthy.

---

# Error Handling Matrix

| Failure                        | User Response                                                                                                               | Background Action                               |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| SMTP timeout                   | Business operation succeeds                                                                                                 | Retry with exponential backoff                  |
| Provider returns 500           | Business operation succeeds                                                                                                 | Retry                                           |
| Invalid recipient email        | Business operation succeeds                                                                                                 | Mark as failed, notify admin if appropriate     |
| Network outage                 | Business operation succeeds                                                                                                 | Retry                                           |
| Database commit fails          | Entire request rolls back                                                                                                   | No email job created                            |
| Worker crashes mid-send        | Worker resumes processing                                                                                                   | Idempotency prevents duplicates                 |
| Email template rendering fails | Business operation succeeds                                                                                                 | Log error, mark failed, alert developers        |
| Attachment generation fails    | Business operation succeeds if attachment is optional; otherwise fail the business operation if the attachment is essential | Retry or move to dead letter based on the error |

# Recommended Stack for Your Go + PostgreSQL NMS

Given your architecture, I'd recommend:

* **PostgreSQL Outbox Table** as the durable queue (avoid introducing RabbitMQ/Kafka unless you truly need them).
* **Background workers** using `FOR UPDATE SKIP LOCKED` for safe concurrent processing.
* **Provider abstraction** so you can switch between SMTP, AWS SES, Postmark, SendGrid, or Resend without changing application logic.
* **HTML templates** with versioning and reusable layouts.
* **Exponential backoff**, dead-letter handling, and structured logging.
* **Webhook processing** for delivery, bounce, and complaint events to keep accurate delivery status.
* **Observability** with metrics, alerts, and an admin page showing pending, sent, failed, and dead-letter emails.

This design is resilient, horizontally scalable, and follows patterns commonly used in production SaaS systems. It can comfortably support anything from a handful of password-reset emails to tens of thousands of monthly invoice emails without making the user wait for email delivery.
