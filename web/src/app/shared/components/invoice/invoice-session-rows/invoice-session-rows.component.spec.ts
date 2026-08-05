import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InvoiceSessionRowsComponent } from './invoice-session-rows.component';
import { InvoiceSession, InvoiceLineAggregate } from '../../../models/invoice-session.models';

describe('InvoiceSessionRowsComponent', () => {
  let component: InvoiceSessionRowsComponent;
  let fixture: ComponentFixture<InvoiceSessionRowsComponent>;

  const sessions: InvoiceSession[] = [
    {
      occurrenceDate: '2026-11-02',
      startMinutes: 480,
      endMinutes: 720,
      durationMinutes: 240,
      sessionTypeName: 'Morning Session',
      sessionAmountMinor: 6000,
    },
    {
      occurrenceDate: '2026-11-09',
      startMinutes: null,
      endMinutes: null,
      durationMinutes: 210,
      sessionTypeName: 'Legacy Session',
      sessionAmountMinor: 5250,
    },
  ];

  const aggregate: InvoiceLineAggregate = {
    description: 'Core childcare',
    sessionCount: 2,
    quantityMinutes: 450,
    totalMinor: 11250,
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceSessionRowsComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceSessionRowsComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('sessions', sessions);
    fixture.componentRef.setInput('unitAmountMinor', 1500);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('renders date/booking/qty/unit/total per row', () => {
    fixture.componentRef.setInput('aggregate', null);
    fixture.detectChanges();

    const rows = fixture.nativeElement.querySelectorAll('tr');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('Mon 2 Nov');
    expect(text).toContain('Morning Session (08:00–12:00)');
    expect(text).toContain('Legacy Session');
    expect(text).toContain('Standard Weekly / Monthly Session Rate');
    expect(text).toContain('4 hrs');
    expect(text).toContain('£60.00');
    expect(text).toContain('3.5 hrs');
    expect(text).toContain('£52.50');
    expect(rows.length).toBe(2);
  });

  it('formats name-only fallback when times absent', () => {
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Legacy Session');
    expect(text).not.toContain('Legacy Session (');

    const timedCell = fixture.nativeElement.querySelectorAll('tr')[0].textContent;
    expect(timedCell).toContain('Morning Session (08:00–12:00)');
  });

  it('renders aggregate row only when aggregate is provided', () => {
    fixture.componentRef.setInput('aggregate', aggregate);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Core childcare · 2 sessions');
    expect(text).toContain('£112.50');
  });

  it('omits aggregate row when aggregate is null', () => {
    fixture.componentRef.setInput('aggregate', null);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('sessions');
  });
});
