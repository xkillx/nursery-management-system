# UK Nursery Management System (NMS) - Complete Funding Logic

> A comprehensive funding model for a UK Nursery Management System (SaaS).

---

# 1. Funding Sources

A child may have one or multiple funding sources.

| Funding Source | Paid By |
|---------------|---------|
| Government Funding | Local Authority |
| Parent Payment | Parent |
| Tax-Free Childcare | HMRC |
| Employer Childcare Scheme | Employer |
| Childcare Voucher (Legacy) | Voucher Provider |
| Charity / Sponsor | Third Party |

### Example

```text
Weekly Fee = £300

Government Funding = £180
Tax-Free Childcare = £60
Parent Pays = £60
```

---

# 2. Government Funding Types

Support the following funding types:

| Funding Type | Description |
|--------------|-------------|
| Working Parent (9 months+) | Government funded childcare |
| FRAS (Families Receiving Additional Support) | Eligible 2-year-olds |
| Universal 15 Hours | All eligible 3–4 year olds |
| Additional 15 Hours | Working parents (total 30 hours) |

---

# 3. Funding Delivery Model

## A. Term-Time Funding

Funding only during school terms.

```text
15 hours/week
38 weeks/year

Total = 570 hours/year
```

or

```text
30 hours/week
38 weeks/year

Total = 1140 hours/year
```

---

## B. Stretched Funding

Same yearly entitlement spread across more weeks.

Example

```text
570 hours
÷ 51 weeks
= 11.18 hours/week
```

Example

```text
1140 hours
÷ 51 weeks
= 22.35 hours/week
```

---

# 4. Nursery Calendar

Nurseries may operate:

- 38 weeks
- 39 weeks
- 50 weeks
- 51 weeks
- 52 weeks

Example

```text
Open:
51 weeks

Closed:
Christmas Week
```

---

# 5. Funding Hours Tracking

Store:

- Allocated Hours
- Used Hours
- Claimed Hours
- Remaining Hours

Example

```text
Allocated : 570
Used      : 415
Remaining : 155
```

---

# 6. Funding Allocation

One booking may be split between funded and private hours.

Example

```text
Booking = 9 hours

Funded = 6 hours
Private = 3 hours
```

Invoice

```text
Government = 6h
Parent = 3h
```

---

# 7. Session Funding

Support different session types.

Morning

```text
08:00 - 13:00
5 Hours
```

Afternoon

```text
13:00 - 18:00
5 Hours
```

Full Day

```text
08:00 - 18:00
10 Hours
```

Funding may cover:

- Entire session
- Partial session
- Multiple sessions

---

# 8. Holiday Funding

Government funding normally covers:

```text
School Term Only
```

Private bookings may continue during holidays.

Example

```text
Summer Holiday

Government Funding = 0
Private Booking = 100%
```

---

# 9. Nursery Closure Logic

Support closures such as:

- Christmas
- Bank Holidays
- Staff Training Days (INSET)
- Emergency Closure

Normally no government funding can be claimed on closed days.

---

# 10. Child Funding Eligibility

Each child should store:

- Funding Start Date
- Funding End Date
- Eligibility Code
- Validation Status
- Renewal Date
- Funding Type

---

# 11. Funding Status

```text
Pending
Approved
Rejected
Expired
Cancelled
```

---

# 12. Funding Claim Status

```text
Draft
Submitted
Accepted
Partially Accepted
Rejected
Paid
```

---

# 13. Local Authority Information

Store:

- Local Authority
- Claim Period
- Funding Rate
- Payment Date
- Claim Reference
- Payment Reference

---

# 14. Funding Rate

Rates vary depending on:

- Child Age
- Funding Type
- Local Authority
- Financial Year

---

# 15. Invoice Calculation

```text
Total Fees

- Government Funding

- Discounts

+ Meals

+ Snacks

+ Consumables

+ Extra Hours

+ Late Collection

+ Registration Fee

+ Deposit

= Parent Balance
```

---

# 16. Consumables

Usually NOT funded.

Examples:

- Meals
- Snacks
- Milk
- Formula
- Nappies
- Wipes
- Sun Cream
- Trips
- Activities

---

# 17. Additional Charges

