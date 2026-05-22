import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Router } from '@angular/router';

import { ApiErrorBody, MappedApiError } from '../models/api-error.models';
import { AuthService } from '../services/auth.service';

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

    if (
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
