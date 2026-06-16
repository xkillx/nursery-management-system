import { Component, DebugElement } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRouteSnapshot, provideRouter, Router, RouterOutlet, Routes } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroHome } from '@ng-icons/heroicons/outline';
import { of } from 'rxjs';

import { PageBreadcrumbComponent } from './page-breadcrumb.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ROLES } from '../../../../core/constants/roles';

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
  const fakeAuth = { currentRole: () => ROLES.manager };
  await TestBed.configureTestingModule({
    imports: [hostComponent, NgIcon],
    providers: [
      provideRouter(routes),
      provideIcons({ heroHome }),
      { provide: AuthService, useValue: fakeAuth },
    ],
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
  it('renders only the Home icon when no ancestor route declares breadcrumb data', async () => {
    const { fixture, runNavigation } = await setupTestBed([
      { path: 'empty', component: HostComponent },
    ]);
    await runNavigation(['/empty']);
    fixture.detectChanges();
    const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
    expect(nav).not.toBeNull();
    const homeLink = nav!.querySelector('a[aria-label="Home"]');
    expect(homeLink).not.toBeNull();
    const otherLinks = Array.from(nav!.querySelectorAll('a')).filter(
      (a) => a.getAttribute('aria-label') !== 'Home',
    );
    expect(otherLinks.length).toBe(0);
  });

  it('renders the trail from ancestor route data with the Home icon prepended', async () => {
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
    const homeLink = nav!.querySelector('a[aria-label="Home"]');
    expect(homeLink).not.toBeNull();
    const items = Array.from(nav!.querySelectorAll('li')).map((li) => (li as HTMLElement).textContent?.trim());
    expect(items.some((t) => t?.includes('Settings'))).toBeTrue();
    expect(items.some((t) => t?.includes('Manager'))).toBeTrue();
    expect(items.some((t) => t?.includes('Rooms'))).toBeTrue();
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
    const nonLeafAnchors = Array.from(
      fixture.nativeElement.querySelectorAll('nav li:not([aria-current="page"]) a') as NodeListOf<HTMLElement>,
    ).filter((a) => a.getAttribute('aria-label') !== 'Home');
    expect(nonLeafAnchors.length).toBe(0);

    await runNavigation(['/detail']);
    fixture.detectChanges();
    const allAnchors = Array.from(
      fixture.nativeElement.querySelectorAll('nav a') as NodeListOf<HTMLElement>,
    );
    const nonHomeAnchors = allAnchors.filter((a) => a.getAttribute('aria-label') !== 'Home');
    expect(nonHomeAnchors.length).toBe(0);
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

    let visibleHome = fixture.nativeElement.querySelector('a[aria-label="Home"]');
    expect(visibleHome).toBeNull();

    const button: DebugElement = fixture.debugElement.query(By.css('button[aria-expanded]'));
    button.nativeElement.click();
    fixture.detectChanges();

    visibleHome = fixture.nativeElement.querySelector('a[aria-label="Home"]');
    expect(visibleHome).not.toBeNull();
  });

  describe('Home icon prepend', () => {
    async function setupWithRole(role: string | null): Promise<TestEnv> {
      TestBed.resetTestingModule();
      const fakeAuth = { currentRole: () => role };
      await TestBed.configureTestingModule({
        imports: [
          HostComponent,
          NgIcon,
        ],
        providers: [
          provideRouter([]),
          provideIcons({ heroHome }),
          { provide: AuthService, useValue: fakeAuth },
        ],
      }).compileComponents();
      const router = TestBed.inject(Router);
      const fixture = TestBed.createComponent(HostComponent);
      fixture.detectChanges();
      return { router, runNavigation: (c) => router.navigate(c), fixture };
    }

    it('prepends a Home icon link as the first crumb when the route declares a breadcrumb', async () => {
      const { fixture, runNavigation } = await setupWithRole(ROLES.manager);
      const router = TestBed.inject(Router);
      router.resetConfig([
        {
          path: 'children',
          data: { breadcrumb: { label: 'Children' } },
          component: HostComponent,
        },
      ]);
      await runNavigation(['/children']);
      fixture.detectChanges();

      const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
      expect(nav).not.toBeNull();
      const homeLink = nav!.querySelector('a[aria-label="Home"]');
      expect(homeLink).not.toBeNull();
      const homeIcon = homeLink!.querySelector('ng-icon');
      expect(homeIcon).not.toBeNull();
    });

    it('routes the Home icon to the role-default landing page', async () => {
      const { fixture, runNavigation } = await setupWithRole(ROLES.owner);
      const router = TestBed.inject(Router);
      router.resetConfig([
        {
          path: 'rooms',
          data: { breadcrumb: { label: 'Rooms' } },
          component: HostComponent,
        },
      ]);
      await runNavigation(['/rooms']);
      fixture.detectChanges();

      const homeLink: HTMLAnchorElement | null = fixture.nativeElement.querySelector('a[aria-label="Home"]');
      expect(homeLink).not.toBeNull();
      expect(homeLink!.getAttribute('href')).toBe('/owner');
    });

    it('renders only the Home icon when no route declares a breadcrumb', async () => {
      const { fixture, runNavigation } = await setupWithRole(ROLES.manager);
      const router = TestBed.inject(Router);
      router.resetConfig([{ path: 'plain', component: HostComponent }]);
      await runNavigation(['/plain']);
      fixture.detectChanges();
      const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
      expect(nav).not.toBeNull();
      const homeLink = nav!.querySelector('a[aria-label="Home"]');
      expect(homeLink).not.toBeNull();
    });

    it('collects breadcrumbs from both parent and child routes (Home > Rooms > New room)', async () => {
      const { fixture, runNavigation } = await setupWithRole(ROLES.manager);
      const router = TestBed.inject(Router);
      router.resetConfig([
        {
          path: 'rooms',
          data: { breadcrumb: { label: 'Rooms', link: ['/rooms'] } },
          children: [
            {
              path: 'new',
              component: HostComponent,
              data: { breadcrumb: { label: 'New room' } },
            },
          ],
        },
      ]);
      await runNavigation(['/rooms/new']);
      fixture.detectChanges();

      const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
      expect(nav).not.toBeNull();
      const labels = Array.from(nav!.querySelectorAll('a, li'))
        .map((el) => (el.textContent ?? '').trim())
        .filter((t) => t.length > 0);
      expect(labels.some((t) => t.includes('Rooms'))).toBeTrue();
      expect(labels.some((t) => t.includes('New room'))).toBeTrue();
    });

    it('does not duplicate a parent breadcrumb that is inherited by a path:"" child', async () => {
      const { fixture, runNavigation } = await setupWithRole(ROLES.manager);
      const router = TestBed.inject(Router);
      router.resetConfig([
        {
          path: 'rooms',
          data: { breadcrumb: { label: 'Rooms', link: ['/rooms'] } },
          children: [
            {
              path: '',
              component: HostComponent,
            },
          ],
        },
      ]);
      await runNavigation(['/rooms']);
      fixture.detectChanges();

      const nav: HTMLElement | null = fixture.nativeElement.querySelector('nav[aria-label="Breadcrumb"]');
      expect(nav).not.toBeNull();
      const liTexts = Array.from(nav!.querySelectorAll('li')).map((li) => (li.textContent ?? '').trim());
      const roomsOccurrences = liTexts.filter((t) => t === 'Rooms').length;
      expect(roomsOccurrences).toBe(1);
    });
  });
});
