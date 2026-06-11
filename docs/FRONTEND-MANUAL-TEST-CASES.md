# Frontend Manual Test Cases

## Document Control

- Product: Nursery Management System web frontend
- Area: Angular web app only
- Scope: Routed frontend modules, role navigation, forms, state handling, responsive behavior, and user-visible errors
- Out of scope: Backend endpoint correctness, database validation, Stripe internals, email delivery internals, server logs, and direct API contract testing
- Created for: Manual QA execution

## Test Approach

Test the frontend as a user would use it. The backend may provide data and responses, but this document does not ask QA to verify API payloads, database rows, jobs, webhooks, or backend authorization rules directly. When a case depends on a specific data condition, use prepared QA data or a frontend/API mock environment and assert only the browser-visible result.

Run the core cases on:

- Desktop: 1440 x 900
- Tablet: 768 x 1024
- Mobile: 390 x 844
- Browsers: Chrome plus one of Safari, Firefox, or Edge

Record evidence for failed cases: page URL, role used, viewport, screenshots, visible message text, and console errors if present.

## Required QA Data

Prepare these accounts and records before execution:

- Manager account with one active nursery membership.
- Manager account with multiple memberships, if membership selection is enabled.
- Practitioner account with access only to attendance.
- Parent account linked to at least one child.
- Active and inactive children.
- Child with complete enrollment and child with missing enrollment requirements.
- Child with no linked guardian and child with at least one linked guardian.
- Active guardian available for linking.
- Funding profiles for missing allowance, zero allowance, under one hour, normal allowance, and above 160 hours.
- Attendance records for not checked in, checked in, absent, incomplete session, and completed session.
- Attendance correction history for at least one session.
- Invoice preflight data with eligible children and blocked children.
- Draft, issued, overdue, payment failed, and paid invoices.
- Parent payable invoice, paid invoice, and payment failed invoice.
- Valid, expired, revoked, and already accepted invitation links.
- Valid, expired, and used password reset links.

## Route Coverage

| Role | Primary routes |
| --- | --- |
| Public | `/signin`, `/signup`, `/forgot-password`, `/reset-password`, `/invite-accept`, unknown route |
| Manager | `/staff/manager/dashboard`, `/staff/manager/children`, `/staff/manager/children/:childId`, `/staff/manager/guardians`, `/staff/manager/invites`, `/staff/practitioner/attendance`, `/staff/manager/attendance-corrections`, `/staff/manager/funding`, `/staff/manager/invoice-run`, `/staff/manager/invoices`, `/staff/manager/invoices/:invoiceId` |
| Practitioner | `/staff/practitioner/attendance` |
| Parent | `/parent/invoices`, `/parent/invoices/:invoiceId` |

## Global, Routing, And Layout

### GL-001 - Unauthenticated route protection

Priority: High

Preconditions:

- User is signed out.

Steps:

1. Open `/staff/manager/dashboard`.
2. Observe the current page.
3. Open `/staff/practitioner/attendance`.
4. Observe the current page.
5. Open `/parent/invoices`.
6. Observe the current page.

Expected result:

- Each protected route redirects to `/signin`.
- Sign-in page is usable and does not show protected page content.
- If a redirect query parameter is present, it references the originally requested URL.

### GL-002 - Role default landing routes

Priority: High

Preconditions:

- Test accounts exist for manager, practitioner, and parent.

Steps:

1. Sign in as manager and open `/`.
2. Sign out.
3. Sign in as practitioner and open `/`.
4. Sign out.
5. Sign in as parent and open `/`.

Expected result:

- Manager lands on `/staff/manager/dashboard`.
- Practitioner lands on `/staff/practitioner/attendance`.
- Parent lands on `/parent/invoices`.

### GL-003 - Wrong-role route redirects

Priority: High

Preconditions:

- User is signed in as practitioner.

Steps:

1. Open `/staff/manager/children`.
2. Observe the redirect target.
3. Open `/parent/invoices`.
4. Observe the redirect target.

Expected result:

- Practitioner cannot view manager or parent content.
- Practitioner is redirected back to the practitioner attendance route.

Repeat as parent and manager:

1. Sign in as parent and open `/staff/manager/dashboard`.
2. Confirm parent is redirected to `/parent/invoices`.
3. Sign in as manager and open `/parent/invoices`.
4. Confirm manager is redirected to manager default route.

### GL-004 - Staff sidebar navigation

Priority: High

Preconditions:

- User is signed in as manager on desktop viewport.

Steps:

1. Confirm the sidebar shows four group headings: Overview, People, Attendance, Billing.
2. Confirm these manager links are visible: Dashboard, Children, Guardians, Invites, Attendance, Attendance corrections, Funding, Invoice run, Invoices.
3. Confirm each link has a distinct icon (no repeated generic user icon).
4. Click each visible link once.
5. Confirm the active link shows a left-rail indicator plus brand color, and has `aria-current="page"`.
6. Navigate to a child detail page (e.g. `/staff/manager/children/:childId`). Confirm `Children` remains highlighted.
7. Navigate to an invoice detail page (e.g. `/staff/manager/invoices/:invoiceId`). Confirm `Invoices` remains highlighted.
8. Navigate to attendance corrections with query params (e.g. `?childId=...`). Confirm `Attendance corrections` remains highlighted.
9. Collapse the sidebar with the header toggle.
10. Confirm collapsed sidebar is 90px wide and shows only icons.
11. Hover over a collapsed link. Confirm a native title tooltip appears with the link label.
12. Confirm collapsed links have accessible names (inspect element for `title` attribute).
13. Hover over the collapsed sidebar area. Confirm labels reappear in grouped layout.

Expected result:

- Only manager-appropriate staff links are visible.
- Links are grouped under headings: Overview, People, Attendance, Billing.
- Each link has a distinct, recognizable icon at 24px.
- Each link routes to the expected page.
- Active state uses color plus left-rail indicator and sets `aria-current="page"`.
- Child detail and invoice detail routes keep their parent links highlighted.
- Query params do not break active state.
- Collapsed mode shows icons with title tooltips and stable 90px width.
- Expanded hover mode shows grouped layout matching the expanded sidebar.

### GL-005 - Practitioner sidebar privacy

Priority: High

Preconditions:

- User is signed in as practitioner.

Steps:

1. Open `/staff/practitioner/attendance`.
2. Inspect sidebar navigation.

Expected result:

- Only Attendance appears.
- Children, Guardians, Invites, Funding, Invoice run, and Invoices are not visible.
- No guardian contact or billing navigation is exposed.

### GL-006 - Parent portal navigation

