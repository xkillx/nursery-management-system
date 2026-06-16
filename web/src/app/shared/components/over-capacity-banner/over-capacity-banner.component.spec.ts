import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OverCapacityBannerComponent, OverCapacityRoom } from './over-capacity-banner.component';

describe('OverCapacityBannerComponent', () => {
  let fixture: ComponentFixture<OverCapacityBannerComponent>;
  let component: OverCapacityBannerComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OverCapacityBannerComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(OverCapacityBannerComponent);
    component = fixture.componentInstance;
  });

  it('is hidden when no rooms are over capacity', () => {
    component.rooms = [];
    fixture.detectChanges();
    expect(component.visible).toBeFalse();
    expect(fixture.nativeElement.querySelector('[data-testid="over-capacity-banner"]')).toBeNull();
  });

  it('renders one over capacity room with a row anchor', () => {
    const rooms: OverCapacityRoom[] = [
      { id: 'room-1', name: 'Sunshine Room', assigned: 14, capacity: 12 },
    ];
    component.rooms = rooms;
    fixture.detectChanges();

    expect(component.visible).toBeTrue();
    const banner = fixture.nativeElement.querySelector('[data-testid="over-capacity-banner"]');
    expect(banner).not.toBeNull();
    expect(banner.textContent).toContain('Sunshine Room is over capacity (14/12)');
    const link = banner.querySelector('a');
    expect(link.getAttribute('href')).toBe('#room-room-1');
  });

  it('switches heading to plural when more than one room is over capacity', () => {
    component.rooms = [
      { id: 'room-1', name: 'Sunshine Room', assigned: 14, capacity: 12 },
      { id: 'room-2', name: 'Sensory Hub', assigned: 9, capacity: 6 },
    ];
    fixture.detectChanges();

    const banner = fixture.nativeElement.querySelector('[data-testid="over-capacity-banner"]');
    expect(banner.textContent).toContain('Rooms over capacity');
    expect(banner.querySelectorAll('a').length).toBe(2);
  });
});