Support:

- Late Pickup
- Extra Hours
- Extra Session
- Holiday Club
- Deposit
- Registration Fee
- Administration Fee
- Emergency Contact Fee
- Transport Fee

---

# 18. Funding Audit Trail

Track every funding change.

Store:

- User
- Date
- Old Value
- New Value
- Reason

Useful for Local Authority audits.

---

# 19. Funding Adjustments

Support:

- Backdated Funding
- Manual Adjustment
- Refund
- Credit Note
- Additional Funded Hours
- Reduced Funded Hours

---

# 20. Funding Reports

Generate reports for:

- Funded Hours
- Used Hours
- Remaining Hours
- Claimed Hours
- Parent Contribution
- Government Contribution
- Funding by Child
- Funding by Room
- Funding by Age
- Funding by Local Authority
- Rejected Claims
- Expired Eligibility
- Forecast Funding Income

---

# 21. Payment Allocation

A single invoice may be paid using multiple payment methods.

Example

```text
Invoice = £350

Government = £200

Tax-Free Childcare = £50

Parent Card = £100
```

Support:

- Split Payments
- Partial Payments
- Overpayments
- Credit Balance

---

# 22. Booking vs Attendance

Government funding is generally based on attendance records.

Track:

- Booked Hours
- Attended Hours
- Absent
- Holiday
- Sick
- No Show

---

# 23. Funding Validation Rules

Before generating an invoice:

- Child eligible?
- Funding dates valid?
- Eligibility code valid?
- Funding hours available?
- Session claimable?
- Nursery open?
- Claim period open?
- Duplicate claim?
- Age eligible?

---

# 24. Funding Priority Rules

Example order:

1. Government Funding
2. Tax-Free Childcare
3. Childcare Voucher
4. Employer Contribution
5. Parent Payment

---

# 25. Financial Year Support

Funding changes every financial year.

Store:

- Financial Year
- Effective Date
- Funding Rate Version

Example

```text
2026/2027

Working Parent Rate

Universal Rate

FRAS Rate
```

---

# 26. Suggested Database Tables

```text
Child

Parent

Booking

BookingSession

Attendance

FundingEligibility

FundingAllocation

FundingRate

FundingRule

FundingClaim

FundingClaimItem

FundingAdjustment

FundingAudit

FundingCalendar

Holiday

Closure

Invoice

InvoiceLine

Payment

PaymentAllocation

LocalAuthority
```

---

# 27. Funding Engine Processing Flow

```text
Child Booking

        │
        ▼

Validate Eligibility

        │
        ▼

Determine Funding Type

        │
        ▼

Determine Term-Time / Stretched

        │
        ▼

Calculate Available Hours

        │
        ▼

Allocate Funded Hours

        │
        ▼

Calculate Private Hours

        │
        ▼

Add Consumables

        │
        ▼

Add Extra Charges

        │
        ▼

Generate Invoice

        │
        ▼

Allocate Payments

        │
        ▼

Generate Funding Claim

        │
        ▼

Create Audit Log
```

---

# 28. Recommended Design Principles

- Rule-based funding engine
- Configurable funding rules
- Local Authority specific settings
- Versioned funding rates
- Full audit trail
- Flexible invoice allocation
- Support multiple funding sources
- Automatic funding hour calculation
- Configurable nursery calendars
- Financial year versioning
- Extensible for future UK government funding changes

---

# Summary

A production-ready UK Nursery Management System should support:

- ✅ Multiple funding sources
- ✅ Government funding (15h, 30h, FRAS, Working Parent)
- ✅ Term-Time and Stretched funding
- ✅ Nursery calendars
- ✅ Holiday and closure rules
- ✅ Funding eligibility validation
- ✅ Hour allocation engine
- ✅ Funding claims
- ✅ Split invoice calculations
- ✅ Tax-Free Childcare
- ✅ Childcare vouchers
- ✅ Additional charges
- ✅ Attendance tracking
- ✅ Audit trail
- ✅ Local Authority management
- ✅ Reporting
- ✅ Financial year versioning
- ✅ Configurable funding rules
- ✅ Multi-payment allocation
- ✅ Extensible funding engine architecture