Priority: High

Preconditions:

- User is signed in as parent.

Steps:

1. Open `/parent/invoices` on desktop.
2. Confirm the parent layout uses top navigation, not the staff sidebar.
3. Confirm `Invoices`, theme toggle, and user dropdown are visible.
4. Resize to mobile width.
5. Confirm the mobile nav still exposes `Invoices`.

Expected result:

- Parent sees parent-only navigation.
- No staff sidebar or manager/practitioner links appear.

### GL-007 - Theme toggle

Priority: Medium

Preconditions:

- User is signed in.

Steps:

1. Click the theme toggle.
2. Confirm colors switch between light and dark mode.
3. Navigate to another route.
4. Confirm the selected theme persists during navigation.
5. Repeat on public auth pages.

Expected result:

- Theme changes are immediate and readable.
- Text, icons, inputs, tables, and alerts retain sufficient contrast.

### GL-008 - User dropdown and sign out

Priority: High

Preconditions:

- User is signed in.

Steps:

1. Click the user avatar/name control.
2. Confirm the dropdown opens.
3. Confirm email/display name and session label match the role.
4. Click outside or close the dropdown.
5. Reopen the dropdown.
6. Click `Sign out`.

Expected result:

- Dropdown opens and closes reliably.
- Session label is Manager session, Practitioner session, or Parent session as appropriate.
- Sign out routes to `/signin`.
- Protected routes are no longer accessible after sign out.

### GL-009 - Unknown route

Priority: Medium

Preconditions:

- None.

Steps:

1. Open `/not-a-real-page`.
2. Click `Back to Home Page`.

Expected result:

- 404 page displays an error illustration and user-friendly copy.
- Back link returns to the correct default route for the current session, or sign-in if signed out.

## Authentication And Onboarding

### AUTH-001 - Successful sign-in, single membership

Priority: Critical

Preconditions:

- User has one active membership.

Steps:

1. Open `/signin`.
2. Enter a valid email.
3. Enter a valid password.
4. Optionally check `Keep me logged in`.
5. Click `Sign in`.

Expected result:

- Button shows `Signing in...` while submitting.
- User lands on the default route for their role.
- No membership picker appears.

### AUTH-002 - Sign-in validation and password visibility

Priority: High

Preconditions:

- User is on `/signin`.

Steps:

1. Submit with blank email and blank password.
2. Confirm field-level or form-level error messaging.
3. Enter a malformed email and submit.
4. Enter a password.
5. Click the password visibility icon.
6. Click it again.

Expected result:

- Errors are visible near the affected fields or in the form alert.
- Password field toggles between hidden and visible.
- Editing email or password clears stale membership challenge state.

### AUTH-003 - Multi-membership selection

Priority: High

Preconditions:

- Account has multiple available memberships.

Steps:

1. Open `/signin`.
2. Enter valid email and password.
3. Click `Sign in`.
4. Observe the membership selection panel.
5. Click `Continue` without selecting a membership.
6. Select one membership choice.
7. Click `Continue`.

Expected result:

- Membership choices show tenant, branch, and role.
- Continue is disabled until a membership is selected.
- Selected membership has visible selected styling and `aria-pressed` behavior.
- User lands on the default route for the selected membership role.

### AUTH-004 - Forgot password request

Priority: High

Preconditions:

- User is on `/forgot-password`.

Steps:

1. Submit with blank email.
2. Enter an invalid email and submit.
3. Enter a valid email and click `Send reset instructions`.
4. Click `Back to sign in`.

Expected result:

- Blank email shows `Enter your email address.`
- Invalid email shows `Enter a valid email address.`
- Valid submit shows `Sending...`, then `Check your email`.
- Confirmation copy does not reveal whether the account exists.
- Back link routes to `/signin`.

### AUTH-005 - Reset password valid token

Priority: Critical

Preconditions:

- Valid reset token exists.

Steps:

1. Open `/reset-password?token=<valid-token>`.
2. Submit blank fields.
3. Enter a password shorter than 8 characters.
4. Enter a valid new password and a different confirmation.
5. Enter matching valid passwords.
6. Click `Reset password`.
7. Click `Back to sign in`.

Expected result:

- Blank new password shows `Enter a new password.`
- Short password shows `Password must be at least 8 characters.`
- Mismatch shows `Passwords do not match.`
- Successful submit shows `Resetting...`, then `Password reset`.
- Session is cleared and user can return to sign-in.

### AUTH-006 - Reset password unusable link

Priority: High

Preconditions:

- No token, expired token, invalid token, or used token scenario exists.

Steps:

1. Open `/reset-password` without token.
2. Open `/reset-password?token=<expired-or-used-token>`.
3. Click `Request a new reset link`.
4. Click `Back to sign in`.

Expected result:

- Page shows `Reset link unusable`.
- Request link routes to `/forgot-password`.
- Back link routes to `/signin`.

### AUTH-007 - Invite acceptance valid token

Priority: Critical

Preconditions:

- Valid invite token exists.

Steps:

1. Open `/invite-accept?token=<valid-token>`.
2. Submit blank fields.
3. Enter a short password.
4. Enter mismatched password confirmation.
5. Enter matching valid passwords.
6. Click `Accept invitation`.
7. Click `Sign in`.

Expected result:

- Validation messages match reset password rules.
- Button shows `Accepting...` while submitting.
- Successful submit shows `Invitation accepted`.
- Sign-in link routes to `/signin`.

### AUTH-008 - Invite terminal states

Priority: High

Preconditions:

- Invalid, expired, already accepted, and revoked invite links exist.

Steps:

1. Open `/invite-accept` without token.
2. Open an expired invite link.
3. Open an already accepted invite link.
4. Open a revoked invite link.
5. Click the sign-in/back links on each state.

Expected result:

- Missing or invalid token shows `Invitation link unusable`.
- Expired token shows `Invitation expired`.
- Already accepted token shows `Invitation already accepted`.
- Revoked token shows `Invitation no longer available`.
- Links route to `/signin`.

### AUTH-009 - Public signup is blocked

Priority: High

Preconditions:

- User is signed out.

Steps:

1. Open `/signup`.
2. Review page content.
3. Click `Return to sign in`.

Expected result:

- No account creation form appears.
- Page states access is invitation-only.
- Link routes to `/signin`.

## Manager Dashboard

### MGR-DASH-001 - Dashboard content and quick navigation

Priority: High

Preconditions:

- User is signed in as manager.

Steps:

