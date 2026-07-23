import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ParentCardComponent } from './parent-card.component';

describe('ParentCardComponent', () => {
  let component: ParentCardComponent;
  let fixture: ComponentFixture<ParentCardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ParentCardComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ParentCardComponent);
    component = fixture.componentInstance;
    component.parent = {
      id: '1',
      parentId: 'p1',
      firstName: 'John',
      lastName: 'Doe',
      email: 'john@example.com',
      phone: '07700900000',
      relationship: 'Father',
      customRelationship: '',
      addressLine1: '',
      addressLine2: '',
      addressCity: '',
      addressPostcode: '',
      hasParentalResponsibility: true,
      canPickUp: true,
      isEmergencyContact: true,
      portalStatus: 'none',
      isEditing: false,
      isNew: false,
    };
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
