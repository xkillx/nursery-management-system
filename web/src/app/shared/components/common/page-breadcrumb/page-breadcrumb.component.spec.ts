import { Component, DebugElement } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRouteSnapshot, provideRouter, Router, RouterOutlet, Routes } from '@angular/router';
import { of } from 'rxjs';

import { PageBreadcrumbComponent } from './page-breadcrumb.component';

@Component({
  standalone: true,
  imports: [PageBreadcrumbComponent, RouterOutlet],
  template: '<app-page-breadcrumb></app-page-breadcrumb><router-outlet></router-outlet>',
})
class HostComponent {}

interface TestEnv {
  router: Router;
  runNavigation: (commands: any[]) => Promise<boolean>;
  fixture: ComponentFixture<HostComponent>;
}

@Component({
  standalone: true,
  imports: [PageBreadcrumbComponent, RouterOutlet],
  template: `<app-page-breadcrumb [mobileCollapseAfter]="1"></app-page-breadcrumb><router-outlet></router-outlet>`,
})
class TestHostComponent1 {}

@Component({
  standalone: true,
  imports: [PageBreadcrumbComponent, RouterOutlet],
  template: `<app-page-breadcrumb [mobileCollapseAfter]="99"></app-page-breadcrumb><router-outlet></router-outlet>`,
})
class TestHostComponent2 {}

async function setupTestBed(routes: Routes, options: { expanded?: boolean } = {}): Promise<TestEnv> {
  TestBed.resetTestingModule();
  const hostComponent = options.expanded ? TestHostComponent2 : TestHostComponent1;
  await TestBed.configureTestingModule({
    imports: [hostComponent],
    providers: [provideRouter(routes)],
  }).compileComponents();

  const router = TestBed.inject(Router);
  const fixture = TestBed.createComponent(hostComponent);
  fixture.detectChanges();
  return {
    router,
    runNavigation: (commands) => router.navigate(commands),
    fixture,
  };
}

