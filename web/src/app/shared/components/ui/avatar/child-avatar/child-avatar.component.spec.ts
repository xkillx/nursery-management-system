import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ChildAvatarComponent } from './child-avatar.component';

describe('ChildAvatarComponent', () => {
  let fixture: ComponentFixture<ChildAvatarComponent>;
  let component: ChildAvatarComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChildAvatarComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ChildAvatarComponent);
    component = fixture.componentInstance;
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  describe('initials', () => {
    it('renders correct initials from first and last name', () => {
      component.name = 'John Smith';
      expect(component.initials).toBe('JS');
    });

    it('uppercases initials', () => {
      component.name = 'jane doe';
      expect(component.initials).toBe('JD');
    });

    it('returns single initial for single-word name', () => {
      component.name = 'Alice';
      expect(component.initials).toBe('A');
    });

    it('uses first and last word for multi-word names', () => {
      component.name = 'Mary Jane Watson';
      expect(component.initials).toBe('MW');
    });

    it('returns empty string for empty name', () => {
      component.name = '';
      expect(component.initials).toBe('');
    });
  });

  describe('photo display', () => {
    it('renders photo when photoUrl is provided', () => {
      component.photoUrl = 'https://example.com/photo.jpg';
      component.name = 'John Smith';
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      expect(img).toBeTruthy();
      expect(img.src).toBe('https://example.com/photo.jpg');
    });

    it('renders initials when photoUrl is null', () => {
      component.photoUrl = null;
      component.name = 'John Smith';
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      const span = fixture.nativeElement.querySelector('span');
      expect(img).toBeFalsy();
      expect(span.textContent.trim()).toBe('JS');
    });

    it('falls back to initials when image fails to load', () => {
      component.photoUrl = 'https://example.com/broken.jpg';
      component.name = 'John Smith';
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      img.dispatchEvent(new Event('error'));
      fixture.detectChanges();

      const span = fixture.nativeElement.querySelector('span');
      expect(fixture.nativeElement.querySelector('img')).toBeFalsy();
      expect(span.textContent.trim()).toBe('JS');
    });
  });

  describe('statusColor', () => {
    it('applies statusColor class when provided', () => {
      component.name = 'John Smith';
      component.statusColor = 'bg-blue-500 text-white';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('bg-blue-500')).toBeTrue();
      expect(div.classList.contains('text-white')).toBeTrue();
    });

    it('applies deterministic color when no statusColor', () => {
      component.name = 'John Smith';
      component.statusColor = null;
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      const classes = Array.from(div.classList) as string[];
      const hasColor = classes.some(
        (c: string) => c.startsWith('bg-') && c !== 'bg-transparent'
      );
      expect(hasColor).toBeTrue();
    });
  });

  describe('size', () => {
    it('applies sm size class', () => {
      component.name = 'John Smith';
      component.size = 'sm';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-8')).toBeTrue();
    });

    it('applies md size class by default', () => {
      component.name = 'John Smith';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-10')).toBeTrue();
    });

    it('applies lg size class', () => {
      component.name = 'John Smith';
      component.size = 'lg';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-12')).toBeTrue();
    });
  });

  describe('shape', () => {
    it('applies circle shape by default', () => {
      component.name = 'John Smith';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('rounded-full')).toBeTrue();
    });

    it('applies rounded shape when set', () => {
      component.name = 'John Smith';
      component.shape = 'rounded';
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('rounded-lg')).toBeTrue();
    });
  });
});