1. Open `/staff/manager/dashboard`.
2. Confirm page title `Manager operations`.
3. Confirm `Open attendance` action is visible.
4. Review `Today's attendance` summary tiles.
5. Review `Incomplete attendance`, `Invoice run status`, `Payment follow-up`, and `Quick actions` sections.
6. Click `Open attendance`.
7. Return to dashboard.
8. Click each enabled quick action.

Expected result:

- Dashboard uses nursery operations terminology, not ecommerce/template content.
- Attendance action routes to practitioner attendance page.
- Correct links route to the relevant manager workflows.
- Disabled quick actions are visibly disabled and not clickable.

### MGR-DASH-002 - Incomplete attendance deep link

Priority: Medium

Preconditions:

- Dashboard contains at least one incomplete attendance row with child/date data.

Steps:

1. Open manager dashboard.
2. In Incomplete attendance, click `Correct`.
3. Observe the attendance correction page.

Expected result:

- User lands on `/staff/manager/attendance-corrections`.
- Child, local date, and session are preselected when provided by the row.
- Correction page loads the matching sessions/history.

## Manager Children

### CHILD-001 - Children list, filters, and pagination

Priority: High

Preconditions:

- Manager has more than 10 children across active/inactive states, or use available data for partial execution.

Steps:

1. Open `/staff/manager/children`.
2. Confirm page title `Children`.
3. Confirm status filter defaults to Active.
4. Change status to Inactive.
5. Change status to All.
6. Confirm table columns: Name, DOB, Start, Rate, Active, Enrollment, Missing requirements, Linked guardians, Actions.
7. Click Next pagination when enabled.
8. Click Previous pagination.

Expected result:

- List reloads when status changes and offset resets.
- Loading row displays while data is loading.
- Empty state says `No children found` when no rows match.
- Rates display in GBP format.
- Active and enrollment badges are visible.
- Pagination buttons enable and disable based on available rows.

### CHILD-002 - Create child

Priority: Critical

Preconditions:

- User is signed in as manager.

Steps:

1. Open `/staff/manager/children`.
2. Click `Add child`.
3. Confirm `Create child` form appears.
4. Enter Full name.
5. Enter Date of birth.
6. Enter Start date.
7. Enter Core hourly rate.
8. Optionally enter End date and Notes.
9. Click `Create child`.

Expected result:

- Form shows required indicators for required fields.
- Button shows saving/loading state while submitting.
- On success, form closes and list reloads.
- New child is visible in the relevant filter.

### CHILD-003 - Create child validation and cancel

Priority: High

Preconditions:

- User is signed in as manager.

Steps:

1. Open the Create child form.
2. Submit with required fields blank.
3. Enter invalid or out-of-policy values, such as negative hourly rate if allowed by browser input.
4. Observe field and server error display.
5. Click `Cancel`.

Expected result:

- Required browser validation or field errors prevent a silent failure.
- Server validation appears as field-level messages or a form alert.
- Cancel closes the form and clears stale errors.

### CHILD-004 - Edit child

Priority: High

Preconditions:

- At least one child exists.

Steps:

1. Open `/staff/manager/children`.
2. Click `Edit` on a child row.
3. Confirm form title is `Edit child`.
4. Confirm existing values are prepopulated.
5. Change Notes or Core hourly rate.
6. Click `Save changes`.

Expected result:

- Existing child data populates accurately.
- Successful save closes the form and refreshes the row.
- Failed save keeps the form open and shows errors.

### CHILD-005 - Child detail navigation

Priority: High

Preconditions:

- At least one child exists.

Steps:

1. Open `/staff/manager/children`.
2. Click `View` on a child row.
3. Click `Back to children`.

Expected result:

- Detail page opens at `/staff/manager/children/:childId`.
- Back link returns to the children list.

## Manager Child Detail And Funding

### CHILD-DETAIL-001 - Child detail summary

Priority: High

Preconditions:

- Child exists with known enrollment state.

Steps:

1. Open the child detail page.
2. Confirm child name appears in the page header.
3. Review Child details section.
4. Review Enrollment status section.
5. If missing requirements exist, confirm each is listed with human-readable labels.

Expected result:

- DOB, start date, end date, active badge, hourly rate, and notes display correctly.
- Enrollment badge shows complete or incomplete state.
- Missing requirements are visible for incomplete enrollment.

### CHILD-DETAIL-002 - Edit child basics from detail

Priority: High

Preconditions:

- Child detail page is open.

Steps:

1. Click `Edit child basics`.
2. Change one editable field.
3. Click `Save changes`.
4. Reopen edit form.
5. Click `Cancel`.

Expected result:

- Same edit form behavior as Children list.
- Save refreshes child detail.
- Cancel closes the form and clears errors.

### CHILD-DETAIL-003 - Linked guardians display and empty state

Priority: High

Preconditions:

- One child has no linked guardians.
- One child has at least one linked guardian.

Steps:

1. Open child detail for a child with no linked guardians.
2. Review Linked guardians section.
3. Open child detail for a child with linked guardians.
4. Review linked guardian names, email/phone, and active badge.

Expected result:

- Empty child shows `No linked guardians` with explanatory copy.
- Linked child shows guardian details without layout overlap.

### CHILD-DETAIL-004 - Link an active guardian

Priority: Critical

Preconditions:

- Child has an available active guardian not already linked.

Steps:

1. Open child detail.
2. In `Link an active guardian`, open the select list.
3. Choose an available guardian.
4. Click `Link guardian`.
5. Observe the linked guardians list.

Expected result:

- Link button is disabled until a guardian is selected.
- Button shows disabled/loading state while linking.
- On success, selected guardian is added to linked guardians.
- The selected guardian is no longer offered in the available list.

### CHILD-DETAIL-005 - No active guardians path

Priority: Medium

Preconditions:

- No active guardians exist.

Steps:

1. Open child detail.
2. Review link guardian area.
3. Click `Go to guardians`.

Expected result:

- Message says no active guardians found.
- Link routes to `/staff/manager/guardians`.

### CHILD-DETAIL-006 - Save monthly funded-hours allowance

Priority: Critical

Preconditions:

- Child exists and selected month overlaps enrollment.

Steps:

1. Open child detail.
2. In Monthly funded-hours allowance, select a billing month.
3. Enter Hours `15`.
4. Enter Minutes `30`.
5. Click `Save allowance`.
6. Change the month away and back.

Expected result:

- Loading state appears while funding profile loads.
- Save button shows `Saving...` while submitting.
- Success message `Saved` appears.
- Last updated timestamp appears when a profile exists.
- Returning to the month shows the saved hours and minutes.

### CHILD-DETAIL-007 - Funding validation

Priority: High

Preconditions:

