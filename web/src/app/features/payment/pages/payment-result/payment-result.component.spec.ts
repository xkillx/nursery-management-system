import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { Meta } from '@angular/platform-browser';

import { PaymentResultComponent } from './payment-result.component';

function createTestBed(queryParams: Record<string, string | null>) {
  TestBed.configureTestingModule({
    imports: [PaymentResultComponent],
    providers: [
      provideRouter([]),
      {
        provide: ActivatedRoute,
        useValue: {
          snapshot: { queryParamMap: { get: (key: string) => queryParams[key] ?? null } },
        },
      },
    ],
  });
  return TestBed.createComponent(PaymentResultComponent);
}

describe('PaymentResultComponent', () => {
  it('renders a provisional success state for outcome=success', () => {
    const fixture = createTestBed({ outcome: 'success', invoice_id: 'inv-1', session_id: 'cs_1' });
    const html = fixture.nativeElement as HTMLElement;
    fixture.detectChanges();

    expect(html.textContent).toContain('Payment received');
    // Provisional wording: does not assert final settlement.
    expect(html.textContent).toContain('We are confirming it with your nursery');
  });

  it('renders the cancelled state for outcome=cancelled', () => {
    const fixture = createTestBed({ outcome: 'cancelled', invoice_id: 'inv-1' });
    const html = fixture.nativeElement as HTMLElement;
    fixture.detectChanges();

    expect(html.textContent).toContain('Payment not completed');
    expect(html.textContent).toContain('No payment was taken');
  });

  it('renders the cancelled state for the Stripe canceled spelling', () => {
    const fixture = createTestBed({ outcome: 'canceled', invoice_id: 'inv-1' });
    const html = fixture.nativeElement as HTMLElement;
    fixture.detectChanges();

    expect(html.textContent).toContain('Payment not completed');
  });

  it('renders a neutral fallback for missing or unknown params without errors', () => {
    const fixture = createTestBed({});
    const html = fixture.nativeElement as HTMLElement;
    fixture.detectChanges();

    expect(html.textContent).toContain('Payment status unavailable');
  });

  it('sets noindex and no-referrer meta tags', () => {
    const fixture = createTestBed({ outcome: 'success', invoice_id: 'inv-1' });
    fixture.detectChanges();

    const meta = TestBed.inject(Meta);
    const robots = meta.getTag('name="robots"');
    const referrer = meta.getTag('http-equiv="Referrer-Policy"');
    expect(robots?.content).toContain('noindex');
    expect(referrer?.content).toBe('no-referrer');
  });

  it('links sign-in with the invoice id for success', () => {
    const fixture = createTestBed({ outcome: 'success', invoice_id: 'inv-9' });
    fixture.detectChanges();

    const component = fixture.componentInstance;
    expect(component.returnParams).toEqual({ return_to: '/parent/invoices/inv-9' });

    const html = fixture.nativeElement as HTMLElement;
    const link = html.querySelector('a') as HTMLAnchorElement;
    expect(link.getAttribute('href') ?? '').toContain('/signin');
  });
});