describe('PageBreadcrumbComponent', () => {
  it('renders nothing when no ancestor route declares breadcrumb data', async () => {
    const { fixture, runNavigation } = await setupTestBed([
      { path: 'empty', component: HostComponent },
    ]);
    await runNavigation(['/empty']);
    fixture.detectChanges();
    const nav = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
    expect(nav).toBeNull();
  });

  it('renders the trail from ancestor route data', async () => {
    const { fixture, runNavigation } = await setupTestBed(
      [
        {
          path: '',
          data: { breadcrumb: { label: 'Settings' } },
          children: [
            {
              path: 'manager',
              data: { breadcrumb: { label: 'Manager' } },
              children: [
                {
                  path: 'rooms',
                  data: { breadcrumb: { label: 'Rooms', link: ['/manager/rooms'] } },
                  component: HostComponent,
                },
              ],
            },
          ],
        },
      ],
      { expanded: true },
    );
    await runNavigation(['/manager/rooms']);
    fixture.detectChanges();
    const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
    expect(nav).not.toBeNull();
    const items = Array.from(nav!.querySelectorAll('li')).map((li) => (li as HTMLElement).textContent?.trim());
    expect(items[0]).toContain('Settings');
    expect(items[1]).toContain('Manager');
    expect(items[2]).toContain('Rooms');
  });

  it('marks the last segment as aria-current="page"', async () => {
    const { fixture, runNavigation } = await setupTestBed(
      [
        {
          path: '',
          data: { breadcrumb: { label: 'Settings' } },
          children: [
            {
              path: 'x',
              data: { breadcrumb: { label: 'X' } },
              component: HostComponent,
            },
          ],
        },
      ],
      { expanded: true },
    );
    await runNavigation(['/x']);
    fixture.detectChanges();
    const last: HTMLElement | null = fixture.nativeElement.querySelector('li[aria-current="page"]');
    expect(last).not.toBeNull();
    expect(last!.textContent.trim()).toBe('X');
  });

  it('renders a routerLink for segments that declare one and a plain span otherwise', async () => {
    const { fixture, runNavigation } = await setupTestBed(
      [
        {
          path: '',
          data: { breadcrumb: { label: 'Settings' } },
          children: [
            {
              path: 'list',
              data: { breadcrumb: { label: 'List', link: ['/list'] } },
              component: HostComponent,
            },
            {
              path: 'detail',
              data: { breadcrumb: { label: 'Detail' } },
              component: HostComponent,
            },
          ],
        },
      ],
      { expanded: true },
    );
    await runNavigation(['/list']);
    fixture.detectChanges();
    const listAnchor = fixture.nativeElement.querySelector('nav a[href="/list"]');
    expect(listAnchor).not.toBeNull();
    expect(listAnchor.textContent.trim()).toBe('List');
    const nonLeafAnchors = fixture.nativeElement.querySelectorAll('nav li:not([aria-current="page"]) a');
    expect(nonLeafAnchors.length).toBe(0);

    await runNavigation(['/detail']);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelectorAll('nav a').length).toBe(0);
  });

  it('resolves a static label returned by a resolve function', async () => {
    const { fixture, runNavigation } = await setupTestBed([
      {
        path: 'static/:id',
        data: {
          breadcrumb: {
            label: 'Loading…',
            resolve: () => 'Resolved label',
          },
        },
        component: HostComponent,
      },
    ]);
    await runNavigation(['/static/42']);
    fixture.detectChanges();
    const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
    expect(nav!.textContent).toContain('Resolved label');
  });

  it('resolves an Observable label returned by a resolve function', async () => {
    const { fixture, runNavigation } = await setupTestBed([
      {
        path: 'async/:id',
        data: {
          breadcrumb: {
            label: 'Loading…',
            resolve: (_route: ActivatedRouteSnapshot, _data: Record<string, unknown>, params: Record<string, string>) =>
              of(`Child ${params['id']}`),
          },
        },
        component: HostComponent,
      },
    ]);
    await runNavigation(['/async/7']);
    fixture.detectChanges();
    const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
    expect(nav!.textContent).toContain('Child 7');
  });

  it('shows a collapse toggle when there are more than 2 crumbs', async () => {
    const { fixture, runNavigation } = await setupTestBed([
      {
        path: '',
        data: { breadcrumb: { label: 'Settings' } },
        children: [
          {
            path: 'a',
            data: { breadcrumb: { label: 'A' } },
            children: [
              {
                path: 'b',
                data: { breadcrumb: { label: 'B' } },
                component: HostComponent,
              },
            ],
          },
        ],
      },
    ]);
    await runNavigation(['/a/b']);
    fixture.detectChanges();
    const button: HTMLButtonElement | null = fixture.nativeElement.querySelector('button[aria-expanded]');
    expect(button).not.toBeNull();
    expect(button!.getAttribute('aria-expanded')).toBe('false');
  });

  it('expands the trail when the collapse toggle is clicked', async () => {
    const { fixture, runNavigation } = await setupTestBed(
      [
        {
          path: '',
          data: { breadcrumb: { label: 'Settings' } },
          children: [
            {
              path: 'a',
              data: { breadcrumb: { label: 'A' } },
              children: [
                {
                  path: 'b',
                  data: { breadcrumb: { label: 'B' } },
                  component: HostComponent,
                },
              ],
            },
          ],
        },
      ],
      { expanded: false },
    );
    await runNavigation(['/a/b']);
    fixture.detectChanges();

    let visibleItems = Array.from(
      fixture.nativeElement.querySelectorAll('nav li:not([aria-hidden="true"])'),
    ).map((li) => (li as HTMLElement).textContent?.trim());
    expect(visibleItems.some((t) => t?.includes('Settings'))).toBeFalse();

    const button: DebugElement = fixture.debugElement.query(By.css('button[aria-expanded]'));
    button.nativeElement.click();
    fixture.detectChanges();

    visibleItems = Array.from(
      fixture.nativeElement.querySelectorAll('nav li:not([aria-hidden="true"])'),
    ).map((li) => (li as HTMLElement).textContent?.trim());
    expect(visibleItems.some((t) => t?.includes('Settings'))).toBeTrue();
  });
});
