import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ButtonComponent } from './button.component';

describe('ButtonComponent', () => {
  let fixture: ComponentFixture<ButtonComponent>;
  let component: ButtonComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ButtonComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ButtonComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  it('emits btnClick when not disabled or loading', () => {
    let emitted = false;
    component.btnClick.subscribe(() => (emitted = true));
    component.onClick(new Event('click'));
    expect(emitted).toBeTrue();
  });

  it('does not emit btnClick when disabled', () => {
    let emitted = false;
    component.btnClick.subscribe(() => (emitted = true));
    component.disabled = true;
    component.onClick(new Event('click'));
    expect(emitted).toBeFalse();
  });

  it('does not emit btnClick when loading', () => {
    let emitted = false;
    component.btnClick.subscribe(() => (emitted = true));
    component.loading = true;
    component.onClick(new Event('click'));
    expect(emitted).toBeFalse();
  });

  it('defaults type to button', () => {
    expect(component.type).toBe('button');
  });

  it('defaults variant to primary', () => {
    expect(component.variant).toBe('primary');
  });

  it('defaults size to md', () => {
    expect(component.size).toBe('md');
  });
});
