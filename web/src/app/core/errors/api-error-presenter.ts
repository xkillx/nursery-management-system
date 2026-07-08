import { MappedApiError } from '../models/api-error.models';

export type ApiErrorContext =
  | 'auth.signin'
  | 'auth.forgotPassword'
  | 'auth.resetPassword'
  | 'auth.inviteAccept'
  | 'auth.managerInvites'
  | 'attendance.list'
  | 'attendance.rowAction'
  | 'attendance.correction'
  | 'people.child'
  | 'invoice.run'
  | 'invoice.managerList'
  | 'invoice.managerDetail'
  | 'invoice.prefill'
  | 'invoice.createDraft'
  | 'invoice.issue'
  | 'payment.managerDiagnostics'
  | 'payment.parentList'
  | 'payment.parentDetail'
  | 'payment.parentCheckout'
  | 'owner.siteSummaries'
  | 'owner.managerAccess'
  | 'sessionTemplates.list'
  | 'sessionTemplates.create'
  | 'sessionTemplates.update'
  | 'sessionTemplates.archive'
  | 'sessionTemplates.reactivate'
  | 'registration.intake'
  | 'registration.profile'
  | 'registration.consent'
  | 'registration.completion';

export interface ApiErrorAction {
  label: string;
  route?: string[];
  queryParams?: Record<string, string>;
  command?: 'retry' | 'refresh' | 'return_to_invoices' | 'sign_in';
}

export interface ApiErrorPresentation {
  message: string;
  title?: string;
  fieldErrors: Record<string, string>;
  showRequestId: boolean;
  requestId: string | null;
  action?: ApiErrorAction;
  kind: 'known' | 'unknown' | 'system';
}

export interface PresentOptions {
  childId?: string;
  billingMonth?: string;
  invoiceId?: string;
}

const GENERIC_MESSAGE = 'Something went wrong. Try again.';

const CODES_WITHOUT_REQUEST_ID: ReadonlySet<string> = new Set([
  'unauthorized',
  'validation_error',
  'rate_limited',
  'membership_selection_required',
  'forbidden_role',
  'forbidden_scope',
  'forbidden_scope_selection',

  'password_reset_token_invalid',
  'password_reset_token_expired',
  'password_reset_token_used',

  'invite_token_invalid',
  'invite_token_expired',
  'invite_token_accepted',
  'invite_token_revoked',
  'invite_role_not_allowed',
  'invite_email_already_registered',
  'invite_scope_conflict',
  'invite_not_pending',
  'invite_already_accepted',
  'invite_not_found',

  'not_found',
  'child_not_found',
  'parent_child_mapping_not_found',
  'membership_not_found',
  'invoice_not_found',
  'funding_profile_not_found',
  'attendance_session_not_found',

  'attendance_session_already_open',
  'attendance_session_not_open',
  'child_enrollment_incomplete',
  'attendance_invalid_time_order',
  'attendance_correction_future_time',
  'attendance_session_overlap',
  'attendance_outside_enrollment_window',
  'attendance_correction_reason_required',
  'attendance_correction_reason_invalid',
  'reason_note_required_for_other',
  'absence_attendance_exists',
  'absence_marker_exists',
  'absence_marker_not_found',

  'membership_not_parent',
  'membership_not_active',
  'child_lifecycle_reason_required',
  'relationship_reason_required',
  'lifecycle_reason_invalid',

  'invoice_not_draft',
  'invoice_not_monthly',
  'invoice_not_in_billing_month',
  'invoice_not_payable',
  'invoice_already_issued',
  'incomplete_attendance',
  'missing_funding_profile',
  'missing_parent_carer_contact',
  'missing_billing_rate',
  'missing_child_name',
  'missing_child_date_of_birth',
  'missing_child_start_date',
  'funding_month_outside_enrollment_window',

  'site_not_found',
  'site_rate_missing',
  'manager_membership_not_found',
  'user_not_found',
  'user_inactive',
]);

function shouldShowRequestId(code: string): boolean {
  return !CODES_WITHOUT_REQUEST_ID.has(code);
}

function isKnown(code: string): boolean {
  return CODES_WITHOUT_REQUEST_ID.has(code);
}

function parentSafeNotFound(context: ApiErrorContext): string {
  if (context.startsWith('payment.parent') || context.startsWith('invoice.parent')) {
    return 'This is no longer available. Return to your invoices and try again.';
  }
  if (context.startsWith('attendance.') || context === 'attendance.list') {
    return 'This is no longer available in your current session. Try refreshing or sign in again.';
  }
  return 'This is no longer available. Try refreshing the page.';
}

