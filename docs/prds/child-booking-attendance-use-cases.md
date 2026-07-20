Using the Pareto Principle (80/20), you do **not** need to build all of these use cases for an MVP. Most UK nurseries operate on a relatively small number of recurring workflows. If you implement those well, you'll cover around **80–90% of daily operations**.

---

# Tier 1 — Core MVP (Build First)

These are the features used almost every day.

## 1. Regular Recurring Booking ⭐⭐⭐⭐⭐

This is the heart of every NMS.

Support:

* Weekly recurring schedule
* Monday–Sunday selection
* Effective From
* Effective Until (optional)
* Session selection
* Room assignment

Examples

* Mon Full Day
* Tue Morning
* Wed Full Day
* Fri School Day

Covers:

* Full Day
* Half Day Morning
* Half Day Afternoon
* School Day
* Weekly recurring booking

**Business value:** Extremely high

---

## 2. One-off (Ad-hoc) Booking ⭐⭐⭐⭐⭐

Allow parents or staff to add a single booking.

Example

Friday only

09:00–15:00

Used for:

* Extra childcare
* Emergency childcare
* School holidays

Covers:

* One-off Booking
* Ad-hoc Booking
* Emergency Booking
* Extra Session

---

## 3. Attendance Register ⭐⭐⭐⭐⭐

Every nursery uses this daily.

Support

* Check In
* Check Out
* Present
* Absent
* Sick
* Holiday

Nice-to-have later

* Late arrival
* Early collection

---

## 4. Funding Allocation ⭐⭐⭐⭐⭐

Support

* No Funding
* 15 Hours
* 30 Hours
* Stretched

Store

* Funding type
* Effective dates
* Hours available

No need for complicated claim engine initially.

---

## 5. Billing Engine ⭐⭐⭐⭐⭐

Generate invoices from bookings.

Invoice lines

* Sessions
* Meals
* Extras
* Discounts

Ignore advanced calculations initially.

---

## 6. Parent Payments ⭐⭐⭐⭐☆

Support

* Card Payment
* Bank Transfer
* Direct Debit (manual reconciliation initially)

---

## 7. Meals & Extras ⭐⭐⭐⭐☆

Simple checkboxes.

Examples

Breakfast

Lunch

Tea

Consumables

Activities

---

## 8. Booking Changes ⭐⭐⭐⭐☆

Support

* Cancel booking
* Edit booking
* Permanent schedule change

---

# Tier 2 — Important (Version 2)

These features add flexibility once the core is stable.

## Wraparound Care

Support

* Early Drop-off
* Late Collection

Internally these can simply be additional paid sessions.

---

## Hourly Booking

Instead of predefined sessions.

Useful for some nurseries but not the majority.

---

## Session Swap

Cancel Tuesday

↓

Attend Thursday

Requires approval workflow.

---

## Holiday Club

Separate seasonal sessions.

---

## School Clubs

* Breakfast Club
* After School Club

---

## Multiple Children Invoice

One invoice for the family.

---

## Waiting List

Enquiry

↓

Waiting List

↓

Offer

↓

Accept

---

## Occupancy Reports

Needed by managers.

---

# Tier 3 — Advanced

Build only when customers request them.

### Shared Custody

Invoice splitting.

---

### Mixed Weekly Funding

Different funding by weekday.

---

### Attendance Segmentation

Example

08:00–09:00 Paid

09:00–15:00 Funded

15:00–18:00 Paid

---

### Transport

Rare.

---

### Forest School

Optional.

---

### Dance

Optional.

---

### Music

Optional.

---

### Languages

Optional.

---

### Holiday Programmes

Summer Club

Half-Term Club

---

### Voucher Integration

Legacy support.

---

### Tax-Free Childcare Automation

Automated reconciliation.

---

### Funding Claim Export

Generate Local Authority claim files.

---

### Staff Ratio Optimisation

Real-time staffing compliance.

---

### Advanced Analytics

Revenue forecasting

Occupancy prediction

Funding utilization

---

# Simplify the Booking Model

Instead of creating separate entities for every booking type, model them with a **single `Booking` entity** plus a reusable **Session Type**.

## Booking

```text
Booking
--------
id
child_id
booking_type
recurrence_type
status
effective_from
effective_until
```

`booking_type`

* Regular
* Ad-hoc

`recurrence_type`

* Weekly
* One-off

## Session Type

```text
Session Type
------------
Full Day
Morning
Afternoon
School Day
Breakfast Club
After School Club
Hourly
```

The `Session Type` defines:

* Start time
* End time
* Duration
* Base price
* Default meals
* Whether funding can apply

This avoids creating separate booking models for Full Day, Half Day, Wraparound, etc.

---

# Attendance Model

```text
Booking
    ↓
Attendance
    ↓
Invoice
```

Attendance records what actually happened, while invoices are generated from bookings (or attendance, depending on nursery policy). This separation supports configurable charging rules without changing the booking model.

---

# Funding Model

Keep funding independent from bookings.

```text
Child
   ↓
Funding Profile
   ↓
Booking
   ↓
Invoice
```

At invoice generation:

1. Calculate booked sessions.
2. Apply eligible funded hours.
3. Bill remaining private hours.
4. Add meals and extras.
5. Apply discounts.
6. Apply vouchers or Tax-Free Childcare.
7. Produce the final invoice.

This separation also allows funding claims and parent invoices to evolve independently.

---

# Recommended MVP Scope

| Feature                   | Priority |
| ------------------------- | -------- |
| Recurring Booking         | ⭐⭐⭐⭐⭐    |
| One-off Booking           | ⭐⭐⭐⭐⭐    |
| Attendance Register       | ⭐⭐⭐⭐⭐    |
| Funding (15/30/Stretched) | ⭐⭐⭐⭐⭐    |
| Invoice Generation        | ⭐⭐⭐⭐⭐    |
| Parent Payments           | ⭐⭐⭐⭐☆    |
| Meals & Extras            | ⭐⭐⭐⭐☆    |
| Booking Amendments        | ⭐⭐⭐⭐☆    |
| Wraparound Care           | ⭐⭐⭐☆☆    |
| Hourly Booking            | ⭐⭐⭐☆☆    |
| Holiday Club              | ⭐⭐☆☆☆    |
| School Clubs              | ⭐⭐☆☆☆    |
| Shared Custody            | ⭐☆☆☆☆    |
| Attendance Segmentation   | ⭐☆☆☆☆    |
| Transport                 | ⭐☆☆☆☆    |
| Advanced Funding Claims   | ⭐☆☆☆☆    |

This prioritization delivers the core workflows that most UK nurseries perform every day while keeping the domain model clean and extensible for future features like holiday clubs, complex funding arrangements, and advanced billing rules.