- Child detail page is open.

Steps:

1. Clear both Hours and Minutes.
2. Click `Save allowance`.
3. Enter Hours `-1`.
4. Click `Save allowance`.
5. Enter Minutes `60`.
6. Click `Save allowance`.
7. Enter an allowance above 744 hours.
8. Click `Save allowance`.

Expected result:

- Empty input shows `Enter an allowance or enter 0 to save no funded hours.`
- Negative hours shows `Hours must be a non-negative whole number.`
- Minutes above 59 shows `Minutes must be a whole number between 0 and 59.`
- Above maximum shows `Total allowance cannot exceed 744 hours (44640 minutes).`
- Invalid entries do not submit silently.

### CHILD-DETAIL-008 - Missing funding profile

Priority: Medium

Preconditions:

- Selected month has no funding profile.

Steps:

1. Select a month with no funding profile.
2. Observe the funding editor.

Expected result:

- Hours and Minutes are blank.
- Page shows `Not set for this month`.
- User can save explicit zero by entering `0` and `0`.

## Manager Guardians

### GUARD-001 - Guardian list, filters, and pagination

Priority: High

Preconditions:

- Manager has active and inactive guardians.

Steps:

1. Open `/staff/manager/guardians`.
2. Confirm page title `Guardians`.
3. Change status filter between Active, Inactive, and All.
4. Confirm table columns: Name, Email, Phone, Active, Deactivated, Reason, Linked children, Actions.
5. Use pagination if available.
6. Click `Children`.

Expected result:

- List reloads on filter changes.
- Loading, empty, and error states work.
- Active badge displays.
- Children link routes to `/staff/manager/children`.

### GUARD-002 - Create guardian

Priority: Critical

Preconditions:

- User is signed in as manager.

Steps:

1. Open `/staff/manager/guardians`.
2. Click `Add guardian`.
3. Confirm `Create guardian` form appears.
4. Enter Full name.
5. Optionally enter Email, Phone, and Notes.
6. Click `Create guardian`.

Expected result:

- Successful save closes the form and refreshes the list.
- Optional fields may remain blank.
- Validation errors appear near fields or in the form alert.

### GUARD-003 - Edit guardian and cancel

Priority: High

Preconditions:

- At least one guardian exists.

Steps:

1. Click `Edit` for a guardian row.
2. Confirm existing values prefill.
3. Change Phone or Notes.
4. Click `Save changes`.
5. Reopen edit.
6. Click `Cancel`.

Expected result:

- Updates are shown after reload.
- Cancel closes the form without saving new changes.

## Manager Invites

### INV-001 - Invite list and status filter

Priority: High

Preconditions:

- Invites exist in pending, accepted, revoked, and expired states.

Steps:

1. Open `/staff/manager/invites`.
2. Confirm page title `User invites`.
3. Change status filter to Pending, All statuses, Accepted, Revoked, and Expired.
4. Review table columns: Email, Role, Status, Expires, Accepted, Revoked, Created, Actions.

Expected result:

- Status filter reloads the list.
- Empty state says `No invitations found` when no rows match.
- Dates display in readable format or `-`.
- Only pending invites show action buttons.

### INV-002 - Send practitioner invite

Priority: Critical

Preconditions:

- Email address is not already invited.

Steps:

1. Open `/staff/manager/invites`.
2. Enter a new email address.
3. Select role `Practitioner`.
4. Click `Send invite`.

Expected result:

- Button shows spinner while submitting.
- Success toast appears.
- Email field clears.
- Role resets to Practitioner.
- Pending list refreshes and includes the new invite.

### INV-003 - Send parent invite

Priority: Critical

Preconditions:

- Email address is not already invited.

Steps:

1. Enter a new email address.
2. Select role `Parent`.
3. Click `Send invite`.

Expected result:

- Invite is created with Parent role.
- Manager role is not available in the role select.

### INV-004 - Invite validation

Priority: High

Preconditions:

- User is on User invites page.

Steps:

1. Submit invite with blank email.
2. Submit invite with malformed email.
3. Submit a duplicate invite if a duplicate fixture exists.

Expected result:

- Field-level email errors appear where applicable.
- Non-field errors appear in the alert.
- Form remains editable after error.

### INV-005 - Resend pending invite

Priority: High

Preconditions:

- Pending invite exists.

Steps:

1. Click `Resend` for a pending invite.
2. Observe row action area.

Expected result:

- Row shows a spinner while resend is pending.
- Success toast appears.
- List reloads.
- Row-level error appears if resend fails.

### INV-006 - Revoke pending invite

Priority: Critical

Preconditions:

- Pending invite exists.

Steps:

1. Click `Revoke`.
2. Confirm dialog title is `Revoke invitation`.
3. Read the confirmation message.
4. Click Cancel.
5. Reopen the dialog.
6. Click `Revoke`.

Expected result:

- Cancel closes the dialog without changing the invite.
- Confirm button shows loading while revoking.
- Success toast appears.
- Invite leaves pending list or shows revoked after list refresh.

## Practitioner Attendance

### ATT-001 - Attendance page initial state

Priority: Critical

Preconditions:

- User is signed in as practitioner or manager.

Steps:

1. Open `/staff/practitioner/attendance`.
2. Confirm page title `Today's attendance`.
3. Confirm Auto refresh toggle is on by default.
4. Confirm Refresh button, search input, and status filters are visible.
5. Wait for data load.

Expected result:

- Loading state appears on first load.
- Last updated time appears after data loads.
- Child cards display name, attendance badge, supporting text, and valid actions.

### ATT-002 - Search and filters

Priority: High

Preconditions:

- Attendance list contains at least one checked-in and one not-in child.

Steps:

1. Type part of a child name in `Search by name...`.
2. Clear the search.
3. Click `Not in`.
4. Click `Checked in`.
5. Click `All`.

Expected result:

- Search filters by child name only.
- Filter counts match visible child states.
- Empty filtered state appears if no child matches.

### ATT-003 - Check in child

Priority: Critical

Preconditions:

- Child is not checked in, not absent, and enrollment is complete.

Steps:

1. Locate the child card.
2. Click `Check in`.
3. Observe pending state.
4. Wait for list reload.

Expected result:

- Button text changes to `Checking in...`.
- Duplicate clicks do not trigger duplicate visible actions.
- After reload, child appears as checked in with check-in time.
- Action changes to `Check out`.

### ATT-004 - Check out child

Priority: Critical

Preconditions:

- Child is currently checked in.

Steps:

1. Click `Check out`.
2. Observe pending state.
3. Wait for list reload.

