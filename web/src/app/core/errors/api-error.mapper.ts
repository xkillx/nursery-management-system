import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Router } from '@angular/router';

import { ApiErrorBody, MappedApiError } from '../models/api-error.models';
import { AuthService } from '../services/auth.service';

const GO_VALIDATION_RE = /Key:\s+'[\w.]+\.(\w+)'\s+Error:Field validation for '\w+' failed on the '(\w+)'/;

const VALIDATION_MESSAGES: Record<string, string> = {
  min: 'Too short.',
  max: 'Too long.',
  required: 'This field is required.',
  email: 'Enter a valid email address.',
};

const FIELD_NAME_MAP: Record<string, string> = {
  Password: 'password',
  Email: 'email',
};

function parseGoValidationDetail(details: string, mapped: MappedApiError): void {
  const match = details.match(GO_VALIDATION_RE);
  if (!match) return;

  const goField = match[1];
  const rule = match[2];
  const field = FIELD_NAME_MAP[goField] || goField.toLowerCase();

  mapped.fieldErrors[field] = VALIDATION_MESSAGES[rule] || mapped.message;
}

const defaultError: MappedApiError = {
  code: 'internal_error',
  message: 'Something went wrong. Please try again.',
  requestId: null,
  fieldErrors: {},
};

@Injectable({ providedIn: 'root' })
export class ApiErrorMapper {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  map(error: unknown): MappedApiError {
    if (!(error instanceof HttpErrorResponse)) {
      return defaultError;
    }

    const payload = error.error as ApiErrorBody | undefined;
    if (!payload || typeof payload !== 'object') {
      return {
        ...defaultError,
        message: error.message || defaultError.message,
      };
    }

    const mapped: MappedApiError = {
      code: payload.code || defaultError.code,
      message: payload.message || defaultError.message,
      requestId: payload.request_id ?? null,
      fieldErrors: {},
    };

    if (mapped.code === 'validation_error' && typeof payload.details === 'string') {
      parseGoValidationDetail(payload.details, mapped);
    } else if (
      mapped.code === 'validation_error' &&
      payload.details &&
      typeof payload.details === 'object' &&
      !Array.isArray(payload.details)
    ) {
      const field = (payload.details as Record<string, unknown>)['field'];
      if (typeof field === 'string' && field.trim() !== '') {
        mapped.fieldErrors[field] = mapped.message;
      }
    }

    return mapped;
  }

  mapAndHandle(error: unknown): MappedApiError {
    const mapped = this.map(error);

    if (mapped.code === 'unauthorized') {
      this.authService.clearSession();
      this.router.navigate(['/signin']);
    }

    return mapped;
  }
}
