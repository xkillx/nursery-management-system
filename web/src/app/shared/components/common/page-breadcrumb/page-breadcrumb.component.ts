import { CommonModule } from '@angular/common';
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  DestroyRef,
  Input,
  OnInit,
  inject,
} from '@angular/core';
import { ActivatedRouteSnapshot, NavigationEnd, Router, RouterModule } from '@angular/router';
import { filter, Observable, of } from 'rxjs';
import { startWith } from 'rxjs/operators';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroHome } from '@ng-icons/heroicons/outline';

import { AuthService } from '../../../../core/services/auth.service';
import { defaultRouteForRole } from '../../../../core/constants/roles';

export type CrumbResolver = (
  route: ActivatedRouteSnapshot,
  data: Record<string, unknown>,
  params: Record<string, string>,
  queryParams: Record<string, string>,
) => string | Observable<string>;

export interface Crumb {
  label: string;
  link?: string[];
  resolve?: CrumbResolver;
}

export const BREADCRUMB_DATA_KEY = 'breadcrumb';

export interface ResolvedCrumb {
  label: string;
  link: string[] | null;
  resolved$: Observable<string>;
  isHome?: boolean;
}

interface CollectedCrumb {
  crumb: Crumb;
  snapshot: ActivatedRouteSnapshot;
}

@Component({
  selector: 'app-page-breadcrumb',
  imports: [CommonModule, RouterModule, NgIcon],
  providers: [provideIcons({ heroHome })],
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './page-breadcrumb.component.html',
})
export class PageBreadcrumbComponent implements OnInit {
  @Input() mobileCollapseAfter = 2;

  crumbs: ResolvedCrumb[] = [];
  visibleCrumbs: ResolvedCrumb[] = [];
  hiddenCrumbs: ResolvedCrumb[] = [];
  showCollapseToggle = false;
  collapsed = true;

  private readonly router = inject(Router);
  private readonly cdr = inject(ChangeDetectorRef);
  private readonly destroyRef = inject(DestroyRef);
  private readonly authService = inject(AuthService);

  ngOnInit(): void {
    const sub = this.router.events
      .pipe(
        filter((e): e is NavigationEnd => e instanceof NavigationEnd),
        startWith(null),
      )
      .subscribe(() => this.rebuild());

    this.destroyRef.onDestroy(() => sub.unsubscribe());
  }

  toggleCollapsed(): void {
    this.collapsed = !this.collapsed;
    this.applyCollapse();
  }

  trackByIndex(index: number, item: ResolvedCrumb): string {
    return `${index}-${item.label}`;
  }

  private rebuild(): void {
    const root = this.router.routerState.snapshot.root;
    const collected: CollectedCrumb[] = [];
    this.collectCrumbs(root, collected);

    const routeCrumbs: ResolvedCrumb[] = collected.map(({ crumb, snapshot }) => ({
      label: crumb.label,
      link: crumb.link ?? null,
      resolved$: this.wrapResolver(crumb, snapshot),
    }));

    const homeCrumb: ResolvedCrumb = {
      label: 'Home',
      link: [defaultRouteForRole(this.authService.currentRole())],
      resolved$: of('Home'),
      isHome: true,
    };

    this.crumbs = [homeCrumb, ...routeCrumbs];
    this.collapsed = true;
    this.applyCollapse();
  }

  private collectCrumbs(
    snapshot: ActivatedRouteSnapshot | null,
    out: CollectedCrumb[],
  ): void {
    if (!snapshot) return;
    const data = snapshot.data ?? {};
    const crumb = data[BREADCRUMB_DATA_KEY] as Crumb | undefined;
    if (crumb && typeof crumb === 'object' && typeof crumb.label === 'string') {
      out.push({ crumb, snapshot });
    }
    if (snapshot.firstChild) {
      this.collectCrumbs(snapshot.firstChild, out);
    }
  }

  private wrapResolver(crumb: Crumb, snapshot: ActivatedRouteSnapshot): Observable<string> {
    if (!crumb.resolve) {
      return of(crumb.label);
    }
    const data: Record<string, unknown> = { ...(snapshot.data ?? {}) };
    const params: Record<string, string> = { ...(snapshot.params ?? {}) };
    const queryParams: Record<string, string> = {};
    snapshot.queryParamMap.keys.forEach((k) => {
      const v = snapshot.queryParamMap.get(k);
      if (v !== null) queryParams[k] = v;
    });
    const result = crumb.resolve(snapshot, data, params, queryParams);
    return typeof result === 'string' ? of(result) : result;
  }

  private applyCollapse(): void {
    if (this.crumbs.length <= 1) {
      this.showCollapseToggle = false;
      this.visibleCrumbs = [...this.crumbs];
      this.hiddenCrumbs = [];
      this.cdr.markForCheck();
      return;
    }

    const n = this.crumbs.length;
    const keep = Math.max(1, this.mobileCollapseAfter);
    const shouldCollapse = this.collapsed && n > keep + 1;
    this.showCollapseToggle = shouldCollapse;
    if (shouldCollapse) {
      this.hiddenCrumbs = this.crumbs.slice(0, n - (keep + 1));
      this.visibleCrumbs = this.crumbs.slice(n - (keep + 1));
    } else {
      this.hiddenCrumbs = [];
      this.visibleCrumbs = [...this.crumbs];
    }
    this.cdr.markForCheck();
  }
}