function presentKnownError(
  mapped: MappedApiError,
  context: ApiErrorContext,
  _options: PresentOptions,
): ApiErrorPresentation {
  const base: ApiErrorPresentation = {
    message: mapped.message,
    fieldErrors: { ...mapped.fieldErrors },
    showRequestId: false,
    requestId: mapped.requestId,
    kind: 'known',
  };

  switch (mapped.code) {
    // Validation — show field errors inline, generic top message for auth
    case 'validation_error':
      if (context === 'auth.signin') {
        base.message = 'Check your email and password, then try again.';
      } else {
        base.message = mapped.message;
      }
      break;

    // Membership selection — component handles picker UI
    case 'membership_selection_required':
      break;

    // Auth
    case 'unauthorized':
      if (context === 'auth.signin') {
        base.message = 'Check your email and password, then try again.';
      }
      break;

    case 'rate_limited':
      base.message = 'Too many attempts. Wait a moment and try again.';
      break;

    // Password reset tokens
    case 'password_reset_token_invalid':
    case 'password_reset_token_expired':
    case 'password_reset_token_used':
      base.message = 'This password reset link is no longer usable. Request a new one.';
      break;

    // Invite tokens
    case 'invite_token_invalid':
    case 'invite_token_expired':
      base.message = 'This invitation link is no longer valid.';
      break;
    case 'invite_token_accepted':
      base.message = 'This invitation has already been accepted.';
      break;
    case 'invite_token_revoked':
      base.message = 'This invitation has been revoked.';
      break;

    // Manager invite codes
    case 'invite_role_not_allowed':
      base.message = 'This role is not available for invitation.';
      base.fieldErrors['role'] = base.message;
      break;
    case 'invite_email_already_registered':
      base.message = 'This email already has access. They can sign in or reset their password.';
      base.fieldErrors['email'] = base.message;
      break;
    case 'invite_scope_conflict':
      base.message = 'A pending invite already exists for this email with a different role.';
      break;
    case 'invite_not_pending':
    case 'invite_already_accepted':
      base.message = 'This invite has changed. Refresh the list to see the current status.';
      base.action = { label: 'Refresh list', command: 'refresh' };
      break;
    case 'invite_not_found':
      base.message = 'This invite no longer exists. Refresh the list.';
      base.action = { label: 'Refresh list', command: 'refresh' };
      break;

    // Authorization / privacy
    case 'forbidden_role':
    case 'forbidden_scope':
    case 'forbidden_scope_selection':
    case 'forbidden_role_unknown':
      if (context.startsWith('payment.parent') || context.startsWith('invoice.manager')) {
        base.message = parentSafeNotFound(context);
      } else if (context.startsWith('attendance.')) {
        base.message = 'You do not have access to this in your current session. Sign in with the right account or switch session.';
      } else {
        base.message = 'You do not have access to this. Try refreshing or sign in again.';
      }
      break;

    case 'not_found':
      base.message = parentSafeNotFound(context);
      break;

    // Attendance
    case 'attendance_session_already_open':
      base.message = 'This child is already checked in. The list will refresh.';
      break;
    case 'attendance_session_not_open':
      base.message = 'There is no open check-in to check out. The list will refresh.';
      break;
    case 'child_enrollment_incomplete':
      base.message = 'This child is not ready for attendance. A manager must complete enrollment.';
      break;
    case 'absence_attendance_exists':
      base.message = 'Attendance already exists for today. The list will refresh.';
      break;
    case 'absence_marker_exists':
      base.message = 'This child is already marked absent today. The list will refresh.';
      break;
    case 'absence_marker_not_found':
      base.message = 'The absence has already been cleared. The list will refresh.';
      break;

    // Attendance corrections — messages set by component; presenter only sets policy
    case 'attendance_session_overlap':
    case 'attendance_outside_enrollment_window':
    case 'attendance_correction_reason_required':
    case 'attendance_correction_reason_invalid':
    case 'attendance_invalid_time_order':
    case 'attendance_correction_future_time':
    case 'attendance_session_not_found':
      break;

    // Child
    case 'child_not_found':
      base.message = 'This child could not be found. Return to the children list.';
      base.action = { label: 'View children', route: ['/manager/children'] };
      break;
    case 'parent_child_mapping_not_found':
      base.message = 'This mapping no longer exists. Refresh the page.';
      break;
    case 'parent_mapping_active_conflict':
      base.message = 'End the current active mapping before creating a new one.';
      break;
    case 'membership_not_parent':
      base.message = 'Select an active parent membership.';
      break;
    case 'membership_not_active':
      base.message = 'Select an active membership.';
      break;
    case 'membership_not_found':
      base.message = 'The selected membership could not be found. Refresh and try again.';
      break;
    case 'funding_month_outside_enrollment_window':
      base.message = 'Choose a billing month within the child\'s enrollment window.';
      break;
    case 'child_lifecycle_reason_required':
    case 'relationship_reason_required':
      base.message = 'A reason is required.';
      base.fieldErrors['reason_code'] = base.message;
      break;
    case 'lifecycle_reason_invalid':
      base.message = 'Select a valid reason.';
      base.fieldErrors['reason_code'] = base.message;
      break;
    case 'reason_note_required_for_other':
      base.message = 'Provide a note when selecting "Other".';
      base.fieldErrors['reason_note'] = base.message;
      break;

    // Invoice
    case 'invoice_not_found':
      if (context.startsWith('payment.parent') || context.startsWith('invoice.')) {
        base.message = 'This invoice is no longer available. Return to your invoices.';
        base.action = { label: 'View invoices', route: [context.startsWith('payment.parent') ? '/parent/invoices' : '/manager/invoices'] };
      } else {
        base.message = 'This invoice could not be found. Return to the invoice list.';
        base.action = { label: 'View invoices', route: ['/manager/invoices'] };
      }
      break;
    case 'invoice_not_draft':
      base.message = 'This invoice is no longer a draft. Refresh the invoice run.';
      break;
    case 'invoice_not_monthly':
      base.message = 'This invoice is not a monthly invoice.';
      break;
    case 'invoice_not_in_billing_month':
      base.message = 'This invoice does not belong to the selected billing month.';
      break;
    case 'invoice_not_payable':
      if (context.startsWith('payment.parent')) {
        base.message = 'This invoice is no longer payable. Refresh to see the current status.';
        base.action = { label: 'Refresh', command: 'refresh' };
      } else {
        base.message = 'This invoice is not in a payable state.';
      }
      break;

    // Invoice run blockers
    case 'incomplete_attendance':
      base.message = 'Attendance data is incomplete. Correct attendance before generating invoices.';
      base.action = { label: 'Go to attendance corrections', route: ['/manager/attendance-corrections'] };
      break;
    case 'missing_funding_profile':
      base.message = 'Funding profile is missing for one or more children.';
      break;
    case 'missing_parent_carer_contact':
      base.message = 'A parent carer contact is missing for one or more children.';
      break;
    case 'invoice_already_issued':
      base.message = 'An invoice has already been issued for this billing period.';
      break;
    case 'missing_billing_rate':
    case 'missing_child_name':
    case 'missing_child_date_of_birth':
    case 'missing_child_start_date':
      break;

    // Funding
    case 'funding_profile_not_found':
      base.message = 'Funding profile not found. Refresh the page.';
      break;

    // Billing setup
    case 'site_rate_missing':
      base.message = 'A site hourly rate must be configured before enrolling a child. Set it in Billing Setup.';
      base.action = { label: 'Go to Billing Setup', route: ['/manager/billing-setup'] };
      break;

    // Owner
    case 'site_not_found':
      base.message = 'Site not found or no longer active. Return to the overview.';
      if (context.startsWith('owner.')) {
        base.action = { label: 'View overview', route: ['/owner'] };
      }
      break;
    case 'manager_membership_not_found':
      base.message = 'Manager membership not found. The list has been refreshed.';
      base.action = { label: 'Refresh', command: 'refresh' };
      break;
    case 'user_not_found':
      base.message = 'User not found.';
      break;
    case 'user_inactive':
      base.message = 'This user account is inactive.';
      break;

    default:
      // Unknown code — fall through to unknown handling
      return presentUnknownError(mapped, context);
  }

  return base;
}