Expected result:

- Button text changes to `Checking out...`.
- Child no longer appears as checked in after successful reload.

### ATT-005 - Mark absent and clear absence

Priority: High

Preconditions:

- Child is not checked in, not absent, and enrollment complete.

Steps:

1. Click `Mark absent`.
2. Wait for reload.
3. Confirm child shows `Marked absent today`.
4. Click `Clear absence`.
5. Wait for reload.

Expected result:

- Mark absent shows `Marking...` while pending.
- Absent child has no Check in action.
- Clear absence shows `Clearing...` while pending.
- After clearing, Check in and Mark absent actions return.

### ATT-006 - Enrollment incomplete child

Priority: High

Preconditions:

- Child is not checked in and enrollment is incomplete.

Steps:

1. Locate incomplete child.
2. Observe supporting text and actions.

Expected result:

- Text shows `Enrollment incomplete`.
- Check in and Mark absent are disabled.
- No guardian or billing information is displayed.

### ATT-007 - Incomplete session warning

Priority: High

Preconditions:

- Child has incomplete session and is not checked in.

Steps:

1. Locate the child card.
2. Observe warning text.

Expected result:

- Card shows `Incomplete session needs manager correction`.
- Frontend does not expose manager-only correction form to practitioner.

### ATT-008 - Row action error

Priority: High

Preconditions:

- Fixture or mock can force one check-in/check-out/absence action to fail.

Steps:

1. Trigger the failing action.
2. Observe affected row after response.
3. Trigger Refresh.

Expected result:

- Error appears below the affected child card only.
- Other child actions remain usable.
- Error clears when the stale child disappears or action succeeds after reload.

### ATT-009 - Manual and automatic refresh

Priority: Medium

Preconditions:

- Attendance page is open.

Steps:

1. Click `Refresh`.
2. Turn Auto refresh off.
3. Wait longer than 30 seconds.
4. Turn Auto refresh on.
5. Wait for background refresh.

Expected result:

- Manual refresh reloads data.
- Auto refresh off stops polling.
- Auto refresh on resumes polling and updates the last updated time.
- Background refresh shows `Refreshing...` but does not block row actions.

## Manager Attendance Corrections

### CORR-001 - Child/date selector and session list

Priority: Critical

Preconditions:

- Manager account has attendance sessions for a child/date.

Steps:

1. Open `/staff/manager/attendance-corrections`.
2. Select a child.
3. Select a local date.
4. Review sessions list.
5. Select a session.

Expected result:

- Children are sorted active first, then by name.
- Loading state appears while sessions load.
- Sessions show status badge, check-in/check-out time, open state when no checkout exists, and duration when available.
- Selected session has visible selected styling.

### CORR-002 - Correct selected session

Priority: Critical

Preconditions:

- A completed or incomplete session exists.

Steps:

1. Select child and date.
2. Select a session.
3. Enter corrected check-in time.
4. Enter corrected check-out time.
5. Select reason `Incorrect time`.
6. Click `Submit correction`.

Expected result:

- Submit is disabled until child, date, times, and reason are present.
- Loading state appears on submit.
- Success alert says `Correction saved successfully.`
- Form fields clear.
- Sessions reload.
- Correction history updates for selected session.

### CORR-003 - Missed-session mode

Priority: Critical

Preconditions:

- Child/date has no sessions, or a session is selected and can switch modes.

Steps:

1. Select child and local date with no sessions.
2. Confirm empty state says `No sessions found`.
3. Enter check-in and check-out time.
4. Select `Missed check-in` or `Missed check-out`.
5. Submit correction.

Expected result:

- Form heading is `Create missed session`.
- Successful submit creates a visible session after reload.

### CORR-004 - Switch from selected session to missed-session mode

Priority: Medium

Preconditions:

- Child/date has at least one session.

Steps:

1. Select a session.
2. Click `Switch to missed-session mode`.

Expected result:

- Selected session clears.
- History panel clears.
- Form heading changes to `Create missed session`.

### CORR-005 - Reason and note validation

Priority: High

Preconditions:

- Correction form is visible.

Steps:

1. Enter valid times and leave reason blank.
2. Observe submit button and field error.
3. Select reason `Other`.
4. Leave Note blank.
5. Enter a note.

Expected result:

- Reason is required.
- `Other` requires Note.
- Field guidance appears near the affected field.
- Submit becomes enabled only when required fields are valid.

### CORR-006 - Time validation and error mapping

Priority: High

Preconditions:

- Correction form is visible.

Steps:

1. Enter check-out time earlier than check-in time.
2. Observe local validation.
3. Use fixtures to trigger overlap, outside enrollment window, future time, and session not found errors.

Expected result:

- Time-order hint says `Set check-out after check-in.`
- Overlap error guides changing check-in/check-out time.
- Enrollment window error displays child date range.
- Future time error says to use a time that has already happened.
- Session not found error asks user to select session again.

### CORR-007 - Issued invoice warning

Priority: High

Preconditions:

- Selected child/date has an issued invoice warning.

Steps:

1. Select the child/date.
2. Observe top alerts.

Expected result:

- Warning alert says an invoice exists for the period.
- Copy states the correction will not change the invoice.

### CORR-008 - Dashboard query params

Priority: Medium

Preconditions:

- Manager dashboard has a correction deep link.

Steps:

1. Open the correction route with `child_id`, `local_date`, and optional `session_id` query params.
2. Observe selected controls and loaded panels.

Expected result:

- Child and date are preselected.
- Sessions load automatically.
- History loads automatically when session id is present.

## Manager Funding Overview

### FUND-001 - Funding overview summary

Priority: High

Preconditions:

- Funding overview data has flagged and unflagged children.

Steps:

1. Open `/staff/manager/funding`.
2. Confirm page title `Funding Overview`.
3. Change Billing month.
4. Review summary counts.

Expected result:

- Loading state appears on month change.
- Summary shows Included, Flagged, Missing, Zero, Under 1h, and Above 160h.
- Error alert appears if overview fails to load.

### FUND-002 - Funding flags and review link

Priority: High

Preconditions:

- Overview contains children with each flag type.

Steps:

1. Review table rows.
2. Confirm flags: Missing allowance, Zero allowance, Under one hour, Above 160 hours.
3. Click `Review allowance`.

Expected result:

- Allowance displays as Not set, 0h 0m, minutes, hours, or hours/minutes as applicable.
- Last updated displays date or dash.
- Review link opens child detail with selected `billing_month` query param.

### FUND-003 - All clear state

Priority: Medium

Preconditions:

- Month has no flagged children.

Steps:

