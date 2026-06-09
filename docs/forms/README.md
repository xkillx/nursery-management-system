# Owner-Provided Registration Forms

Source forms:

- `child-application-form.md`
- `parental-consent-form.md`

These forms are implementation inputs for post-MVP registration/enrolment work. The raw Markdown is intentionally preserved as received; use the backlog tickets for product scope and field normalization.

## Backlog Links

- API: `docs/MVP-30D-API-BACKEND-BACKLOG.md`
  - `API-PM-01` registration/enrolment profile data
  - `API-PM-02` consent and acknowledgement ledger
  - `API-PM-03` office-use enrolment checklist metadata
- Frontend: `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`
  - `FE-PM-01` manager registration/enrolment editor
  - `FE-PM-02` parent/guardian digital registration and consent journey
  - `FE-PM-03` manager consent review/history UI
  - `FE-PM-04` office-use checklist UI

## Field Groups

Normalize the forms into these product sections before implementation:

- Child identity and demographics: full name, date of birth, sex, religion, ethnicity, first language, other languages, disability, access requirements, home address, postcode, telephone.
- Medical profile: allergies/conditions, medication, immunisation state/country, illness or diagnosis history, dietary requirements, side effects.
- Health contacts: doctor name/address/telephone, health visitor name/address/telephone.
- Social care and development: social-services referral/contact, concerns about walking, speech/language, hearing, sight, emotional wellbeing, behaviour, referrals to paediatrician, speech/language therapist, early support service, or other professionals.
- Parents/carers and responsibility: parent/carer names, parental responsibility, addresses, phone numbers, emails, work addresses.
- Funding/support notes: benefits contributing to nursery fees, working tax credit, college/university payment routing, 2-year-old and 3-year-old funding, other notes.
- Emergency and collection: emergency contacts, authorised collectors, relationship, address/telephone, collection password policy, over-18 collection rule.
- Child routine notes: special words, comforter, routines, and other care notes.
- Declarations and consent: urgent medical treatment, plasters, safeguarding reporting acknowledgement, GDPR declaration, SENCO discussion, health visitor discussion, transition documents, outings, face painting, sun cream, nappy cream, photograph purposes, promotional/website use, coursework, social media.
- Office-use enrolment checks: deposit, application date, start date, date left, sessions/days requested, term-time-only space, contract, handbook, Red Book, birth certificate/passport, proof of address.

## Implementation Notes

- Keep registration and consent fields tenant/branch scoped and linked to a child.
- Treat medical, social-care, collection-password, and consent data as sensitive. Do not expose these through practitioner attendance, parent invoice, or billing APIs.
- Keep consent history rather than overwriting prior signed decisions.
- Actual document upload/storage is a separate follow-up unless explicitly prioritized; `API-PM-03` tracks metadata/checklist status only.
