import { Component, inject } from '@angular/core';
import { SidebarService } from '../../services/sidebar.service';
import { CommonModule } from '@angular/common';
import { AppSidebarComponent } from '../app-sidebar/app-sidebar.component';
import { BackdropComponent } from '../backdrop/backdrop.component';
import { RouterModule } from '@angular/router';
import { AppHeaderComponent } from '../app-header/app-header.component';
import { ToastContainerComponent } from '../../components/ui/toast/toast-container.component';
import { PageBreadcrumbComponent } from '../../components/common/page-breadcrumb/page-breadcrumb.component';

@Component({
  selector: 'app-layout',
  imports: [
    CommonModule,
    RouterModule,
    AppHeaderComponent,
    AppSidebarComponent,
    BackdropComponent,
    ToastContainerComponent,
    PageBreadcrumbComponent,
  ],
  templateUrl: './app-layout.component.html',
})

export class AppLayoutComponent {
  readonly sidebarService = inject(SidebarService);
  readonly isExpanded$ = this.sidebarService.isExpanded$;
  readonly isHovered$ = this.sidebarService.isHovered$;
  readonly isMobileOpen$ = this.sidebarService.isMobileOpen$;

  get containerClasses() {
    return [
      'flex-1',
      'transition-all',
      'duration-300',
      'ease-in-out',
      (this.isExpanded$ || this.isHovered$) ? 'xl:ml-[290px]' : 'xl:ml-[90px]',
      this.isMobileOpen$ ? 'ml-0' : ''
    ];
  }

}
