import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

import { ManagerParentsComponent } from './manager-parents.component';

describe('ManagerParentsComponent', () => {
  let component: ManagerParentsComponent;
  let fixture: ComponentFixture<ManagerParentsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerParentsComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerParentsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