1. Select all-clear month.

Expected result:

- Empty state title is `All clear`.
- Message includes included child count and selected month.

## Manager Invoice Run

### INV-RUN-001 - Preflight summary and blockers

Priority: Critical

Preconditions:

- Manager invoice run data exists for a completed billing month.

Steps:

1. Open `/staff/manager/invoice-run`.
2. Confirm default billing month label.
3. Review progress indicator.
4. Review Preflight summary tiles.
5. Review Exceptions table if blockers exist.
6. Click each available Next action link in blockers.

Expected result:

- Summary shows children in month, ready for draft, exceptions, sessions included, attended time, funded deduction, and estimated total.
- Blocker rows show child, issue, detail, and next action.
- Next action links route to relevant child detail, corrections, or non-action label.

### INV-RUN-002 - Change billing month resets workflow

Priority: High

Preconditions:

- Invoice run page has generated drafts or selected rows.

Steps:

1. Generate drafts or select some draft rows.
2. Change Billing month.

Expected result:

- Workflow resets to preflight.
- Drafts, issue result, selected invoices, expanded details, and confirmation dialogs clear.
- New month preflight loads.

### INV-RUN-003 - Generate draft invoices

Priority: Critical

Preconditions:

- Preflight has eligible children.

Steps:

1. Click `Generate draft invoices`.
2. Observe loading state.
3. Review success alert.
4. Review Draft invoices table.

Expected result:

- Button shows spinner while generating.
- Success message includes generated, updated, and blocked counts.
- Draft list loads.
- Ready draft invoices are selected by default.
- Toast confirms generated count.

### INV-RUN-004 - No eligible children

Priority: High

Preconditions:

- Preflight has zero eligible children.

Steps:

1. Open invoice run for a blocked month.

Expected result:

- Generate button is not shown.
- Empty state says `No eligible children`.
- User can inspect blockers and next actions.

### INV-RUN-005 - Draft review, select all, and detail expansion

Priority: Critical

Preconditions:

- Drafts exist.

Steps:

1. Confirm draft columns: checkbox, Child, Status, Attended, Funded deduction, Extras, Net due, Actions.
2. Clear Select all.
3. Select one draft row.
4. Click `Detail`.
5. Click `Hide`.

Expected result:

- Select all selects or clears ready drafts only.
- Non-draft rows are not selectable.
- Issue selected count and total update with selected rows.
- Detail panel shows line description, quantity, unit, and amount.

### INV-RUN-006 - Bulk issue selected invoices

Priority: Critical

Preconditions:

- At least two ready drafts are selected.

Steps:

1. Click `Issue selected`.
2. Review confirmation dialog.
3. Click Cancel.
4. Reopen dialog.
5. Click `Issue invoices`.

Expected result:

- Dialog states invoice count, total, billing month, and immutable warning.
- Cancel closes without issuing.
- Confirm shows loading state.
- Result summary appears with issued count, total issued, billing month, issued rows, and skipped rows if any.
- Toast confirms issued count.

### INV-RUN-007 - Single issue fallback

Priority: High

Preconditions:

- At least one ready draft exists.

Steps:

1. Click `Issue` on one draft row.
2. Review confirmation dialog.
3. Confirm issue.

Expected result:

- Dialog identifies child and invoice total.
- Result summary updates.
- Draft list reloads and selections update.

### INV-RUN-008 - Invoice run error states

Priority: High

Preconditions:

- Fixture or mock can force preflight, generate, list drafts, and issue failures.

Steps:

1. Trigger each failure scenario.
2. Observe the page.

Expected result:

- Error alert appears with actionable user-facing message.
- Dialogs close only when appropriate.
- User can retry the action after error.

## Manager Invoices

### MGR-INV-001 - Invoice list filters

Priority: Critical

Preconditions:

- Invoices exist in draft, issued, payment failed, overdue, and paid statuses.

Steps:

1. Open `/staff/manager/invoices`.
2. Change Billing month.
3. Click each status filter: All, Draft, Issued, Payment failed, Overdue, Paid.
4. Review table.

Expected result:

- List reloads on month or status change.
- Table shows invoice identity, billing month, status, due status, payment, subtotal, funded deduction, net due, dates, and View action.
- Empty state shows selected month and status when no rows match.

### MGR-INV-002 - Invoice list pagination

Priority: Medium

Preconditions:

- More than 50 invoices exist for a filter, or use mock pagination.

Steps:

1. Open invoice list.
2. Click Next.
3. Confirm offset changes.
4. Click Previous.

Expected result:

- Previous disabled on first page.
- Next enabled only when current page is full.
- Offset text updates.

### MGR-INV-003 - Payment cue labels

Priority: High

Preconditions:

- List contains draft, unpaid, overdue, payment failed, paid, and no-payment-due examples.

Steps:

1. Review Payment column for each invoice state.

Expected result:

- Draft shows `Not issued`.
- Paid shows `Paid` and paid timestamp when available.
- Payment failed shows failure timestamp when available.
- Overdue unpaid shows unpaid/overdue cue.
- No payment action button appears on manager list.

### MGR-INV-004 - Manager invoice detail summary

Priority: Critical

Preconditions:

- Invoice detail exists.

Steps:

1. Open `/staff/manager/invoices`.
2. Click `View` on an invoice.
3. Confirm back link.
4. Review header, child name, billing month, status badges, summary, dates, calculation, and line items.

Expected result:

- Detail displays invoice identity or draft fallback.
- Back link returns to invoice list.
- Currency and minute formatting are readable.
- Line items show description, kind, quantity, unit amount, and line amount.

### MGR-INV-005 - Immutable issued invoice notice

Priority: High

Preconditions:

- Issued, overdue, paid, or payment failed invoice exists.

Steps:

1. Open detail for immutable invoice.

Expected result:

- Locked badge appears.
- Notice says issued invoice is locked and cannot be edited.
- No direct edit controls are visible.

### MGR-INV-006 - Draft invoice read-only review

Priority: High

Preconditions:

- Draft invoice exists.

Steps:

1. Open draft invoice detail.

Expected result:

- Page shows read-only review notice.
- Payment review section does not appear.
- No issue/edit action appears on detail; user is directed conceptually to Invoice run.

### MGR-INV-007 - Payment diagnostics

Priority: Critical

Preconditions:

- Issued-or-later invoice with payment diagnostics exists.

Steps:

1. Open invoice detail.
2. Review Payment review section.
3. Review Payment state, Amount paid, Balance due, Paid at, and Last updated where available.
4. Review Latest payment attempt.
5. Review Latest payment event.
6. Review Payment history.

