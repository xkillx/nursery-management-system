import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { tap } from 'rxjs';
import { environment } from '../../../environments/environment';

let correlationId: string | null = null;

function getOrCreateCorrelationId(): string {
  if (!correlationId) {
    correlationId = crypto.randomUUID();
  }
  return correlationId;
}

function generateId(bytes: number): string {
  const buf = new Uint8Array(bytes);
  crypto.getRandomValues(buf);
  return Array.from(buf)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

export const diagnosticsInterceptor: HttpInterceptorFn = (req, next) => {
  if (!environment.enableApiDiagnostics) {
    return next(req);
  }

  const requestId = req.headers.has('X-Request-ID')
    ? req.headers.get('X-Request-ID')!
    : generateId(16);

  const enrichedReq = req.clone({
    setHeaders: {
      'X-Request-ID': requestId,
      'X-Correlation-ID': req.headers.has('X-Correlation-ID')
        ? req.headers.get('X-Correlation-ID')!
        : getOrCreateCorrelationId(),
      traceparent: req.headers.has('traceparent')
        ? req.headers.get('traceparent')!
        : `00-${generateId(16)}-${generateId(8)}-01`,
    },
  });

  return next(enrichedReq).pipe(
    tap({
      error: (error: unknown) => {
        if (error instanceof HttpErrorResponse) {
          logFailedDiagnostic(error, requestId);
        }
      },
    }),
  );
};

function logFailedDiagnostic(resp: HttpErrorResponse, requestId: string): void {
  const body = resp.error as Record<string, unknown> | undefined;

  const info: Record<string, unknown> = {
    url: sanitizePath(resp.url ?? ''),
    status: resp.status,
  };

  if (body && typeof body === 'object' && !Array.isArray(body)) {
    if (typeof body['code'] === 'string') {
      info['errorCode'] = body['code'];
    }
    info['requestId'] =
      typeof body['request_id'] === 'string' ? body['request_id'] : requestId;
  } else {
    info['requestId'] = requestId;
  }

  info['correlationId'] = correlationId;

  const traceparent = resp.headers?.get('traceparent') ?? '';
  if (traceparent) {
    const parts = traceparent.split('-');
    if (parts.length >= 2) {
      info['traceId'] = parts[1];
    }
  }

  console.debug('API_DIAGNOSTICS', info);
}

function sanitizePath(url: string): string {
  try {
    const u = new URL(url);
    return u.pathname;
  } catch {
    const qIndex = url.indexOf('?');
    return qIndex >= 0 ? url.slice(0, qIndex) : url;
  }
}
