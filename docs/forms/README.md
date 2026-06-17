# Owner-Provided Child Management Forms

Source forms:

- `child-application-form.md`
- `parental-consent-form.md`

These forms are implementation inputs for post-MVP child management work. The raw Markdown is intentionally preserved as received; use the backlog tickets for product scope and field normalization.

## Backlog Links

- API: `docs/MVP-30D-API-BACKEND-BACKLOG.md`
  - `API-PM-08` child management refactor — atomic create + per-resource records
  - `API-PM-02` consent and acknowledgement ledger (separate from the current single-row consent record)
- Frontend: `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`
  - `FE-PM-08` manager child edit stepper (replaces the legacy guided registration intake)
  - `FE-PM-02` parent/guardian digital registration and consent journey

## Field Groups

The paper forms normalize into the new child sub-record tables. Field-by-field mapping:

- **Child identity** (`children`): first name, middle name, last name, date of birth, start date, end date, notes.
- **Profile** (`child_profiles`): sex, religion, ethnic origin, first language, other languages, home address, postcode, telephone, disability status/notes, access requirements, routine care notes, GDPR declaration metadata, registration date, 7 section-review booleans.
- **Health** (`child_health_profiles`): allergies/conditions, prescribed medication, immunisation status/country, illness or diagnosis history, dietary requirements, side effects, doctor name/address/telephone, health visitor name/address/telephone.
- **Safeguarding** (`child_safeguarding_profiles`): social-services referral/contact, concerns about walking / speech-language / hearing / sight / emotional wellbeing / behaviour, professional referrals, restricted notes.
- **Contacts** (`child_contacts`): parent/carers, emergency contacts, authorised collectors — with `contact_type` discriminator. Each row has full_name, relationship, address, telephone, email, work_address, has_parental_responsibility.
- **Funding** (`child_funding_records`): benefits contributing to fees, working tax credit, college/university payment routing, 2-year-old and 3-year-old funding, support notes, reviewed flag.
- **Consent** (`child_consent_records`): urgent medical treatment, plasters, safeguarding reporting acknowledgement, GDPR data processing, SENCO discussion, health visitor discussion, transition documents, outings, face painting, sun cream, nappy cream, photograph purposes, promotional/website use, coursework, social media. Single current row per child.
- **Collection** (`child_collection_settings`): over-18 collection acknowledgement, hashed collection password + metadata.
- **Room** (`child_room_assignments`): opening room assignment on create (room_id, start_date).
- **Billing** (`child_billing_profiles`): site-rate or custom billing basis.

## Implementation Notes

- All sub-records are tenant/branch scoped and linked to a child.
- Treat medical, social-care, collection-password, and consent data as sensitive. Do not expose these through practitioner attendance, parent invoice, or billing APIs.
- The current consent record is a single current row per child (PUT replaces it). A versioned consent-history ledger is a separate follow-up (`API-PM-02`).
- Actual document upload/storage is a separate follow-up unless explicitly prioritized; the paper-form `on_file` flag captures metadata only.
