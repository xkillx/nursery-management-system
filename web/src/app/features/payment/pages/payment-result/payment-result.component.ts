import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { Meta } from '@angular/platform-browser';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCheckCircle, heroXCircle, heroExclamationTriangle, heroArrowRight } from '@ng-icons/heroicons/outline';

import { AuthPageLayoutComponent } from '../../../../shared/layout/auth-page-layout/auth-page-layout.component';

export type PaymentOutcomeKind = 'success' | 'cancelled' | 'unknown';

@Component({
  selector: 'app-payment-result',
  imports: [
    AuthPageLayoutComponent,
    RouterLink,
    NgIcon,
  ],
  providers: [
    provideIcons({ heroCheckCircle, heroXCircle, heroExclamationTriangle, heroArrowRight }),
  ],
  templateUrl: './payment-result.component.html',
})
export class PaymentResultComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly meta = inject(Meta);

  outcome: PaymentOutcomeKind = 'unknown';
  invoiceId: string | null = null;
  sessionId: string | null = null;

  ngOnInit(): void {
    const params = this.route.snapshot.queryParamMap;
    this.invoiceId = params.get('invoice_id');
    this.sessionId = params.get('session_id');

    const raw = params.get('outcome');
    if (raw === 'success') {
      this.outcome = 'success';
    } else if (raw === 'cancelled' || raw === 'canceled') {
      this.outcome = 'cancelled';
    } else {
      this.outcome = 'unknown';
    }

    // Invoice ids must not leak into referrers or search indexes (U5).
    this.meta.addTag({ name: 'robots', content: 'noindex, nofollow' });
    this.meta.addTag({ 'http-equiv': 'Referrer-Policy', content: 'no-referrer' });
  }

  get returnParams(): Record<string, string> {
    return this.invoiceId ? { return_to: `/parent/invoices/${this.invoiceId}` } : {};
  }
}