Expected result:

- Loading payment information appears until diagnostics load.
- Attempt labels are human-readable.
- Provider IDs appear only as secondary detail.
- Event outcome, webhook status, transitions, and timestamps are readable.
- No parent checkout action is exposed to manager.

### MGR-INV-008 - Parent retry visibility and pending notice

Priority: High

Preconditions:

- Payment status examples exist for retry available, retry unavailable, and open checkout attempt.

Steps:

1. Open invoice detail with retry available.
2. Open invoice detail with retry unavailable reason.
3. Open invoice detail with open checkout attempt.

Expected result:

- Retry available notice says parent can retry payment.
- Retry unavailable shows human-readable reason.
- Open checkout attempt shows `Awaiting provider update`.

### MGR-INV-009 - Payment history pagination

Priority: Medium

Preconditions:

- More than 50 payment events exist for an invoice, or use mock pagination.

Steps:

1. Open invoice detail payment history.
2. Click Next.
3. Click Previous.

Expected result:

- Payment event list updates by page.
- Previous and Next enable/disable correctly.

## Parent Portal Invoices

### PARENT-001 - Parent invoices list

Priority: Critical

Preconditions:

- Parent is linked to one or more children with invoices.

Steps:

1. Sign in as parent.
2. Open `/parent/invoices`.
3. Confirm title `Invoices`.
4. Review Needs attention section.
5. Review child-grouped history sections.

Expected result:

- Parent sees only parent portal layout.
- Needs attention appears before history when overdue, payment failed, or payable issued invoices exist.
- History is grouped by child.
- No staff links or manager data appear.

### PARENT-002 - Empty parent invoices state

Priority: Medium

Preconditions:

- Parent has no issued-or-later invoices.

Steps:

1. Open `/parent/invoices`.

Expected result:

- Empty state title is `No invoices yet`.
- Message says issued invoices for linked children will appear here.

### PARENT-003 - Parent invoice pay from list

Priority: Critical

Preconditions:

- Parent has a payable issued, overdue, or payment failed invoice with balance due.

Steps:

1. Open parent invoices list.
2. In Needs attention, locate payable invoice.
3. Click `Pay now`.
4. Observe button state before redirect.

Expected result:

- Button changes to `Paying...`.
- Duplicate clicks are blocked.
- Browser redirects to hosted checkout URL.
- No card fields or custom payment form appear in Angular.

### PARENT-004 - Parent invoice list load more

Priority: Medium

Preconditions:

- Parent has more than 200 invoices, or use mock pagination.

Steps:

1. Open parent invoices list.
2. Click `Load more`.
3. Observe appended items.

Expected result:

- Button shows `Loading more...`.
- New invoices append without duplicate invoice IDs.
- Existing attention/history grouping remains coherent.

### PARENT-005 - Parent invoice detail summary

Priority: Critical

Preconditions:

- Parent invoice detail exists.

Steps:

1. Open `/parent/invoices`.
2. Click `View`.
3. Review back link, header, status badges, summary, billing period/dates, calculation, and line items.

Expected result:

- Detail is parent-safe and does not show manager diagnostics.
- Summary shows subtotal, funded deduction, total due, amount paid, and balance when relevant.
- Calculation and line item sections are readable.

### PARENT-006 - Parent payment from detail

Priority: Critical

Preconditions:

- Detail invoice is payable.

Steps:

1. Open parent invoice detail.
2. Confirm Balance due panel and `Pay now` button.
3. Click `Pay now`.

Expected result:

- Button changes to `Processing...`.
- Browser redirects to hosted checkout URL.
- No card-entry UI exists in Angular.

### PARENT-007 - Non-payable parent invoice

Priority: High

Preconditions:

- Paid invoice or zero-balance invoice exists.

Steps:

1. Open parent invoice detail.
2. Review action area.

Expected result:

- Pay action is hidden.
- Paid amount and balance display accurately.

### PARENT-008 - Payment return success

Priority: Critical

Preconditions:

- Parent returns from checkout with success query params.

Steps:

1. Open `/parent/invoices/:invoiceId?checkout=success&session_id=<id>`.
2. Observe alert.
3. Observe URL after page loads.
4. If invoice is not terminal yet, wait up to 20 seconds.

Expected result:

- Query params are removed using replace navigation.
- If paid, success alert says payment received.
- If still processing, info alert says payment is still processing and page shows `Checking payment status...`.
- Polling stops when status becomes paid or payment failed, or after the configured polling window.

### PARENT-009 - Payment return cancelled

Priority: High

Preconditions:

- Parent returns from checkout with cancelled/canceled query param.

Steps:

1. Open `/parent/invoices/:invoiceId?checkout=cancelled`.
2. Repeat with `checkout=canceled` if supported in environment.

Expected result:

- Warning alert says payment canceled and no payment was taken.
- Query params are removed.
- Pay now remains available if invoice is still payable.

### PARENT-010 - Payment return failed or processing

Priority: High

Preconditions:

- Invoice status can be payment failed or non-terminal after success return.

Steps:

1. Open parent detail after failed checkout/webhook result.
2. Open parent detail after success return while status is still non-terminal.

Expected result:

- Failed state shows error alert and allows retry if payable.
- Processing state shows info alert and polling copy.
- Error loading detail shows a top error alert.

## Responsive, Accessibility, And Visual QA

### QA-001 - Mobile practitioner attendance

Priority: High

Preconditions:

- Mobile viewport is active.

Steps:

1. Open practitioner attendance on 390 x 844.
2. Search, filter, check in, check out, mark absent, and clear absence.

Expected result:

- Child cards stack cleanly.
- Buttons are full-width or large enough for touch.
- Text does not overlap or clip.
- Required actions remain visible without horizontal scrolling.

### QA-002 - Desktop manager data tables

Priority: High

Preconditions:

- Desktop viewport is active.

Steps:

1. Open Children, Guardians, Invites, Funding, Invoice run, and Invoices.
2. Inspect table headers, row actions, filters, and pagination.

Expected result:

- Tables are readable and horizontally scroll only when necessary.
- Filters and primary actions remain visible.
- No TailAdmin/demo copy appears in production routes.

### QA-003 - Parent mobile payment flow

Priority: High

Preconditions:

- Mobile viewport and parent account with payable invoice.

Steps:

1. Open parent invoices list.
2. Open detail.
3. Start payment.
4. Return with cancelled and success query states.

Expected result:

- Attention cards and detail summary fit mobile width.
- Pay action remains visible and touch-friendly.
- Return alerts are readable and do not cover content.

### QA-004 - Keyboard navigation

