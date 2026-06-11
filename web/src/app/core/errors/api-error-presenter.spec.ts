import { MappedApiError } from '../models/api-error.models';
import {
  ApiErrorContext,
  presentApiError,
  formatPresentedApiError,
  ApiErrorPresentation,
} from './api-error-presenter';

function mapped(overrides: Partial<MappedApiError> = {}): MappedApiError {
  return {
    code: 'internal_error',
    message: 'Something went wrong.',
    requestId: null,
    fieldErrors: {},
    ...overrides,
  };
}

describe('presentApiError', () => {
  describe('unknown / system codes', () => {
    it('shows request ID for internal_error', () => {
      const result = presentApiError(mapped({ code: 'internal_error', requestId: 'req-1' }), 'auth.signin');
      expect(result.showRequestId).toBe(true);
      expect(result.requestId).toBe('req-1');
      expect(result.kind).toBe('system');
    });

    it('shows request ID for db_unavailable', () => {
      const result = presentApiError(mapped({ code: 'db_unavailable', requestId: 'req-2' }), 'auth.signin');
      expect(result.showRequestId).toBe(true);
      expect(result.kind).toBe('system');
    });

    it('shows request ID for unrecognized code', () => {
      const result = presentApiError(mapped({ code: 'some_new_code', requestId: 'req-3' }), 'auth.signin');
      expect(result.showRequestId).toBe(true);
      expect(result.kind).toBe('unknown');
      expect(result.message).toBe('Something went wrong. Try again.');
    });

    it('hides backend message for unknown codes', () => {
      const result = presentApiError(
        mapped({ code: 'weird_error', message: 'Internal DB stack trace', requestId: 'req-4' }),
        'auth.signin',
      );
      expect(result.message).not.toContain('DB stack trace');
    });

    it('shows request ID for payment_provider_error', () => {
      const result = presentApiError(
        mapped({ code: 'payment_provider_error', requestId: 'req-5' }),
        'payment.parentCheckout',
      );
      expect(result.showRequestId).toBe(true);
      expect(result.requestId).toBe('req-5');
    });

    it('shows request ID for payment_provider_unconfigured', () => {
      const result = presentApiError(
        mapped({ code: 'payment_provider_unconfigured', requestId: 'req-6' }),
        'payment.parentCheckout',
      );
      expect(result.showRequestId).toBe(true);
    });

    it('parent provider error hides provider name', () => {
      const result = presentApiError(
        mapped({ code: 'payment_provider_error', message: 'Stripe failed', requestId: 'req-7' }),
        'payment.parentCheckout',
      );
      expect(result.message).toBe('Payment is unavailable right now. Try again later.');
      expect(result.message).not.toContain('Stripe');
    });

    it('manager provider error uses generic message', () => {
      const result = presentApiError(
        mapped({ code: 'payment_provider_error', requestId: 'req-8' }),
        'payment.managerDiagnostics',
      );
      expect(result.showRequestId).toBe(true);
      expect(result.kind).toBe('unknown');
    });
  });

  describe('known codes — no request ID', () => {
    it('unauthorized on signin shows credentials message', () => {
      const result = presentApiError(
        mapped({ code: 'unauthorized', requestId: 'req-10' }),
        'auth.signin',
      );
      expect(result.message).toBe('Check your email and password, then try again.');
      expect(result.showRequestId).toBe(false);
    });

    it('rate_limited shows wait message', () => {
      const result = presentApiError(mapped({ code: 'rate_limited' }), 'auth.forgotPassword');
      expect(result.message).toContain('Wait a moment');
      expect(result.showRequestId).toBe(false);
    });

    it('password reset token codes show unusable link', () => {
      for (const code of ['password_reset_token_invalid', 'password_reset_token_expired', 'password_reset_token_used']) {
        const result = presentApiError(mapped({ code }), 'auth.resetPassword');
        expect(result.message).toContain('no longer usable');
        expect(result.showRequestId).toBe(false);
      }
    });

    it('invite token codes show terminal states', () => {
      const result = presentApiError(mapped({ code: 'invite_token_expired' }), 'auth.inviteAccept');
      expect(result.message).toContain('no longer valid');
      expect(result.showRequestId).toBe(false);
    });

    it('invite_already_accepted shows terminal state', () => {
      const result = presentApiError(mapped({ code: 'invite_token_accepted' }), 'auth.inviteAccept');
      expect(result.message).toContain('already been accepted');
      expect(result.showRequestId).toBe(false);
    });

    it('manager invite_role_not_allowed sets field error', () => {
      const result = presentApiError(mapped({ code: 'invite_role_not_allowed' }), 'auth.managerInvites');
      expect(result.fieldErrors['role']).toBeDefined();
      expect(result.showRequestId).toBe(false);
    });

    it('manager invite_email_already_registered sets field error', () => {
      const result = presentApiError(mapped({ code: 'invite_email_already_registered' }), 'auth.managerInvites');
      expect(result.fieldErrors['email']).toBeDefined();
      expect(result.message).toContain('sign in');
      expect(result.showRequestId).toBe(false);
    });

    it('manager invite_scope_conflict shows pending invite message', () => {
      const result = presentApiError(mapped({ code: 'invite_scope_conflict' }), 'auth.managerInvites');
      expect(result.message).toContain('pending invite');
      expect(result.showRequestId).toBe(false);
    });

    it('invite_not_pending shows refresh action', () => {
      const result = presentApiError(mapped({ code: 'invite_not_pending' }), 'auth.managerInvites');
      expect(result.action?.command).toBe('refresh');
      expect(result.showRequestId).toBe(false);
    });
  });

  describe('attendance codes', () => {
    it('attendance_session_already_open shows check-in message', () => {
      const result = presentApiError(mapped({ code: 'attendance_session_already_open' }), 'attendance.rowAction');
      expect(result.message).toContain('already checked in');
      expect(result.showRequestId).toBe(false);
    });

    it('attendance_session_not_open shows no check-in message', () => {
      const result = presentApiError(mapped({ code: 'attendance_session_not_open' }), 'attendance.rowAction');
      expect(result.message).toContain('no open check-in');
      expect(result.showRequestId).toBe(false);
    });

    it('child_enrollment_incomplete shows manager message', () => {
      const result = presentApiError(mapped({ code: 'child_enrollment_incomplete' }), 'attendance.rowAction');
      expect(result.message).toContain('manager must complete enrollment');
      expect(result.showRequestId).toBe(false);
    });

    it('absence_attendance_exists shows already exists message', () => {
      const result = presentApiError(mapped({ code: 'absence_attendance_exists' }), 'attendance.rowAction');
      expect(result.message).toContain('already exists');
      expect(result.showRequestId).toBe(false);
    });

    it('absence_marker_exists shows already marked message', () => {
      const result = presentApiError(mapped({ code: 'absence_marker_exists' }), 'attendance.rowAction');
      expect(result.message).toContain('already marked absent');
      expect(result.showRequestId).toBe(false);
    });

    it('absence_marker_not_found shows cleared message', () => {
      const result = presentApiError(mapped({ code: 'absence_marker_not_found' }), 'attendance.rowAction');
      expect(result.message).toContain('already been cleared');
      expect(result.showRequestId).toBe(false);
    });

    it('attendance correction codes remain known without request ID', () => {
      const codes = [
        'attendance_session_overlap',
        'attendance_outside_enrollment_window',
        'attendance_correction_reason_required',
        'attendance_correction_reason_invalid',
        'reason_note_required_for_other',
        'attendance_invalid_time_order',
        'attendance_correction_future_time',
      ];
      for (const code of codes) {
        const result = presentApiError(mapped({ code }), 'attendance.correction');
        expect(result.showRequestId).toBe(false);
        expect(result.kind).toBe('known');
      }
    });
  });

  describe('people codes', () => {
    it('child_not_found shows return-to-list action', () => {
      const result = presentApiError(mapped({ code: 'child_not_found' }), 'people.child');
      expect(result.action?.route).toEqual(['/staff/manager/children']);
      expect(result.showRequestId).toBe(false);
    });

    it('guardian_not_found shows return-to-list action', () => {
      const result = presentApiError(mapped({ code: 'guardian_not_found' }), 'people.guardian');
      expect(result.action?.route).toEqual(['/staff/manager/guardians']);
      expect(result.showRequestId).toBe(false);
    });

    it('guardian_not_active shows reactivation guidance', () => {
      const result = presentApiError(mapped({ code: 'guardian_not_active' }), 'people.guardianLink');
      expect(result.message).toContain('Reactivate');
      expect(result.showRequestId).toBe(false);
    });

    it('parent_mapping_active_conflict shows end-current guidance', () => {
      const result = presentApiError(mapped({ code: 'parent_mapping_active_conflict' }), 'people.guardianLink');
      expect(result.message).toContain('End the current active mapping');
      expect(result.showRequestId).toBe(false);
    });

    it('funding_month_outside_enrollment_window shows enrollment message', () => {
      const result = presentApiError(
        mapped({ code: 'funding_month_outside_enrollment_window' }),
        'people.child',
      );
      expect(result.message).toContain('enrollment window');
      expect(result.showRequestId).toBe(false);
    });

    it('reason codes set field errors', () => {
      const codes = ['child_lifecycle_reason_required', 'guardian_deactivation_reason_required', 'relationship_reason_required'];
      for (const code of codes) {
        const result = presentApiError(mapped({ code }), 'people.child');
        expect(result.fieldErrors['reason_code']).toBeDefined();
        expect(result.showRequestId).toBe(false);
      }
    });
  });

  describe('invoice codes', () => {
    it('invoice_not_found for parent shows parent-safe message', () => {
      const result = presentApiError(mapped({ code: 'invoice_not_found' }), 'payment.parentDetail');
      expect(result.message).toContain('no longer available');
      expect(result.showRequestId).toBe(false);
      expect(result.action?.route).toEqual(['/parent/invoices']);
    });

    it('invoice_not_found for manager shows manager message', () => {
      const result = presentApiError(mapped({ code: 'invoice_not_found' }), 'invoice.managerDetail');
      expect(result.action?.route).toEqual(['/staff/manager/invoices']);
      expect(result.showRequestId).toBe(false);
    });

    it('invoice_not_payable for parent shows refresh action', () => {
      const result = presentApiError(mapped({ code: 'invoice_not_payable' }), 'payment.parentCheckout');
      expect(result.message).toContain('no longer payable');
      expect(result.showRequestId).toBe(false);
      expect(result.action?.command).toBe('refresh');
    });

    it('invoice_not_draft shows refresh guidance', () => {
      const result = presentApiError(mapped({ code: 'invoice_not_draft' }), 'invoice.run');
      expect(result.message).toContain('no longer a draft');
      expect(result.showRequestId).toBe(false);
    });

    it('incomplete_attendance shows corrections link', () => {
      const result = presentApiError(mapped({ code: 'incomplete_attendance' }), 'invoice.run');
      expect(result.action?.route).toEqual(['/staff/manager/attendance-corrections']);
      expect(result.showRequestId).toBe(false);
    });
  });

  describe('privacy — parent and practitioner', () => {
    it('parent forbidden does not reveal entity existence', () => {
      const result = presentApiError(mapped({ code: 'forbidden_role' }), 'payment.parentDetail');
      expect(result.message).not.toContain('child');
      expect(result.message).not.toContain('guardian');
      expect(result.message).not.toContain('does not exist');
    });

    it('parent not_found uses safe copy', () => {
      const result = presentApiError(mapped({ code: 'not_found' }), 'payment.parentDetail');
      expect(result.message).toContain('no longer available');
      expect(result.message).not.toContain('does not exist');
    });

    it('practitioner forbidden does not reveal entity existence', () => {
      const result = presentApiError(mapped({ code: 'forbidden_scope' }), 'attendance.rowAction');
      expect(result.message).not.toContain('child');
      expect(result.message).not.toContain('guardian');
    });
  });

  describe('validation_error with field details', () => {
    it('passes field errors through', () => {
      const result = presentApiError(
        mapped({ code: 'validation_error', message: 'Invalid email', fieldErrors: { email: 'Invalid email' } }),
        'auth.signin',
      );
      expect(result.fieldErrors['email']).toBe('Invalid email');
      expect(result.showRequestId).toBe(false);
    });
  });
});