function presentUnknownError(
  mapped: MappedApiError,
  _context: ApiErrorContext,
): ApiErrorPresentation {
  return {
    message: GENERIC_MESSAGE,
    fieldErrors: {},
    showRequestId: true,
    requestId: mapped.requestId,
    kind: mapped.code === 'internal_error' || mapped.code === 'db_unavailable' ? 'system' : 'unknown',
  };
}

export function presentApiError(
  mapped: MappedApiError,
  context: ApiErrorContext,
  options?: PresentOptions,
): ApiErrorPresentation {
  const opts = options ?? {};

  // Provider/service codes always get request ID + generic message
  if (
    mapped.code === 'internal_error' ||
    mapped.code === 'db_unavailable' ||
    mapped.code === 'payment_provider_error' ||
    mapped.code === 'payment_provider_unconfigured'
  ) {
    if (mapped.code === 'payment_provider_unconfigured' || mapped.code === 'payment_provider_error') {
      if (context.startsWith('payment.parent')) {
        return {
          message: 'Payment is unavailable right now. Try again later.',
          fieldErrors: {},
          showRequestId: true,
          requestId: mapped.requestId,
          kind: 'system',
        };
      }
    }
    return presentUnknownError(mapped, context);
  }

  if (isKnown(mapped.code)) {
    return presentKnownError(mapped, context, opts);
  }

  return presentUnknownError(mapped, context);
}

export function formatPresentedApiError(presentation: ApiErrorPresentation): string {
  if (presentation.showRequestId && presentation.requestId) {
    return `${presentation.message} (Request: ${presentation.requestId})`;
  }
  return presentation.message;
}
