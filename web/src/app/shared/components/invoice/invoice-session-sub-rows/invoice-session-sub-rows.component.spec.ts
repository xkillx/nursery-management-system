import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InvoiceSessionSubRowsComponent } from './invoice-session-sub-rows.component';
import { InvoiceSession } from '../../../models/invoice-session.models';

describe('InvoiceSessionSubRowsComponent', () => {
  let component: InvoiceSessionSubRowsComponent;
  let fixture: ComponentFixture<InvoiceSessionSubRowsComponent>;

  const sessions: InvoiceSession[] = [
    {
      occurrenceDate: '2026-11-02',
      startMinutes: 480,
      endMinutes: 720,
      durationMinutes: 240,
      sessionTypeName: 'Morning Session',
      sessionAmountMinor: 6000,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceSessionSubRowsComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceSessionSubRowsComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('sessions', sessions);
    fixture.componentRef.setInput('unitAmountMinor', 1500);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('renders date + booking + qty + unit + total per session', () => {
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Mon 2 Nov');
    expect(text).toContain('Morning Session (08:00–12:00)');
    expect(text).toContain('4 hrs');
    expect(text).toContain('£15.00');
    expect(text).toContain('£60.00');
  });

  it('renders nothing when sessions is empty', () => {
    fixture.componentRef.setInput('sessions', []);
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll('tr');
    expect(rows.length).toBe(0);
  });
});
