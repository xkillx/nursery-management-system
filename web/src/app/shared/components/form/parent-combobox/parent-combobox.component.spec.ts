import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ParentComboboxComponent } from './parent-combobox.component';

describe('ParentComboboxComponent', () => {
  let component: ParentComboboxComponent;
  let fixture: ComponentFixture<ParentComboboxComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ParentComboboxComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ParentComboboxComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