describe('formatPresentedApiError', () => {
  it('appends request ID when showRequestId is true', () => {
    const p: ApiErrorPresentation = {
      message: 'Something went wrong. Try again.',
      fieldErrors: {},
      showRequestId: true,
      requestId: 'req-abc',
      kind: 'system',
    };
    expect(formatPresentedApiError(p)).toBe('Something went wrong. Try again. (Request: req-abc)');
  });

  it('omits request ID when showRequestId is false', () => {
    const p: ApiErrorPresentation = {
      message: 'Check your credentials.',
      fieldErrors: {},
      showRequestId: false,
      requestId: 'req-xyz',
      kind: 'known',
    };
    expect(formatPresentedApiError(p)).toBe('Check your credentials.');
  });

  it('omits request ID when requestId is null', () => {
    const p: ApiErrorPresentation = {
      message: 'Something went wrong. Try again.',
      fieldErrors: {},
      showRequestId: true,
      requestId: null,
      kind: 'unknown',
    };
    expect(formatPresentedApiError(p)).toBe('Something went wrong. Try again.');
  });
});

describe('presentApiError — owner contexts', () => {
  it('site_not_found shows owner-safe message with overview action', () => {
    const result = presentApiError(mapped({ code: 'site_not_found' }), 'owner.managerAccess');
    expect(result.kind).toBe('known');
    expect(result.message).toContain('Site not found');
    expect(result.action?.route).toEqual(['/owner']);
  });

  it('manager_membership_not_found shows refresh action', () => {
    const result = presentApiError(mapped({ code: 'manager_membership_not_found' }), 'owner.managerAccess');
    expect(result.kind).toBe('known');
    expect(result.message).toContain('not found');
    expect(result.action?.command).toBe('refresh');
  });

  it('user_not_found shows user not found message', () => {
    const result = presentApiError(mapped({ code: 'user_not_found' }), 'owner.managerAccess');
    expect(result.kind).toBe('known');
    expect(result.message).toContain('User not found');
  });

  it('user_inactive shows inactive message', () => {
    const result = presentApiError(mapped({ code: 'user_inactive' }), 'owner.managerAccess');
    expect(result.kind).toBe('known');
    expect(result.message).toContain('inactive');
  });
});
