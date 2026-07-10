import { HttpClient, HttpRequest, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { diagnosticsInterceptor } from './diagnostics.interceptor';

describe('diagnosticsInterceptor', () => {
  let httpTesting: HttpTestingController;
  let http: HttpClient;

  function configure(enabled: boolean): void {
    (environment as { enableApiDiagnostics: boolean }).enableApiDiagnostics = enabled;

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([diagnosticsInterceptor])),
        provideHttpClientTesting(),
      ],
    });

    httpTesting = TestBed.inject(HttpTestingController);
    http = TestBed.inject(HttpClient);
  }

  afterEach(() => {
    httpTesting.verify();
  });

  describe('when enabled', () => {
    beforeEach(() => configure(true));

    it('adds X-Request-ID header when absent', () => {
      http.get('/api/test').subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.has('X-Request-ID')).toBeTrue();
      expect(req.request.headers.get('X-Request-ID')!.length).toBe(32);
      req.flush({});
    });

    it('adds X-Correlation-ID header when absent', () => {
      http.get('/api/test').subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.has('X-Correlation-ID')).toBeTrue();
      expect(req.request.headers.get('X-Correlation-ID')!.length).toBeGreaterThan(0);
      req.flush({});
    });

    it('preserves existing X-Request-ID when supplied', () => {
      http.get('/api/test', { headers: { 'X-Request-ID': 'client-req-id' } }).subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.get('X-Request-ID')).toBe('client-req-id');
      req.flush({});
    });

    it('preserves existing X-Correlation-ID when supplied', () => {
      http.get('/api/test', { headers: { 'X-Correlation-ID': 'client-corr-id' } }).subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.get('X-Correlation-ID')).toBe('client-corr-id');
      req.flush({});
    });

    it('logs redacted metadata for failed API responses', () => {
      const spy = spyOn(console, 'debug').and.stub();

      http.get('/api/test').subscribe({
        error: () => void 0,
      });

      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      req.flush(
        { code: 'internal_error', message: 'Something went wrong.', request_id: 'req-abc' },
        { status: 500, statusText: 'Internal Server Error' },
      );

      expect(spy).toHaveBeenCalledTimes(1);
      const args = spy.calls.mostRecent().args;
      expect(args[0]).toBe('API_DIAGNOSTICS');
      const info = args[1] as Record<string, unknown>;
      expect(info['status']).toBe(500);
      expect(info['requestId']).toBe('req-abc');
      expect(info['correlationId']).toBeDefined();
      expect(Object.keys(info)).not.toContain('body');
    });

    it('does not log request or response body', () => {
      const spy = spyOn(console, 'debug').and.stub();

      http.post('/api/test', { sensitive: 'data' }).subscribe({
        error: () => void 0,
      });

      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      req.flush(
        { code: 'error', children: [{ name: 'child-name' }] },
        { status: 500, statusText: 'Internal Server Error' },
      );

      const info = spy.calls.mostRecent().args[1] as Record<string, unknown>;
      expect(Object.keys(info)).not.toContain('body');
      expect(Object.keys(info)).not.toContain('children');
      expect(Object.keys(info)).not.toContain('sensitive');
    });

    it('adds traceparent header when absent', () => {
      http.get('/api/test').subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.has('traceparent')).toBeTrue();
      expect(req.request.headers.get('traceparent')!).toMatch(/^00-[0-9a-f]{32}-[0-9a-f]{16}-01$/);
      req.flush({});
    });
  });

  describe('when disabled', () => {
    beforeEach(() => configure(false));

    it('does not add diagnostic headers', () => {
      http.get('/api/test').subscribe();
      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      expect(req.request.headers.has('X-Request-ID')).toBeFalse();
      expect(req.request.headers.has('X-Correlation-ID')).toBeFalse();
      expect(req.request.headers.has('traceparent')).toBeFalse();
      req.flush({});
    });

    it('does not log failed responses', () => {
      const spy = spyOn(console, 'debug').and.stub();

      http.get('/api/test').subscribe({
        error: () => void 0,
      });

      const req = httpTesting.expectOne((r: HttpRequest<unknown>) => r.url === '/api/test');
      req.flush(
        { code: 'error' },
        { status: 500, statusText: 'Internal Server Error' },
      );

      expect(spy).not.toHaveBeenCalled();
    });
  });
});