Priority: High

Preconditions:

- Use keyboard only.

Steps:

1. Tab through sign-in form.
2. Tab through staff sidebar and header dropdown.
3. Tab through child/guardian forms.
4. Tab through invoice run confirmation dialogs.
5. Tab through parent Pay now flow until redirect.

Expected result:

- Focus order is logical.
- Focus indicator is visible.
- Buttons, links, inputs, selects, and custom controls are reachable.
- Dialog actions are keyboard-operable.

### QA-005 - Form labels and error messages

Priority: High

Preconditions:

- Forms are available in auth, child, guardian, invites, funding, corrections.

Steps:

1. Use screen reader or browser accessibility inspector.
2. Inspect labels for all visible inputs/selects.
3. Trigger validation errors.

Expected result:

- Inputs have visible labels or accessible labels.
- Errors are adjacent to fields or announced in alerts.
- Required inputs are visually clear.

### QA-006 - Loading, empty, and error states

Priority: High

Preconditions:

- Mock or QA environment can return loading delay, empty list, and error response.

Steps:

1. Trigger loading state for each routed module.
2. Trigger empty state where applicable.
3. Trigger generic error state.

Expected result:

- Loading indicators are visible and not mistaken for empty data.
- Empty states explain what happened and next action when available.
- Error states are user-facing, non-technical, and include request ID when the frontend presents one.

## Manager Registration/Enrolment Editor

### Test RE-01: Open registration from child detail

Steps:

- Log in as a manager.
- Navigate to a child detail page.
- Click the Registration/Enrolment link.
- If the profile or checklist shows as incomplete, a link to the registration editor is present.

Expected result:

- The registration editor opens at `/staff/manager/children/{childId}/registration`.
- Child name and date of birth are displayed.
- Registration profile and office-use checklist completion badges are visible.

### Test RE-02: Save demographics/home section

Steps:

- Open the registration editor for a child.
- Edit sex, religion, ethnic origin, first language, other languages, home address, postcode, telephone, disability status, disability notes, and access requirements in the Demographics and home section.
- Check the "Section reviewed" checkbox.
- Click "Save section".

Expected result:

- The section header shows a success message.
- Reload the page: the saved values persist.

### Test RE-03: Save medical/dietary section with explicit no answers

Steps:

- Set medical conditions, prescribed medication, and dietary requirements to "No".
- Set immunisation status to "Up to date".
- Add optional notes.
- Click "Save section".

Expected result:

- Success message appears.
- Save completes without validation errors.

### Test RE-04: Save health contacts

Steps:

- Enter doctor and health visitor names, addresses, and phone numbers.
- Click "Save section".

Expected result:

- Values persist after reload.

### Test RE-05: Save social/development section

Steps:

- Set social services involvement and all six development concerns to "No".
- Add optional notes or social worker contact.
- Click "Save section".

Expected result:

- Values persist after reload.

### Test RE-06: Add and save parent/carer, emergency contact, and authorised collector

Steps:

- In the Parent/carer section, click "+ Add parent/carer".
- Fill in name, relationship, telephone, and optionally check "Has parental responsibility".
- Click "Save parent/carers".
- Repeat for Emergency contacts and Authorised collectors sections.

Expected result:

- Each contact type saves independently.
- After reload, the saved contacts appear.

### Test RE-07: Remove a contact entry

Steps:

- Add a contact row with data.
- Click "Remove".
- Save the section.

Expected result:

- The removed contact does not appear after reload.

### Test RE-08: Set collection password

Steps:

- In the Collection password section, enter a new password in the password field.
- Click "Set password".

Expected result:

- The password input clears.
- The password value is not displayed anywhere on the page.
- "Password is set" status shows with last updated timestamp.

### Test RE-09: Save collection review flags

Steps:

- Check "Over-18 collection acknowledged" and "Emergency collection reviewed".
- Click "Save collection flags".

Expected result:

- Flags persist after reload.

### Test RE-10: Save funding support section

Steps:

- Set each funding question to "No".
- Add optional notes.
- Click "Save section".

Expected result:

- Values persist after reload.

### Test RE-11: Save routine care section

Steps:

- Enter free-text notes.
- Click "Save section".

Expected result:

- Values persist after reload.

### Test RE-12: Save GDPR declaration

Steps:

- Enter declaring person name and a declaration date.
- Click "Save GDPR declaration".

Expected result:

- Declaration timestamp appears under the form.
- Values persist after reload.

### Test RE-13: Save office-use checklist

Steps:

- Set deposit status, application date status, start date status, sessions/days requested status, and contract status to "Complete" with appropriate dates.
- Set term-time-only space to "No".
- Set Red Book, birth certificate/passport, and proof of address to "Complete" with checked dates.
- Click "Save checklist".

Expected result:

- Office-use checklist completion badge updates to "Complete".
- Values persist after reload.

### Test RE-14: Verify registration profile completion on child detail

Steps:

- After completing all registration sections, navigate back to the child detail page.
- Reload the page.

Expected result:

- The registration profile completion badge shows "Complete" status.

### Test RE-15: Verify office-use checklist completion on child detail

Steps:

- After completing the office-use checklist, navigate back to the child detail page.
- Reload the page.

Expected result:

- The office-use checklist completion badge shows "Complete" status.

### Test RE-16: Confirm no document upload control

Steps:

- Inspect the office-use checklist section.

Expected result:

- No file upload, document attachment, or "Upload" controls are present.
- Checklist items are status/date fields only.

### Test RE-17: API validation preserves draft values

Steps:

- Enter an invalid value in a field that triggers API validation (e.g., an invalid date format in a date field).
- Click "Save section".

Expected result:

- A section-level error message appears.
- The draft values the manager entered are not lost or reset.

### Test RE-18: Desktop and mobile layout

Steps:

- Open the registration editor on a desktop viewport (1440 x 900).
- Open the same page on a mobile viewport (390 x 844).

Expected result:

- All sections are readable.
- Buttons and controls are not clipped or overlapping.
- On mobile, fields stack in a single column.

## Final Regression Checklist

Before sign-off, confirm:

- Public signup remains invitation-only.
- Issued invoices have no direct edit controls.
- Manager UI does not expose parent-only checkout actions.
- Practitioner UI does not expose guardian contact details or billing data.
- Parent UI does not expose staff navigation, manager diagnostics, or unrelated children.
- All production routes use nursery domain terminology.
- Mobile and desktop layouts have no clipped text, overlapping buttons, or unreachable actions.
- Theme toggle works on public, staff, and parent layouts.
- Sign out clears protected access.
